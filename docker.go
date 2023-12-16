package composer

import (
	"bytes"
	"context"
	"io/ioutil"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
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
	cli         *client.Client
	coordinator *dockerImageCoordinator
	updater     Updater
}

func newDockerProvider(updater Updater) *dockerProvider {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	return &dockerProvider{
		cli:         cli,
		coordinator: newDockerImageCoordinator(cli),
		updater:     updater,
	}
}

func (d *dockerProvider) Reattach(project, id string, handle *proto.ServiceState_Handle) {
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

func (d *dockerProvider) Create(c *ComputeResource) (*proto.ServiceState_Handle, error) {
	ctx := context.Background()

	if err := d.createImage(c.Image); err != nil {
		return nil, err
	}

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
	hostConfig := &container.HostConfig{
		NetworkMode: container.NetworkMode(c.Service.NetworkMode),
	}

	netConfig := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{},
	}

	for _, mount := range c.Mounts {
		hostConfig.Binds = append(hostConfig.Binds, mount.HostPath+":"+mount.TaskPath)
	}

	resp, err := d.cli.ContainerCreate(ctx, config, hostConfig, netConfig, nil, "")
	if err != nil {
		return nil, err
	}

	if err := d.cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return nil, err
	}

	container, err := d.cli.ContainerInspect(ctx, resp.ID)
	if err != nil {
		return nil, err
	}

	go d.containerWait(c.Project, c.Name, resp.ID)

	handle := &proto.ServiceState_Handle{
		ContainerId: resp.ID,
		Ip:          container.NetworkSettings.IPAddress,
	}
	return handle, nil
}

func (d *dockerProvider) Exec(containerID string, args []string) (*proto.ExecTaskResult, error) {
	ctx := context.Background()

	config := types.ExecConfig{
		AttachStderr: true,
		AttachStdout: true,
		Cmd:          args,
	}

	exec, err := d.cli.ContainerExecCreate(ctx, containerID, config)
	if err != nil {
		return nil, err
	}

	resp, err := d.cli.ContainerExecAttach(ctx, exec.ID, types.ExecStartCheck{})
	if err != nil {
		return nil, err
	}
	defer resp.Close()

	// read the output
	var outBuf, errBuf bytes.Buffer
	outputDone := make(chan error)

	go func() {
		// StdCopy demultiplexes the stream into two buffers
		_, err = stdcopy.StdCopy(&outBuf, &errBuf, resp.Reader)
		outputDone <- err
	}()

	select {
	case err := <-outputDone:
		if err != nil {
			return nil, err
		}
		break

	case <-ctx.Done():
		return nil, ctx.Err()
	}

	stdout, err := ioutil.ReadAll(&outBuf)
	if err != nil {
		return nil, err
	}
	stderr, err := ioutil.ReadAll(&errBuf)
	if err != nil {
		return nil, err
	}

	res, err := d.cli.ContainerExecInspect(ctx, exec.ID)
	if err != nil {
		return nil, err
	}

	execResult := &proto.ExecTaskResult{
		ExitCode: uint64(res.ExitCode),
		Stdout:   string(stdout),
		Stderr:   string(stderr),
	}
	return execResult, nil
}

func (d *dockerProvider) createImage(image string) error {
	_, dockerImageRaw, _ := d.cli.ImageInspectWithRaw(context.Background(), image)
	if dockerImageRaw != nil {
		// already available
		return nil
	}
	if _, err := d.coordinator.PullImage(image); err != nil {
		return err
	}
	return nil
}
