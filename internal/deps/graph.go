package deps

import (
	"errors"
	"fmt"
)

var ErrCircularDependency = errors.New("circular dependency detected")

type DependencyGraph struct {
	Adjacency map[string][]string
}

func NewGraph() *DependencyGraph {
	return &DependencyGraph{
		Adjacency: make(map[string][]string),
	}
}

func (g *DependencyGraph) AddDependency(pkg, dep string) {
	g.Adjacency[pkg] = append(g.Adjacency[pkg], dep)
}

func (g *DependencyGraph) CheckCircularity() error {
	visited := make(map[string]bool)
	onStack := make(map[string]bool)

	for node := range g.Adjacency {
		if !visited[node] {
			if g.hasCycle(node, visited, onStack) {
				return ErrCircularDependency
			}
		}
	}
	return nil
}

func (g *DependencyGraph) hasCycle(node string, visited, onStack map[string]bool) bool {
	visited[node] = true
	onStack[node] = true

	for _, neighbor := range g.Adjacency[node] {
		if !visited[neighbor] {
			if g.hasCycle(neighbor, visited, onStack) {
				return true
			}
		} else if onStack[neighbor] {
			return true
		}
	}

	onStack[node] = false
	return false
}

func (g *DependencyGraph) TopologicalSort() ([]string, error) {
	if err := g.CheckCircularity(); err != nil {
		return nil, err
	}

	visited := make(map[string]bool)
	var stack []string

	for node := range g.Adjacency {
		if !visited[node] {
			g.sortVisit(node, visited, &stack)
		}
	}

	// We also need to include nodes that are dependencies but don't have their own dependencies
	for _, deps := range g.Adjacency {
		for _, dep := range deps {
			if !visited[dep] {
				g.sortVisit(dep, visited, &stack)
			}
		}
	}

	return stack, nil
}

func (g *DependencyGraph) sortVisit(node string, visited map[string]bool, stack *[]string) {
	visited[node] = true
	for _, neighbor := range g.Adjacency[node] {
		if !visited[neighbor] {
			g.sortVisit(neighbor, visited, stack)
		}
	}
	*stack = append(*stack, node)
}

func (g *DependencyGraph) Resolve(pkg string, getDeps func(string) ([]string, error)) ([]string, error) {
	visited := make(map[string]bool)
	onStack := make(map[string]bool)
	var result []string

	var visit func(string) error
	visit = func(p string) error {
		if onStack[p] {
			return fmt.Errorf("%w: %s involved in cycle", ErrCircularDependency, p)
		}
		if visited[p] {
			return nil
		}

		onStack[p] = true
		deps, err := getDeps(p)
		if err != nil {
			return err
		}

		for _, d := range deps {
			if err := visit(d); err != nil {
				return err
			}
		}

		onStack[p] = false
		visited[p] = true
		result = append(result, p)
		return nil
	}

	if err := visit(pkg); err != nil {
		return nil, err
	}

	return result, nil
}
