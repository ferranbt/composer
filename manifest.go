package composer

import (
	"fmt"
	"strings"
	"text/template"
)

type Manifest struct {
	specs []*Spec
}

func NewManifest() *Manifest {
	return &Manifest{
		specs: []*Spec{},
	}
}

func (m *Manifest) AddSpec(spec *Spec) *Manifest {
	m.specs = append(m.specs, spec)
	return m
}

func (m *Manifest) Specs() []*Spec {
	return m.specs
}

func (m *Manifest) Validate() error {
	// First pass: parse templates to auto-populate ports and connectsTo
	for _, spec := range m.specs {
		if err := m.parseTemplatesInSpec(spec); err != nil {
			return err
		}
	}

	specNames := make(map[string]bool)
	specPorts := make(map[string]map[string]bool)

	for _, spec := range m.specs {
		specNames[spec.Name()] = true
		specPorts[spec.Name()] = make(map[string]bool)
		for _, port := range spec.ports {
			specPorts[spec.Name()][port.Name] = true
		}
	}

	for _, spec := range m.specs {
		for _, dep := range spec.dependsOn {
			if !specNames[dep.Name] {
				return fmt.Errorf("spec '%s' depends on '%s' which does not exist", spec.Name(), dep.Name)
			}
			// Validate healthy condition requires health check
			if dep.Condition == "healthy" {
				depSpec := m.getSpec(dep.Name)
				if depSpec.healthCheck == nil {
					return fmt.Errorf("spec '%s' depends on '%s' with condition 'healthy' but '%s' does not have a health check", spec.Name(), dep.Name, dep.Name)
				}
			}
		}

		for _, conn := range spec.connectsTo {
			if !specNames[conn.Service] {
				return fmt.Errorf("spec '%s' connects to '%s' which does not exist", spec.Name(), conn.Service)
			}
			if !specPorts[conn.Service][conn.Port] {
				return fmt.Errorf("spec '%s' connects to port '%s' on '%s' which does not exist", spec.Name(), conn.Port, conn.Service)
			}
		}
	}

	return nil
}

func (m *Manifest) parseTemplatesInSpec(spec *Spec) error {
	funcs := &templateFuncs{
		Service: func(serviceName, portName string) string {
			// Auto-populate connectsTo
			exists := false
			for _, conn := range spec.connectsTo {
				if conn.Service == serviceName && conn.Port == portName {
					exists = true
					break
				}
			}
			if !exists {
				spec.connectsTo = append(spec.connectsTo, &ConnectsTo{
					Service: serviceName,
					Port:    portName,
				})
			}
			return ""
		},
		Port: func(portName string, defaultPort int) int {
			// Check if port exists and validate it matches
			for _, p := range spec.ports {
				if p.Name == portName {
					if p.Port != defaultPort {
						// Store error to be returned later
						spec.validationError = fmt.Errorf("port '%s' defined with conflicting values: %d and %d", portName, p.Port, defaultPort)
					}
					return defaultPort
				}
			}
			// Auto-populate if not exists
			spec.ports = append(spec.ports, &Port{
				Name: portName,
				Port: defaultPort,
			})
			return defaultPort
		},
	}

	// Parse args
	for _, arg := range spec.args {
		if _, err := applyArgsTemplate(arg, funcs); err != nil {
			return err
		}
	}

	// Parse env
	for _, v := range spec.env {
		if _, err := applyArgsTemplate(v, funcs); err != nil {
			return err
		}
	}

	if spec.validationError != nil {
		return spec.validationError
	}

	return nil
}

type templateFuncs struct {
	Service func(serviceName, portName string) string
	Port    func(portName string, defaultPort int) int
}

func applyArgsTemplate(arg string, tfuncs *templateFuncs) (string, error) {
	funcs := template.FuncMap{
		"Service": func(serviceName, portName string) string {
			return tfuncs.Service(serviceName, portName)
		},
		"Port": func(portName string, defaultPort int) int {
			return tfuncs.Port(portName, defaultPort)
		},
	}

	tpl, err := template.New("").Funcs(funcs).Parse(arg)
	if err != nil {
		return "", err
	}

	var out strings.Builder
	if err := tpl.Execute(&out, nil); err != nil {
		return "", err
	}

	return out.String(), nil
}

func (m *Manifest) getSpec(name string) *Spec {
	for _, spec := range m.specs {
		if spec.Name() == name {
			return spec
		}
	}
	return nil
}
