# Contributing to spinlint

## Adding a Rule

All lint rules implement the two-method [`Rule` interface](pkg/rules/rule.go):

```go
type Rule interface {
    Name() string
    Check(pipeline schema.Pipeline) []Violation
}
```

`Violation` carries the rule name, the affected stage's `refId`, a human-readable message, and an optional severity:

```go
type Violation struct {
    Rule     string `json:"rule"`
    Stage    string `json:"stage,omitempty"`
    Message  string `json:"message"`
    Severity string `json:"severity,omitempty"` // "warning" or "" (treated as "error")
}
```

Follow these steps to add a new rule. For more context, see the [Adding a Rule](README.md#adding-a-rule) section of the README.

### 1. Create the rule file

Create `pkg/rules/<your_rule>.go`:

```go
package rules

import (
    "fmt"
    "github.com/jakeva/spinlint/pkg/schema"
)

type YourRule struct{}

func (r YourRule) Name() string { return "your-rule-name" }

func (r YourRule) Check(pipeline schema.Pipeline) []Violation {
    var violations []Violation
    for _, stage := range pipeline.Stages {
        if /* condition */ {
            violations = append(violations, Violation{
                Rule:    r.Name(),
                Stage:   stage.RefID,
                Message: fmt.Sprintf("..."),
                // Severity: "warning", // omit for error (the default)
            })
        }
    }
    return violations
}
```

Set `Severity: "warning"` for advisory checks; omit it for hard errors. Errors cause a non-zero exit code and map to `"error"` level in SARIF output; warnings map to `"warning"` level.

### 2. Register the rule

Add your rule to `pkg/rules/registry.go`:

```go
var All = []Rule{
    RequiredStageFields{},
    BrokenRequisiteRefs{},
    DuplicateRefIDs{},
    CircularDependencies{},
    OrphanedStages{},
    YourRule{}, // add here
}
```

Rules run in the order listed.

### 3. Add a schema field (if needed)

If your rule needs data not yet modelled, add it to `pkg/schema/pipeline.go`:

```go
type Stage struct {
    // ...existing fields...
    YourNewField bool `json:"yourNewField"`
}
```

### 4. Write tests

Create `pkg/rules/<your_rule>_test.go` using the `rules_test` package. See existing `*_test.go` files for the pattern. At minimum test the happy path (no violations) and one violation case.

### 5. Update the README

Add your rule to the **Rules** section of `README.md` following the same format as existing rules.

---

## Development workflow

```bash
make lint && make test
```

Requirements: Go 1.22+, `golangci-lint` (for the `lint` target only).
