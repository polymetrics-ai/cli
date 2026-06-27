# SPEC — Action Step (Phase 1)

## Package layout

```
internal/flow/
    manifest.go        — extend: add KindAction, ActionStep fields, validation
    engine.go          — extend: dispatch KindAction via ActionRunner
    action.go          — NEW: ActionRunner, all 7 safety features
    action_test.go     — NEW: all action tests (httptest.Server fake destination)
    errors.go          — extend: ErrApprovalRequired, ErrSchemaDrift, ErrDLQFull

internal/cli/
    flow_cli.go        — extend: --per-action flag; flow-level approval token flow
```

No new packages. No new go.mod dependencies.

## FlowStep extension (action kind)

```go
const KindAction StepKind = "action"

// ActionConfig is embedded in FlowStep for kind=action.
type ActionConfig struct {
    SourceTable           string            `json:"source_table"`
    DestinationConnector  string            `json:"destination_connector"`
    DestinationCredential string            `json:"destination_credential"`
    DestinationConfig     map[string]string `json:"destination_config,omitempty"`
    Action                string            `json:"action"` // upsert | create | delete
    Mappings              map[string]string `json:"mappings"`
    MaxRetries            int               `json:"max_retries,omitempty"` // default 3
    BatchSize             int               `json:"batch_size,omitempty"`  // default 100
}
```

## ActionRunner interface

```go
// ActionRunner executes a single action step with all safety invariants.
type ActionRunner interface {
    Plan(ctx context.Context, step FlowStep) (ActionPlan, error)
    Preview(ctx context.Context, plan ActionPlan) (ActionPreview, error)
    Execute(ctx context.Context, plan ActionPlan, token string) (ActionResult, error)
}
```

## ActionPlan / ActionPreview / ActionResult

```go
type ActionPlan struct {
    StepID        string
    FlowName      string
    Token         string    // approval token (shown once)
    TokenHash     string    // stored
    RecordCount   int
    SchemaDrift   bool
    DriftDetail   string
    CreatedAt     time.Time
    ExpiresAt     time.Time
}

type ActionPreview struct {
    Plan        ActionPlan
    Sample      []map[string]any // up to 3 redacted records
    RecordCount int
}

type ActionResult struct {
    RecordsAttempted int
    RecordsSucceeded int
    RecordsFailed    int
    DLQPath          string
    ReceiptIDs       []string
}
```

## Safety feature implementations

### 1. Idempotent writes (action.go)
- `deterministicRecordID(flowName, stepID, record) string` — SHA-256 of sorted JSON of
  primary-key fields. Written into the mapped record as `_pm_id`.
- Before write: check in-memory set of sent IDs for this run. Skip if already sent.

### 2. Identity mapping (action.go)
- File: `.polymetrics/state/identity_map.json` — `map[string]string` (pm_id → external_id).
- After successful write: store external_id returned by connector in the map.
- Before write: if pm_id already in map, skip (idempotency layer 2).

### 3. Dedupe/merge (action.go)
- Before writing batch: deduplicate on `email`, `domain`, or `external_id` fields if present.
- Keep last-seen record per dedup key.

### 4. Rate-limit / backoff (action.go)
- Inline exponential backoff: `base=500ms, max=30s, jitter=±25%`, cap at `MaxRetries`.
- Detect 429 via connector returning `*connsdk.HTTPError{Status: 429}` or sentinel.
- On 429: wait `Retry-After` header if present, else backoff; retry up to MaxRetries.

### 5. DLQ + bounded retries (action.go)
- Failed records (after MaxRetries exhausted) written to:
  `.polymetrics/dlq/<flow>/<step>/<run-id>.ndjson`
- Each line: `{"pm_id":"...","error":"...","attempts":N,"record":{...redacted...}}`
- `ActionResult.DLQPath` is set. `RecordsFailed` is incremented.

### 6. Schema-drift detection (action.go)
- Snapshot stored at `.polymetrics/state/schema_snap_<flow>_<step>.json`.
- At step start: fetch live catalog from destination connector via `Catalog()`.
- Compare field names and types against stored snapshot.
- Breaking change = field removed or type changed. On breaking change: return `ErrSchemaDrift`
  **before any write**.
- New additive fields are non-breaking (warning only).

### 7. Receipts/audit (action.go)
- After each successful batch write: call `ledger.Append` with:
  `{Mode:"action", Operation:"<flow>/<step>", Status:"receipt", ...}`.
- Record count and redacted sample go in `RecordsRead`/`RecordsWritten`.

## Approval flow (flow-level token)

Flow `plan` → dry-runs action steps → surfaces token in RunResult.
Flow `run --token <tok>` → validates token → executes all action steps.
`--per-action` flag → each action step gets its own token prompt (same mechanics, per step).

## Test strategy

All tests in `internal/flow/action_test.go` using `httptest.Server` as the fake destination.
The fake server tracks received record IDs, returns 429 on demand, returns 400 on permanent failure.

## Error sentinels (extend errors.go)

```go
var (
    ErrApprovalRequired = errors.New("flow: action step requires approval token")
    ErrSchemaDrift      = errors.New("flow: schema drift detected — step paused")
    ErrTokenExpired     = errors.New("flow: approval token has expired")
    ErrTokenInvalid     = errors.New("flow: approval token is invalid")
)
```
