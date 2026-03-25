package rules_test

import (
	"testing"

	"github.com/jakeva/spinlint/pkg/rules"
	"github.com/jakeva/spinlint/pkg/schema"
)

func TestDuplicateRefIDs_Valid(t *testing.T) {
	pipeline := schema.Pipeline{
		Stages: []schema.Stage{
			{Type: "wait", Name: "A", RefID: "1"},
			{Type: "wait", Name: "B", RefID: "2"},
			{Type: "wait", Name: "C", RefID: "3"},
		},
	}
	v := rules.DuplicateRefIDs{}.Check(pipeline)
	if len(v) != 0 {
		t.Errorf("expected no violations, got %d: %v", len(v), v)
	}
}

func TestDuplicateRefIDs_SingleDuplicate(t *testing.T) {
	pipeline := schema.Pipeline{
		Stages: []schema.Stage{
			{Type: "wait", Name: "A", RefID: "1"},
			{Type: "deploy", Name: "B", RefID: "1"},
		},
	}
	v := rules.DuplicateRefIDs{}.Check(pipeline)
	if len(v) != 1 {
		t.Errorf("expected 1 violation, got %d: %v", len(v), v)
	}
}

func TestDuplicateRefIDs_TriplicateReportsOnce(t *testing.T) {
	pipeline := schema.Pipeline{
		Stages: []schema.Stage{
			{Type: "wait", Name: "A", RefID: "1"},
			{Type: "wait", Name: "B", RefID: "1"},
			{Type: "wait", Name: "C", RefID: "1"},
		},
	}
	// Only one violation per duplicated refId, not one per extra occurrence.
	v := rules.DuplicateRefIDs{}.Check(pipeline)
	if len(v) != 1 {
		t.Errorf("expected 1 violation for triplicate refId, got %d: %v", len(v), v)
	}
}

func TestDuplicateRefIDs_SkipsEmptyRefID(t *testing.T) {
	pipeline := schema.Pipeline{
		Stages: []schema.Stage{
			{Type: "wait", Name: "A", RefID: ""},
			{Type: "wait", Name: "B", RefID: ""},
		},
	}
	v := rules.DuplicateRefIDs{}.Check(pipeline)
	if len(v) != 0 {
		t.Errorf("empty refIds should not trigger duplicate check, got %v", v)
	}
}
