# Verification — #3987 four-path warehouse conformance

## Automated evidence

| Check | Result |
| --- | --- |
| `go test -count=1 -timeout 20m ./internal/app ./internal/synctransport -run 'TestWarehouseMediated\|TestOrchestratorDispatchesFourClosedPairingsWithoutPairBranches'` | Pass — exact production composition and four-direction sealed-workset topology. |
| `go test -count=1 -timeout 20m ./internal/app ./internal/synctransport` | Pass. |
| `go test -count=1 -timeout 20m ./internal/connectors ./internal/synccontract` | Pass. |
| `go test -count=1 -timeout 20m ./internal/cli` | Pass. |
| `go test -count=1 -timeout 20m ./cmd/connectorgen` | Pass — required consumer package gate. |
| `go vet ./internal/app ./internal/synctransport ./internal/connectors ./internal/synccontract` | Pass. |
| `go build ./cmd/pm` | Pass. |
| `go run ./cmd/agentcontractgen check` | Pass. |
| `go run ./cmd/connectorgen validate internal/connectors/defs` | Pass before and after the scratch proof. |
| `go run ./cmd/connectorgen surface-sync --check` | Pass. |
| `go run ./cmd/connectorgen certification-matrix --check` | Pass. |
| `make tidy-check lint agent-contract-check connectorgen-validate connectorgen-surface-sync github-parity-artifacts-check connectorgen-certification-matrix connector-boundary connector-runtime-preflight connector-canon-check release-workflow-check` | Pass. |
| `make docs-check-no-build smoke-no-build` | Pass. |
| `pnpm --dir website run gen:docs` twice, each followed by `git diff --exit-code` | Pass and byte-stable. |

The full `go test -timeout 20m ./...` and `make verify` aggregate targets were not run as single local commands. Repository policy says their duration is routinely cut off under the command harness; the changed packages, their consumers, `internal/cli`, and each non-test `make verify` gate above were run independently. CI remains the full-suite authority.

## Scratch failure proof

After a passing `connectorgen validate`, the test temporarily changed only the GitHub API destination's admitted source stream from `issues` to `pull_requests`. Validation remained green, but the named `api_to_api` subtest failed before endpoint I/O because the destination no longer admitted the source executor for `issues`. The exact binding was restored through `apply_patch`; validation and the named test then passed again. This proves the matrix detects a schema-valid declaration defect in its own direction rather than relying on a case count.

## Review and UAT

- Inline/manual GSD fallback is recorded because canonical delivery forbids spawning a role executor in this task environment.
- Manual code review is recorded in `REVIEW.md`; no production behavior or certification roll-up was altered.
- The deterministic suite covers the transport topology. Existing fresh-binary/live route proofs remain the evidence for external GitHub/PostgreSQL behavior and are not reclassified as a new live certification here.

Status: complete. This record distinguishes deterministic production-composition conformance from existing opt-in fresh-binary/live route proofs and does not promote either to #3978’s final certification.
