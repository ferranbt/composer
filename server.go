package composer

import (
	"context"

	"github.com/ferranbt/composer/proto"
)

type Config struct {
	DbPath string
}

func DefaultConfig() *Config {
	return &Config{
		DbPath: "runner.db",
	}
}

type Option func(*Config)

type Server struct {
	runners map[string]*ProjectRunner
	store   *BoltdbStore
	docker  *dockerProvider
}

func NewServer(config ...Option) (*Server, error) {
	cfg := DefaultConfig()
	for _, opt := range config {
		opt(cfg)
	}

	r := &Server{
		runners: map[string]*ProjectRunner{},
	}

	store, err := NewBoltdbStore(cfg.DbPath)
	if err != nil {
		return nil, err
	}
	r.store = store

	docker := newDockerProvider(r)
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
		runner := newProjectRunner(p, r.docker, r.store)
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
		runner = newProjectRunner(req, r.docker, r.store)

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

func (r *Server) Update(c *ComputeUpdate) error {
	return r.runners[c.Project].Update(c)
}
