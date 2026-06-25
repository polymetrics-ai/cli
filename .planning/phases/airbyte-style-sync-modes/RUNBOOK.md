# Runbook

## Verify

```bash
go test ./...
go vet ./...
go build ./cmd/pm
make smoke
./pm docs generate --dir docs/cli
./pm skills generate --dir docs/skills
```

## Rollback

This phase is local-file only. To roll back a test project, restore the previous binary and keep or remove `.polymetrics/warehouse/_pm_raw` depending on whether raw history should be retained.

## Recovery

If an overwrite run fails, retry the sync. The previous final JSONL table should remain readable because temp files are swapped only after success.

