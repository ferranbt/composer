package composer

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ferranbt/composer/internal"
)

type RuntimeConfig struct {
	LogsLocation string
	Logger       *slog.Logger
}

type Runtime struct {
	logger   *slog.Logger
	config   *RuntimeConfig
	manifest *Manifest

	docker *Docker
	exec   *Exec

	dag *internal.DAG

	tasks       map[string]*taskRef
	taskUpdates chan struct{}

	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
	mu     sync.Mutex

	reservedPorts map[int]bool
}

func NewRuntime(manifest *Manifest, config *RuntimeConfig) (*Runtime, error) {
	if config == nil {
		config = &RuntimeConfig{}
	}

	// Set up logger
	logger := config.Logger
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
	}

	docker, err := NewDocker()
	if err != nil {
		return nil, err
	}

	r := &Runtime{
		manifest:      manifest,
		tasks:         make(map[string]*taskRef),
		taskUpdates:   make(chan struct{}, 10),
		config:        config,
		reservedPorts: make(map[int]bool),
		logger:        logger,
		exec:          NewExec(),
		docker:        docker,
	}

	return r, nil
}

func (r *Runtime) NotifyTaskUpdated() {
	select {
	case r.taskUpdates <- struct{}{}:
	default:
	}
}

func (r *Runtime) ApplyTemplate(spec *Spec) ([]string, map[string]string, error) {
	resolvePort := func(targetSpec *Spec, portName string, defaultPort int) int {
		if spec.driverType == DriverTypeExec {
			for _, p := range targetSpec.ports {
				if p.Name == portName {
					return p.Port
				}
			}
			return defaultPort
		}
		return defaultPort
	}

	resolveAddr := func(targetSpec *Spec, portName string) string {
		var port int
		for _, p := range targetSpec.ports {
			if p.Name == portName {
				port = p.Port
				break
			}
		}

		// Service runs on exec
		if spec.driverType == DriverTypeExec {
			if targetSpec.driverType == DriverTypeExec {
				// exec -> exec: localhost:port
				return fmt.Sprintf("localhost:%d", port)
			}
			// exec -> docker: localhost:port
			return fmt.Sprintf("localhost:%d", port)
		}

		// Service runs on docker
		if targetSpec.driverType == DriverTypeExec {
			// docker -> exec: host.docker.internal:port
			return fmt.Sprintf("host.docker.internal:%d", port)
		}
		// docker -> docker: DNS service:port
		return fmt.Sprintf("%s:%d", targetSpec.Name(), port)
	}

	funcs := &templateFuncs{
		Service: func(serviceName, portName string) string {
			targetSpec := r.manifest.getSpec(serviceName)
			if targetSpec == nil {
				return ""
			}
			return resolveAddr(targetSpec, portName)
		},
		Port: func(portName string, defaultPort int) int {
			return resolvePort(spec, portName, defaultPort)
		},
	}

	var argsResult []string
	for _, arg := range spec.args {
		newArg, err := applyArgsTemplate(arg, funcs)
		if err != nil {
			return nil, nil, err
		}
		argsResult = append(argsResult, newArg)
	}

	envs := map[string]string{}
	for k, v := range spec.env {
		newV, err := applyArgsTemplate(v, funcs)
		if err != nil {
			return nil, nil, err
		}
		envs[k] = newV
	}

	return argsResult, envs, nil
}

// reservePort finds the first available port from the startPort and reserves it.
// Note that we have to keep track of the port in 'reservedPorts' because
// the port allocation happens before the services use it and bind to it.
func (r *Runtime) reservePort(startPort int) (int, error) {
	for i := startPort; i < startPort+1000; i++ {
		if r.reservedPorts[i] {
			continue
		}

		// Try to bind to the port to verify it's available
		listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", i))
		if err != nil {
			continue
		}
		listener.Close()

		r.reservedPorts[i] = true
		return i, nil
	}

	return 0, fmt.Errorf("could not reserve a port in range %d-%d", startPort, startPort+1000)
}

func (r *Runtime) preValidateManifest() error {
	// Reserve ports for all specs that need them
	for _, spec := range r.manifest.Specs() {
		for _, port := range spec.ports {
			reservedPort, err := r.reservePort(port.Port)
			if err != nil {
				return err
			}
			port.Port = reservedPort
		}
	}

	// Apply templates and pre-populate taskRefs for all specs
	for _, spec := range r.manifest.Specs() {
		args, env, err := r.ApplyTemplate(spec)
		if err != nil {
			return err
		}
		spec.args = args
		spec.env = env
	}

	// Build DAG of services and validate it
	r.dag = internal.NewDAG()
	for _, spec := range r.manifest.Specs() {
		var deps []string
		for _, dep := range spec.dependsOn {
			deps = append(deps, dep.Name)
		}
		r.dag.AddNode(spec.Name(), deps)
	}

	if err := r.dag.Validate(); err != nil {
		return err
	}

	return nil
}

func (r *Runtime) Run(ctx context.Context) error {
	// Create internal context for task lifetime
	r.mu.Lock()
	r.ctx, r.cancel = context.WithCancel(context.Background())
	r.mu.Unlock()

	err := r.preValidateManifest()
	if err != nil {
		return err
	}

	for _, spec := range r.manifest.Specs() {
		var taskServiceLogger io.Writer
		if r.config.LogsLocation != "" {
			taskServiceLogger, err = r.startLogStreamForTask(spec.Name())
			if err != nil {
				return err
			}
		}

		var driver Driver
		if spec.driverType == DriverTypeExec {
			driver = r.exec
		} else {
			driver = r.docker
		}

		logger := r.logger.With("service", spec.Name())
		task := newTaskRef(logger, spec, driver, taskServiceLogger, r)

		r.tasks[spec.Name()] = task
	}

	r.logger.Debug("starting runtime", "tasks", len(r.tasks))

	// Channel to receive task completion notifications
	// Any task finishing (with or without error) during deployment is bad
	taskDoneCh := make(chan error, len(r.tasks))

	// Start all task goroutines with internal context
	for _, task := range r.tasks {
		r.wg.Add(1)

		go func(t *taskRef) {
			defer r.wg.Done()

			err := t.Run(r.ctx)
			// Notify that task finished (with or without error)
			// Any task finishing during deployment is a failure
			select {
			case taskDoneCh <- err:
			default:
			}
		}(task)
	}

	// trigger an initial task update to start the
	// first iteration of the reconcile loop
	r.taskUpdates <- struct{}{}

	// Reconciliation loop for deployment phase only
	for {
		select {
		case <-ctx.Done():
			// User cancelled deployment, trigger cleanup
			r.cancel()
			r.wg.Wait()
			return ctx.Err()
		case <-r.ctx.Done():
			// Internal context cancelled (Stop() was called during deployment)
			r.wg.Wait()
			return fmt.Errorf("runtime stopped during deployment")
		case err := <-taskDoneCh:
			// Task finished during deployment - this is a failure
			r.cancel()
			r.wg.Wait()
			if err != nil {
				r.logger.Error("task failed during deployment", "error", err)
				return fmt.Errorf("task failed: %w", err)
			}
			r.logger.Error("task exited during deployment")
			return fmt.Errorf("task exited unexpectedly during deployment")
		case <-r.taskUpdates:
			done, err := r.reconcileLoop()
			if err != nil {
				// Deployment error - stop all tasks and return error
				r.cancel()
				r.wg.Wait()
				return err
			}
			if done {
				// All tasks deployed successfully - return and let tasks run
				r.logger.Info("all tasks deployed, runtime ready")
				return nil
			}
		}
	}
}

func (r *Runtime) Stop() {
	r.mu.Lock()
	cancel := r.cancel
	r.mu.Unlock()

	if cancel != nil {
		cancel()
	}
	// Wait for all task goroutines to finish
	r.wg.Wait()
}

func (r *Runtime) reconcileLoop() (bool, error) {
	completed := make(map[string]bool)
	for name, task := range r.tasks {
		if status := task.GetStatus(); status == TaskStatusDeployed {
			completed[name] = true
		}
	}

	if len(completed) == len(r.tasks) {
		return true, nil
	}

	ready := r.dag.GetReadyNodes(completed)
	r.logger.Debug("checking ready tasks", "ready", ready)

	toStart := []string{}
	for _, name := range ready {
		// skip task that is starting already
		status := r.tasks[name].GetStatus()
		if status == TaskStatusStarting {
			continue
		}

		r.logger.Debug("evaluating task", "task", name)

		spec := r.manifest.getSpec(name)
		if spec == nil {
			return false, fmt.Errorf("spec '%s' not found", name)
		}

		// Check if all dependencies are satisfied (including healthy condition)
		allDepsReady := true
		for _, dep := range spec.dependsOn {
			depTask := r.tasks[dep.Name]

			r.logger.Debug("checking dependency", "task", name, "dependency", dep.Name, "condition", dep.Condition, "healthy", depTask.IsHealthy())
			if dep.Condition == "healthy" && !depTask.IsHealthy() {
				r.logger.Debug("dependency not ready", "task", name, "dependency", dep.Name, "reason", "not healthy")
				allDepsReady = false
				break
			}
		}

		if !allDepsReady {
			continue
		}

		toStart = append(toStart, name)
	}

	// mark the start as ready now
	for _, startTask := range toStart {
		r.tasks[startTask].MarkAsReady()
	}

	return false, nil
}

func (r *Runtime) startLogStreamForTask(name string) (io.Writer, error) {
	logPath := filepath.Join(r.config.LogsLocation, name+".log")

	if err := os.MkdirAll(r.config.LogsLocation, 0755); err != nil {
		return nil, fmt.Errorf("failed to create logs directory: %w", err)
	}

	logFile, err := os.Create(logPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create log file: %w", err)
	}
	return logFile, nil
}

type taskRef struct {
	healthy     atomic.Bool
	logger      *slog.Logger
	driver      Driver
	spec        *Spec
	status      TaskStatus
	statusLock  sync.Mutex
	prober      *internal.Prober
	logWriter   io.Writer
	taskUpdater TaskUpdater
	readyCh     chan struct{}
	readyOnce   sync.Once
	stoppedCh   chan struct{}
}

func newTaskRef(logger *slog.Logger, spec *Spec, driver Driver, logWriter io.Writer, taskUpdater TaskUpdater) *taskRef {
	return &taskRef{
		logger:      logger,
		spec:        spec,
		driver:      driver,
		status:      TaskStatusPending,
		readyCh:     make(chan struct{}),
		logWriter:   logWriter,
		taskUpdater: taskUpdater,
		stoppedCh:   make(chan struct{}),
	}
}

func (t *taskRef) IsHealthy() bool {
	return t.healthy.Load()
}

func (t *taskRef) MarkAsReady() {
	// Close the ready channel if not already closed. This handles the race
	// where multiple reconciliation loops see a task as ready before it's
	// marked as starting, causing duplicate run attempts.
	t.readyOnce.Do(func() {
		close(t.readyCh)
	})
}

func (t *taskRef) GetStatus() TaskStatus {
	t.statusLock.Lock()
	status := t.status
	t.statusLock.Unlock()
	return status
}

func (t *taskRef) setStatus(status TaskStatus) {
	t.statusLock.Lock()
	t.status = status
	t.statusLock.Unlock()
	t.taskUpdater.NotifyTaskUpdated()
}

func (t *taskRef) TaskUpdated(status TaskStatus) {
	if status == TaskStatusStopped {
		close(t.stoppedCh)
	}

	t.setStatus(status)
}

func (t *taskRef) Run(ctx context.Context) error {
	// wait for the task to be ready or stop
	select {
	case <-t.readyCh:
	case <-ctx.Done():
		// the task was closed before it was scheduled to run
		t.setStatus(TaskStatusStopped)
		return nil
	}

	t.setStatus(TaskStatusStarting)

	// the task is ready to run
	handle, err := t.driver.RunTask(ctx, t.spec, t)
	if err != nil {
		if ctx.Err() != nil {
			// context is done, error is likely due to cancellation
			// update the status of the task to stopped
			t.setStatus(TaskStatusStopped)
			return nil
		}
		return err
	}

	// if context was closed already, do not event start the post-start tasks
	// and go right away to clean up tasks
	if ctx.Err() != nil {
		goto CLEANUP
	}

	if t.logWriter != nil {
		// log storing enabled
		go func() {
			logReader, err := handle.Logs()
			if err != nil {
				panic(err)
			}
			io.Copy(t.logWriter, logReader)
		}()
	}

	// if the task has a health probe, start to track it
	if healthcheck := t.spec.healthCheck; healthcheck != nil {
		t.prober = internal.NewProber(healthcheck, func(healthy bool) {
			t.healthy.Store(healthy)
			t.taskUpdater.NotifyTaskUpdated()
		})
		t.prober.Start()
	}

	// wait for context to close or to be notified that the task was stopped
	// if the channel was closed
	select {
	case <-ctx.Done():
	case <-t.stoppedCh:
		// the task stopped, return an error and notify, the status
		// has already being modified
		return handle.Err()
	}

CLEANUP:
	// the runtime stopped, stop the tasks and wait for it to stop
	if err := handle.Stop(); err != nil {
		t.logger.Error("failed to stop task", "err", err)
	}

	select {
	case <-t.stoppedCh:
	case <-time.After(20 * time.Second):
	}

	return nil
}
