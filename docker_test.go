package composer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ferranbt/composer/proto"
	"github.com/stretchr/testify/require"
)

type mockUpdater struct {
}

func (m *mockUpdater) Update(c *ComputeUpdate) error {
	return nil
}

func TestDriver_BindMount(t *testing.T) {
	d := newDockerProvider(&mockUpdater{})

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
	require.NoError(t, err)

	defer d.Kill(handle.ContainerId)

	// touch the file /var/file.txt should be visible
	// on the bind folder as {bindDir}/file.txt
	res, err := d.Exec(handle.ContainerId, []string{"touch", "/var/file.txt"})
	require.NoError(t, err)
	require.Zero(t, res.ExitCode)

	_, err = os.Stat(filepath.Join(bindDir, "file.txt"))
	require.NoError(t, err)
}
