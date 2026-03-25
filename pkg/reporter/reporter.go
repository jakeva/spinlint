package reporter

import (
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"

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
	results []Result // buffered for json/sarif output
}

// New creates a Reporter that writes to out using the given format ("text", "json", or "sarif").
func New(out io.Writer, format string) *Reporter {
	return &Reporter{out: out, format: format}
}

// Add records the lint result for a single file. For text format it writes
// immediately; for json/sarif it buffers until Flush is called.
func (r *Reporter) Add(file string, violations []rules.Violation) {
	if r.format == "text" {
		if len(violations) == 0 {
			fmt.Fprintf(r.out, "%s: OK\n", file)
			return
		}
		for _, v := range violations {
			fmt.Fprintf(r.out, "%s: [%s] %s\n", file, v.Rule, v.Message)
		}
		return
	}

	vv := violations
	if vv == nil {
		vv = []rules.Violation{}
	}
	r.results = append(r.results, Result{File: file, Violations: vv})
}

// Flush writes buffered output. It is a no-op for text format.
func (r *Reporter) Flush() error {
	switch r.format {
	case "json":
		enc := json.NewEncoder(r.out)
		enc.SetIndent("", "  ")
		return enc.Encode(r.results)
	case "sarif":
		return r.flushSARIF()
	}
	return nil
}

// --- SARIF 2.1.0 types -------------------------------------------------

type sarifLog struct {
	Schema  string     `json:"$schema"`
	Version string     `json:"version"`
	Runs    []sarifRun `json:"runs"`
}

type sarifRun struct {
	Tool    sarifTool     `json:"tool"`
	Results []sarifResult `json:"results"`
}

type sarifTool struct {
	Driver sarifDriver `json:"driver"`
}

type sarifDriver struct {
	Name           string      `json:"name"`
	InformationURI string      `json:"informationUri"`
	Rules          []sarifRule `json:"rules"`
}

type sarifRule struct {
	ID               string       `json:"id"`
	ShortDescription sarifMessage `json:"shortDescription"`
}

type sarifResult struct {
	RuleID    string          `json:"ruleId"`
	Level     string          `json:"level"`
	Message   sarifMessage    `json:"message"`
	Locations []sarifLocation `json:"locations"`
}

type sarifLocation struct {
	PhysicalLocation sarifPhysicalLocation `json:"physicalLocation"`
}

type sarifPhysicalLocation struct {
	ArtifactLocation sarifArtifactLocation `json:"artifactLocation"`
}

type sarifArtifactLocation struct {
	URI       string `json:"uri"`
	URIBaseID string `json:"uriBaseId"`
}

type sarifMessage struct {
	Text string `json:"text"`
}

// -----------------------------------------------------------------------

func (r *Reporter) flushSARIF() error {
	// Collect unique rule IDs from violations to populate tool.driver.rules.
	seenRules := map[string]bool{}
	var driverRules []sarifRule
	var results []sarifResult

	for _, res := range r.results {
		for _, v := range res.Violations {
			if !seenRules[v.Rule] {
				seenRules[v.Rule] = true
				driverRules = append(driverRules, sarifRule{
					ID:               v.Rule,
					ShortDescription: sarifMessage{Text: v.Rule},
				})
			}
			results = append(results, sarifResult{
				RuleID:  v.Rule,
				Level:   "error",
				Message: sarifMessage{Text: v.Message},
				Locations: []sarifLocation{
					{
						PhysicalLocation: sarifPhysicalLocation{
							ArtifactLocation: sarifArtifactLocation{
								URI:       toSARIFURI(res.File),
								URIBaseID: "%SRCROOT%",
							},
						},
					},
				},
			})
		}
	}

	// Ensure non-null arrays so GitHub Code Scanning parses the document correctly.
	if driverRules == nil {
		driverRules = []sarifRule{}
	}
	if results == nil {
		results = []sarifResult{}
	}

	log := sarifLog{
		Schema:  "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
		Version: "2.1.0",
		Runs: []sarifRun{
			{
				Tool: sarifTool{
					Driver: sarifDriver{
						Name:           "spinlint",
						InformationURI: "https://github.com/jakeva/spinlint",
						Rules:          driverRules,
					},
				},
				Results: results,
			},
		},
	}

	enc := json.NewEncoder(r.out)
	enc.SetIndent("", "  ")
	return enc.Encode(log)
}

// toSARIFURI converts a file path to a forward-slash relative URI suitable
// for SARIF artifactLocation.uri. Absolute paths are made relative to the
// working directory.
func toSARIFURI(path string) string {
	if filepath.IsAbs(path) {
		if rel, err := filepath.Rel(".", path); err == nil {
			path = rel
		}
	}
	return filepath.ToSlash(path)
}
