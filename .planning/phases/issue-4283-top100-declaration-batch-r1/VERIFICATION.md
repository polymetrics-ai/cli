# Issue #4283 — Verification Checklist

- [x] Source-lock file exists and parses for each increment-1 connector; `SOURCE-LOCK-VERIFICATION.json` confirms raw byte and SHA-256 agreement (10 / 10).
- [x] Source-lock operation inventory and `api_surface.json` method/path inventory are reconciled: 4,378 / 4,378 (100%).
- [x] `go run ./cmd/connectorgen validate` passes: 552 connectors, zero findings.
- [x] `go run ./cmd/connectorgen surface-sync --check` passes: zero fields filled or corrected.
- [x] `make connector-runtime-preflight`, `make connector-canon-check`, and `make connector-boundary` pass.
- [x] Fixture-backed conformance runs for the ten selected bundles pass via `go test -timeout 20m ./internal/connectors/conformance -run 'TestConformance/(dockerhub|gitlab|jira|vercel|notion|stripe|bitbucket|circleci|sentry|asana)$'`.
- [x] Generated non-live sweep artifacts were generated and byte-checked for every selected connector.
- [x] `make docs-check`, `make smoke-no-build`, `make agent-contract-check`, `make tidy-check`, `make lint`, and `go build ./cmd/pm` pass.
- [x] No provider credential is requested, read, printed, or stored.
- [x] Live certification is recorded `pending` for every connector.
- [x] Transport-parity blocker is explicit: 10 `sync_transport` entries are `foundation-gap` and `recoverable: true`; `TRANSPORT-GAP.md` has file-and-line evidence plus a smallest safe recovery. No GitHub-only transport evidence or destination contract was copied.

## Docker Hub full-parity proof (2026-08-19)

- [x] The pinned 54-operation lock and 54-row `api_surface.json` have an exact
  connector-local crosswalk with no source-only, surface-only, or duplicate
  identity.
- [x] `operations.json` materializes 49 source-backed inventory contracts: 23
  `rest_read`, 26 `rest_write`, and 6 typed delete contracts. They remain
  non-terminal; no command or provider call is introduced.
- [x] `dockerhub-declaration-disposition.json` has one disposition per pinned
  operation: four existing stream bindings; 46 elevated-scope disabled rows;
  three recoverable HEAD foundation gaps; and one source-deprecated disabled
  login row.
- [x] Docker Hub-only `connectorgen validate`, `surface-sync --check`, runtime
  preflight, source-lock byte check, fixture-backed conformance, generated
  docs/golden checks, and the repository verification gates are rerun after
  this materialization:
  - `go run ./cmd/connectorgen validate internal/connectors/defs/dockerhub --json` — 0 findings.
  - `go run ./cmd/connectorgen certification-sweep . --connector dockerhub --check` — current (5 rows; 4 CLI commands).
  - `go test -timeout 20m ./internal/connectors/conformance -run 'TestConformance/dockerhub$' -count=1` — pass.
  - raw source SHA-256/byte check — `99d9d53c…53d0756`, 148322 bytes, matching the lock.
  - `make connector-runtime-preflight`, `make connector-canon-check`, and `make connector-boundary` — pass.
  - `go run ./cmd/connectorgen surface-sync internal/connectors/defs --check` and `make connectorgen-surface-sync` — 552 bundles, 0 changes.
  - `go test -timeout 20m ./internal/cli`, `go test -timeout 20m ./cmd/connectorgen`, and `go test -timeout 20m ./internal/connectors/engine` — pass.
  - `go build ./cmd/pm`, `./pm docs validate --connectors-dir docs/connectors`, `make docs-check`, `make smoke-no-build`, `make lint`, `go vet ./...`, `make tidy-check`, and `make agent-contract-check` — pass.
  - `make github-parity-artifacts-check`, `make connectorgen-certification-matrix`, `make connectorgen-certification-candidates`, `make connectorgen-certification-sweep`, and `make release-workflow-check` — pass.
