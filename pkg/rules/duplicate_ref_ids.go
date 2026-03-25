package rules

import (
	"fmt"

	"github.com/jakeva/spinlint/pkg/schema"
)

// DuplicateRefIDs flags any refId value that appears on more than one stage.
// Duplicate refIds cause Spinnaker to silently misroute stage dependencies.
type DuplicateRefIDs struct{}

func (r DuplicateRefIDs) Name() string { return "duplicate-ref-ids" }

func (r DuplicateRefIDs) Check(pipeline schema.Pipeline) []Violation {
	seen := make(map[string]int, len(pipeline.Stages)) // refId → first-seen index
	reported := make(map[string]bool)

	var violations []Violation
	for i, stage := range pipeline.Stages {
		if stage.RefID == "" {
			continue
		}
		if firstIdx, dup := seen[stage.RefID]; dup {
			if !reported[stage.RefID] {
				violations = append(violations, Violation{
					Rule:  r.Name(),
					Stage: stage.RefID,
					Message: fmt.Sprintf(
						"refId %q is used by stages at index %d and %d",
						stage.RefID, firstIdx, i,
					),
				})
				reported[stage.RefID] = true
			}
		} else {
			seen[stage.RefID] = i
		}
	}

	return violations
}
