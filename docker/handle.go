package docker

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log/slog"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/ferranbt/composer/proto"
)

type taskHandle struct {
	client      *client.Client
	logger      *slog.Logger
	containerID string

	waitCh chan struct{}

	exitResult     *proto.ExitResult
	exitResultLock sync.Mutex
}

func (t *taskHandle) run() {
	t.logger.Debug("handle running", "id", t.containerID)

	statusCh, errCh := t.client.ContainerWait(context.Background(), t.containerID, container.WaitConditionNotRunning)

	var status container.WaitResponse
	select {
	case err := <-errCh:
		if err != nil {
			// TODO: unit test
			t.logger.Error("failed to wait container", "id", t.containerID, "err", err)
		}
	case status = <-statusCh:
	}

	fmt.Println("- task is done in provider -")

	t.exitResultLock.Lock()
	t.exitResult = &proto.ExitResult{
		ExitCode: uint64(status.StatusCode),
	}
	t.exitResultLock.Unlock()
	close(t.waitCh)
}

func (t *taskHandle) Kill(killTimeout time.Duration) error {
	seconds := int(killTimeout.Seconds())
	if err := t.client.ContainerStop(context.Background(), t.containerID, container.StopOptions{Timeout: &seconds}); err != nil {
		return err
	}
	return nil
}

func (h *taskHandle) Exec(ctx context.Context, args []string) (*proto.ExecTaskResult, error) {
	config := types.ExecConfig{
		AttachStderr: true,
		AttachStdout: true,
		Cmd:          args,
	}

	exec, err := h.client.ContainerExecCreate(ctx, h.containerID, config)
	if err != nil {
		return nil, err
	}

	resp, err := h.client.ContainerExecAttach(ctx, exec.ID, types.ExecStartCheck{})
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

	res, err := h.client.ContainerExecInspect(ctx, exec.ID)
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
