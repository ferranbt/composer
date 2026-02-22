package composer

import (
	"context"
	"io"
	"os/exec"
	"sync"
)

var _ Driver = (*Exec)(nil)

type Exec struct {
}

func NewExec() *Exec {
	return &Exec{}
}

func (e *Exec) RunTask(ctx context.Context, s *Spec, updater DriverUpdater) (TaskHandle, error) {
	cmd := exec.CommandContext(ctx, s.args[0], s.args[1:]...)

	var env []string
	for k, v := range s.env {
		env = append(env, k+"="+v)
	}
	cmd.Env = env

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	task := &execTask{
		cmd:     cmd,
		stdout:  stdout,
		stderr:  stderr,
		spec:    s,
		updater: updater,
	}
	go task.run()

	return task, nil
}

var _ TaskHandle = (*execTask)(nil)

type execTask struct {
	cmd     *exec.Cmd
	stdout  io.ReadCloser
	stderr  io.ReadCloser
	spec    *Spec
	updater DriverUpdater
	err     error
	errLock sync.RWMutex
}

func (t *execTask) run() {
	t.updater.TaskUpdated(TaskStatusDeployed)

	err := t.cmd.Wait()
	if err != nil {
		t.errLock.Lock()
		t.err = err
		t.errLock.Unlock()
	}
	t.updater.TaskUpdated(TaskStatusStopped)
}

func (t *execTask) Stop() error {
	if t.cmd.Process != nil {
		return t.cmd.Process.Kill()
	}
	return nil
}

func (t *execTask) Logs() (io.ReadCloser, error) {
	return io.NopCloser(io.MultiReader(t.stdout, t.stderr)), nil
}

func (t *execTask) Err() error {
	t.errLock.RLock()
	defer t.errLock.RUnlock()
	return t.err
}
