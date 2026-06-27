# RUNBOOK — Phase 2: RLM Deterministic Backend

## Running RLM manually

### Prerequisites
- `pm init` has been run in the project root (creates `~/.polymetrics/<project>/`).
- For deterministic mode: an InTable NDJSON file exists in the warehouse directory.
- For fixture mode: no prerequisites — fixture rows are hardcoded.

### Score a table (deterministic mode)

```bash
pm rlm run \
  --spec path/to/spec.json \
  --in contacts \
  --out lead_scores \
  --mode deterministic
```

Expected output (human-readable):
```
rlm: mode=deterministic in=contacts out=lead_scores records_read=42 records_scored=42 records_failed=0 duration=1.2ms
```

### Score without writing (dry run)

```bash
pm rlm run --spec spec.json --in contacts --out lead_scores --mode deterministic --dry-run
```

No OutTable file is created. The JSON envelope reports `dry_run: true`.

### Use fixture mode (no InTable needed)

```bash
pm rlm run --spec spec.json --out lead_scores --mode fixture
```

Useful for pipeline testing, CI, and demos without real data.

### Machine-readable output

```bash
pm rlm run --spec spec.json --in contacts --out lead_scores --mode deterministic --json
```

Outputs a single JSON object to stdout:
```json
{"mode":"deterministic","in_table":"contacts","out_table":"lead_scores","records_read":42,"records_scored":42,"records_failed":0,"duration_ns":1234567,"dry_run":false}
```

---

## Spec file format (JSON)

```json
{
  "name": "likely-customers",
  "description": "Score contacts by engagement signals",
  "features": [
    {"name": "email",   "weight": 0.3, "score_if_set": 1.0, "default": 0.0},
    {"name": "company", "weight": 0.4, "score_if_set": 1.0, "default": 0.0},
    {"name": "title",   "weight": 0.3, "score_if_set": 1.0, "default": 0.0}
  ]
}
```

`score_if_gt` example:
```json
{"name": "followers_count", "weight": 0.5, "score_if_gt": 1.0, "threshold": 100, "default": 0.0}
```

---

## OutTable format

Each row in `lead_scores.ndjson`:
```json
{"_rlm_score":0.87,"_rlm_mode":"deterministic","_rlm_spec":"likely-customers","_rlm_scored_at":"2026-06-27T10:00:00Z","email":"alice@example.com","company":"Acme","title":"CTO"}
```

Rows are sorted by `_rlm_score` descending, then by `_polymetrics_raw_id` ascending.

---

## In a flow manifest

```yaml
steps:
  - id: sync_contacts
    kind: sync
    connection: hubspot
    stream: contacts
    out: contacts

  - id: score_contacts
    kind: rlm
    spec: specs/likely_customers.json
    in: contacts
    out: lead_scores
    mode: deterministic
    depends_on: [sync_contacts]
```

---

## Troubleshooting

### "table name contains invalid characters"
Table names must be bare identifiers: `[a-zA-Z0-9_-]+`. No slashes, dots, or path separators.

### "records_failed > 0 in result"
Some InTable rows were malformed JSON. Check stderr for `[rlm] parse error line=N`. The run still completes; failed rows are skipped. Inspect the source table with `pm query --table contacts`.

### "rlm: model backend not implemented"
`--mode model` is a stub. Phase 4 has not been implemented. Use `--mode deterministic` or `--mode fixture`.

### OutTable appears to have fewer rows than InTable
Records with a `json.Unmarshal` error are counted in `records_failed` and excluded from OutTable. Also check that InTable has the `record` field (it should if produced by `pm etl run`).

### OutTable already exists — will it be overwritten?
Yes, `pm rlm run` overwrites OutTable on each run (atomic rename). Back it up first if needed. The flow engine (Phase 0) handles idempotency at the pipeline level.

---

## Development: running tests

```bash
cd /path/to/polymetrics-v2
go test ./internal/rlm/... -v
go test ./internal/cli/... -run TestRLM -v
```

## Verification gate

```bash
export GOTOOLCHAIN=auto
gofmt -w internal/rlm internal/cli
go vet ./internal/rlm/... ./internal/cli/...
go test ./internal/rlm/... ./internal/cli/...
go build ./cmd/pm
make verify
```
