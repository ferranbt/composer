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

func NewEvent(typ string) *ServiceState_Event {
	return &ServiceState_Event{
		Type:    typ,
		Details: map[string]string{},
	}
}

func (t *ServiceState) AddEvent(event *ServiceState_Event) {
	t.Events = append(t.Events, event)
}

func (s *Service) Copy() *Service {
	return proto.Clone(s).(*Service)
}

func (r *ExitResult) Successful() bool {
	return r.ExitCode == 0
}

func (t *ServiceState_Event) SetExitCode(c int64) *ServiceState_Event {
	t.Details["exit_code"] = fmt.Sprintf("%d", c)
	return t
}

func (t *ServiceState_Event) SetSignal(s int64) *ServiceState_Event {
	t.Details["signal"] = fmt.Sprintf("%d", s)
	return t
}

func (t *ServiceState_Event) FailsTask() bool {
	_, ok := t.Details["fails_task"]
	return ok
}

func (t *ServiceState_Event) SetFailsTask() *ServiceState_Event {
	t.Details["fails_task"] = "true"
	return t
}

func (t *ServiceState_Event) SetTaskFailed(name string) *ServiceState_Event {
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
