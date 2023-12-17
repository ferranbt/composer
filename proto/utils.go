package proto

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"

	"google.golang.org/protobuf/proto"
)

func (s *Service) Hash() (string, error) {
	raw, err := proto.Marshal(s)
	if err != nil {
		return "", err
	}

	hash := sha512.Sum512(raw)
	return hex.EncodeToString(hash[:]), nil
}

func NewEvent(project, service, typ string) *Event {
	return &Event{
		Project: project,
		Service: service,
		Type:    typ,
		Details: map[string]string{},
	}
}

func (s *Service) Copy() *Service {
	return proto.Clone(s).(*Service)
}

func (r *ExitResult) Successful() bool {
	return r.ExitCode == 0
}

func (t *Event) SetExitCode(c int64) *Event {
	t.Details["exit_code"] = fmt.Sprintf("%d", c)
	return t
}

func (t *Event) SetSignal(s int64) *Event {
	t.Details["signal"] = fmt.Sprintf("%d", s)
	return t
}

func (t *Event) FailsTask() bool {
	_, ok := t.Details["fails_task"]
	return ok
}

func (t *Event) SetFailsTask() *Event {
	t.Details["fails_task"] = "true"
	return t
}

func (t *Event) SetTaskFailed(name string) *Event {
	t.Details["failed_task"] = name
	return t
}

const (
	TaskStarted       = "Started"
	TaskTerminated    = "Terminated"
	TaskRestarting    = "Restarting"
	TaskNotRestarting = "Not-restarting"
	TaskSiblingFailed = "Sibling-failed"
	TaskKilling       = "Task-Killing"
)
