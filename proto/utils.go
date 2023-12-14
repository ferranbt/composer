package proto

import (
	"crypto/sha512"
	"encoding/hex"

	gproto "google.golang.org/protobuf/proto"
)

func (s *Service) Hash() (string, error) {
	raw, err := gproto.Marshal(s)
	if err != nil {
		return "", err
	}

	hash := sha512.Sum512(raw)
	return hex.EncodeToString(hash[:]), nil
}

func NewEvent(typ string) *TaskState_Event {
	return &TaskState_Event{
		Type: typ,
	}
}

func (t *TaskState) AddEvent(event *TaskState_Event) {
	t.Events = append(t.Events, event)
}
