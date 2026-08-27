# Verification — Stripe provider-dialect tolerance foundation

## Focused and changed-package proof

- `go test -timeout 20m ./cmd/connectorgen -run '^(TestSourceImportStripeReferenceDepthOperationLocal|TestSourceImportRejectsUnsafeReferences|TestSourceReferenceResolverCachesNormalizedTargetsWithoutBypassingCounts)$' -count=1` — PASS (`ok polymetrics.ai/cmd/connectorgen 1.095s`).
- `go test -timeout 20m ./cmd/connectorgen -count=1` — PASS.
- `go test -timeout 20m ./internal/connectors/engine -count=1` — PASS (`ok polymetrics.ai/internal/connectors/engine 15.313s`).
- `go test -timeout 20m ./internal/connectors/commandrunner -count=1` — PASS.
- `go vet ./...` — PASS.
- `go build ./cmd/connectorgen` — PASS.
- `go build ./cmd/pm` — PASS.
- `git diff --check` — PASS.

The focused red command before the production change is recorded in
[`TDD-LEDGER.md`](TDD-LEDGER.md): it returned the connector-wide preflight
`reference depth limit exceeded` failure. The green test proves complete
Stripe GET/DELETE identity, exact nested projection, operation-local retained
failure, source-projection blocking, normalized pointer memoization, and
unchanged reference accounting.

## Repository verification gates

- `make tidy-check` — PASS.
- `make lint` — PASS (the initial run correctly reported the deleted global-preflight helpers as unused; removal of that dead architecture made the final run clean).
- `make docs-check-no-build` — PASS (`Validated connector docs in docs/connectors`).
- `make smoke-no-build` — PASS.
- `make agent-contract-check` — PASS.
- `make connectorgen-validate` — PASS (553 connectors, 0 findings).
- `make connectorgen-surface-sync` — PASS (553 scanned; 0 changes).
- `make connectorgen-declaration-admission` — PASS (1 connector, 1 source operation, 0 findings).
- `make connectorgen-operation-evidence` — PASS (1774 rows; 5 rollups; fixed-100 passed).
- `make github-parity-artifacts-check` — PASS.
- `make connectorgen-certification-subject` — PASS.
- `make connectorgen-certification-matrix` — PASS.
- `make connectorgen-certification-candidates` — PASS.
- `make connectorgen-certification-sweep` — PASS.
- `make connector-boundary` — PASS.
- `make connector-canon-check` — PASS.
- `make release-workflow-check` — PASS.

## Delivery constraints and deliberately inapplicable checks

- This foundation changes no `pm` command, flags, manual/help surface,
  connector definition, generated Stripe artifact, or website page. Help,
  website, and `pm stripe …` missing-credential evidence are therefore not
  applicable; recording any command as runnable would be false.
- No credentials, live provider requests, source downloads, certificates,
  hashes, delete actions, reverse-ETL actions, or generic provider-I/O escape
  hatch were used or added.
- The retained Stripe source artifact is absent from current `main`. The
  hermetic fixture records its known source URL and exact source operation
  identities; the historical artifact hash and follow-up request-contract
  blockers are documented in `CONTEXT.md`.
- Full `go test -timeout 20m ./...` and monolithic `make verify` are not run
  in this per-command worker because repository guidance says to run changed
  packages and each make gate independently; CI owns the full monolithic
  suite.
