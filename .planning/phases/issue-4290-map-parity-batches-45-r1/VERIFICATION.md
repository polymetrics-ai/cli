# Verification Checklist — Issue #4290

- [x] Batch 4 source lock/map inventory equality is green (`materialize-parity-maps.mjs check batch4`).
- [ ] Batch 5 source lock/map inventory equality is green.
- [ ] Every map has valid six-class parity classification and DELETE coverage.
- [ ] Every unauthored operation is `declaration-pending`, not `foundation-gap`.
- [x] `go run ./cmd/connectorgen validate` is green (552 connectors, 0 findings; Batch 4 checkpoint).
- [x] `go run ./cmd/connectorgen surface-sync --check` is green (552 connectors, 0 fields corrected; Batch 4 checkpoint).
- [ ] Connector-boundary, focused tests, generated-file, and documentation checks have recorded results.
- [ ] The PR base is read back from the GitHub API after opening.
