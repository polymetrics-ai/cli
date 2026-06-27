# DATA-MODEL — Action Step (Phase 1)

## New state files (all under .polymetrics/)

### identity_map.json
```json
{
  "<pm_id>": "<external_system_id>",
  "sha256-abc123": "ext-id-456"
}
```
Path: `.polymetrics/state/identity_map.json`
Backing: `statestore.JSONStore[map[string]string]` pattern (same as existing stores).

### schema_snap_<flow>_<step>.json
```json
{
  "flow": "myflow",
  "step": "send-emails",
  "connector": "sendgrid",
  "snapshot_at": "2026-06-27T10:00:00Z",
  "fields": [
    {"name": "to",      "type": "string"},
    {"name": "subject", "type": "string"},
    {"name": "body",    "type": "string"}
  ]
}
```
Path: `.polymetrics/state/schema_snap_<flow>_<step>.json`

### DLQ entries (.polymetrics/dlq/<flow>/<step>/<run-id>.ndjson)
Each line is:
```json
{
  "pm_id": "sha256-abc123",
  "error": "http 400 for ...: invalid email",
  "attempts": 3,
  "record": {"to": "[REDACTED]", "subject": "[REDACTED]"}
}
```
Values are redacted; only field names are stored.

## Extended StepResult (in-memory, serialized in RunResult)

```go
type StepResult struct {
    ID             string `json:"id"`
    Kind           string `json:"kind"`
    Status         string `json:"status"`
    RecordsRead    int    `json:"records_read"`
    RecordsWritten int    `json:"records_written"`
    RecordsFailed  int    `json:"records_failed,omitempty"`
    DurationNs     int64  `json:"duration_ns"`
    Error          string `json:"error,omitempty"`
    DLQPath        string `json:"dlq_path,omitempty"`
    SchemaDrift    bool   `json:"schema_drift,omitempty"`
}
```

## Ledger receipt record

Uses existing `LedgerRecord` shape from `internal/flow/engine.go`:
```go
LedgerRecord{
    Mode:      "action",
    Operation: "<flow>/<step>",
    Status:    "receipt",
    Error:     "",  // empty on success
}
```
`RecordsRead` = records attempted; `RecordsWritten` = records succeeded.
