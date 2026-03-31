package rules

import (
	"fmt"

	"github.com/jakeva/spinlint/pkg/schema"
)

// OrphanedStages flags stages that are completely disconnected from the pipeline
// graph: no other stage depends on them AND they have no prerequisites of their own.
// These are isolated islands that Spinnaker will execute independently, which is
// usually unintentional. Severity: warning.
type OrphanedStages struct{}

func (r OrphanedStages) Name() string { return "orphaned-stages" }

func (r OrphanedStages) Check(pipeline schema.Pipeline) []Violation {
	// Build the set of refIds that at least one other stage depends on.
	dependedOn := make(map[string]bool)
	for _, stage := range pipeline.Stages {
		for _, dep := range stage.RequisiteStageRefIds {
			dependedOn[dep] = true
		}
	}

	var violations []Violation
	for _, stage := range pipeline.Stages {
		if stage.RefID == "" {
			continue // empty refId is caught by required-stage-fields
		}
		// A stage is an orphaned island if nothing depends on it AND it has no
		// prerequisites of its own — i.e. it is completely disconnected.
		if !dependedOn[stage.RefID] && len(stage.RequisiteStageRefIds) == 0 {
			violations = append(violations, Violation{
				Rule:     r.Name(),
				Stage:    stage.RefID,
				Severity: "warning",
				Message:  fmt.Sprintf("stage %q is an orphaned island: nothing depends on it and it has no prerequisites", stage.RefID),
			})
		}
	}
	return violations
}
