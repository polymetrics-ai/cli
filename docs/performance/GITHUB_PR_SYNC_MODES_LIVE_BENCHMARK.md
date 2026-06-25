# GitHub PR Sync Modes Live Benchmark

Date: 2026-06-25

Repository: `rails/rails`

Stream: `github.pull_requests`

Connector path: real `github` connector over GitHub REST API to local JSONL warehouse.

Auth: no PAT used; public repository read.

Config:

- `per_page=30`
- `max_pages=2`
- `state=all`
- `sort=updated`
- `direction=desc`
- batch size: `15`

## Two-Run Semantic Check

Each sync mode ran twice against its own warehouse table.

| Mode | First Read | First Loaded | First Time | Second Read | Second Loaded | Second Time | Final Rows | Result |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| `full_refresh_append` | 60 | 60 | 1409.64 ms | 60 | 60 | 210.46 ms | 120 | passed; append duplicates intentionally |
| `full_refresh_overwrite` | 60 | 60 | 208.84 ms | 60 | 60 | 197.08 ms | 60 | passed; final table replaced |
| `full_refresh_overwrite_deduped` | 60 | 60 | 244.17 ms | 60 | 60 | 235.62 ms | 60 | passed; final table deduped/replaced |
| `incremental_append` | 60 | 60 | 188.03 ms | 60 | 1 | 208.95 ms | 61 | passed; inclusive cursor row appended |
| `incremental_append_deduped` | 60 | 60 | 220.06 ms | 60 | 60 | 225.55 ms | 60 | passed; final table stayed deduped |

Notes:

- `incremental_append` second run loaded one record because PM uses inclusive cursor semantics for retry safety.
- `incremental_append_deduped` reports `loaded` as final materialized row count after dedupe. Its second run accepted only the inclusive cursor batch, then rematerialized 60 final rows.
- GitHub stream defaults were applied: `primary_key=[node_id]`, `cursor_field=updated_at`.

## Resource Benchmark

Fresh project. One ETL run per mode wrapped with `/usr/bin/time -l`.

| Mode | Read | Loaded | Batches | Final Rows | Real Time | Records/sec | Max RSS |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| `full_refresh_append` | 60 | 60 | 4 | 60 | 0.20 s | 300.00 | 26.50 MB |
| `full_refresh_overwrite` | 60 | 60 | 4 | 60 | 0.19 s | 315.79 | 26.08 MB |
| `full_refresh_overwrite_deduped` | 60 | 60 | 4 | 60 | 0.19 s | 315.79 | 26.34 MB |
| `incremental_append` | 60 | 60 | 4 | 60 | 0.19 s | 315.79 | 25.77 MB |
| `incremental_append_deduped` | 60 | 60 | 4 | 60 | 0.19 s | 315.79 | 26.23 MB |

## Verification

Commands run:

```bash
go build -o pm ./cmd/pm
go test ./internal/app -run 'TestGithubPullRequestsETLSupportsAllSyncModes|Test.*Sync'
go test ./...
```

Result: passed.

