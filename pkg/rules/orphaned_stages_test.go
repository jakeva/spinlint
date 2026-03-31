package rules_test

import (
	"testing"

	"github.com/jakeva/spinlint/pkg/rules"
	"github.com/jakeva/spinlint/pkg/schema"
)

func TestOrphanedStages_NoOrphans(t *testing.T) {
	// Linear chain: 1 ← 2. Stage 1 is a root (depended on), stage 2 is a leaf (has prereqs).
	pipeline := schema.Pipeline{
		Stages: []schema.Stage{
			{Type: "wait", Name: "A", RefID: "1", RequisiteStageRefIds: []string{}},
			{Type: "wait", Name: "B", RefID: "2", RequisiteStageRefIds: []string{"1"}},
		},
	}
	v := rules.OrphanedStages{}.Check(pipeline)
	if len(v) != 0 {
		t.Errorf("expected no violations, got %d: %v", len(v), v)
	}
}

func TestOrphanedStages_OrphanedIsland(t *testing.T) {
	// Stage 3 has no prerequisites and nothing depends on it — orphaned island.
	pipeline := schema.Pipeline{
		Stages: []schema.Stage{
			{Type: "wait", Name: "A", RefID: "1", RequisiteStageRefIds: []string{}},
			{Type: "wait", Name: "B", RefID: "2", RequisiteStageRefIds: []string{"1"}},
			{Type: "wait", Name: "C", RefID: "3", RequisiteStageRefIds: []string{}},
		},
	}
	v := rules.OrphanedStages{}.Check(pipeline)
	if len(v) != 1 {
		t.Fatalf("expected 1 violation, got %d: %v", len(v), v)
	}
	if v[0].Stage != "3" {
		t.Errorf("expected Stage %q, got %q", "3", v[0].Stage)
	}
	if v[0].Severity != "warning" {
		t.Errorf("expected Severity %q, got %q", "warning", v[0].Severity)
	}
}

func TestOrphanedStages_TerminalStageNotFlagged(t *testing.T) {
	// Stage 2 is a leaf node (nothing depends on it) but has prerequisites — valid terminal stage.
	pipeline := schema.Pipeline{
		Stages: []schema.Stage{
			{Type: "wait", Name: "A", RefID: "1", RequisiteStageRefIds: []string{}},
			{Type: "wait", Name: "B", RefID: "2", RequisiteStageRefIds: []string{"1"}},
		},
	}
	v := rules.OrphanedStages{}.Check(pipeline)
	if len(v) != 0 {
		t.Errorf("terminal stage should not be flagged, got %d: %v", len(v), v)
	}
}

func TestOrphanedStages_SkipsEmptyRefID(t *testing.T) {
	// A stage with no refId is skipped — caught by required-stage-fields, not us.
	pipeline := schema.Pipeline{
		Stages: []schema.Stage{
			{Type: "wait", Name: "A", RefID: "", RequisiteStageRefIds: []string{}},
		},
	}
	v := rules.OrphanedStages{}.Check(pipeline)
	if len(v) != 0 {
		t.Errorf("stage with empty refId should be skipped, got %d: %v", len(v), v)
	}
}

func TestOrphanedStages_MultipleOrphans(t *testing.T) {
	// Two disconnected islands alongside a valid chain.
	pipeline := schema.Pipeline{
		Stages: []schema.Stage{
			{Type: "wait", Name: "Chain start", RefID: "1", RequisiteStageRefIds: []string{}},
			{Type: "wait", Name: "Chain end", RefID: "2", RequisiteStageRefIds: []string{"1"}},
			{Type: "wait", Name: "Orphan A", RefID: "3", RequisiteStageRefIds: []string{}},
			{Type: "wait", Name: "Orphan B", RefID: "4", RequisiteStageRefIds: []string{}},
		},
	}
	v := rules.OrphanedStages{}.Check(pipeline)
	if len(v) != 2 {
		t.Errorf("expected 2 violations, got %d: %v", len(v), v)
	}
}

func TestOrphanedStages_ComplexTopology(t *testing.T) {
	// Pipeline layout:
	//
	//   Diamond subgraph:
	//     "1" ──► "2" ──► "4" ──► "5"  (terminal — has prereqs, not orphaned)
	//     "1" ──► "3" ──► "4"
	//
	//   Separate independent chain:
	//     "6" ──► "7"  (root "6" is depended-on; terminal "7" has prereqs)
	//
	//   Orphaned islands (no prereqs, nothing depends on them):
	//     "8", "9"
	//
	//   Stage with empty refId — must be skipped.
	//
	// Expected: exactly 2 warnings, for stages "8" and "9".
	pipeline := schema.Pipeline{
		Stages: []schema.Stage{
			{Type: "jenkins", Name: "Build",          RefID: "1", RequisiteStageRefIds: []string{}},
			{Type: "deploy",  Name: "Deploy canary",  RefID: "2", RequisiteStageRefIds: []string{"1"}},
			{Type: "deploy",  Name: "Deploy baseline", RefID: "3", RequisiteStageRefIds: []string{"1"}},
			{Type: "manualJudgment", Name: "Approve", RefID: "4", RequisiteStageRefIds: []string{"2", "3"}},
			{Type: "deploy",  Name: "Deploy prod",    RefID: "5", RequisiteStageRefIds: []string{"4"}},
			{Type: "bake",    Name: "Bake image",     RefID: "6", RequisiteStageRefIds: []string{}},
			{Type: "deploy",  Name: "Deploy staging", RefID: "7", RequisiteStageRefIds: []string{"6"}},
			{Type: "wait",    Name: "Stale wait",     RefID: "8", RequisiteStageRefIds: []string{}},
			{Type: "webhook", Name: "Dead webhook",   RefID: "9", RequisiteStageRefIds: []string{}},
			{Type: "wait",    Name: "No refId",       RefID: "",  RequisiteStageRefIds: []string{}},
		},
	}

	v := rules.OrphanedStages{}.Check(pipeline)
	if len(v) != 2 {
		t.Fatalf("expected 2 violations, got %d: %v", len(v), v)
	}

	orphaned := map[string]bool{}
	for _, viol := range v {
		orphaned[viol.Stage] = true
		if viol.Severity != "warning" {
			t.Errorf("stage %q: expected severity %q, got %q", viol.Stage, "warning", viol.Severity)
		}
	}
	for _, id := range []string{"8", "9"} {
		if !orphaned[id] {
			t.Errorf("expected stage %q to be flagged as orphaned", id)
		}
	}
	for _, id := range []string{"1", "2", "3", "4", "5", "6", "7"} {
		if orphaned[id] {
			t.Errorf("stage %q should not be flagged as orphaned", id)
		}
	}
}
