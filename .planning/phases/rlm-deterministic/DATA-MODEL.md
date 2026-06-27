# DATA-MODEL — Phase 2: RLM Deterministic Backend

## Warehouse storage

RLM uses the same local warehouse format as `internal/app` (local NDJSON files). No new storage format is introduced.

### InTable format (read)

File: `<warehouse_dir>/<in_table>.ndjson`
Each line is a `localRawRecord` (defined in `internal/app/local_warehouse.go`):

```json
{
  "_polymetrics_raw_id": "abc123",
  "_polymetrics_run_id": "run-456",
  "_polymetrics_sync_id": "conn-789",
  "_polymetrics_generation_id": 1,
  "_polymetrics_extracted_at": "2026-06-27T09:00:00Z",
  "_polymetrics_loaded_at": "2026-06-27T09:00:01Z",
  "_polymetrics_cursor": "",
  "_polymetrics_primary_key": "abc123",
  "_polymetrics_deleted": false,
  "record": {
    "email": "alice@example.com",
    "company": "Acme Corp",
    "title": "CTO",
    "followers_count": 250
  }
}
```

RLM reads the `record` field from each line. The `_polymetrics_*` fields are preserved in the output for lineage.

### OutTable format (write)

File: `<warehouse_dir>/<out_table>.ndjson`
Each line is a flat JSON object merging RLM metadata with the source `record` fields:

```json
{
  "_rlm_score": 0.87,
  "_rlm_mode": "deterministic",
  "_rlm_spec": "likely-customers",
  "_rlm_scored_at": "2026-06-27T10:00:00Z",
  "_polymetrics_raw_id": "abc123",
  "email": "alice@example.com",
  "company": "Acme Corp",
  "title": "CTO",
  "followers_count": 250
}
```

Ordering: rows sorted by `_rlm_score` DESC, then `_polymetrics_raw_id` ASC.

### Write atomicity

1. Write to `<out_table>.ndjson.tmp` (temp file in same directory).
2. `os.Rename(<tmp>, <out_table>.ndjson)` — atomic on POSIX.
3. On error during write: temp file removed; original OutTable unchanged.

---

## Ledger record

When a `LedgerAppender` is provided, one `LedgerRecord` is appended per run:

```json
{
  "id": "<run_uuid>",
  "mode": "rlm",
  "operation": "deterministic",
  "status": "success",
  "records_read": 42,
  "records_written": 42,
  "duration_ns": 1234567,
  "created_at": "2026-06-27T10:00:00Z"
}
```

`mode = "rlm"`, `operation = <backend mode string>`.
Maps to `ledger.RunRecord` (existing type). No schema migration required.

---

## Spec file (disk, authored by user)

File: any path, passed via `--spec` flag. JSON format. Not stored in warehouse.

```json
{
  "name": "likely-customers",
  "description": "optional",
  "features": [
    {
      "name": "email",
      "weight": 0.3,
      "score_if_set": 1.0,
      "default": 0.0
    }
  ]
}
```

Constraints:
- `name`: non-empty string, <= 256 chars, `[a-zA-Z0-9_-]` only (used as `_rlm_spec` field value).
- `features`: non-empty list, max 1000 items.
- `weight`: float64, must be >= 0.0 (negative weights rejected).
- `name` per feature: non-empty, <= 256 chars.
- `score_if_gt` requires `threshold` to be set (non-nil).

---

## In-memory scoring representation

During a run, records are held as `[]scoredRecord`:

```go
type scoredRecord struct {
    RawID  string         // from _polymetrics_raw_id for tie-breaking
    Score  float64        // normalized [0.0, 1.0]
    Fields map[string]any // merged: source record fields + _rlm_* fields
}
```

This is not persisted; it exists only during a single `Run` call.

---

## No new Postgres schema

This phase does not introduce any new database tables. The ledger already has `polymetrics_run_ledger` (managed by `internal/ledger`). If a Postgres ledger is in use, the existing table is sufficient.
