package composer

import (
	"strings"
	"testing"
	"time"

	"github.com/ferranbt/composer/docker"
	"github.com/ferranbt/composer/proto"
	"github.com/stretchr/testify/require"
)

func TestRunner_MultipleServices_Simple(t *testing.T) {
	project := &proto.Project{
		Name: "test",
		Services: map[string]*proto.Service{
			"a": {
				Image: "busybox:1.29.3",
				Args:  []string{"sleep", "30"},
			},
			"a1": {
				Image: "busybox:1.29.3",
				Args:  []string{"sleep", "30"},
			},
		},
	}

	r := newTestRunner(t, project)
	go r.run()

	WaitUntil(t, func() bool {
		status := r.Status()
		if !status.Complete {
			return false
		}

		events := r.notifier.(*eventSink).events

		createdEvents := 0
		for _, event := range events {
			if event.Type == proto.TaskStarted {
				createdEvents++
			}
		}
		if createdEvents != 2 {
			return false
		}

		// all of the containers have assigned IPs
		for _, state := range status.State {
			if state.Handle.Ip == "" {
				return false
			}
		}

		return true
	})
}

func TestRunner_MultipleServices_Depends(t *testing.T) {
	project := &proto.Project{
		Name: "test",
		Services: map[string]*proto.Service{
			"a": {
				Image: "busybox:1.29.3",
				Args:  []string{"sleep", "30"},
			},
			"a1": {
				Image:   "busybox:1.29.3",
				Args:    []string{"sleep", "30"},
				Depends: []string{"a"},
			},
		},
	}

	r := newTestRunner(t, project)
	go r.run()

	WaitUntil(t, func() bool {
		status := r.Status()
		if !status.Complete {
			return false
		}

		events := r.notifier.(*eventSink).events

		createdEvents := 0
		for _, event := range events {
			if event.Type == proto.TaskStarted {
				createdEvents++
			}
		}
		if createdEvents != 2 {
			return false
		}

		// all of the containers have assigned IPs
		for _, state := range status.State {
			if state.Handle.Ip == "" {
				return false
			}
		}

		return true
	})
}

func TestRunner_UpdateDependencies(t *testing.T) {
	// Initial: a -> a1 -> a2
	// Update: a -------------> a1
	//         |--> a2 -> a3 -> a1
	project := &proto.Project{
		Name: "test",
		Services: map[string]*proto.Service{
			"a": {
				Image: "busybox:1.29.3",
				Args:  []string{"sleep", "30"},
			},
			"a1": {
				Image:   "busybox:1.29.3",
				Args:    []string{"sleep", "30"},
				Depends: []string{"a"},
			},
			"a2": {
				Image:   "busybox:1.29.3",
				Args:    []string{"sleep", "30"},
				Depends: []string{"a1"},
			},
		},
	}

	r := newTestRunner(t, project)
	go r.run()

	WaitUntil(t, func() bool {
		status := r.Status()
		if !status.Complete {
			return false
		}

		createdEvents := 0
		events := r.notifier.(*eventSink).events
		for _, event := range events {
			if event.Type == proto.TaskStarted {
				createdEvents++
			}
		}
		if createdEvents != 3 {
			return false
		}

		// all of the containers have assigned IPs
		for _, state := range status.State {
			if state.Handle.Ip == "" {
				return false
			}
		}

		return true
	})

	project.Services["a3"] = &proto.Service{
		Image:   "busybox:1.29.3",
		Args:    []string{"sleep", "30"},
		Depends: []string{"a2"},
	}
	project.Services["a2"].Depends = []string{"a"}
	project.Services["a1"].Depends = []string{"a", "a3"}

	r.UpdateProject(project)

	WaitUntil(t, func() bool {
		status := r.Status()
		if !status.Complete {
			return false
		}

		var (
			completedEvents uint64
			createdEvents   uint64
		)
		events := r.notifier.(*eventSink).events
		for _, event := range events {
			if event.Type == proto.TaskStarted {
				createdEvents++
			}
			if event.Type == proto.TaskTerminated {
				completedEvents++
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

func TestRunner_SharedIP(t *testing.T) {
	project := &proto.Project{
		Name: "test",
		Services: map[string]*proto.Service{
			"a": {
				Image: "busybox:1.29.3",
				Args:  []string{"sleep", "30"},
			},
			"a1": {
				Image:       "busybox:1.29.3",
				Args:        []string{"sleep", "30"},
				Depends:     []string{"a"},
				NetworkMode: "service:a",
			},
		},
	}

	r := newTestRunner(t, project)
	go r.run()

	var status *Status
	WaitUntil(t, func() bool {
		status = r.Status()

		createdEvents := 0
		events := r.notifier.(*eventSink).events
		for _, event := range events {
			if event.Type == proto.TaskStarted {
				createdEvents++
			}
		}

		return createdEvents == 2
	})

	ipCommand := strings.Split("/bin/ip route", " ")

	aState := status.State["a"]
	a1State := status.State["a1"]

	resA, err := r.docker.Exec(aState.Handle.ContainerId, ipCommand)
	require.NoError(t, err)

	resA1, err := r.docker.Exec(a1State.Handle.ContainerId, ipCommand)
	require.NoError(t, err)

	require.NotEmpty(t, resA.Stdout)
	require.Equal(t, resA.Stdout, resA1.Stdout)

	// a1 does not have an ip assigned
	require.NotEmpty(t, aState.Handle.Ip)
	require.Empty(t, a1State.Handle.Ip)
}

func newTestRunner(t *testing.T, p *proto.Project) *ProjectRunner {
	store := newInmemStore(t)
	require.NoError(t, store.PutProject(p))

	sink := &eventSink{}
	runner := newProjectRunner(p, docker.NewProvider(), store, sink, nil)
	return runner
}

func WaitUntil(t *testing.T, cond func() bool) {
	timeout := time.After(20 * time.Second)

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
