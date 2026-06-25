# Runbook

## Verify

```bash
go test ./...
go build ./cmd/pm
./pm docs validate --connectors-dir docs/connectors
make verify
```

## Smoke Native Connector

```bash
pm init --root /tmp/pm-native
pm etl check --connector source-stripe --root /tmp/pm-native --json
pm etl catalog --connector source-stripe --root /tmp/pm-native --json
pm etl read --connector source-stripe --stream records --limit 1 --root /tmp/pm-native --json
```

## Rollback

Revert the native catalog adapter, catalog enablement derivation, registry registration, CLI direct ETL commands, and regenerated docs.
