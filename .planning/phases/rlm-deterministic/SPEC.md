# SPEC — Phase 2: RLM Deterministic Backend

## 1. Package layout

```
internal/rlm/
  rlm.go          — Analyzer interface, RunRequest, RunResult, backend registry, ErrNotImplemented
  spec.go         — Spec struct (feature definitions + weights), ParseSpec([]byte) (*Spec, error)
  deterministic.go — DeterministicAnalyzer: weighted-feature SQL scorer, offline, reproducible
  fixture.go      — FixtureAnalyzer: canned table rows, credential-free
  model.go        — ModelAnalyzer stub: method set only, all methods return ErrNotImplemented
  rlm_test.go     — shared table-driven tests
  spec_test.go    — spec parse tests
  deterministic_test.go — determinism + scoring tests
  fixture_test.go — fixture parity tests
  testdata/
    likely_customers.yaml  — example spec for end-to-end demo
```

## 2. Core interfaces and types

### Analyzer interface

```go
// Analyzer is the strategy interface for all RLM backends.
// Implementations must be safe for concurrent use from a single goroutine per Run call.
type Analyzer interface {
    // Run scores every record in req.InTable and materializes results to req.OutTable.
    // It must be deterministic: identical InTable + Spec always produces identical OutTable.
    Run(ctx context.Context, req RunRequest) (RunResult, error)
    // Mode returns the backend identifier ("deterministic", "fixture", "model").
    Mode() string
}
```

### RunRequest

```go
type RunRequest struct {
    Spec     *Spec  // parsed scoring spec
    InTable  string // source warehouse table name (no path — resolved via WarehouseDir)
    OutTable string // destination warehouse table name
    WarehouseDir string // resolved from app runtime config
    DryRun   bool   // if true, compute scores but do not write OutTable
}
```

### RunResult

```go
type RunResult struct {
    Mode           string        `json:"mode"`
    InTable        string        `json:"in_table"`
    OutTable       string        `json:"out_table"`
    RecordsRead    int           `json:"records_read"`
    RecordsScored  int           `json:"records_scored"`
    RecordsFailed  int           `json:"records_failed"`
    Duration       time.Duration `json:"duration_ns"`
    DryRun         bool          `json:"dry_run"`
}
```

### Spec

```go
type Spec struct {
    Name        string    `yaml:"name"`
    Description string    `yaml:"description,omitempty"`
    Features    []Feature `yaml:"features"`
}

type Feature struct {
    Name       string  `yaml:"name"`        // source field name in warehouse record
    Weight     float64 `yaml:"weight"`      // relative weight; all weights summed = raw total
    ScoreIfSet float64 `yaml:"score_if_set,omitempty"` // score to assign if field is non-empty/non-zero
    ScoreIfGT  *float64 `yaml:"score_if_gt,omitempty"` // score if numeric field > threshold
    Threshold  *float64 `yaml:"threshold,omitempty"`   // used with ScoreIfGT
    Default    float64  `yaml:"default,omitempty"`     // score when condition is false
}
```

### Materialized output record

Every row written to OutTable is a JSON object with the following reserved fields plus all original source record fields:

```json
{
  "_rlm_score":       0.87,
  "_rlm_mode":        "deterministic",
  "_rlm_spec":        "likely-customers",
  "_rlm_scored_at":   "2026-06-27T00:00:00Z",
  "<original fields>": "..."
}
```

## 3. Deterministic backend algorithm

1. Open InTable NDJSON file (one JSON object per line, same format as `internal/app` local warehouse).
2. For each record, iterate Features in spec order:
   - Compute feature score (ScoreIfSet, ScoreIfGT, Default).
   - Multiply by Weight.
3. Sum weighted scores; normalize by sum of all weights to produce `_rlm_score` in [0.0, 1.0].
4. Sort output records by `_rlm_score` descending, then by `_polymetrics_raw_id` ascending as tie-breaker (ensures identical ordering on identical input).
5. Write OutTable NDJSON atomically (temp file → rename).

## 4. Fixture backend

- Returns a hardcoded slice of `connectors.Record` (defined in `fixture.go` as package-level `var DefaultFixtureRows`).
- Scores each row using the same algorithm as deterministic (reuses scoring logic).
- Writes OutTable same as deterministic.
- Does not read InTable (ignores it). This is for credential-free CI and demos.

## 5. Model backend stub

```go
// ModelAnalyzer is a placeholder for the Phase 4 model backend.
// HUMAN GATE: do not implement until Phase 4 is approved.
type ModelAnalyzer struct{}

func (m *ModelAnalyzer) Run(_ context.Context, _ RunRequest) (RunResult, error) {
    return RunResult{}, ErrNotImplemented
}
func (m *ModelAnalyzer) Mode() string { return "model" }
```

`ErrNotImplemented = errors.New("rlm: model backend not implemented (requires Phase 4 approval)")`.

## 6. CLI verb

```
pm rlm run --spec <yaml-file> --in <table> --out <table> --mode <deterministic|fixture|model>
pm rlm run --spec <yaml-file> --in <table> --out <table> --mode deterministic --dry-run
pm rlm run --spec <yaml-file> --in <table> --out <table> --mode fixture
```

Dispatch is added to `internal/cli/cli.go` `switch cmd` as `case "rlm":`.
Implementation in `internal/cli/rlm_cli.go` (new file, follows existing patterns in `cli.go`).

Flags:
- `--spec` (required): path to YAML spec file
- `--in` (required for deterministic): source warehouse table name
- `--out` (required): destination warehouse table name
- `--mode` (required): deterministic | fixture | model
- `--dry-run`: score but do not materialize
- `--json`: machine-readable output (existing global flag)

Output (JSON envelope):
```json
{"mode":"deterministic","in_table":"contacts","out_table":"lead_scores","records_read":42,"records_scored":42,"records_failed":0,"duration_ns":1234567,"dry_run":false}
```

Exit codes follow existing `internal/cli` conventions (0=success, 1=user error, 2=internal error).

## 7. Flow step kind integration

In `internal/flow` (Phase 0 package), add `"rlm"` as a recognized step kind. A flow step of kind `rlm` has:
```yaml
kind: rlm
spec: likely_customers.yaml
in: contacts
out: lead_scores
mode: deterministic
```

The flow engine invokes `pm rlm run` arguments (or calls the `internal/rlm` package directly — implementation detail for the flow phase). The `internal/rlm` package must NOT import `internal/flow` (dependency must be one-way).

## 8. Warehouse format compatibility

- Local warehouse: `~/.polymetrics/<project>/warehouse/<table>.ndjson` — matches `internal/app/local_warehouse.go`'s `localWarehouseTablePath`.
- WarehouseDir is resolved the same way as `app.App` resolves it (via `localWarehouseDir(destRuntime)`).
- RLM reads the raw `.ndjson` file and parses the `record` field from each `localRawRecord` JSON object.

## 9. Observability

- Structured log lines to stderr: `[rlm] mode=deterministic in=contacts out=lead_scores records_read=42 records_scored=42 duration=1.2ms`
- `RunResult` is the primary machine-readable output (JSON envelope via `--json`).
- Ledger: append a `ledger.RunRecord` with `Operation="rlm"`, `Mode=<backend mode>`.
