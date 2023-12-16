package proto

import (
	"crypto/sha512"
	"encoding/hex"

	"google.golang.org/protobuf/proto"
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

func NewEvent(typ string) *ServiceState_Event {
	return &ServiceState_Event{
		Type: typ,
	}
}

func (t *ServiceState) AddEvent(event *ServiceState_Event) {
	t.Events = append(t.Events, event)
}

func (s *Service) Copy() *Service {
	return proto.Clone(s).(*Service)
}
