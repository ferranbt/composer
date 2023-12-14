package runner

import (
	"testing"
	"time"

	"github.com/ferranbt/composer/proto"
	"github.com/stretchr/testify/require"
)

func TestRunner_Up(t *testing.T) {
	project := &proto.Project{
		Name: "test",
		Services: map[string]*proto.Service{
			"a": {
				Image: "redis:alpine",
			},
			"a1": {
				Image:   "redis:alpine",
				Depends: []string{"a"},
			},
			"a2": {
				Image:   "redis:alpine",
				Depends: []string{"a1"},
			},
		},
	}

	r := newTestRunner(t, project)
	go r.run()

	WaitUntil(t, func() bool {
		if !r.complete {
			return false
		}

		createdEvents := 0
		for _, state := range r.containers {
			for _, event := range state.Events {
				if event.Type == "running" {
					createdEvents++
				}
			}
		}
		if createdEvents != 3 {
			return false
		}

		return true
	})

	project.Services["a3"] = &proto.Service{
		Image:   "redis:alpine",
		Depends: []string{"a2"},
	}
	project.Services["a1"].Depends = []string{"a", "a3"}

	r.UpdateProject(project)

	WaitUntil(t, func() bool {
		if !r.complete {
			return false
		}

		var (
			completedEvents uint64
			createdEvents   uint64
		)
		for _, state := range r.containers {
			for _, event := range state.Events {
				if event.Type == "running" {
					createdEvents++
				}
				if event.Type == "completed" {
					completedEvents++
				}
			}
		}

		if createdEvents != 6 {
			return false
		}
		if completedEvents != 2 {
			return false
		}

		return true
	})
}

func newTestRunner(t *testing.T, p *proto.Project) *ProjectRunner {
	store := newInmemStore(t)
	require.NoError(t, store.PutProject(p))

	updater := &localUpdater{}
	runner := newProjectRunner(p, newDockerProvider(updater), store)
	updater.p = runner

	return runner
}

type localUpdater struct {
	p *ProjectRunner
}

func (l *localUpdater) Update(c *ComputeUpdate) error {
	return l.p.Update(c)
}

func WaitUntil(t *testing.T, cond func() bool) {
	timeout := time.After(3 * time.Second)

	for {
		select {
		case <-timeout:
			t.Fatal("timeout")
		default:
		}

		if cond() {
			return
		}
		time.Sleep(1 * time.Second)
	}
}