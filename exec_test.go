package composer

import (
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type testExecUpdater struct {
	updates []TaskStatus
	mu      sync.Mutex
	done    chan struct{}
}

func newTestExecUpdater() *testExecUpdater {
	return &testExecUpdater{
		done: make(chan struct{}),
	}
}

func (t *testExecUpdater) TaskUpdated(status TaskStatus) {
	t.mu.Lock()
	t.updates = append(t.updates, status)
	t.mu.Unlock()
	if status == TaskStatusStopped {
		close(t.done)
	}
}

func TestExecBasicCommand(t *testing.T) {
	updater := newTestExecUpdater()
	e := NewExec()

	spec := NewSpec("test").
		UseExec().
		Args("echo", "hello exec")

	handle, err := e.RunTask(t.Context(), spec, updater)
	require.NoError(t, err)
	defer handle.Stop()

	select {
	case <-updater.done:
		updater.mu.Lock()
		require.Contains(t, updater.updates, TaskStatusDeployed)
		require.Contains(t, updater.updates, TaskStatusStopped)
		updater.mu.Unlock()
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for task completion")
	}
}

func TestExecLogs(t *testing.T) {
	updater := newTestExecUpdater()
	e := NewExec()

	spec := NewSpec("test").
		UseExec().
		Args("sh", "-c", "echo stdout && echo stderr >&2")

	handle, err := e.RunTask(t.Context(), spec, updater)
	require.NoError(t, err)
	defer handle.Stop()

	// Get logs before process completes
	logs, err := handle.Logs()
	require.NoError(t, err)
	defer logs.Close()

	data, err := io.ReadAll(logs)
	require.NoError(t, err)

	output := string(data)
	require.Contains(t, output, "stdout")
	require.Contains(t, output, "stderr")
}

func TestExecEnv(t *testing.T) {
	updater := newTestExecUpdater()
	e := NewExec()

	spec := NewSpec("test").
		UseExec().
		Args("sh", "-c", "echo $TEST_VAR").
		Env("TEST_VAR", "test_value_exec")

	handle, err := e.RunTask(t.Context(), spec, updater)
	require.NoError(t, err)
	defer handle.Stop()

	logs, err := handle.Logs()
	require.NoError(t, err)
	defer logs.Close()

	data, err := io.ReadAll(logs)
	require.NoError(t, err)

	output := strings.TrimSpace(string(data))
	require.Contains(t, output, "test_value_exec")
}

func TestExecStop(t *testing.T) {
	updater := newTestExecUpdater()
	e := NewExec()

	spec := NewSpec("test").
		UseExec().
		Args("sleep", "30")

	handle, err := e.RunTask(t.Context(), spec, updater)
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	err = handle.Stop()
	require.NoError(t, err)

	select {
	case <-updater.done:
		// Process stopped successfully
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for process to stop")
	}
}

func TestExecError(t *testing.T) {
	updater := newTestExecUpdater()
	e := NewExec()

	spec := NewSpec("test").
		UseExec().
		Args("sh", "-c", "exit 42")

	handle, err := e.RunTask(t.Context(), spec, updater)
	require.NoError(t, err)
	defer handle.Stop()

	select {
	case <-updater.done:
		err := handle.Err()
		require.Error(t, err)
		require.Contains(t, err.Error(), "exit status 42")
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for task completion")
	}
}

func TestExecMultipleArgs(t *testing.T) {
	updater := newTestExecUpdater()
	e := NewExec()

	spec := NewSpec("test").
		UseExec().
		Args("sh", "-c", "echo arg1 && echo arg2 && echo arg3")

	handle, err := e.RunTask(t.Context(), spec, updater)
	require.NoError(t, err)
	defer handle.Stop()

	logs, err := handle.Logs()
	require.NoError(t, err)
	defer logs.Close()

	data, err := io.ReadAll(logs)
	require.NoError(t, err)

	output := string(data)
	require.Contains(t, output, "arg1")
	require.Contains(t, output, "arg2")
	require.Contains(t, output, "arg3")
}

func TestExecMultipleEnvVars(t *testing.T) {
	updater := newTestExecUpdater()
	e := NewExec()

	spec := NewSpec("test").
		UseExec().
		Args("sh", "-c", "echo $VAR1:$VAR2:$VAR3").
		Env("VAR1", "value1").
		Env("VAR2", "value2").
		Env("VAR3", "value3")

	handle, err := e.RunTask(t.Context(), spec, updater)
	require.NoError(t, err)
	defer handle.Stop()

	logs, err := handle.Logs()
	require.NoError(t, err)
	defer logs.Close()

	data, err := io.ReadAll(logs)
	require.NoError(t, err)

	output := strings.TrimSpace(string(data))
	require.Equal(t, "value1:value2:value3", output)
}

func TestExecLongRunningProcess(t *testing.T) {
	updater := newTestExecUpdater()
	e := NewExec()

	spec := NewSpec("test").
		UseExec().
		Args("sh", "-c", "for i in 1 2 3 4 5; do echo line$i; sleep 0.1; done")

	handle, err := e.RunTask(t.Context(), spec, updater)
	require.NoError(t, err)
	defer handle.Stop()

	logs, err := handle.Logs()
	require.NoError(t, err)
	defer logs.Close()

	// Read logs concurrently
	var output string
	var readErr error
	done := make(chan struct{})
	go func() {
		data, err := io.ReadAll(logs)
		readErr = err
		output = string(data)
		close(done)
	}()

	select {
	case <-updater.done:
		<-done
		require.NoError(t, readErr)
		for i := 1; i <= 5; i++ {
			require.Contains(t, output, "line"+string(rune('0'+i)))
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for task completion")
	}
}
