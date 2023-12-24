package composer

import (
	"context"
	"strings"

	"github.com/ferranbt/composer/docker"
	"github.com/ferranbt/composer/hooks"
	"github.com/ferranbt/composer/proto"
)

type AllocUpdater interface {
}

type ProjectRunner struct {
	project *proto.Project

	// services is the list of running services
	services map[string]*serviceRunner

	notifier Notifier

	docker   *docker.Provider
	store    *BoltdbStore
	complete bool
	closeCh  chan struct{}
	updateCh chan struct{}
	hooks    []hooks.ServiceHookFactory
}

func newProjectRunner(project *proto.Project, docker *docker.Provider, store *BoltdbStore, notifier Notifier, hooks []hooks.ServiceHookFactory) *ProjectRunner {
	p := &ProjectRunner{
		project:  project,
		docker:   docker,
		store:    store,
		services: map[string]*serviceRunner{},
		closeCh:  make(chan struct{}),
		updateCh: make(chan struct{}, 10),
		notifier: notifier,
		hooks:    hooks,
	}
	return p
}

func (p *ProjectRunner) Restore() error {
	tasksIds, err := p.store.GetTasks(p.project.Name)
	if err != nil {
		return err
	}

	for _, id := range tasksIds {
		runner := newServiceRunner(p.project, id, p.project.Services[id], p.docker, p.store, p.taskStateUpdated, p.notifier, p.hooks)
		p.services[id] = runner

		go runner.Run()
	}

	return nil
}

func (p *ProjectRunner) taskStateUpdated() {
	select {
	case p.updateCh <- struct{}{}:
	default:
	}
}

func (r *ProjectRunner) run() {
	for {
		r.runIteration()

		select {
		case <-r.updateCh:
		case <-r.closeCh:
			return
		}
	}
}

type Status struct {
	Complete bool
	State    map[string]*proto.ServiceState
}

func (r *ProjectRunner) Status() *Status {
	containersState := map[string]*proto.ServiceState{}
	for name, state := range r.services {
		containersState[name] = state.TaskState()
	}

	s := &Status{
		Complete: r.complete,
		State:    containersState,
	}
	return s
}

func (r *ProjectRunner) runIteration() {
	containersState := r.Status().State
	res := newReconciler(containersState, r.project).compute()

	for name := range res.create {
		service := r.project.Services[name]

		// rewrite the network host config if it refers to another running container
		if strings.HasPrefix(service.NetworkMode, "service:") {
			dep := strings.TrimPrefix(service.NetworkMode, "service:")
			depState, ok := containersState[dep]

			if ok && depState.State == proto.ServiceState_Running {
				service.NetworkMode = "container:" + depState.Handle.ContainerId
			}
		}

		runner := newServiceRunner(r.project, name, service, r.docker, r.store, r.taskStateUpdated, r.notifier, r.hooks)
		r.services[name] = runner

		go runner.Run()
	}

	for name := range res.remove {
		r.services[name].Kill(context.Background())
	}

	for name := range res.taint {
		r.services[name].Taint()
	}

	r.complete = res.complete
}

func (r *ProjectRunner) UpdateProject(p *proto.Project) {
	r.project = p
	r.complete = false
	r.taskStateUpdated()
}
