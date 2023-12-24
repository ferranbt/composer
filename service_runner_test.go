package composer

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/ferranbt/composer/docker"
	"github.com/ferranbt/composer/hooks"
	"github.com/ferranbt/composer/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func destroyRunner(tr *serviceRunner) {
	tr.Kill(context.Background())
}

func testWaitForTaskToDie(t *testing.T, tr *serviceRunner) {
	waitForResult(func() (bool, error) {
		ts := tr.TaskState()
		return ts.State == proto.ServiceState_Dead, fmt.Errorf("expected task to be dead, got %v", ts.State)
	}, func(err error) {
		require.NoError(t, err)
	})
}

func testWaitForTaskToStart(t *testing.T, tr *serviceRunner) {
	waitForResult(func() (bool, error) {
		ts := tr.TaskState()
		return ts.State == proto.ServiceState_Running, fmt.Errorf("expected task to be running, got %v", ts.State)
	}, func(err error) {
		require.NoError(t, err)
	})
}

type eventSink struct {
	events []*proto.Event
}

func (e *eventSink) Notify(event *proto.Event) {
	if e.events == nil {
		e.events = []*proto.Event{}
	}
	e.events = append(e.events, event)
}

func setupServiceRunner(t *testing.T, task *proto.Service, state *BoltdbStore) *serviceRunner {
	name := "test-task"
	driver := docker.NewProvider()

	project := &proto.Project{
		Name: "test-project",
		Services: map[string]*proto.Service{
			name: task,
		},
	}

	if state == nil {
		var err error

		state, err = NewBoltdbStore(filepath.Join(t.TempDir(), "my.db"))
		assert.NoError(t, err)
		assert.NoError(t, state.PutProject(project))
	}

	sink := &eventSink{}
	return newServiceRunner(project, name, task, driver, state, func() {}, sink, nil)
}

func TestTaskRunner_Stop_ExitCode(t *testing.T) {
	tt := &proto.Service{
		Image: "busybox:1.29.3",
		Args:  []string{"sleep", "3"},
	}
	r := setupServiceRunner(t, tt, nil)
	go r.Run()

	testWaitForTaskToStart(t, r)

	err := r.Kill(context.Background())
	require.NoError(t, err)

	events := r.notifier.(*eventSink).events
	terminatedEvent := events[1]
	require.Equal(t, terminatedEvent.Type, proto.TaskTerminated)
	require.Equal(t, terminatedEvent.Details["exit_code"], "137")
}

func TestTaskRunner_Restore_AlreadyRunning(t *testing.T) {
	// Restoring a running task should not re run the task
	tt := &proto.Service{
		Image: "busybox:1.29.3",
		Args:  []string{"sleep", "3"},
	}

	oldRunner := setupServiceRunner(t, tt, nil)
	go oldRunner.Run()

	testWaitForTaskToStart(t, oldRunner)

	// stop the task runner
	oldRunner.Shutdown()

	// start another task runner with the same state
	newRunner := setupServiceRunner(t, tt, oldRunner.store)

	// restore the task
	require.NoError(t, newRunner.Restore())
	defer destroyRunner(newRunner)

	go newRunner.Run()

	// wait for the process to finish
	testWaitForTaskToDie(t, newRunner)

	// assert the process only started once
	events := newRunner.notifier.(*eventSink).events
	require.Len(t, events, 1)
	require.Equal(t, events[0].Type, proto.TaskTerminated)
}

func TestTaskRunner_Hooks(t *testing.T) {
	name := "test-task"
	driver := docker.NewProvider()

	task := &proto.Service{
		Image: "busybox:1.29.3",
		Args:  []string{"sleep", "5"},
	}

	project := &proto.Project{
		Name: "test-project",
		Services: map[string]*proto.Service{
			name: task,
		},
	}

	state := newInmemStore(t)
	require.NoError(t, state.PutProject(project))

	if state == nil {
		var err error

		state, err = NewBoltdbStore(filepath.Join(t.TempDir(), "my.db"))
		assert.NoError(t, err)
		assert.NoError(t, state.PutProject(project))
	}

	hook := &testHook{}
	hooks := []hooks.ServiceHook{
		hook,
	}

	sink := &eventSink{}
	runner := newServiceRunner(project, name, task, driver, state, func() {}, sink, hooks)

	go runner.Run()

	// wait for the process to finish
	testWaitForTaskToDie(t, runner)

	// assert the state of the hooks
	require.NotEmpty(t, hook.service)
	require.NotEmpty(t, hook.ip)
}

type testHook struct {
	service *proto.Service
	ip      string
}

func (t *testHook) Name() string {
	return "test"
}

func (t *testHook) Prestart(ctx context.Context, req *hooks.ServicePrestartHookRequest) error {
	t.service = req.Service
	return nil
}

func (t *testHook) Poststart(ctx context.Context, req *hooks.ServicePoststartHookRequest) error {
	t.ip = req.Ip
	return nil
}

func (t *testHook) Stop(context.Context, *hooks.ServiceStopRequest) error {
	return nil
}

type testFn func() (bool, error)
type errorFn func(error)

func waitForResult(test testFn, error errorFn) {
	waitForResultRetries(500, test, error)
}

func waitForResultRetries(retries int64, test testFn, error errorFn) {
	for retries > 0 {
		time.Sleep(10 * time.Millisecond)
		retries--

		success, err := test()
		if success {
			return
		}

		if retries == 0 {
			error(err)
		}
	}
}
