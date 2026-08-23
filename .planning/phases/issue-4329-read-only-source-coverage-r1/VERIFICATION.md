# Verification — issue 4329

## Status

Implementation is green for the focused engine, source-projection, and
operation-evidence behavior tests. A full serial `make verify` pass completed
after the initial implementation checkpoint; the final evidence-rollup commit
is queued for the same full verification before push.

## Required final checks

- `GOFLAGS=-p=3 go test -count=1 -timeout 20m ./internal/connectors/engine -run '^TestParseAPISurfaceOperationModelEnumRemainsClosedWithReadOnly$'` — passed
- `GOFLAGS=-p=3 go test -count=1 -timeout 20m ./cmd/connectorgen -run '^(TestSourceProjectionRequiresExplicitReadOnlyNonMutationDeclaration|TestOperationEvidenceSeparatesDeclaredReadOnlyFromFoundations)$'` — passed
- `GOFLAGS=-p=3 go run ./cmd/connectorgen operation-evidence --check` — passed; 1,525 rows, 5 foundation rollups, fixed-100 passed
- `GOFLAGS=-p=3 go run ./cmd/connectorgen validate internal/connectors/defs/sentry` — passed; 1 connector, 0 findings
- `GOFLAGS=-p=3 go run ./cmd/connectorgen validate internal/connectors/defs/vercel` — passed; 1 connector, 0 findings
- `GOFLAGS=-p=3 go vet ./cmd/connectorgen ./internal/connectors/engine` — passed
- `GOFLAGS=-p=3 go build ./cmd/pm` — passed
- Frozen GitHub artifacts measured: source lock `3,420,025` bytes / `281b1cfcc67eb63e19ef83daf06197bf3d3b23db0b6bc9b73e02fc18ee278fb6`; descriptor `43,354,021` bytes / `d1978c0c6fd0eb66e9fcd4d78d637864a6e486f558aaad1e51550bc43758b899`
- `GOFLAGS=-p=3 make verify` — passed serially: format, tidy, vet, all tests, build, docs, smoke, lint, agent contract, connector validation/sync/evidence, GitHub parity artifacts, certification, boundary, canon, and release checks

The Sentry/Vercel source locks are intentionally no longer in the production
tree after the source-lock embed slimming work, so source-import checks for
them are not locally runnable on this branch. Their historical source evidence
is used only to hand off actual mutation findings; it is not restored or
modified by this issue.
