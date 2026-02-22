package composer

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestManifest_ValidateDepends(t *testing.T) {
	t.Run("valid dependencies", func(t *testing.T) {
		m := NewManifest()
		m.AddSpec(NewSpec("service1"))
		m.AddSpec(NewSpec("service2").DependsOn("service1", "started"))

		err := m.Validate()
		require.NoError(t, err)
	})

	t.Run("invalid dependency - service not found", func(t *testing.T) {
		m := NewManifest()
		m.AddSpec(NewSpec("service1").DependsOn("nonexistent", "started"))

		err := m.Validate()
		require.Error(t, err)
		require.Equal(t, "spec 'service1' depends on 'nonexistent' which does not exist", err.Error())
	})

	t.Run("no dependencies", func(t *testing.T) {
		m := NewManifest()
		m.AddSpec(NewSpec("service1"))
		m.AddSpec(NewSpec("service2"))

		err := m.Validate()
		require.NoError(t, err)
	})

	t.Run("healthy dependency requires health check", func(t *testing.T) {
		m := NewManifest()
		m.AddSpec(NewSpec("db"))
		m.AddSpec(NewSpec("api").DependsOn("db", "healthy"))

		err := m.Validate()
		require.Error(t, err)
		require.Equal(t, "spec 'api' depends on 'db' with condition 'healthy' but 'db' does not have a health check", err.Error())
	})

	t.Run("healthy dependency with health check is valid", func(t *testing.T) {
		m := NewManifest()
		m.AddSpec(NewSpec("db").HealthCheckHTTP("http://localhost:5432", http.StatusOK, 1*time.Second, 1*time.Second, 1*time.Second))
		m.AddSpec(NewSpec("api").DependsOn("db", "healthy"))

		err := m.Validate()
		require.NoError(t, err)
	})
}

func TestManifest_ValidateConnect(t *testing.T) {
	t.Run("valid connects to", func(t *testing.T) {
		m := NewManifest()
		m.AddSpec(NewSpec("service1").Port("http", 8080))
		m.AddSpec(NewSpec("service2").ConnectsTo("service1", "http"))

		err := m.Validate()
		require.NoError(t, err)
	})

	t.Run("connects to nonexistent service", func(t *testing.T) {
		m := NewManifest()
		m.AddSpec(NewSpec("service1").ConnectsTo("nonexistent", "http"))

		err := m.Validate()
		require.Error(t, err)
		require.Equal(t, "spec 'service1' connects to 'nonexistent' which does not exist", err.Error())
	})

	t.Run("connects to nonexistent port", func(t *testing.T) {
		m := NewManifest()
		m.AddSpec(NewSpec("service1").Port("http", 8080))
		m.AddSpec(NewSpec("service2").ConnectsTo("service1", "grpc"))

		err := m.Validate()
		require.Error(t, err)
		require.Equal(t, "spec 'service2' connects to port 'grpc' on 'service1' which does not exist", err.Error())
	})
}

func TestManifest_ValidateTemplates(t *testing.T) {
	t.Run("Port template auto-populates ports", func(t *testing.T) {
		m := NewManifest()

		service := NewSpec("service").
			Args("--port", "{{Port \"http\" 8080}}")

		m.AddSpec(service)

		err := m.Validate()
		require.NoError(t, err)

		require.Len(t, service.ports, 1)
		require.Equal(t, "http", service.ports[0].Name)
		require.Equal(t, 8080, service.ports[0].Port)
	})

	t.Run("Service template auto-populates connectsTo", func(t *testing.T) {
		m := NewManifest()

		service1 := NewSpec("service1").Port("http", 8080)
		service2 := NewSpec("service2").
			Args("curl", "{{Service \"service1\" \"http\"}}")

		m.AddSpec(service1)
		m.AddSpec(service2)

		err := m.Validate()
		require.NoError(t, err)

		require.Len(t, service2.connectsTo, 1)
		require.Equal(t, "service1", service2.connectsTo[0].Service)
		require.Equal(t, "http", service2.connectsTo[0].Port)
	})

	t.Run("template in env auto-populates", func(t *testing.T) {
		m := NewManifest()

		service1 := NewSpec("service1").Port("http", 8080)
		service2 := NewSpec("service2").
			Env("API_URL", "{{Service \"service1\" \"http\"}}")

		m.AddSpec(service1)
		m.AddSpec(service2)

		err := m.Validate()
		require.NoError(t, err)

		require.Len(t, service2.connectsTo, 1)
		require.Equal(t, "service1", service2.connectsTo[0].Service)
		require.Equal(t, "http", service2.connectsTo[0].Port)
	})

	t.Run("port defined manually and in template with same value does not duplicate", func(t *testing.T) {
		m := NewManifest()

		service := NewSpec("service").
			Port("http", 8080).
			Args("--port", "{{Port \"http\" 8080}}")

		m.AddSpec(service)

		err := m.Validate()
		require.NoError(t, err)

		require.Len(t, service.ports, 1)
		require.Equal(t, "http", service.ports[0].Name)
		require.Equal(t, 8080, service.ports[0].Port)
	})

	t.Run("port defined manually and in template with different value fails", func(t *testing.T) {
		m := NewManifest()

		service := NewSpec("service").
			Port("http", 8080).
			Args("--port", "{{Port \"http\" 9090}}")

		m.AddSpec(service)

		err := m.Validate()
		require.Error(t, err)
		require.Contains(t, err.Error(), "port 'http' defined with conflicting values")
	})
}
