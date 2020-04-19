package dag

import (
	"errors"
)

// Graph is a directed acyclic graph.
type Graph []*Node

// Node represents a single graph node.
type Node struct {
	Value   interface{}
	Edges   Graph
	visited bool
	current bool
}

// Resolve sorts the graph using depth-first search.
func (g *Graph) Resolve() error {
	unresolved := *g
	if len(unresolved) == 0 {
		return errors.New("empty graph")
	}

	var resolved Graph
	prev := len(unresolved)
	for len(unresolved) > 0 {
		if err := visit(unresolved[0], &unresolved, &resolved); err != nil {
			return err
		}
		if len(unresolved) == prev {
			return errors.New("invalid graph")
		}
		prev = len(unresolved)
	}

	*g = resolved

	return nil
}

func (g *Graph) index(n *Node) (int, bool) {
	for i, node := range *g {
		if node == n {
			return i, true
		}
	}
	return 0, false
}

func visit(n *Node, unresolved *Graph, resolved *Graph) error {
	if n.visited {
		return nil
	} else if n.current {
		return errors.New("cycle detected") // TODO give more information
	}

	n.current = true
	for _, edge := range n.Edges {
		if err := visit(edge, unresolved, resolved); err != nil {
			return err
		}
	}
	n.current = false
	n.visited = true

	if i, found := unresolved.index(n); found {
		// Remove node.
		*unresolved = append((*unresolved)[:i], (*unresolved)[i+1:]...)
	}

	*resolved = append(*resolved, n)
	return nil
}
