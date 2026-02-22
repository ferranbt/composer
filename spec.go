package composer

import (
	"time"

	"github.com/ferranbt/composer/internal"
)

type DependsOn struct {
	Name      string
	Condition string
}

type Port struct {
	Name string
	Port int
}

type ConnectsTo struct {
	Service string
	Port    string
}

type DriverType string

const (
	DriverTypeDocker DriverType = "docker"
	DriverTypeExec   DriverType = "exec"
)

type Spec struct {
	name            string
	image           string
	tag             string
	args            []string
	env             map[string]string
	dependsOn       []*DependsOn
	ports           []*Port
	connectsTo      []*ConnectsTo
	driverType      DriverType
	healthCheck     *internal.HealthCheck
	validationError error
}

func NewSpec(name string) *Spec {
	return &Spec{
		name:       name,
		env:        make(map[string]string),
		dependsOn:  []*DependsOn{},
		ports:      []*Port{},
		connectsTo: []*ConnectsTo{},
		driverType: DriverTypeDocker,
	}
}

func (s *Spec) Args(args ...string) *Spec {
	s.args = args
	return s
}

func (s *Spec) Env(key, value string) *Spec {
	s.env[key] = value
	return s
}

func (s *Spec) DependsOn(name, condition string) *Spec {
	s.dependsOn = append(s.dependsOn, &DependsOn{
		Name:      name,
		Condition: condition,
	})
	return s
}

func (s *Spec) Port(name string, port int) *Spec {
	s.ports = append(s.ports, &Port{
		Name: name,
		Port: port,
	})
	return s
}

func (s *Spec) ConnectsTo(service, port string) *Spec {
	s.connectsTo = append(s.connectsTo, &ConnectsTo{
		Service: service,
		Port:    port,
	})
	return s
}

func (s *Spec) Name() string {
	return s.name
}

func (s *Spec) Image(image, tag string) *Spec {
	s.image = image
	s.tag = tag
	return s
}

func (s *Spec) UseExec() *Spec {
	s.driverType = DriverTypeExec
	return s
}

func (s *Spec) DriverType() DriverType {
	return s.driverType
}

func (s *Spec) GetPort(name string) *Port {
	for _, p := range s.ports {
		if p.Name == name {
			return p
		}
	}
	return nil
}

func (s *Spec) HealthCheckHTTP(url string, expectedStatus int, initialDelay, interval, timeout time.Duration) *Spec {
	s.healthCheck = &internal.HealthCheck{
		HTTP: &internal.HTTPCheck{
			URL:            url,
			ExpectedStatus: expectedStatus,
		},
		InitialDelay: initialDelay,
		Interval:     interval,
		Timeout:      timeout,
	}
	return s
}
