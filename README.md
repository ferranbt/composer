# Composer

A lightweight SDK for orchestrating multi-service systems in Go tests and development environments.

Modern systems are complex. Running end-to-end tests or local dev environments often means juggling databases, APIs, message queues, and services that depend on each other.

Composer gives you a semantic Go API to define, orchestrate, and manage these systems. Write your service definitions in code, get type safety, and integrate directly with your test suites.

## Quick Example

```go
m := composer.NewManifest()

// Database with health check
m.AddSpec(composer.NewSpec("postgres").
    Image("postgres", "15").
    Port("pg", 5432).
    Env("POSTGRES_PASSWORD", "test").
    HealthCheckHTTP("http://localhost:5432", 200, 1*time.Second, 500*time.Millisecond, 2*time.Second))

// API that depends on healthy database
m.AddSpec(composer.NewSpec("api").
    Image("myapi", "latest").
    Args("--db={{Service \"postgres\" \"pg\"}}").
    DependsOn("postgres", "healthy"))

// Run everything
runtime, _ := composer.NewRuntime(m, nil, &composer.RuntimeConfig{
    LogsLocation: "./logs",
})
runtime.Run(context.Background())
```

## Features

- **Dependency management**: Services start in the right order based on dependencies
- **Health checks**: Wait for services to be ready before starting dependents
- **Templates**: Reference services and ports using `{{Service "name" "port"}}` syntax
- **Multiple drivers**: Run Docker containers or local processes
- **Log streaming**: Automatically capture logs to files
