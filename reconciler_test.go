package composer

import (
	"fmt"
	"testing"

	"github.com/ferranbt/composer/proto"
	"github.com/stretchr/testify/require"
)

type resultExpectation struct {
	create  int
	destroy int
	taint   int
	healthy bool
}

func assertResults(t *testing.T, r *reconcileResults, exp *resultExpectation) {
	require.Equal(t, exp.create, len(r.create))
	require.Equal(t, exp.destroy, len(r.remove))
	require.Equal(t, exp.taint, len(r.taint))
	require.Equal(t, exp.healthy, r.complete)
}

func TestReconciler_WithDependency(t *testing.T) {
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
		},
	}

	nodes := map[string]*proto.TaskState{}

	res := newReconciler(nodes, project)
	assertResults(t, res.compute(), &resultExpectation{
		create:  1,
		destroy: 0,
		taint:   0,
		healthy: false,
	})
}

func TestReconciler_BulkNoDependencies(t *testing.T) {
	project := &proto.Project{
		Name:     "test",
		Services: map[string]*proto.Service{},
	}

	for i := 0; i < 10; i++ {
		project.Services[fmt.Sprintf("a%d", i)] = &proto.Service{
			Image: "redis:alpine",
		}
	}

	nodes := map[string]*proto.TaskState{}

	res := newReconciler(nodes, project)
	assertResults(t, res.compute(), &resultExpectation{
		create:  10,
		destroy: 0,
		taint:   0,
		healthy: false,
	})
}
