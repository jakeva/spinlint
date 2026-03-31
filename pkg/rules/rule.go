package rules

import "github.com/jakeva/spinlint/pkg/schema"

// Violation describes a single rule failure found in a pipeline.
type Violation struct {
	Rule     string `json:"rule"`
	Stage    string `json:"stage,omitempty"`    // refId or "index N" if refId is absent
	Message  string `json:"message"`
	Severity string `json:"severity,omitempty"` // "warning" or "" (treated as "error")
}

// Rule is implemented by every lint check.
type Rule interface {
	Name() string
	Check(pipeline schema.Pipeline) []Violation
}
