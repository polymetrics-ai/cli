package flow

import (
	"fmt"
)

// BuildDAG returns step IDs in topological order (Kahn's algorithm).
// Returns ErrCyclicDependency if a cycle is detected.
func BuildDAG(manifest FlowManifest) ([]string, error) {
	steps := manifest.Steps
	n := len(steps)
	if n == 0 {
		return nil, nil
	}

	// Map table name -> step IDs that produce it.
	tableProducer := map[string][]int{} // table -> slice of step indices
	for i, s := range steps {
		for _, t := range s.Out {
			tableProducer[t] = append(tableProducer[t], i)
		}
	}

	// Build adjacency list (edges: producer -> consumer) and in-degree.
	adj := make([][]int, n)
	inDeg := make([]int, n)
	for j, s := range steps {
		for _, t := range s.In {
			for _, i := range tableProducer[t] {
				if i == j {
					continue // self-loop guard
				}
				adj[i] = append(adj[i], j)
				inDeg[j]++
			}
		}
	}

	// Kahn's algorithm.
	queue := []int{}
	for i, d := range inDeg {
		if d == 0 {
			queue = append(queue, i)
		}
	}

	order := make([]string, 0, n)
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		order = append(order, steps[cur].ID)
		for _, next := range adj[cur] {
			inDeg[next]--
			if inDeg[next] == 0 {
				queue = append(queue, next)
			}
		}
	}

	if len(order) != n {
		// Collect remaining node IDs for the error message.
		remaining := []string{}
		for i, s := range steps {
			if inDeg[i] > 0 {
				remaining = append(remaining, s.ID)
			}
		}
		return nil, fmt.Errorf("%w: nodes in cycle: %v", ErrCyclicDependency, remaining)
	}

	return order, nil
}
