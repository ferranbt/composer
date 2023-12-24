package hooks

import (
	"context"

	"github.com/ferranbt/composer/proto"
)

type ServiceHook interface {
	Name() string
}

type ServiceHookFactory func(project *proto.Project, service *proto.Service) ServiceHook

type ServicePrestartHookRequest struct {
	Service *proto.Service
}

type ServicePrestartHook interface {
	ServiceHook

	Prestart(context.Context, *ServicePrestartHookRequest) error
}

type ServicePoststartHookRequest struct {
	Ip string
}

type ServicePoststartHook interface {
	ServiceHook

	Poststart(context.Context, *ServicePoststartHookRequest) error
}

type ServiceStopRequest struct {
}

type ServiceStopHook interface {
	ServiceHook

	Stop(context.Context, *ServiceStopRequest) error
}
