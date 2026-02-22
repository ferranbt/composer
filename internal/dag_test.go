package internal

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDAG(t *testing.T) {
	t.Run("no cycle in simple graph", func(t *testing.T) {
		dag := NewDAG()
		dag.AddNode("a", []string{})
		dag.AddNode("b", []string{"a"})
		dag.AddNode("c", []string{"b"})

		require.False(t, dag.HasCycle())
		require.NoError(t, dag.Validate())
	})

	t.Run("detects simple cycle", func(t *testing.T) {
		dag := NewDAG()
		dag.AddNode("a", []string{"b"})
		dag.AddNode("b", []string{"a"})

		require.True(t, dag.HasCycle())
		require.Error(t, dag.Validate())
		require.Contains(t, dag.Validate().Error(), "cycle detected")
	})

	t.Run("detects indirect cycle", func(t *testing.T) {
		dag := NewDAG()
		dag.AddNode("a", []string{"b"})
		dag.AddNode("b", []string{"c"})
		dag.AddNode("c", []string{"a"})

		require.True(t, dag.HasCycle())
		require.Error(t, dag.Validate())
	})

	t.Run("no cycle in complex graph", func(t *testing.T) {
		dag := NewDAG()
		dag.AddNode("a", []string{})
		dag.AddNode("b", []string{})
		dag.AddNode("c", []string{"a", "b"})
		dag.AddNode("d", []string{"c"})

		require.False(t, dag.HasCycle())
		require.NoError(t, dag.Validate())
	})

	t.Run("GetReadyNodes returns nodes with no dependencies", func(t *testing.T) {
		dag := NewDAG()
		dag.AddNode("a", []string{})
		dag.AddNode("b", []string{})
		dag.AddNode("c", []string{"a", "b"})

		ready := dag.GetReadyNodes(map[string]bool{})
		require.Len(t, ready, 2)
		require.Contains(t, ready, "a")
		require.Contains(t, ready, "b")
	})

	t.Run("GetReadyNodes returns next available nodes", func(t *testing.T) {
		dag := NewDAG()
		dag.AddNode("a", []string{})
		dag.AddNode("b", []string{})
		dag.AddNode("c", []string{"a", "b"})
		dag.AddNode("d", []string{"c"})

		completed := map[string]bool{
			"a": true,
			"b": true,
		}

		ready := dag.GetReadyNodes(completed)
		require.Len(t, ready, 1)
		require.Equal(t, "c", ready[0])
	})

	t.Run("GetReadyNodes waits for all dependencies", func(t *testing.T) {
		dag := NewDAG()
		dag.AddNode("a", []string{})
		dag.AddNode("b", []string{})
		dag.AddNode("c", []string{"a", "b"})

		completed := map[string]bool{
			"a": true,
		}

		ready := dag.GetReadyNodes(completed)
		require.Len(t, ready, 1)
		require.Equal(t, "b", ready[0])
	})

	t.Run("validates dependency exists", func(t *testing.T) {
		dag := NewDAG()
		dag.AddNode("a", []string{"nonexistent"})

		err := dag.Validate()
		require.Error(t, err)
		require.Contains(t, err.Error(), "does not exist")
	})
}
