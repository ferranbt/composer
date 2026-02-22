package composer

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

var (
	_ Driver     = (*Docker)(nil)
	_ TaskHandle = (*dockerTask)(nil)
)

type ImagePuller struct {
	client  *client.Client
	pulling map[string]chan struct{}
	mu      sync.Mutex
}

func NewImagePuller(client *client.Client) *ImagePuller {
	return &ImagePuller{
		client:  client,
		pulling: make(map[string]chan struct{}),
	}
}

func (p *ImagePuller) Pull(ctx context.Context, imageRef string) error {
	p.mu.Lock()

	// Check if image is already being pulled
	if ch, exists := p.pulling[imageRef]; exists {
		p.mu.Unlock()
		// Wait for existing pull to complete
		<-ch
		return nil
	}

	// Create channel to signal completion
	ch := make(chan struct{})
	p.pulling[imageRef] = ch
	p.mu.Unlock()

	// Perform the pull
	defer func() {
		p.mu.Lock()
		delete(p.pulling, imageRef)
		close(ch)
		p.mu.Unlock()
	}()

	reader, err := p.client.ImagePull(ctx, imageRef, image.PullOptions{})
	if err != nil {
		return err
	}
	defer reader.Close()

	// Drain the reader to ensure pull completes
	_, err = io.Copy(io.Discard, reader)
	return err
}

type Docker struct {
	client      *client.Client
	imagePuller *ImagePuller
}

func NewDocker() (*Docker, error) {
	client, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	d := &Docker{
		client:      client,
		imagePuller: NewImagePuller(client),
	}
	return d, nil
}

func (d *Docker) RunTask(ctx context.Context, s *Spec, updater DriverUpdater) (TaskHandle, error) {
	imageRef := fmt.Sprintf("%s:%s", s.image, s.tag)

	// Pull image if not available
	if _, err := d.client.ImageInspect(ctx, imageRef); err != nil {
		if err := d.imagePuller.Pull(ctx, imageRef); err != nil {
			return nil, fmt.Errorf("failed to pull image: %w", err)
		}
	}

	var env []string
	for k, v := range s.env {
		env = append(env, k+"="+v)
	}

	exposedPorts := nat.PortSet{}
	portBindings := nat.PortMap{}

	for _, port := range s.ports {
		containerPort := nat.Port(fmt.Sprintf("%d/tcp", port.Port))
		exposedPorts[containerPort] = struct{}{}
		portBindings[containerPort] = []nat.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: fmt.Sprintf("%d", port.Port),
			},
		}
	}

	config := &container.Config{
		Image:        imageRef,
		Cmd:          s.args,
		Env:          env,
		ExposedPorts: exposedPorts,
	}

	hostConfig := &container.HostConfig{
		PortBindings: portBindings,
	}

	resp, err := d.client.ContainerCreate(ctx, config, hostConfig, nil, nil, "")
	if err != nil {
		return nil, err
	}

	if err := d.client.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return nil, err
	}

	task := &dockerTask{
		client:      d.client,
		spec:        s,
		containerID: resp.ID,
		updater:     updater,
	}
	go task.run()

	return task, nil
}

type dockerTask struct {
	client      *client.Client
	spec        *Spec
	containerID string
	updater     DriverUpdater
	err         error
	errLock     sync.RWMutex
}

func (t *dockerTask) run() {
	t.updater.TaskUpdated(TaskStatusDeployed)

	ctx := context.Background()
	statusCh, errCh := t.client.ContainerWait(ctx, t.containerID, container.WaitConditionNotRunning)

	select {
	case status := <-statusCh:
		t.errLock.Lock()
		if status.Error != nil {
			t.err = fmt.Errorf("container error: %s", status.Error.Message)
		} else if status.StatusCode != 0 {
			t.err = fmt.Errorf("container exited with code %d", status.StatusCode)
		}
		t.errLock.Unlock()
		t.updater.TaskUpdated(TaskStatusStopped)
	case err := <-errCh:
		t.errLock.Lock()
		t.err = err
		t.errLock.Unlock()
		t.updater.TaskUpdated(TaskStatusStopped)
	}
}

func (t *dockerTask) Stop() error {
	ctx := context.Background()
	timeout := 0
	return t.client.ContainerStop(ctx, t.containerID, container.StopOptions{
		Timeout: &timeout,
	})
}

func (t *dockerTask) Logs() (io.ReadCloser, error) {
	ctx := context.Background()
	return t.client.ContainerLogs(ctx, t.containerID, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
	})
}

func (t *dockerTask) Err() error {
	t.errLock.RLock()
	defer t.errLock.RUnlock()
	return t.err
}
