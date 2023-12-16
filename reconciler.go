package composer

import (
	"fmt"

	"github.com/ferranbt/composer/dag"
	"github.com/ferranbt/composer/proto"
)

type reconciler struct {
	containers map[string]*proto.TaskState
	project    *proto.Project
	dag        *dag.Dag
}

type reconcileResults struct {
	create   map[string]struct{}
	complete bool
	taint    map[string]struct{}
	remove   map[string]struct{}
}

func (r *reconcileResults) String() string {
	return fmt.Sprintf("create: %d, destroy: %d, taint: %d, complete: %t", len(r.create), len(r.remove), len(r.taint), r.complete)
}

func (r *reconcileResults) noUpdates() bool {
	return len(r.create) == 0 && len(r.taint) == 0 && len(r.remove) == 0
}

func newReconciler(containers map[string]*proto.TaskState, project *proto.Project) *reconciler {
	return &reconciler{
		containers: containers,
		project:    project,
	}
}

type serviceWithName struct {
	Name string
	*proto.Service
}

func (r *reconciler) buildDag() {
	d := &dag.Dag{}

	// add all the nodes to the dag
	for name := range r.project.Services {
		d.AddVertex(name)
	}

	// add all the edges to the dag
	for name, srv := range r.project.Services {
		for _, dep := range srv.Depends {
			d.AddEdge(dag.Edge{
				Src: name,
				Dst: dep,
			})
		}
	}

	r.dag = d
}

func (r *reconciler) compute() *reconcileResults {
	r.buildDag()

	res := &reconcileResults{
		create: map[string]struct{}{},
		taint:  map[string]struct{}{},
		remove: map[string]struct{}{},
	}

	// get the services from the project which have all their
	// dependencies healthy and are either:
	// 1. not running => Create it
	// 2. tainted => Remove it
	// 2. running but with a different spec => remove it and taint the nodes that depend on it
	for name, service := range r.project.Services {
		node, exists := r.containers[name]

		// check if the dependences are healthy
		allDependenciesRunning := true
		for _, dep := range service.Depends {
			dependency, ok := r.containers[dep]
			if !ok || dependency.State != proto.TaskState_Running {
				allDependenciesRunning = false
				break
			}
		}

		// All the depencies are resolved, we can create it!
		if !allDependenciesRunning {
			continue
		}

		// if the node does not exist, we need to create it
		if !exists || node.State == proto.TaskState_Dead {
			res.create[name] = struct{}{}
		} else {
			// if the node is tainted, remove it
			if node.State == proto.TaskState_Tainted {
				res.remove[name] = struct{}{}
				break
			}

			// if the node exists, but the spec is different, we need to remove it
			// and taint the nodes that depend on it.
			serviceHash, err := service.Hash()
			if err != nil {
				panic(err)
			}
			if node.Hash != serviceHash {
				res.remove[name] = struct{}{}

				// taint the nodes that depend on this one
				for _, dep := range r.dag.GetInbound(name) {
					res.taint[dep.(string)] = struct{}{}
				}
			}
		}
	}

	// are all the nodes running?
	allRunning := true
	for _, n := range r.containers {
		if n.State != proto.TaskState_Running {
			allRunning = false
			break
		}
	}

	// if all the nodes are running and no resources need creation
	// the deployment is completed
	if allRunning && res.noUpdates() {
		res.complete = true
	}

	return res
}
