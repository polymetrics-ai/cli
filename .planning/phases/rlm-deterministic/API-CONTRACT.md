# API-CONTRACT — Phase 2: RLM Deterministic Backend

## Go package API (`internal/rlm`)

This is the stable package surface. Do not change signatures without a versioned ADR.

### Types

```go
package rlm

import (
    "context"
    "errors"
    "time"
    "polymetrics.ai/internal/connectors"
)

// ErrNotImplemented is returned by backends that are not yet implemented.
var ErrNotImplemented = errors.New("rlm: model backend not implemented (requires Phase 4 approval)")

// ErrUnknownMode is returned when an unrecognized mode string is passed to NewAnalyzer.
var ErrUnknownMode = errors.New("rlm: unknown mode")

// ErrInvalidTableName is returned when a table name contains disallowed characters.
var ErrInvalidTableName = errors.New("rlm: invalid table name")

// Analyzer is the strategy interface for all RLM backends.
type Analyzer interface {
    Run(ctx context.Context, req RunRequest) (RunResult, error)
    Mode() string
}

// RunRequest is the input to Analyzer.Run.
type RunRequest struct {
    Spec         *Spec
    InTable      string
    OutTable     string
    WarehouseDir string
    DryRun       bool
}

// RunResult is the output of Analyzer.Run.
type RunResult struct {
    Mode          string        `json:"mode"`
    InTable       string        `json:"in_table"`
    OutTable      string        `json:"out_table"`
    RecordsRead   int           `json:"records_read"`
    RecordsScored int           `json:"records_scored"`
    RecordsFailed int           `json:"records_failed"`
    Duration      time.Duration `json:"duration_ns"`
    DryRun        bool          `json:"dry_run"`
}

// Spec is a parsed and validated scoring specification.
type Spec struct {
    Name        string    `json:"name"`
    Description string    `json:"description,omitempty"`
    Features    []Feature `json:"features"`
}

// Feature is one weighted scoring rule within a Spec.
type Feature struct {
    Name       string   `json:"name"`
    Weight     float64  `json:"weight"`
    ScoreIfSet float64  `json:"score_if_set,omitempty"`
    ScoreIfGT  float64  `json:"score_if_gt,omitempty"`
    Threshold  *float64 `json:"threshold,omitempty"`
    Default    float64  `json:"default,omitempty"`
}

// ParseSpec parses and validates a JSON spec from raw bytes.
func ParseSpec(data []byte) (*Spec, error)

// NewAnalyzer returns an Analyzer for the given mode string.
// Valid modes: "deterministic", "fixture", "model".
// "model" returns a stub that always errors with ErrNotImplemented.
func NewAnalyzer(mode string) (Analyzer, error)
```

### Concrete types (exported for testing; use via Analyzer interface in production)

```go
type DeterministicAnalyzer struct {
    Ledger LedgerAppender // optional; nil = no ledger write
}

type FixtureAnalyzer struct {
    Rows []connectors.Record // defaults to DefaultFixtureRows if nil
}

type ModelAnalyzer struct{} // stub only

var DefaultFixtureRows []connectors.Record // at least 5 hardcoded records
```

### LedgerAppender interface (thin wrapper to avoid importing ledger package directly)

```go
type LedgerAppender interface {
    Append(ctx context.Context, record LedgerRecord) error
}

type LedgerRecord struct {
    ID             string
    Mode           string
    Operation      string
    Status         string
    RecordsRead    int
    RecordsWritten int
    Duration       int64 // nanoseconds
}
```

---

## CLI contract

### Command

```
pm rlm run [flags]
```

### Flags

| Flag | Type | Required | Description |
|---|---|---|---|
| `--spec` | string | yes | Path to JSON spec file |
| `--in` | string | yes (deterministic) | Source warehouse table name |
| `--out` | string | yes | Destination warehouse table name |
| `--mode` | string | yes | `deterministic`, `fixture`, or `model` |
| `--dry-run` | bool | no | Score but do not write OutTable |
| `--json` | bool | no | Global flag; machine-readable output |

### Exit codes

| Code | Meaning |
|---|---|
| 0 | Success |
| 1 | User error (bad flag, missing file, validation failure) |
| 2 | Internal error (warehouse I/O failure) |

### stdout (--json mode)

```json
{
  "mode": "deterministic",
  "in_table": "contacts",
  "out_table": "lead_scores",
  "records_read": 42,
  "records_scored": 42,
  "records_failed": 0,
  "duration_ns": 1234567,
  "dry_run": false
}
```

### stderr

Human-readable progress and error lines. Never contains secret values. Never required for machine consumption.

---

## OutTable NDJSON record schema

Each line of `<warehouse_dir>/<out_table>.ndjson` is a JSON object with:

| Field | Type | Description |
|---|---|---|
| `_rlm_score` | float64 [0.0,1.0] | Normalized weighted score |
| `_rlm_mode` | string | Backend mode that produced the score |
| `_rlm_spec` | string | `spec.Name` value |
| `_rlm_scored_at` | string (RFC3339) | UTC timestamp of the run |
| `<source fields>` | any | All fields from the original warehouse record |

The file is sorted: highest `_rlm_score` first; ties broken by `_polymetrics_raw_id` ascending.

---

## Flow step YAML (reference for Phase 0 integration)

```yaml
kind: rlm
id: score_contacts
spec: specs/likely_customers.json
in: contacts
out: lead_scores
mode: deterministic   # or fixture
dry_run: false        # optional
```

This is consumed by `internal/flow` (not `internal/rlm`). `internal/rlm` does not parse flow manifests.
