package reporter

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/jakeva/spinlint/pkg/rules"
)

// Result holds the lint outcome for a single file.
type Result struct {
	File       string           `json:"file"`
	Violations []rules.Violation `json:"violations"`
}

// Reporter writes lint results to an output stream in the chosen format.
type Reporter struct {
	out     io.Writer
	format  string
	results []Result // buffered for JSON output
}

// New creates a Reporter that writes to out using the given format ("text" or "json").
func New(out io.Writer, format string) *Reporter {
	return &Reporter{out: out, format: format}
}

// Add records the lint result for a single file. For text format it writes
// immediately; for JSON it buffers until Flush is called.
func (r *Reporter) Add(file string, violations []rules.Violation) {
	if r.format == "json" {
		vv := violations
		if vv == nil {
			vv = []rules.Violation{}
		}
		r.results = append(r.results, Result{File: file, Violations: vv})
		return
	}

	// text (default)
	if len(violations) == 0 {
		fmt.Fprintf(r.out, "%s: OK\n", file)
		return
	}
	for _, v := range violations {
		fmt.Fprintf(r.out, "%s: [%s] %s\n", file, v.Rule, v.Message)
	}
}

// Flush writes buffered JSON output. It is a no-op for text format.
func (r *Reporter) Flush() error {
	if r.format != "json" {
		return nil
	}
	enc := json.NewEncoder(r.out)
	enc.SetIndent("", "  ")
	return enc.Encode(r.results)
}
