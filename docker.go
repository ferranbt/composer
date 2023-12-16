package composer

import (
	"context"
	"io"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/ferranbt/composer/proto"
)

type ComputeResource struct {
	*proto.Service

	Name    string
	Project string
}

type ComputeUpdate struct {
	Name    string
	Project string

	// The different update events
	Completed *ComputeUpdateResourceCompleted
	Failed    *ComputeUpdateResourceFailed
	Created   *ComputeUpdateResourceCreated
}

type ComputeUpdateResourceCompleted struct {
	exitResult *proto.ExitResult
}

type ComputeUpdateResourceFailed struct {
}

type ComputeUpdateResourceCreated struct {
	IP string
}

type Updater interface {
	Update(c *ComputeUpdate) error
}

type Provider interface {
	Create(c *ComputeResource) error
}

type dockerProvider struct {
	cli     *client.Client
	updater Updater
}

func newDockerProvider(updater Updater) *dockerProvider {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	return &dockerProvider{cli: cli, updater: updater}
}

func (d *dockerProvider) Reattach(project, id string, handle *proto.TaskState_Handle) {
	go d.containerWait(project, id, handle.ContainerId)
}

func (d *dockerProvider) containerWait(project, name, containerID string) {
	statusCh, errCh := d.cli.ContainerWait(context.Background(), containerID, container.WaitConditionNotRunning)

	var status container.WaitResponse
	select {
	case err := <-errCh:
		if err != nil {
			// TODO: unit test
			panic(err)
		}
	case status = <-statusCh:
	}

	exitResult := &proto.ExitResult{
		ExitCode: uint64(status.StatusCode),
	}

	d.updater.Update(&ComputeUpdate{Name: name, Project: project, Completed: &ComputeUpdateResourceCompleted{
		exitResult: exitResult,
	}})
}

func (d *dockerProvider) Kill(id string) error {
	return d.cli.ContainerKill(context.Background(), id, "SIGKILL")
}

func (d *dockerProvider) Create(c *ComputeResource) (*proto.TaskState_Handle, error) {
	// check if the image exists
	_, _, err := d.cli.ImageInspectWithRaw(context.Background(), c.Image)
	if err != nil {
		// pull the image
		reader, err := d.cli.ImagePull(context.Background(), c.Image, types.ImagePullOptions{})
		if err != nil {
			return nil, err
		}
		defer reader.Close()
		io.Copy(os.Stdout, reader)
	}

	ctx := context.Background()

	env := []string{}
	for k, v := range c.Env {
		env = append(env, k+"="+v)
	}

	labels := map[string]string{
		"compose": "true",
		"project": c.Project,
		"name":    c.Name,
	}

	config := &container.Config{
		Image:  c.Image,
		Cmd:    c.Args,
		Env:    env,
		Labels: labels,
	}
	hostConfig := &container.HostConfig{}

	netConfig := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{},
	}

	resp, err := d.cli.ContainerCreate(ctx, config, hostConfig, netConfig, nil, "")
	if err != nil {
		return nil, err
	}

	if err := d.cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return nil, err
	}

	go d.containerWait(c.Project, c.Name, resp.ID)

	handle := &proto.TaskState_Handle{
		ContainerId: resp.ID,
	}
	return handle, nil
}
