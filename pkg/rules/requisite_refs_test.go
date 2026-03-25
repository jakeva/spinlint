package rules_test

import (
	"testing"

	"github.com/jakeva/spinlint/pkg/rules"
	"github.com/jakeva/spinlint/pkg/schema"
)

func TestBrokenRequisiteRefs_Valid(t *testing.T) {
	pipeline := schema.Pipeline{
		Stages: []schema.Stage{
			{Type: "wait", Name: "Wait", RefID: "1", RequisiteStageRefIds: []string{}},
			{Type: "deploy", Name: "Deploy", RefID: "2", RequisiteStageRefIds: []string{"1"}},
		},
	}
	violations := rules.BrokenRequisiteRefs{}.Check(pipeline)
	if len(violations) != 0 {
		t.Errorf("expected no violations, got %d: %v", len(violations), violations)
	}
}

func TestBrokenRequisiteRefs_BrokenRef(t *testing.T) {
	pipeline := schema.Pipeline{
		Stages: []schema.Stage{
			{Type: "wait", Name: "Wait", RefID: "1", RequisiteStageRefIds: []string{"99"}},
		},
	}
	violations := rules.BrokenRequisiteRefs{}.Check(pipeline)
	if len(violations) != 1 {
		t.Errorf("expected 1 violation, got %d: %v", len(violations), violations)
	}
	if violations[0].Rule != "broken-requisite-refs" {
		t.Errorf("unexpected rule name: %s", violations[0].Rule)
	}
}

func TestBrokenRequisiteRefs_MultipleBroken(t *testing.T) {
	pipeline := schema.Pipeline{
		Stages: []schema.Stage{
			{Type: "deploy", Name: "Deploy", RefID: "1", RequisiteStageRefIds: []string{"42", "99"}},
		},
	}
	violations := rules.BrokenRequisiteRefs{}.Check(pipeline)
	if len(violations) != 2 {
		t.Errorf("expected 2 violations, got %d: %v", len(violations), violations)
	}
}

func TestBrokenRequisiteRefs_EmptyPipeline(t *testing.T) {
	violations := rules.BrokenRequisiteRefs{}.Check(schema.Pipeline{})
	if len(violations) != 0 {
		t.Errorf("expected no violations for empty pipeline, got %d", len(violations))
	}
}
