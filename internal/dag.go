package internal

import "fmt"

type DAG struct {
	nodes map[string]*dagNode
}

type dagNode struct {
	name         string
	dependencies []string
}

func NewDAG() *DAG {
	return &DAG{
		nodes: make(map[string]*dagNode),
	}
}

func (d *DAG) AddNode(name string, dependencies []string) {
	d.nodes[name] = &dagNode{
		name:         name,
		dependencies: dependencies,
	}
}

func (d *DAG) HasCycle() bool {
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	for name := range d.nodes {
		if d.hasCycleUtil(name, visited, recStack) {
			return true
		}
	}
	return false
}

func (d *DAG) hasCycleUtil(name string, visited, recStack map[string]bool) bool {
	if recStack[name] {
		return true
	}
	if visited[name] {
		return false
	}

	visited[name] = true
	recStack[name] = true

	node := d.nodes[name]
	if node != nil {
		for _, dep := range node.dependencies {
			if d.hasCycleUtil(dep, visited, recStack) {
				return true
			}
		}
	}

	recStack[name] = false
	return false
}

func (d *DAG) GetReadyNodes(completed map[string]bool) []string {
	var ready []string

	for name, node := range d.nodes {
		if completed[name] {
			continue
		}

		allDepsCompleted := true
		for _, dep := range node.dependencies {
			if !completed[dep] {
				allDepsCompleted = false
				break
			}
		}

		if allDepsCompleted {
			ready = append(ready, name)
		}
	}

	return ready
}

func (d *DAG) Validate() error {
	if d.HasCycle() {
		return fmt.Errorf("dependency cycle detected")
	}

	// Check all dependencies exist
	for name, node := range d.nodes {
		for _, dep := range node.dependencies {
			if _, exists := d.nodes[dep]; !exists {
				return fmt.Errorf("node '%s' depends on '%s' which does not exist", name, dep)
			}
		}
	}

	return nil
}
