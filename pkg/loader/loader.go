package loader

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jakeva/spinlint/pkg/schema"
)

// Glob expands a file path or glob pattern into a list of matching paths.
// Returns an empty slice (no error) when nothing matches.
func Glob(pattern string) ([]string, error) {
	paths, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid glob pattern %q: %w", pattern, err)
	}
	return paths, nil
}

// LoadFile reads and parses a Spinnaker pipeline JSON file.
func LoadFile(path string) (schema.Pipeline, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return schema.Pipeline{}, fmt.Errorf("reading %s: %w", path, err)
	}

	var p schema.Pipeline
	if err := json.Unmarshal(data, &p); err != nil {
		return schema.Pipeline{}, fmt.Errorf("parsing %s: %w", path, err)
	}

	return p, nil
}
