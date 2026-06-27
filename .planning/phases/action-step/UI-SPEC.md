# UI-SPEC — Action Step (Phase 1)

## CLI only (no web UI in this phase)

### pm flow plan <manifest> [--json]

Human mode (stderr):
```
flow: planning myflow
  step sync-contacts (sync) — 500 records in contacts
  step send-emails (action) — 42 records → sendgrid [sg-prod]
    sample: {"to":"a***@example.com", "subject":"[REDACTED]"}
    sample: {"to":"b***@example.com", "subject":"[REDACTED]"}
    sample: {"to":"c***@example.com", "subject":"[REDACTED]"}
  approval_token: eyXXXXXXXXXXXXXXXXXXXXXXX
  (token valid for 24h; pass to: pm flow run --token <tok>)
```

JSON mode (stdout, --json):
```json
{"flow_name":"myflow","status":"planned","approval_token":"eyXXX","steps":[...]}
```

### pm flow run <manifest> --token <tok> [--json]

Human mode (stderr):
```
flow: running myflow
  [1/2] sync-contacts ... ok (500 records, 1.2s)
  [2/2] send-emails ... ok (42 sent, 0 failed, 0.8s)
flow: done (ok)
```

On schema drift:
```
flow: send-emails PAUSED — schema drift detected
  field "email": type changed string → integer (breaking change)
  to resume: delete .polymetrics/state/schema_snap_myflow_send-emails.json
             then re-run pm flow plan
flow: failed
```

On DLQ:
```
  [2/2] send-emails ... partial (40 sent, 2 failed → .polymetrics/dlq/myflow/send-emails/run-abc.ndjson)
```

### pm flow run <manifest> (no token, has action steps)

```
error: flow contains action steps that require approval
run: pm flow plan <manifest> to generate an approval token
exit 1
```

## Exit codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Error (approval missing, schema drift, step failure) |
| 2 | Lease held (another run in progress) |
