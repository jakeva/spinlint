package rules

import (
	"fmt"

	"github.com/jakeva/spinlint/pkg/schema"
)

// BrokenRequisiteRefs checks that every refId listed in requisiteStageRefIds
// points to a stage that actually exists in the pipeline.
type BrokenRequisiteRefs struct{}

func (r BrokenRequisiteRefs) Name() string { return "broken-requisite-refs" }

func (r BrokenRequisiteRefs) Check(pipeline schema.Pipeline) []Violation {
	known := make(map[string]bool, len(pipeline.Stages))
	for _, stage := range pipeline.Stages {
		if stage.RefID != "" {
			known[stage.RefID] = true
		}
	}

	var violations []Violation
	for _, stage := range pipeline.Stages {
		for _, dep := range stage.RequisiteStageRefIds {
			if !known[dep] {
				violations = append(violations, Violation{
					Rule:  r.Name(),
					Stage: stage.RefID,
					Message: fmt.Sprintf(
						"stage %q references unknown refId %q in requisiteStageRefIds",
						stage.RefID, dep,
					),
				})
			}
		}
	}

	return violations
}
