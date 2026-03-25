package loader_test

import (
	"testing"

	"github.com/jakeva/spinlint/pkg/loader"
)

func TestLoadFile_Valid(t *testing.T) {
	p, err := loader.LoadFile("../../testdata/valid.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Name != "Deploy to Production" {
		t.Errorf("unexpected pipeline name: %q", p.Name)
	}
	if len(p.Stages) != 3 {
		t.Errorf("expected 3 stages, got %d", len(p.Stages))
	}
}

func TestLoadFile_Invalid(t *testing.T) {
	p, err := loader.LoadFile("../../testdata/invalid.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(p.Stages) != 3 {
		t.Errorf("expected 3 stages, got %d", len(p.Stages))
	}
}

func TestLoadFile_NotFound(t *testing.T) {
	_, err := loader.LoadFile("../../testdata/nonexistent.json")
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}

func TestGlob_MatchesFiles(t *testing.T) {
	paths, err := loader.Glob("../../testdata/*.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(paths) < 2 {
		t.Errorf("expected at least 2 matches, got %d", len(paths))
	}
}

func TestGlob_NoMatches(t *testing.T) {
	paths, err := loader.Glob("../../testdata/*.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(paths) != 0 {
		t.Errorf("expected 0 matches, got %d", len(paths))
	}
}
