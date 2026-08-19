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
