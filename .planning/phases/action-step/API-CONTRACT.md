# API-CONTRACT — Action Step (Phase 1)

## pm flow plan <manifest> --json

```json
{
  "flow_name": "myflow",
  "status": "planned",
  "approval_token": "<random-18-byte-hex>",
  "steps": [
    {"id": "sync-contacts", "kind": "sync", "status": "dry_run"},
    {"id": "send-emails",   "kind": "action", "status": "dry_run",
     "records_preview": 42, "sample": [...3 redacted records...]}
  ]
}
```

Token is present only in `plan` output and only once. Not stored in cleartext.

## pm flow preview <manifest> --json

Same shape as `plan` but `status` is `"preview"` and `sample` is populated.

## pm flow run <manifest> --token <tok> --json

```json
{
  "flow_name": "myflow",
  "status": "ok",
  "steps": [
    {"id": "sync-contacts", "kind": "sync",   "status": "ok",
     "records_read": 500, "records_written": 500, "duration_ns": 1234567},
    {"id": "send-emails",   "kind": "action", "status": "ok",
     "records_read": 42, "records_written": 40, "records_failed": 2,
     "dlq_path": ".polymetrics/dlq/myflow/send-emails/run-abc.ndjson",
     "duration_ns": 9876543}
  ]
}
```

## StepResult extension for action steps

```go
type StepResult struct {
    // existing fields ...
    RecordsFailed int    `json:"records_failed,omitempty"`
    DLQPath       string `json:"dlq_path,omitempty"`
    SchemaDrift   bool   `json:"schema_drift,omitempty"`
}
```

## Error envelope (non-zero exit)

```json
{"error": "flow: schema drift detected — step paused", "step": "send-emails"}
```

## FlowStep JSON (action kind)

```json
{
  "id": "send-emails",
  "kind": "action",
  "source_table": "lead_outreach",
  "destination_connector": "sendgrid",
  "destination_credential": "sg-prod",
  "action": "create",
  "mappings": {"to": "email", "subject": "subject_line", "body": "email_body"},
  "max_retries": 3,
  "batch_size": 50,
  "in": ["lead_outreach"],
  "out": []
}
```
