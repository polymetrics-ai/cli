# Issue #4292 — verification checklist

## Planned checks

- [ ] Per-batch red map-integrity assertion fails before its source artifacts
  are added, then passes after the complete batch map is present.
- [ ] JSON integrity assertion: source lock, crosswalk, and disposition IDs
  agree exactly; every row has one class and class totals agree; every
  reverse-ETL disposition uses the locked generic-destination foundation gap.
- [ ] `go run ./cmd/connectorgen validate internal/connectors/defs/<connector> --json`
  for each changed connector.
- [ ] `go run ./cmd/connectorgen surface-sync internal/connectors/defs --check`.
- [ ] Targeted Go tests identified from the validation/generator code, with
  `-timeout 20m`.
- [ ] Repository generated-file/snapshot checks applicable to changed bundle
  metadata.
- [ ] `go run ./cmd/connectorgen boundary . --json` in detached capture,
  polling to a recorded exit result.
- [ ] Standard review of the final changed files and an automated review route
  recorded in the PR body.

## Results

Pending execution.
