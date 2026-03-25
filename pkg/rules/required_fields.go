package rules

import (
	"fmt"

	"github.com/jakeva/spinlint/pkg/schema"
)

// RequiredStageFields enforces that every stage has a non-empty type, name, and refId.
type RequiredStageFields struct{}

func (r RequiredStageFields) Name() string { return "required-stage-fields" }

func (r RequiredStageFields) Check(pipeline schema.Pipeline) []Violation {
	var violations []Violation

	for i, stage := range pipeline.Stages {
		ref := stage.RefID
		if ref == "" {
			ref = fmt.Sprintf("index %d", i)
		}

		if stage.Type == "" {
			violations = append(violations, Violation{
				Rule:    r.Name(),
				Stage:   ref,
				Message: fmt.Sprintf("stage %q is missing required field 'type'", ref),
			})
		}
		if stage.Name == "" {
			violations = append(violations, Violation{
				Rule:    r.Name(),
				Stage:   ref,
				Message: fmt.Sprintf("stage %q is missing required field 'name'", ref),
			})
		}
		if stage.RefID == "" {
			violations = append(violations, Violation{
				Rule:    r.Name(),
				Stage:   fmt.Sprintf("index %d", i),
				Message: fmt.Sprintf("stage at index %d is missing required field 'refId'", i),
			})
		}
	}

	return violations
}
