# RUNBOOK — Action Step (Phase 1)

## Typical action step workflow

```
# 1. Plan (dry-run; generates approval token)
pm flow plan myflow.json --json
# → {"flow_name":"myflow","status":"planned","approval_token":"<tok>","steps":[...]}

# 2. Preview (shows records + redacted sample; consumes no writes)
pm flow preview myflow.json --json

# 3. Execute (requires token from step 1)
pm flow run myflow.json --token <tok> --json
# → {"flow_name":"myflow","status":"ok","steps":[...]}
```

## Per-action approval (opt-in)

```
pm flow run myflow.json --per-action --json
# Each action step generates its own token; executor prompts for each.
```

## Checking the DLQ

Failed records land in `.polymetrics/dlq/<flow>/<step>/<run-id>.ndjson`.

```
# List DLQ files
ls .polymetrics/dlq/

# Inspect a DLQ file
cat .polymetrics/dlq/myflow/send-emails/run-abc123.ndjson | jq .

# Retry DLQ records (future: pm flow dlq retry)
# For now: fix the root cause, then re-run the flow (idempotency ensures safe retry).
```

## Checking the identity map

```
cat .polymetrics/state/identity_map.json | jq .
# Keys are pm_ids (deterministic); values are external system IDs.
```

## Schema snapshot management

Snapshots live at `.polymetrics/state/schema_snap_<flow>_<step>.json`.
If a legitimate schema change breaks a flow:

```
# 1. Update the destination schema to be compatible, OR
# 2. Delete the snapshot to force a fresh baseline on next run:
rm .polymetrics/state/schema_snap_myflow_send-emails.json
# 3. Re-run plan → preview → approve → execute
```

## DLQ cleanup

```
# Remove DLQ files older than 30 days
find .polymetrics/dlq -name "*.ndjson" -mtime +30 -delete
```

## Observability

Ledger receipts: `.polymetrics/logs/ledger.ndjson` (JSONLedger path).

```
# Count action receipts
grep '"mode":"action"' .polymetrics/logs/ledger.ndjson | jq .
```

## Error codes

| Error | Meaning | Remedy |
|-------|---------|--------|
| `flow: action step requires approval token` | No --token provided | Run `pm flow plan` first to get token |
| `flow: schema drift detected — step paused` | Breaking schema change | Delete snapshot; update schema; re-plan |
| `flow: approval token has expired` | Token > 24h old | Re-run `pm flow plan` |
| `flow: approval token is invalid` | Wrong token | Re-run `pm flow plan` |
