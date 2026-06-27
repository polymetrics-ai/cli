# RUNBOOK — Flow Engine (Phase 0)

## Prerequisites

- `pm init` has been run in the project root.
- At least one connection exists (`pm connections list`).
- Flow manifests live under `.polymetrics/flows/<name>.json` (or specify `--file`).

## Common operations

### Create and run a flow

```bash
# Write a manifest
cat > .polymetrics/flows/daily-sync.json <<'EOF'
{
  "version": 1,
  "name": "daily-sync",
  "steps": [
    {
      "id": "sync-hubspot",
      "kind": "sync",
      "connection": "hubspot-prod",
      "streams": ["contacts"],
      "out": ["contacts"]
    },
    {
      "id": "score-contacts",
      "kind": "query",
      "sql": "SELECT * FROM contacts WHERE email IS NOT NULL",
      "in": ["contacts"],
      "out": ["scored_contacts"]
    }
  ]
}
EOF

# Validate and plan (runs read-only steps)
pm flow plan --file .polymetrics/flows/daily-sync.json --json

# Run fully (same as plan in Phase 0 — no action steps yet)
pm flow run --file .polymetrics/flows/daily-sync.json --json
```

### Preview without executing

```bash
pm flow preview --file .polymetrics/flows/daily-sync.json --json
```

### Check status of last run

```bash
pm flow status daily-sync --json
```

### List all flows

```bash
pm flow list --json
```

### Force re-run (ignore checkpoints)

```bash
pm flow run --file .polymetrics/flows/daily-sync.json --force --json
```

## Troubleshooting

### "flow: another run is already in progress" (ErrLeaseHeld)

The lock file `.polymetrics/locks/flow-<name>.lock` exists.

1. Check if another `pm flow run` process is active: `ps aux | grep 'pm flow'`
2. If no process is running, the lock is stale (process crashed).
3. Remove the stale lock: `rm .polymetrics/locks/flow-<name>.lock`
4. Re-run the flow.

The engine automatically removes stale locks (PID check) on startup, but if this fails,
manual removal is safe.

### "flow: cyclic dependency detected"

Two or more steps form a cycle via their `in`/`out` table declarations.

1. Run `pm flow plan --file <manifest> --json` to see the error with step IDs.
2. Restructure the manifest to eliminate the cycle.

### "flow: manifest invalid"

1. Check the error message for the specific field.
2. Ensure `version: 1`, all step IDs are unique, `kind` is `sync` or `query`.
3. Verify every table in `in` is produced by some step's `out`.

### Step fails mid-flow

The flow stops at the failed step. Successful steps are checkpointed.

1. Inspect the error in the JSON output `steps[].error`.
2. Fix the underlying issue (connection, SQL, etc.).
3. Re-run with `pm flow run` — completed steps are skipped automatically.
4. Use `--force` to restart from scratch.

## Ledger location

`.polymetrics/logs/ledger.jsonl` — append-only, one JSON object per line.

## Lock file location

`.polymetrics/locks/flow-<name>.lock`

## Checkpoint file location

`.polymetrics/state/flow-checkpoints.json`
