package rules_test

import (
	"testing"

	"github.com/jakeva/spinlint/pkg/rules"
	"github.com/jakeva/spinlint/pkg/schema"
)

func TestRequiredStageFields_Valid(t *testing.T) {
	pipeline := schema.Pipeline{
		Name: "test",
		Stages: []schema.Stage{
			{Type: "wait", Name: "Wait", RefID: "1"},
		},
	}
	violations := rules.RequiredStageFields{}.Check(pipeline)
	if len(violations) != 0 {
		t.Errorf("expected no violations, got %d: %v", len(violations), violations)
	}
}

func TestRequiredStageFields_MissingAll(t *testing.T) {
	pipeline := schema.Pipeline{
		Stages: []schema.Stage{
			{Type: "", Name: "", RefID: ""},
		},
	}
	violations := rules.RequiredStageFields{}.Check(pipeline)
	if len(violations) != 3 {
		t.Errorf("expected 3 violations (type, name, refId), got %d: %v", len(violations), violations)
	}
}

func TestRequiredStageFields_MissingName(t *testing.T) {
	pipeline := schema.Pipeline{
		Stages: []schema.Stage{
			{Type: "wait", Name: "", RefID: "1"},
		},
	}
	violations := rules.RequiredStageFields{}.Check(pipeline)
	if len(violations) != 1 {
		t.Errorf("expected 1 violation, got %d: %v", len(violations), violations)
	}
	if violations[0].Rule != "required-stage-fields" {
		t.Errorf("unexpected rule name: %s", violations[0].Rule)
	}
}
