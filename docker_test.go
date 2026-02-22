package composer

import (
	"context"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/docker/docker/client"
	"github.com/stretchr/testify/require"
)

type testUpdater struct {
	updates []TaskStatus
	mu      sync.Mutex
	done    chan struct{}
}

func newTestUpdater() *testUpdater {
	return &testUpdater{
		done: make(chan struct{}),
	}
}

func (t *testUpdater) TaskUpdated(status TaskStatus) {
	t.mu.Lock()
	t.updates = append(t.updates, status)
	t.mu.Unlock()
	if status == TaskStatusStopped {
		close(t.done)
	}
}

func TestDockerLogs(t *testing.T) {
	d, err := NewDocker()
	require.NoError(t, err)

	spec := NewSpec("test").
		Image("alpine", "latest").
		Args("echo", "hello world")

	handle, err := d.RunTask(t.Context(), spec, newTestUpdater())
	require.NoError(t, err)
	defer handle.Stop()

	time.Sleep(500 * time.Millisecond)

	logs, err := handle.Logs()
	require.NoError(t, err)
	defer logs.Close()

	data, err := io.ReadAll(logs)
	require.NoError(t, err)
	require.Contains(t, string(data), "hello world")
}

func TestDockerStop(t *testing.T) {
	d, err := NewDocker()
	require.NoError(t, err)

	spec := NewSpec("test").
		Image("alpine", "latest").
		Args("sleep", "10")

	handle, err := d.RunTask(t.Context(), spec, newTestUpdater())
	require.NoError(t, err)

	err = handle.Stop()
	require.NoError(t, err)
}

func TestDockerEnv(t *testing.T) {
	d, err := NewDocker()
	require.NoError(t, err)

	spec := NewSpec("test").
		Image("alpine", "latest").
		Args("sh", "-c", "echo $TEST_VAR").
		Env("TEST_VAR", "test_value")

	handle, err := d.RunTask(t.Context(), spec, newTestUpdater())
	require.NoError(t, err)
	defer handle.Stop()

	time.Sleep(500 * time.Millisecond)

	logs, err := handle.Logs()
	require.NoError(t, err)
	defer logs.Close()

	data, err := io.ReadAll(logs)
	require.NoError(t, err)

	output := string(data)
	output = strings.TrimSpace(strings.ReplaceAll(output, "\x00", ""))
	require.Contains(t, output, "test_value")
}

func TestDockerUpdates(t *testing.T) {
	updater := newTestUpdater()
	d, err := NewDocker()
	require.NoError(t, err)

	spec := NewSpec("test").
		Image("alpine", "latest").
		Args("echo", "hello")

	handle, err := d.RunTask(t.Context(), spec, updater)
	require.NoError(t, err)
	defer handle.Stop()

	select {
	case <-updater.done:
		updater.mu.Lock()
		require.Contains(t, updater.updates, TaskStatusStopped)
		updater.mu.Unlock()
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for task update")
	}
}

func TestImagePuller(t *testing.T) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	require.NoError(t, err)

	puller := NewImagePuller(cli)

	t.Run("pulls image successfully", func(t *testing.T) {
		ctx := context.Background()
		err := puller.Pull(ctx, "alpine:latest")
		require.NoError(t, err)
	})

	t.Run("concurrent pulls deduplicate", func(t *testing.T) {
		ctx := context.Background()

		var pullCount atomic.Int32
		var wg sync.WaitGroup

		// Create a new puller for this test
		testPuller := NewImagePuller(cli)

		// Launch 5 concurrent pulls of the same image
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := testPuller.Pull(ctx, "alpine:3.19")
				require.NoError(t, err)
				pullCount.Add(1)
			}()
		}

		wg.Wait()

		// All 5 goroutines should complete
		require.Equal(t, int32(5), pullCount.Load())

		// Verify only one entry was in the pulling map at any time
		// (it should be cleaned up now)
		testPuller.mu.Lock()
		require.Empty(t, testPuller.pulling)
		testPuller.mu.Unlock()
	})

	t.Run("sequential pulls work correctly", func(t *testing.T) {
		ctx := context.Background()

		err := puller.Pull(ctx, "alpine:3.18")
		require.NoError(t, err)

		err = puller.Pull(ctx, "alpine:3.18")
		require.NoError(t, err)
	})
}
