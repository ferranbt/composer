package composer

import (
	"context"
	"io"
)

type TaskStatus string

const (
	TaskStatusPending  TaskStatus = "pending"
	TaskStatusStarting TaskStatus = "starting"
	TaskStatusDeployed TaskStatus = "deployed"
	TaskStatusStopped  TaskStatus = "stopped"
)

type DriverUpdater interface {
	TaskUpdated(event TaskStatus)
}

type TaskUpdater interface {
	NotifyTaskUpdated()
}

type Driver interface {
	RunTask(ctx context.Context, s *Spec, updater DriverUpdater) (TaskHandle, error)
}

type TaskHandle interface {
	Stop() error
	Logs() (io.ReadCloser, error)
	Err() error
}
