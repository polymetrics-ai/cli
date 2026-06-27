# API-CONTRACT — Flow Engine (Phase 0)

All `--json` outputs are machine-stable contracts. Do not change field names or remove fields
without a version bump.

## pm flow plan / pm flow run

Exit 0 on success, non-zero on error.

### Success response

```json
{
  "flow_name": "daily-sync",
  "status": "ok",
  "steps": [
    {
      "id": "sync-hubspot",
      "kind": "sync",
      "status": "ok",
      "records_read": 1423,
      "records_written": 1423,
      "duration_ns": 83000000
    },
    {
      "id": "score-contacts",
      "kind": "query",
      "status": "ok",
      "records_read": 1423,
      "records_written": 891,
      "duration_ns": 12000000
    }
  ]
}
```

### Step skipped (checkpoint hit)

```json
{
  "id": "sync-hubspot",
  "kind": "sync",
  "status": "skipped",
  "records_read": 0,
  "records_written": 0,
  "duration_ns": 0
}
```

### Error response (exit non-zero)

```json
{
  "error": "flow: step failed",
  "step_id": "sync-hubspot",
  "detail": "connection hubspot-prod not found"
}
```

## pm flow preview

```json
{
  "flow_name": "daily-sync",
  "status": "dry_run",
  "steps": [
    {
      "id": "sync-hubspot",
      "kind": "sync",
      "status": "dry_run",
      "records_read": 0,
      "records_written": 0,
      "duration_ns": 0
    }
  ]
}
```

## pm flow list

```json
{
  "flows": [
    {
      "name": "daily-sync",
      "path": ".polymetrics/flows/daily-sync.json",
      "steps": 2
    }
  ]
}
```

## pm flow status

```json
{
  "flow_name": "daily-sync",
  "last_run": {
    "status": "ok",
    "started_at": "2026-06-27T10:00:00Z",
    "finished_at": "2026-06-27T10:01:23Z",
    "steps": [...]
  }
}
```

If no run exists:

```json
{
  "flow_name": "daily-sync",
  "last_run": null
}
```

## Error envelope (all commands, non-zero exit)

```json
{
  "error": "<sentinel error message>",
  "detail": "<optional human-readable context>"
}
```

## Stability policy

- All fields above are stable from Phase 0 onward.
- New fields may be added in later phases (additive).
- Existing fields will not be renamed or removed without a major version indicator.
