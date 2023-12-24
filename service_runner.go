package composer

import (
	"context"
	"crypto/rand"
	"log/slog"
	"sync"
	"time"

	"github.com/ferranbt/composer/docker"
	"github.com/ferranbt/composer/hooks"
	"github.com/ferranbt/composer/proto"
)

type serviceRunner struct {
	logger           *slog.Logger
	project          *proto.Project
	service          *proto.Service
	name             string
	driver           *docker.Provider
	shutdownCh       chan struct{}
	killCtx          context.Context
	killCtxCancel    context.CancelFunc
	killErr          error
	killed           bool
	waitCh           chan struct{}
	handle           *proto.ServiceState_Handle
	store            *BoltdbStore
	statusLock       sync.Mutex
	status           *proto.ServiceState
	taskStateUpdated func()
	restartCount     uint64
	notifier         Notifier
	runnerHooks      []hooks.ServiceHook
}

func newServiceRunner(project *proto.Project, name string, service *proto.Service, driver *docker.Provider, store *BoltdbStore, taskStateUpdated func(), notifier Notifier, hooks []hooks.ServiceHook) *serviceRunner {
	killCtx, killCancel := context.WithCancel(context.Background())

	hash, err := service.Hash()
	if err != nil {
		panic(err)
	}

	return &serviceRunner{
		logger:           slog.Default(),
		driver:           driver,
		project:          project,
		service:          service,
		status:           &proto.ServiceState{Hash: hash},
		name:             name,
		shutdownCh:       make(chan struct{}),
		killCtx:          killCtx,
		killCtxCancel:    killCancel,
		waitCh:           make(chan struct{}),
		taskStateUpdated: taskStateUpdated,
		store:            store,
		notifier:         notifier,
		runnerHooks:      hooks,
	}
}

func (t *serviceRunner) SetLogger(logger *slog.Logger) {
	t.logger = logger
}

func (t *serviceRunner) Run() {
	defer close(t.waitCh)
	var result *proto.ExitResult
	_ = result

MAIN:
	for {
		select {
		case <-t.killCtx.Done():
			break MAIN
		case <-t.shutdownCh:
			return
		default:
		}

		if err := t.preStart(); err != nil {
			t.logger.Error("prestart failed", "error", err)
			goto RESTART
		}

		select {
		case <-t.killCtx.Done():
			break MAIN
		case <-t.shutdownCh:
			return
		default:
		}

		if err := t.runDriver(); err != nil {
			goto RESTART
		}

		if err := t.postStart(); err != nil {
			t.logger.Error("poststart failed", "error", err)
			goto RESTART
		}

		{
			result = nil

			resultCh, err := t.driver.WaitTask(context.Background(), t.handle.ContainerId)
			if err != nil {
				t.logger.Error("failed to wait for task", "err", err)
			} else {
				select {
				case <-t.killCtx.Done():
					result = t.handleKill(resultCh)
				case <-t.shutdownCh:
					return
				case result = <-resultCh:
				}

				t.emitExitResultEvent(result)
			}
		}

		t.clearDriverHandle()

	RESTART:
		restart, delay := t.shouldRestart()
		if !restart {
			break MAIN
		}

		select {
		case <-t.shutdownCh:
			return
		case <-time.After(delay):
		}
	}

	// task is dead
	t.UpdateStatus(proto.ServiceState_Dead, nil)

	// Run the stop hooks
	if err := t.stop(); err != nil {
		t.logger.Error("stop failed", "error", err)
	}
}

func (s *serviceRunner) shouldRestart() (bool, time.Duration) {
	if s.killed {
		return false, 0
	}

	if s.service.RestartPolicy == nil {
		// no restart
		return false, 0
	}

	s.restartCount++
	if s.restartCount > 5 {
		// too many restarts, consider this task dead and do not realocate
		s.UpdateStatus(proto.ServiceState_Dead, proto.NewEvent(s.project.Name, s.name, proto.TaskNotRestarting).SetFailsTask())
		return false, 0
	}

	s.UpdateStatus(proto.ServiceState_Pending, proto.NewEvent(s.project.Name, s.name, proto.TaskRestarting))
	return true, time.Duration(2 * time.Second)
}

func (s *serviceRunner) handleKill(resultCh <-chan *proto.ExitResult) *proto.ExitResult {
	s.killed = true

	// Check if it is still running
	select {
	case result := <-resultCh:
		return result
	default:
	}

	if err := s.driver.StopTask(s.handle.ContainerId, 0); err != nil {
		s.killErr = err
	}

	select {
	case result := <-resultCh:
		return result
	case <-s.shutdownCh:
		return nil
	}
}

func (s *serviceRunner) Taint() {
	s.UpdateStatus(proto.ServiceState_Tainted, nil)
}

func (s *serviceRunner) WaitCh() <-chan struct{} {
	return s.waitCh
}

func (t *serviceRunner) Kill(ctx context.Context) error {
	t.killCtxCancel()

	select {
	case <-t.WaitCh():
	case <-ctx.Done():
		return ctx.Err()
	}

	return t.killErr
}

func (t *serviceRunner) emitExitResultEvent(result *proto.ExitResult) {
	if result == nil {
		return
	}
	event := proto.NewEvent(t.project.Name, t.name, proto.TaskTerminated).
		SetExitCode(int64(result.ExitCode)).
		SetSignal(0)

	t.notifier.Notify(event)
}

func (t *serviceRunner) runDriver() error {
	if t.handle != nil {
		t.UpdateStatus(proto.ServiceState_Running, nil)
		return nil
	}

	invocationId := make([]byte, 16)
	rand.Read(invocationId)

	tt := &docker.ComputeResource{
		Service: t.service,
	}

	handle, err := t.driver.Create(tt)
	if err != nil {
		return err
	}
	t.handle = handle
	if err := t.store.PutTaskHandle(t.project.Name, t.name, handle); err != nil {
		return err
	}
	t.UpdateStatus(proto.ServiceState_Running, proto.NewEvent(t.project.Name, t.name, proto.TaskStarted))
	return nil
}

func (t *serviceRunner) Shutdown() {
	close(t.shutdownCh)
	<-t.WaitCh()
	t.taskStateUpdated()
}

func (t *serviceRunner) clearDriverHandle() {
	if t.handle != nil {
		t.driver.DestroyTask(t.handle.ContainerId, true)
	}
	t.handle = nil
}

func (t *serviceRunner) TaskState() *proto.ServiceState {
	t.statusLock.Lock()
	defer t.statusLock.Unlock()

	t.status.Handle = t.handle
	return t.status
}

func (t *serviceRunner) Restore() error {
	state, handle, err := t.store.GetTaskState(t.project.Name, t.name)
	if err != nil {
		return err
	}
	t.status = state

	if err := t.driver.RecoverTask(handle.ContainerId, handle); err != nil {
		t.UpdateStatus(proto.ServiceState_Pending, nil)
		return nil
	}

	// the handle was restored
	t.handle = handle
	return nil
}

func (t *serviceRunner) UpdateStatus(status proto.ServiceState_State, ev *proto.Event) {
	t.statusLock.Lock()
	defer t.statusLock.Unlock()

	t.logger.Info("Update status", "status", status.String())
	t.status.State = status

	if ev != nil {
		if ev.FailsTask() {
			t.status.Failed = true
		}
		t.notifier.Notify(ev)
	}

	if err := t.store.PutTaskState(t.project.Name, t.name, t.status); err != nil {
		t.logger.Warn("failed to persist task state during update status", "err", err)
	}
	t.taskStateUpdated()
}
