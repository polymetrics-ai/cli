# Verification Checklist — Issue #4290

- [ ] Batch 4 source lock/map inventory equality is green.
- [ ] Batch 5 source lock/map inventory equality is green.
- [ ] Every map has valid six-class parity classification and DELETE coverage.
- [ ] Every unauthored operation is `declaration-pending`, not `foundation-gap`.
- [ ] `go run ./cmd/connectorgen validate` is green.
- [ ] `go run ./cmd/connectorgen surface-sync --check` is green.
- [ ] Connector-boundary, focused tests, generated-file, and documentation checks have recorded results.
- [ ] The PR base is read back from the GitHub API after opening.
