# spinlint

A static linter for [Spinnaker](https://spinnaker.io/) pipeline JSON definitions. Catch misconfigured stages before they reach your Spinnaker instance and cause silent failures or runtime deadlocks.

---

## Table of Contents

- [Installation](#installation)
- [Usage](#usage)
- [Output Formats](#output-formats)
- [Rules](#rules)
- [Exit Codes](#exit-codes)
- [Project Structure](#project-structure)
- [Adding a Rule](#adding-a-rule)
- [Development](#development)

---

## Installation

**From source** (requires Go 1.22+):

```bash
git clone https://github.com/jakeva/spinlint.git
cd spinlint
make build
# binary is at ./bin/spinlint
```

**With `go install`:**

```bash
go install github.com/jakeva/spinlint/cmd/spinlint@latest
```

---

## Usage

```
spinlint validate <file|glob> [...]
```

Validate a single file:

```bash
spinlint validate pipelines/deploy.json
```

Validate all JSON files in a directory:

```bash
spinlint validate 'pipelines/*.json'
```

Validate multiple explicit files:

```bash
spinlint validate deploy.json canary.json rollback.json
```

Mix globs and explicit paths:

```bash
spinlint validate 'pipelines/**/*.json' overrides/hotfix.json
```

### Flags

| Flag | Short | Default | Description |
|---|---|---|---|
| `--format` | `-f` | `text` | Output format: `text`, `json`, or `sarif` |

---

## Output Formats

### text (default)

One line per violation. Clean files print `OK`.

```
pipelines/deploy.json: OK
pipelines/broken.json: [required-stage-fields] stage "2" is missing required field 'type'
pipelines/broken.json: [broken-requisite-refs] stage "2" references unknown refId "99" in requisiteStageRefIds
pipelines/broken.json: [required-stage-fields] stage at index 2 is missing required field 'refId'
```

Each line follows the pattern:

```
<file>: [<rule>] <message>
```

### json

Buffers all results and emits a single JSON array. Useful for CI systems, dashboards, or piping into `jq`. Clean files produce `"violations": []` (never `null`).

```bash
spinlint validate --format json 'pipelines/*.json'
```

```json
[
  {
    "file": "pipelines/deploy.json",
    "violations": []
  },
  {
    "file": "pipelines/broken.json",
    "violations": [
      {
        "rule": "required-stage-fields",
        "stage": "2",
        "message": "stage \"2\" is missing required field 'type'"
      },
      {
        "rule": "broken-requisite-refs",
        "stage": "2",
        "message": "stage \"2\" references unknown refId \"99\" in requisiteStageRefIds"
      }
    ]
  }
]
```

**Example: count total violations with `jq`**

```bash
spinlint validate --format json 'pipelines/*.json' \
  | jq '[.[].violations[]] | length'
```

**Example: list only files that failed**

```bash
spinlint validate --format json 'pipelines/*.json' \
  | jq -r '.[] | select(.violations | length > 0) | .file'
```

### sarif

Emits a [SARIF 2.1.0](https://docs.oasis-open.org/sarif/sarif/v2.1.0/sarif-v2.1.0.html) document. Upload it to [GitHub Code Scanning](https://docs.github.com/en/code-security/code-scanning) to surface violations as inline annotations directly on PR diffs — no log diving required.

```bash
spinlint validate --format sarif 'pipelines/*.json' > results.sarif
```

**GitHub Actions — upload to Code Scanning:**

```yaml
- name: Run spinlint (SARIF)
  run: ./bin/spinlint validate --format sarif 'pipelines/*.json' > results.sarif || true

- name: Upload SARIF to GitHub Code Scanning
  uses: github/codeql-action/upload-sarif@v3
  with:
    sarif_file: results.sarif
```

> The `|| true` prevents a non-zero exit from blocking the upload step. The job must have `permissions: security-events: write`.

Each violation maps to a SARIF `result` with `ruleId`, `level: "error"`, the violation message, and a `physicalLocation` URI relative to the repository root.

---

## Rules

Rules are run in the order listed below. Every rule is applied to every file independently.

### `required-stage-fields`

Every stage in a Spinnaker pipeline must declare three fields: `type`, `name`, and `refId`. Missing any of them causes Spinnaker to either reject the pipeline on import or behave unpredictably at runtime.

**Checks:**
- `type` is non-empty
- `name` is non-empty
- `refId` is non-empty

**Example violation:**

```
pipelines/deploy.json: [required-stage-fields] stage "3" is missing required field 'name'
```

---

### `broken-requisite-refs`

`requisiteStageRefIds` controls execution order — a stage only runs after all of its listed prerequisite stages have completed. If a `refId` listed there does not correspond to any stage in the pipeline, the dependency can never be satisfied.

**Checks:** every entry in `requisiteStageRefIds` matches the `refId` of a stage in the same pipeline.

**Example violation:**

```
pipelines/deploy.json: [broken-requisite-refs] stage "4" references unknown refId "99" in requisiteStageRefIds
```

---

### `duplicate-ref-ids`

`refId` values must be unique within a pipeline. Spinnaker uses `refId` as the primary key for resolving stage dependencies. Duplicate values cause Spinnaker to silently misroute execution — one stage may unexpectedly inherit the dependencies of another.

**Checks:** no two stages share the same `refId`. Stages with an empty `refId` are skipped (caught by `required-stage-fields` instead).

**Example violation:**

```
pipelines/deploy.json: [duplicate-ref-ids] refId "2" is used by stages at index 1 and 3
```

---

### `circular-dependencies`

Detects cycles in the stage dependency graph using depth-first search with three-color marking (unvisited / in-stack / done). A cycle means two or more stages are each waiting on each other, causing the pipeline to deadlock at runtime — Spinnaker will spin forever with no error message.

Broken refs are skipped here (handled by `broken-requisite-refs`) to avoid false positives.

**Checks:** the directed graph formed by `requisiteStageRefIds` edges is acyclic.

**Example violation:**

```
pipelines/deploy.json: [circular-dependencies] circular dependency: 1 → 2 → 3 → 1
```

Self-loops (`requisiteStageRefIds: ["1"]` on stage `"1"`) are also detected.

---

## Exit Codes

| Code | Meaning |
|---|---|
| `0` | All files passed — no violations found |
| `1` | One or more violations found, or a file could not be read/parsed |

This makes spinlint suitable for use as a pre-commit hook or CI gate:

```bash
spinlint validate 'pipelines/*.json' || exit 1
```

---

## Project Structure

```
spinlint/
├── cmd/
│   └── spinlint/
│       └── main.go          # CLI entry point (cobra root + validate command)
├── pkg/
│   ├── schema/
│   │   └── pipeline.go      # Pipeline and Stage types
│   ├── rules/
│   │   ├── rule.go          # Rule interface and Violation type
│   │   ├── registry.go      # var All []Rule — the active rule set
│   │   ├── required_fields.go
│   │   ├── requisite_refs.go
│   │   ├── duplicate_ref_ids.go
│   │   └── circular_deps.go
│   ├── loader/
│   │   └── loader.go        # Glob expansion and JSON file loading
│   └── reporter/
│       └── reporter.go      # Text and JSON output formatting
├── testdata/
│   ├── valid.json           # A well-formed 3-stage pipeline
│   └── invalid.json         # Pipeline with violations across multiple rules
├── go.mod
├── Makefile
└── README.md
```

---

## Adding a Rule

All rules implement a two-method interface defined in `pkg/rules/rule.go`:

```go
type Rule interface {
    Name() string
    Check(pipeline schema.Pipeline) []Violation
}
```

`Violation` carries the rule name, the affected stage's `refId` (optional), and a human-readable message:

```go
type Violation struct {
    Rule    string `json:"rule"`
    Stage   string `json:"stage,omitempty"`
    Message string `json:"message"`
}
```

To add a new rule:

**1. Create the rule file** in `pkg/rules/`:

```go
// pkg/rules/no_disabled_stages.go
package rules

import (
    "fmt"
    "github.com/jakeva/spinlint/pkg/schema"
)

type NoDisabledStages struct{}

func (r NoDisabledStages) Name() string { return "no-disabled-stages" }

func (r NoDisabledStages) Check(pipeline schema.Pipeline) []Violation {
    var violations []Violation
    for _, stage := range pipeline.Stages {
        if stage.IsDisabled {
            violations = append(violations, Violation{
                Rule:    r.Name(),
                Stage:   stage.RefID,
                Message: fmt.Sprintf("stage %q is disabled", stage.RefID),
            })
        }
    }
    return violations
}
```

**2. Register it** in `pkg/rules/registry.go`:

```go
var All = []Rule{
    RequiredStageFields{},
    BrokenRequisiteRefs{},
    DuplicateRefIDs{},
    CircularDependencies{},
    NoDisabledStages{}, // add here
}
```

**3. Add a schema field** in `pkg/schema/pipeline.go` if the rule needs data not yet modelled:

```go
type Stage struct {
    // ...existing fields...
    IsDisabled bool `json:"isDisabled"`
}
```

The rule will automatically be included in `text`, `json`, and `sarif` output with no further changes.

---

## Development

Requirements: Go 1.22+, `golangci-lint` (for the `lint` target only).

| Target | Command | Description |
|---|---|---|
| Build | `make build` | Compile to `./bin/spinlint` |
| Test | `make test` | Run all tests with `-v` |
| Vet | `make vet` | Run `go vet` |
| Lint | `make lint` | Run `go vet` then `golangci-lint` |
| Clean | `make clean` | Remove `./bin/` |

Run the full check suite before opening a PR:

```bash
make lint && make test
```
