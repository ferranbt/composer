package runner

import (
	"github.com/ferranbt/composer/proto"
)

type ProjectRunner struct {
	project *proto.Project

	// containers is a map of the state for the containers in this project
	containers map[string]*proto.TaskState

	docker   *dockerProvider
	store    *BoltdbStore
	complete bool
	closeCh  chan struct{}
	updateCh chan struct{}
}

func newProjectRunner(project *proto.Project, docker *dockerProvider, store *BoltdbStore) *ProjectRunner {
	p := &ProjectRunner{
		project:    project,
		docker:     docker,
		store:      store,
		containers: map[string]*proto.TaskState{},
		closeCh:    make(chan struct{}),
		updateCh:   make(chan struct{}, 10),
	}
	return p
}

func (p *ProjectRunner) Restore() error {
	tasksIds, err := p.store.GetTasks(p.project.Name)
	if err != nil {
		return err
	}

	for _, id := range tasksIds {
		state, err := p.store.GetTaskState(p.project.Name, id)
		if err != nil {
			return err
		}

		p.docker.Reattach(p.project.Name, state.Handle.ContainerId, state.Handle)
		p.containers[id] = state
	}

	return nil
}

func (p *ProjectRunner) notify() {
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

func (r *ProjectRunner) runIteration() {
	res := newReconciler(r.containers, r.project).compute()

	for name := range res.create {
		service := r.project.Services[name]
		res := &ComputeResource{
			Name:    name,
			Project: r.project.Name,
			Service: service,
		}
		handle, err := r.docker.Create(res)
		if err != nil {
			panic(err)
		}

		hash, err := service.Hash()
		if err != nil {
			panic(err)
		}
		state := &proto.TaskState{
			State:  proto.TaskState_Running,
			Handle: handle,
			Hash:   hash,
		}

		previous, ok := r.containers[name]
		if ok {
			state.Events = previous.Events
		}

		state.AddEvent(proto.NewEvent("running"))
		if err := r.store.PutTaskState(r.project.Name, name, state); err != nil {
			panic(err)
		}
		r.containers[name] = state
	}

	for name := range res.remove {
		state := r.containers[name]
		if err := r.docker.Kill(state.Handle.ContainerId); err != nil {
			panic(err)
		}
	}

	for name := range res.taint {
		state := r.containers[name]
		state.State = proto.TaskState_Tainted
		state.AddEvent(proto.NewEvent("taint"))

		if err := r.store.PutTaskState(r.project.Name, name, state); err != nil {
			panic(err)
		}
		r.containers[name] = state
	}

	if len(res.create) != 0 {
		r.notify()
	}

	r.complete = res.complete
}

func (r *ProjectRunner) UpdateProject(p *proto.Project) {
	r.project = p
	r.complete = false
	r.notify()
}

func (r *ProjectRunner) Update(c *ComputeUpdate) error {
	node, ok := r.containers[c.Name]
	if !ok {
		panic("?")
	}

	if c.Created != nil {
		node.State = proto.TaskState_Running
	} else if c.Completed != nil {
		node.State = proto.TaskState_Dead
		node.AddEvent(proto.NewEvent("completed"))
	}

	r.containers[c.Name] = node
	if err := r.store.PutTaskState(r.project.Name, c.Name, node); err != nil {
		panic(err)
	}

	select {
	case r.updateCh <- struct{}{}:
	default:
	}

	return nil
}
