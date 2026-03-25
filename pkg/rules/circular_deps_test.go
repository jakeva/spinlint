package rules_test

import (
	"testing"

	"github.com/jakeva/spinlint/pkg/rules"
	"github.com/jakeva/spinlint/pkg/schema"
)

func TestCircularDependencies_Valid(t *testing.T) {
	// Linear chain: 1 ← 2 ← 3 (no cycle)
	pipeline := schema.Pipeline{
		Stages: []schema.Stage{
			{Type: "wait", Name: "A", RefID: "1", RequisiteStageRefIds: []string{}},
			{Type: "wait", Name: "B", RefID: "2", RequisiteStageRefIds: []string{"1"}},
			{Type: "wait", Name: "C", RefID: "3", RequisiteStageRefIds: []string{"2"}},
		},
	}
	v := rules.CircularDependencies{}.Check(pipeline)
	if len(v) != 0 {
		t.Errorf("expected no violations for linear chain, got %d: %v", len(v), v)
	}
}

func TestCircularDependencies_DirectCycle(t *testing.T) {
	// A depends on B, B depends on A
	pipeline := schema.Pipeline{
		Stages: []schema.Stage{
			{Type: "wait", Name: "A", RefID: "1", RequisiteStageRefIds: []string{"2"}},
			{Type: "wait", Name: "B", RefID: "2", RequisiteStageRefIds: []string{"1"}},
		},
	}
	v := rules.CircularDependencies{}.Check(pipeline)
	if len(v) == 0 {
		t.Error("expected at least one violation for direct cycle, got none")
	}
}

func TestCircularDependencies_ThreeNodeCycle(t *testing.T) {
	// 1 → 2 → 3 → 1
	pipeline := schema.Pipeline{
		Stages: []schema.Stage{
			{Type: "wait", Name: "A", RefID: "1", RequisiteStageRefIds: []string{"3"}},
			{Type: "wait", Name: "B", RefID: "2", RequisiteStageRefIds: []string{"1"}},
			{Type: "wait", Name: "C", RefID: "3", RequisiteStageRefIds: []string{"2"}},
		},
	}
	v := rules.CircularDependencies{}.Check(pipeline)
	if len(v) == 0 {
		t.Error("expected violations for 3-node cycle, got none")
	}
}

func TestCircularDependencies_SelfLoop(t *testing.T) {
	pipeline := schema.Pipeline{
		Stages: []schema.Stage{
			{Type: "wait", Name: "A", RefID: "1", RequisiteStageRefIds: []string{"1"}},
		},
	}
	v := rules.CircularDependencies{}.Check(pipeline)
	if len(v) == 0 {
		t.Error("expected violation for self-loop, got none")
	}
}

func TestCircularDependencies_IgnoresBrokenRefs(t *testing.T) {
	// Ref "99" doesn't exist — broken-requisite-refs handles it, not us.
	pipeline := schema.Pipeline{
		Stages: []schema.Stage{
			{Type: "wait", Name: "A", RefID: "1", RequisiteStageRefIds: []string{"99"}},
		},
	}
	v := rules.CircularDependencies{}.Check(pipeline)
	if len(v) != 0 {
		t.Errorf("broken refs should not trigger circular-dependencies, got %v", v)
	}
}
