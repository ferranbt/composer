package docker

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ferranbt/composer/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func NewTestDockerDriver(t *testing.T) *Provider {
	d := NewProvider()
	return d
}

func TestDriver_Start_WaitFinished(t *testing.T) {
	d := NewTestDockerDriver(t)

	tt := &ComputeResource{
		Service: &proto.Service{
			Image: "busybox:1.29.3",
			Args:  []string{"echo", "hello"},
		},
	}
	handle, err := d.Create(tt)
	assert.NoError(t, err)

	defer d.DestroyTask(handle.ContainerId, true)

	waitCh, _ := d.WaitTask(context.Background(), handle.ContainerId)
	select {
	case res := <-waitCh:
		assert.True(t, res.Successful())
	case <-time.After(time.Second):
		t.Fatalf("timeout")
	}
}

func TestDriver_Start_Wait(t *testing.T) {
	d := NewTestDockerDriver(t)

	tt := &ComputeResource{
		Service: &proto.Service{
			Image: "busybox:1.29.3",
			Args:  []string{"nc", "-l", "-p", "3000", "127.0.0.1"},
		},
	}
	handle, err := d.Create(tt)
	assert.NoError(t, err)

	defer d.DestroyTask(handle.ContainerId, true)

	waitCh, _ := d.WaitTask(context.Background(), handle.ContainerId)
	select {
	case res := <-waitCh:
		t.Fatalf("it should not finish yet: %v", res)
	case <-time.After(time.Second):
	}
}

func TestDriver_Start_Kill_Wait(t *testing.T) {
	d := NewTestDockerDriver(t)

	tt := &ComputeResource{
		Service: &proto.Service{
			Image: "busybox:1.29.3",
			Args:  []string{"echo", "hello"},
		},
	}
	handle, err := d.Create(tt)
	assert.NoError(t, err)

	defer d.DestroyTask(handle.ContainerId, true)

	waitCh, _ := d.WaitTask(context.Background(), handle.ContainerId)

	err = d.StopTask(handle.ContainerId, time.Second)
	assert.NoError(t, err)

	select {
	case res := <-waitCh:
		assert.True(t, res.Successful())
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}
}

func TestDriver_Start_Kill_Timeout(t *testing.T) {
	d := NewTestDockerDriver(t)

	tt := &ComputeResource{
		Service: &proto.Service{
			Image: "busybox:1.29.3",
			Args:  []string{"sleep", "10"},
		},
	}
	handle, err := d.Create(tt)
	assert.NoError(t, err)

	defer d.DestroyTask(handle.ContainerId, true)

	waitCh, _ := d.WaitTask(context.Background(), handle.ContainerId)

	err = d.StopTask(handle.ContainerId, time.Second)
	assert.NoError(t, err)

	select {
	case res := <-waitCh:
		assert.Equal(t, res.ExitCode, uint64(137))
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}
}

func TestDriver_Exec(t *testing.T) {
	d := NewTestDockerDriver(t)

	tt := &ComputeResource{
		Service: &proto.Service{
			Image: "busybox:1.29.3",
			Args:  []string{"sleep", "10"},
		},
	}
	handle, err := d.Create(tt)
	assert.NoError(t, err)

	defer d.DestroyTask(handle.ContainerId, true)

	// send a command that returns true
	res, err := d.Exec(handle.ContainerId, []string{"echo", "a"})
	require.NoError(t, err)

	require.Zero(t, res.ExitCode)
	require.Empty(t, res.Stderr)
	require.Equal(t, strings.TrimSpace(string(res.Stdout)), "a")

	// send a command that should fail (command not found)
	res, err = d.Exec(handle.ContainerId, []string{"curl"})
	require.NoError(t, err)
	require.NotZero(t, res.ExitCode)
}

func TestDriver_BindMount(t *testing.T) {
	d := NewTestDockerDriver(t)

	bindDir, err := os.MkdirTemp("/tmp", "driver-")
	require.NoError(t, err)

	tt := &ComputeResource{
		Service: &proto.Service{
			Image: "busybox:1.29.3",
			Args:  []string{"sleep", "10"},
			Mounts: []*proto.Service_MountConfig{
				{HostPath: bindDir, TaskPath: "/var"},
			},
		},
	}

	handle, err := d.Create(tt)
	assert.NoError(t, err)

	// touch the file /var/file.txt should be visible
	// on the bind folder as {bindDir}/file.txt
	res, err := d.Exec(handle.ContainerId, []string{"touch", "/var/file.txt"})
	require.NoError(t, err)
	require.Zero(t, res.ExitCode)

	_, err = os.Stat(filepath.Join(bindDir, "file.txt"))
	require.NoError(t, err)
}
