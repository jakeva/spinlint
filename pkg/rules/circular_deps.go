package rules

import (
	"fmt"
	"strings"

	"github.com/jakeva/spinlint/pkg/schema"
)

// CircularDependencies detects cycles in the stage dependency graph using
// depth-first search. A cycle causes Spinnaker to deadlock at runtime.
type CircularDependencies struct{}

func (r CircularDependencies) Name() string { return "circular-dependencies" }

func (r CircularDependencies) Check(pipeline schema.Pipeline) []Violation {
	// Build adjacency list: refId → requisiteStageRefIds (known refIds only).
	// Unknown refs are handled by broken-requisite-refs, so we skip them here
	// to avoid false positives.
	known := make(map[string]bool, len(pipeline.Stages))
	adj := make(map[string][]string, len(pipeline.Stages))
	for _, stage := range pipeline.Stages {
		if stage.RefID != "" {
			known[stage.RefID] = true
			adj[stage.RefID] = stage.RequisiteStageRefIds
		}
	}

	const (
		unvisited = 0
		inStack   = 1
		done      = 2
	)
	color := make(map[string]int, len(adj))

	var violations []Violation

	var dfs func(node string, path []string)
	dfs = func(node string, path []string) {
		color[node] = inStack
		currentPath := append(append([]string(nil), path...), node)

		for _, dep := range adj[node] {
			if !known[dep] {
				continue // broken ref handled elsewhere
			}
			switch color[dep] {
			case unvisited:
				dfs(dep, currentPath)
			case inStack:
				// Back edge: locate where the cycle starts in currentPath.
				start := 0
				for i, n := range currentPath {
					if n == dep {
						start = i
						break
					}
				}
				cycle := append(currentPath[start:], dep)
				violations = append(violations, Violation{
					Rule:    r.Name(),
					Stage:   node,
					Message: fmt.Sprintf("circular dependency: %s", strings.Join(cycle, " → ")),
				})
			}
		}
		color[node] = done
	}

	for refID := range adj {
		if color[refID] == unvisited {
			dfs(refID, nil)
		}
	}

	return violations
}
