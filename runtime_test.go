package composer

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/docker/docker/api/types/image"
	"github.com/ferranbt/composer/internal"
	"github.com/stretchr/testify/require"
)

func TestRuntime_CycleDetection(t *testing.T) {
	m := NewManifest()
	m.AddSpec(NewSpec("service1").DependsOn("service2", "started"))
	m.AddSpec(NewSpec("service2").DependsOn("service1", "started"))

	runtime, err := NewRuntime(m, nil)
	require.NoError(t, err)

	err = runtime.Run(context.Background())
	require.Error(t, err)
	require.Contains(t, err.Error(), "cycle detected")
}

func TestApplyTemplate(t *testing.T) {
	t.Run("docker -> docker uses service name", func(t *testing.T) {
		m := NewManifest()
		m.AddSpec(NewSpec("db").Image("postgres", "latest").Port("pg", 5432))
		m.AddSpec(NewSpec("api").Image("myapi", "latest").Args("--db={{Service \"db\" \"pg\"}}"))

		runtime, err := NewRuntime(m, nil)
		require.NoError(t, err)
		runtime.manifest = m

		apiSpec := m.getSpec("api")
		args, _, err := runtime.ApplyTemplate(apiSpec)
		require.NoError(t, err)
		require.Equal(t, []string{"--db=db:5432"}, args)
	})

	t.Run("docker -> exec uses host.docker.internal", func(t *testing.T) {
		m := NewManifest()
		m.AddSpec(NewSpec("host-service").UseExec().Args("./server").Port("http", 8080))
		m.AddSpec(NewSpec("api").Image("myapi", "latest").Args("--upstream={{Service \"host-service\" \"http\"}}"))

		runtime, err := NewRuntime(m, nil)
		require.NoError(t, err)
		runtime.manifest = m

		apiSpec := m.getSpec("api")
		args, _, err := runtime.ApplyTemplate(apiSpec)
		require.NoError(t, err)
		require.Equal(t, []string{"--upstream=host.docker.internal:8080"}, args)
	})

	t.Run("exec -> docker uses localhost", func(t *testing.T) {
		m := NewManifest()
		m.AddSpec(NewSpec("db").Image("postgres", "latest").Port("pg", 5432))
		m.AddSpec(NewSpec("cli").UseExec().Args("./client", "--db={{Service \"db\" \"pg\"}}"))

		runtime, err := NewRuntime(m, nil)
		require.NoError(t, err)
		runtime.manifest = m

		cliSpec := m.getSpec("cli")
		args, _, err := runtime.ApplyTemplate(cliSpec)
		require.NoError(t, err)
		require.Equal(t, []string{"./client", "--db=localhost:5432"}, args)
	})

	t.Run("exec -> exec uses localhost", func(t *testing.T) {
		m := NewManifest()
		m.AddSpec(NewSpec("server").UseExec().Args("./server").Port("http", 8080))
		m.AddSpec(NewSpec("client").UseExec().Args("./client", "{{Service \"server\" \"http\"}}"))

		runtime, err := NewRuntime(m, nil)
		require.NoError(t, err)
		runtime.manifest = m

		clientSpec := m.getSpec("client")
		args, _, err := runtime.ApplyTemplate(clientSpec)
		require.NoError(t, err)
		require.Equal(t, []string{"./client", "localhost:8080"}, args)
	})

	t.Run("Port template in env variables", func(t *testing.T) {
		m := NewManifest()
		m.AddSpec(NewSpec("server").UseExec().Port("http", 8080).
			Env("LISTEN", "0.0.0.0:{{Port \"http\" 3000}}"))

		runtime, err := NewRuntime(m, nil)
		require.NoError(t, err)
		runtime.manifest = m

		serverSpec := m.getSpec("server")
		_, env, err := runtime.ApplyTemplate(serverSpec)
		require.NoError(t, err)
		require.Equal(t, "0.0.0.0:8080", env["LISTEN"])
	})

	t.Run("Port template uses default when not found", func(t *testing.T) {
		m := NewManifest()
		m.AddSpec(NewSpec("server").UseExec().
			Env("LISTEN", "0.0.0.0:{{Port \"http\" 3000}}"))

		runtime, err := NewRuntime(m, nil)
		require.NoError(t, err)
		runtime.manifest = m

		serverSpec := m.getSpec("server")
		_, env, err := runtime.ApplyTemplate(serverSpec)
		require.NoError(t, err)
		require.Equal(t, "0.0.0.0:3000", env["LISTEN"])
	})
}

func TestRuntime_HealthCheck(t *testing.T) {
	server, _ := internal.MockHealthServer(2)
	defer server.Close()

	m := NewManifest()
	m.AddSpec(longRunningSpec("db").
		HealthCheckHTTP(server.URL, http.StatusOK, 10*time.Millisecond, 50*time.Millisecond, 1*time.Second))
	m.AddSpec(longRunningSpec("api").
		DependsOn("db", "healthy"))

	runtime, err := NewRuntime(m, nil)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = runtime.Run(ctx)
	require.NoError(t, err)

	dbTask := runtime.tasks["db"]
	apiTask := runtime.tasks["api"]

	require.True(t, dbTask.IsHealthy())
	require.Equal(t, TaskStatusDeployed, apiTask.status)
}

func TestRuntime_LogStreaming(t *testing.T) {
	tmpDir := t.TempDir()

	m := NewManifest()
	m.AddSpec(NewSpec("service1").
		Image("alpine", "latest").
		Args("sh", "-c", "echo hello && echo world && sleep 3600"))

	runtime, err := NewRuntime(m, &RuntimeConfig{
		LogsLocation: tmpDir,
	})
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = runtime.Run(ctx)
	require.NoError(t, err)

	// Give time for logs to be written
	time.Sleep(200 * time.Millisecond)

	// Check log file was created
	logPath := filepath.Join(tmpDir, "service1.log")
	require.FileExists(t, logPath)

	// Read log file
	data, err := os.ReadFile(logPath)
	require.NoError(t, err)

	content := string(data)
	require.Contains(t, content, "hello")
	require.Contains(t, content, "world")
}

func TestRuntime_Success_SimpleService(t *testing.T) {
	m := NewManifest()
	m.AddSpec(longRunningSpec("hello"))

	runTestRuntime(t, m)
}

func TestRuntime_Success_TwoIndependentServices(t *testing.T) {
	m := NewManifest()
	m.AddSpec(longRunningSpec("service1"))
	m.AddSpec(longRunningSpec("service2"))

	runTestRuntime(t, m)
}

func TestRuntime_Success_ServiceWithDependency(t *testing.T) {
	m := NewManifest()
	m.AddSpec(longRunningSpec("db"))
	m.AddSpec(longRunningSpec("api").
		DependsOn("db", "started"))

	runTestRuntime(t, m)
}

func TestRuntime_Success_HealthCheckDependency(t *testing.T) {
	server, _ := internal.MockHealthServer(2)
	defer server.Close()

	m := NewManifest()
	m.AddSpec(longRunningSpec("db").
		HealthCheckHTTP(server.URL, http.StatusOK, 10*time.Millisecond, 50*time.Millisecond, 2*time.Second))
	m.AddSpec(longRunningSpec("api").
		DependsOn("db", "healthy"))

	runTestRuntime(t, m)
}

func TestRuntime_Success_TemplateResolution(t *testing.T) {
	m := NewManifest()
	m.AddSpec(longRunningSpec("server").
		Port("http", 8080))
	m.AddSpec(longRunningSpec("client").
		Args("sh", "-c", "echo 'Connecting to {{Service \"server\" \"http\"}}' && sleep 3600").
		DependsOn("server", "started"))

	runTestRuntime(t, m)
}

func TestRuntime_Success_ComplexGraph(t *testing.T) {
	m := NewManifest()
	m.AddSpec(longRunningSpec("service1"))
	m.AddSpec(longRunningSpec("service2"))
	m.AddSpec(longRunningSpec("service3").
		DependsOn("service1", "started").
		DependsOn("service2", "started"))
	m.AddSpec(longRunningSpec("service4").
		DependsOn("service3", "started"))

	runTestRuntime(t, m)
}

func runTestRuntime(t *testing.T, m *Manifest) {
	t.Helper()

	// Create log directory with test name and timestamp
	timestamp := time.Now().Format("20060102-150405")
	logsDir := fmt.Sprintf("e2e-logs/%s-%s", t.Name(), timestamp)

	runtime, err := NewRuntime(m, &RuntimeConfig{
		LogsLocation: logsDir,
	})
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = runtime.Run(ctx)
	require.NoError(t, err)
	runtime.Stop()
}

func verifyAllTasksStopped(t *testing.T, tasks map[string]*taskRef) {
	for _, task := range tasks {
		status := task.GetStatus()
		require.Contains(t, []TaskStatus{TaskStatusPending, TaskStatusStopped}, status,
			"task should be either pending (never started) or stopped")
	}
}

func TestRuntime_Cleanup_ContextCancelledDuringDeployment(t *testing.T) {
	m := NewManifest()
	// Add a service that takes time to pull (not in cache)
	m.AddSpec(longRunningSpec("service1"))
	m.AddSpec(longRunningTaskImage(t, "service2"))

	runtime, err := NewRuntime(m, nil)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(t.Context())

	// Cancel context shortly after starting deployment
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	err = runtime.Run(ctx)
	require.Error(t, err)
	require.Equal(t, context.Canceled, err)
	verifyAllTasksStopped(t, runtime.tasks)
}

func TestRuntime_Cleanup_TaskFailsDuringDeployment(t *testing.T) {
	m := NewManifest()
	// Create a task that will fail immediately while the other
	// is in RunTask state (pulling image)
	m.AddSpec(NewSpec("failing-task").
		Image("alpine", "latest").
		Args("sh", "-c", "exit 1")) // Exits with error code 1
	m.AddSpec(longRunningTaskImage(t, "service2"))

	runtime, err := NewRuntime(m, nil)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = runtime.Run(ctx)
	require.Error(t, err)
	verifyAllTasksStopped(t, runtime.tasks)
}

func TestRuntime_Cleanup_TaskExitsDuringDeployment(t *testing.T) {
	m := NewManifest()
	// Create a task that exits cleanly (without error) but shouldn't exit during deployment
	m.AddSpec(NewSpec("exiting-task").
		Image("alpine", "latest").
		Args("sh", "-c", "echo 'done' && exit 0")) // Exits cleanly
	m.AddSpec(longRunningTaskImage(t, "service2"))

	runtime, err := NewRuntime(m, nil)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = runtime.Run(ctx)
	require.Error(t, err)

	require.Contains(t, err.Error(), "exited unexpectedly during deployment")
	verifyAllTasksStopped(t, runtime.tasks)
}

func TestRuntime_Cleanup_StopCalledDuringDeployment(t *testing.T) {
	m := NewManifest()
	m.AddSpec(longRunningSpec("service1"))
	m.AddSpec(longRunningSpec("service2"))

	runtime, err := NewRuntime(m, nil)
	require.NoError(t, err)

	ctx := context.Background()

	// Call Stop() shortly after starting deployment
	go func() {
		time.Sleep(100 * time.Millisecond)
		runtime.Stop()
	}()

	err = runtime.Run(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "runtime stopped during deployment")
	verifyAllTasksStopped(t, runtime.tasks)
}

type mockTaskUpdater struct {
	updateCh chan struct{}
	mu       sync.Mutex
}

func newMockTaskUpdater() *mockTaskUpdater {
	return &mockTaskUpdater{
		updateCh: make(chan struct{}, 10),
	}
}

func (m *mockTaskUpdater) NotifyTaskUpdated() {
	m.updateCh <- struct{}{}
}

func runTaskRef(ctx context.Context, task *taskRef) <-chan error {
	errCh := make(chan error)

	go func() {
		errCh <- task.Run(ctx)
	}()

	return errCh
}

func waitForTaskReady(t *testing.T, task *taskRef, updater *mockTaskUpdater, errCh <-chan error) {
	for {
		if task.GetStatus() == TaskStatusDeployed {
			break
		}
		select {
		case <-updater.updateCh:
		case err := <-errCh:
			t.Fatal(err)
		case <-time.After(20 * time.Second):
			t.Fatal("task failed to deploy")
		}
	}
}

func TestTaskRef_SuccessfulWorkflow(t *testing.T) {
	spec := longRunningSpec("test-task")

	driver, err := NewDocker()
	require.NoError(t, err)

	updater := newMockTaskUpdater()
	logger := newTestLogger()

	ctx, cancel := context.WithCancel(t.Context())

	task := newTaskRef(logger, spec, driver, nil, updater)
	errCh := runTaskRef(ctx, task)

	// MarkAsReady closes the readyCh and starts the task
	task.MarkAsReady()

	waitForTaskReady(t, task, updater, errCh)

	// check that calling MarkAsReady twice does not create any side effect
	// and the task does not get deployed again
	task.MarkAsReady()

	select {
	case <-time.After(3 * time.Second):
	case <-updater.updateCh:
		t.Fatal("mark as ready created side effects")
	}

	// cancel the context and wait for the Run to finish and return no error
	cancel()
	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("unexpected error cleaning up the task: %v", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("failed to close task")
	}

	// the status of the task is stopped now
	require.Equal(t, task.GetStatus(), TaskStatusStopped)
}

func TestTaskRef_ContextCancelled_BeforeReady(t *testing.T) {
	spec := longRunningSpec("test-task")
	driver, err := NewDocker()
	require.NoError(t, err)

	updater := newMockTaskUpdater()
	logger := newTestLogger()

	task := newTaskRef(logger, spec, driver, nil, updater)

	ctx, cancel := context.WithCancel(context.Background())

	errCh := runTaskRef(ctx, task)
	cancel() // Cancel immediately

	err = <-errCh
	require.NoError(t, err)
	require.Equal(t, task.GetStatus(), TaskStatusStopped)
}

func TestTaskRef_ContextCancelled_WhileTaskStarting(t *testing.T) {
	// use an image that we do not have locally, it will need to download it
	spec := longRunningTaskImage(t, "test-task")
	driver, err := NewDocker()
	require.NoError(t, err)

	updater := newMockTaskUpdater()
	logger := newTestLogger()

	task := newTaskRef(logger, spec, driver, nil, updater)

	ctx, cancel := context.WithCancel(context.Background())

	errCh := runTaskRef(ctx, task)
	task.MarkAsReady()

	// wait a little bit for ready to close first and then cancel right away
	// such that context cancel happens inside the driver on RunTask
	time.Sleep(100 * time.Millisecond)
	cancel()

	err = <-errCh
	require.NoError(t, err)
	require.Equal(t, TaskStatusStopped, task.GetStatus())
}

func TestTaskRef_ContextCancelled_AfterStarted(t *testing.T) {
	// use an image that we do not have locally, it will need to download it
	spec := longRunningTaskImage(t, "test-task")
	driver, err := NewDocker()
	require.NoError(t, err)

	updater := newMockTaskUpdater()
	logger := newTestLogger()

	task := newTaskRef(logger, spec, driver, nil, updater)

	ctx, cancel := context.WithCancel(context.Background())

	errCh := runTaskRef(ctx, task)
	task.MarkAsReady()

	waitForTaskReady(t, task, updater, errCh)
	cancel()

	err = <-errCh
	require.NoError(t, err)
	require.Equal(t, TaskStatusStopped, task.GetStatus())
}

func TestTaskRef_Run_WithHealthCheck(t *testing.T) {
	server, getCallCount := internal.MockHealthServer(2)
	defer server.Close()

	spec := longRunningSpec("test-task-health").
		HealthCheckHTTP(server.URL, http.StatusOK, 10*time.Millisecond, 50*time.Millisecond, 1*time.Second)

	driver, err := NewDocker()
	require.NoError(t, err)

	updater := newMockTaskUpdater()
	logger := newTestLogger()

	task := newTaskRef(logger, spec, driver, nil, updater)

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	errCh := runTaskRef(ctx, task)

	// Mark as ready and wait for the task to be deployed, then the probe will start
	task.MarkAsReady()
	waitForTaskReady(t, task, updater, errCh)

	// wait some time for the probe to start
	time.Sleep(100 * time.Millisecond)

	require.True(t, task.IsHealthy())
	require.Greater(t, getCallCount(), 1)
	require.NotNil(t, task.prober)

	// Cleanup
	cancel()
	<-errCh
}

func TestTaskRef_Run_TaskFinishes(t *testing.T) {
	spec := longRunningSpec("test-task-cleanup").
		Args("sh", "-c", "echo 'api started' && sleep 2")
	driver, err := NewDocker()
	require.NoError(t, err)

	updater := newMockTaskUpdater()
	logger := newTestLogger()

	task := newTaskRef(logger, spec, driver, nil, updater)

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	errCh := runTaskRef(ctx, task)

	task.MarkAsReady()
	waitForTaskReady(t, task, updater, errCh)

	// the task should stop after after 2 seconds
	select {
	case <-updater.updateCh:
		// the only update we expect is to notify that the task has stopped
	case <-time.After(3 * time.Second):
		t.Fatal("task should have stopped")
	}

	status := task.GetStatus()
	require.Equal(t, status, TaskStatusStopped)

	err = <-errCh
	require.Nil(t, err)
}

func TestTaskRef_Run_TaskFails(t *testing.T) {
	spec := longRunningSpec("test-task-cleanup").
		Args("sh", "-c", "echo 'running' && sleep 2 && exit 1")
	driver, err := NewDocker()
	require.NoError(t, err)

	updater := newMockTaskUpdater()
	logger := newTestLogger()

	task := newTaskRef(logger, spec, driver, nil, updater)

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	errCh := runTaskRef(ctx, task)
	task.MarkAsReady()

	err = <-errCh
	require.NotNil(t, err)
	require.Equal(t, task.GetStatus(), TaskStatusStopped)
}

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, nil))
}

func longRunningSpec(name string) *Spec {
	return NewSpec(name).
		Image("alpine", "latest").
		Args("sh", "-c", "echo 'api started' && sleep 30000")
}

func longRunningTaskImage(t *testing.T, name string) *Spec {
	// this function returns an spec that will take some time to be in deployment
	// state since it will try to use an image that is not found locally
	// it is useful to test failure/cleanup scenarios in which one of the tasks
	// fails or stops (not this one) while the Run function is still waiting
	// for other tasks to finish (this one)
	t.Helper()

	docker, err := NewDocker()
	require.NoError(t, err)

	imageName, tag := "ubuntu", "latest"

	// Remove the image, ignore errors if it doesn't exist
	_, _ = docker.client.ImageRemove(context.Background(), fmt.Sprintf("%s:%s", imageName, tag), image.RemoveOptions{
		Force:         true,
		PruneChildren: true,
	})

	return NewSpec(name).
		Image(imageName, tag).
		Args("sh", "-c", "echo 'api started' && sleep 30000")
}
