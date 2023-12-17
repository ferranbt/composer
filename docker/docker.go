package docker

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/ferranbt/composer/proto"
)

type ComputeResource struct {
	*proto.Service
}

type Provider struct {
	cli         *client.Client
	coordinator *dockerImageCoordinator
	store       *taskStore
}

func NewProvider() *Provider {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	return &Provider{
		cli:         cli,
		coordinator: newDockerImageCoordinator(cli),
		store:       newTaskStore(),
	}
}

func (d *Provider) RecoverTask(taskID string, task *proto.ServiceState_Handle) error {
	if _, ok := d.store.Get(taskID); ok {
		return nil
	}

	container, err := d.cli.ContainerInspect(context.Background(), task.ContainerId)
	if err != nil {
		return err
	}
	if !container.State.Running {
		return fmt.Errorf("container is not running")
	}

	h := &taskHandle{
		logger:      slog.Default(),
		client:      d.cli,
		containerID: task.ContainerId,
		waitCh:      make(chan struct{}),
	}

	d.store.Set(taskID, h)
	go h.run()

	return nil
}

var ErrTaskNotFound = fmt.Errorf("task not found")

func (d *Provider) WaitTask(ctx context.Context, taskID string) (<-chan *proto.ExitResult, error) {
	handle, ok := d.store.Get(taskID)
	if !ok {
		return nil, ErrTaskNotFound
	}
	ch := make(chan *proto.ExitResult)
	go func() {
		defer close(ch)

		select {
		case <-handle.waitCh:
			ch <- handle.exitResult
		case <-ctx.Done():
			ch <- &proto.ExitResult{}
		}
	}()
	return ch, nil
}

func (d *Provider) StopTask(taskID string, timeout time.Duration) error {
	fmt.Println("_ STOP TASK _")
	h, ok := d.store.Get(taskID)
	if !ok {
		return ErrTaskNotFound
	}

	return h.Kill(timeout)
}

func (d *Provider) DestroyTask(taskID string, force bool) error {
	fmt.Println("_ DESTROY TASK _")
	h, ok := d.store.Get(taskID)
	if !ok {
		return ErrTaskNotFound
	}

	c, err := d.cli.ContainerInspect(context.Background(), h.containerID)
	if err != nil {
		return err
	} else {
		if c.State.Running {
			if !force {
				return fmt.Errorf("cannot destroy if force not set to true")
			}

			if err := d.cli.ContainerStop(context.Background(), h.containerID, container.StopOptions{}); err != nil {
				h.logger.Warn("failed to stop container", "err", err)
			}
		}
	}

	d.store.Delete(taskID)
	return nil
}

func (d *Provider) Create(c *ComputeResource) (*proto.ServiceState_Handle, error) {
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

	h := &taskHandle{
		logger:      slog.Default(),
		client:      d.cli,
		containerID: resp.ID,
		waitCh:      make(chan struct{}),
	}

	d.store.Set(resp.ID, h)
	go h.run()

	handle := &proto.ServiceState_Handle{
		ContainerId: resp.ID,
		Ip:          container.NetworkSettings.IPAddress,
	}
	return handle, nil
}

func (d *Provider) Exec(taskID string, cmd []string) (*proto.ExecTaskResult, error) {
	h, ok := d.store.Get(taskID)
	if !ok {
		return nil, fmt.Errorf("task not found")
	}

	return h.Exec(context.Background(), cmd)
}

func (d *Provider) createImage(image string) error {
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

type taskStore struct {
	store map[string]*taskHandle
	lock  sync.RWMutex
}

func newTaskStore() *taskStore {
	return &taskStore{store: map[string]*taskHandle{}}
}

func (ts *taskStore) Set(id string, handle *taskHandle) {
	ts.lock.Lock()
	defer ts.lock.Unlock()
	ts.store[id] = handle
}

func (ts *taskStore) Get(id string) (*taskHandle, bool) {
	ts.lock.RLock()
	defer ts.lock.RUnlock()
	t, ok := ts.store[id]
	return t, ok
}

func (ts *taskStore) Delete(id string) {
	ts.lock.Lock()
	defer ts.lock.Unlock()
	delete(ts.store, id)
}
