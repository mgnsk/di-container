package dag

import (
	"errors"
	"fmt"
)

// Graph is a directed acyclic graph.
type Graph []*Node

// Node represents a single graph node.
type Node struct {
	Value   interface{}
	Edges   []*Node
	visited bool
	current bool
}

func (n *Node) visit(unresolved *Graph, resolved *Graph) error {
	if n.visited {
		return nil
	} else if n.current {
		return fmt.Errorf("cycle detected on type '%T'", n.Value)
	}

	n.current = true
	for _, edge := range n.Edges {
		if err := edge.visit(unresolved, resolved); err != nil {
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

// Resolve sorts the graph using depth-first search.
func (g *Graph) Resolve() error {
	var resolved Graph

	prev := len(*g)
	for prev > 0 {
		if err := (*g)[0].visit(g, &resolved); err != nil {
			return err
		}
		newLen := len(*g)
		if newLen == prev {
			return errors.New("invalid graph")
		}
		prev = newLen
	}

	*g = resolved

	return nil
}

func (g *Graph) index(n *Node) (int, bool) {
	graph := *g
	for i := range graph {
		if n == graph[i] {
			return i, true
		}
	}
	return 0, false
}
