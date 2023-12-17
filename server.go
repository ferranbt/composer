package composer

import (
	"context"

	"github.com/ferranbt/composer/docker"
	"github.com/ferranbt/composer/proto"
)

type Notifier interface {
	Notify(event *proto.Event)
}

type Config struct {
	DbPath   string
	Notifier Notifier
}

type mockNotifier struct {
}

func (m *mockNotifier) Notify(event *proto.Event) {
}

func WithNotifier(n Notifier) Option {
	return func(c *Config) {
		c.Notifier = n
	}
}

func DefaultConfig() *Config {
	return &Config{
		DbPath:   "runner.db",
		Notifier: &mockNotifier{},
	}
}

type Option func(*Config)

type Server struct {
	config  *Config
	runners map[string]*ProjectRunner
	store   *BoltdbStore
	docker  *docker.Provider
}

func NewServer(configOpts ...Option) (*Server, error) {
	cfg := DefaultConfig()
	for _, opt := range configOpts {
		opt(cfg)
	}

	r := &Server{
		config:  cfg,
		runners: map[string]*ProjectRunner{},
	}

	store, err := NewBoltdbStore(cfg.DbPath)
	if err != nil {
		return nil, err
	}
	r.store = store

	docker := docker.NewProvider()
	r.docker = docker

	if err := r.initialLoad(); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *Server) initialLoad() error {
	projects, err := r.store.GetProjects()
	if err != nil {
		return err
	}

	for _, p := range projects {
		runner := newProjectRunner(p, r.docker, r.store, r.config.Notifier)
		if err := runner.Restore(); err != nil {
			return err
		}

		go runner.run()
		r.runners[p.Name] = runner
	}

	return nil
}

func (r *Server) Up(ctx context.Context, req *proto.Project) (*proto.Project_Ref, error) {
	runner, ok := r.runners[req.Name]
	if !ok {
		runner = newProjectRunner(req, r.docker, r.store, r.config.Notifier)

		if err := r.store.PutProject(req); err != nil {
			return nil, err
		}

		go runner.run()
		r.runners[req.Name] = runner
	} else {
		runner.UpdateProject(req)
	}
	return nil, nil
}
