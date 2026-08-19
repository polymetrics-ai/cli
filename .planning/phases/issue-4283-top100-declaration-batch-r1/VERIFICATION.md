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

## Docker Hub full-parity proof — secret-policy retrofit (2026-08-19)

- [x] The pinned 54-operation lock and 54-row `api_surface.json` retain an
  exact connector-local crosswalk; the raw artifact is
  `99d9d53c…53d0756` at 148322 bytes, matching the lock.
- [x] `operations.json` now has 50 source-backed contracts (23 `rest_read`, 27
  `rest_write`). The additional login operation is source-derived and all five
  credential-minting/exchange contracts declare `secret_sensitive` plus a
  redacting `sensitive_policy`.
- [x] `dockerhub-declaration-disposition.json` accounts for all 54 source rows:
  41 runnable command bindings and 13 disabled, recoverable rows. The disabled
  set is ten named `foundation-gap` rows and three schema/media rows;
  `unsafe-to-exercise` is zero. All six documented DELETE operations now have
  typed delete actions.
- [x] `go run ./cmd/connectorgen validate internal/connectors/defs/dockerhub --json` — 0 findings after generation.
- [x] `go run ./cmd/connectorgen surface-sync internal/connectors/defs --check` — 552 bundles, 0 changes.
- [x] `go run ./cmd/connectorgen certification-sweep . --connector dockerhub --check` — current (43 rows; 41 CLI commands).
- [x] `go test -timeout 20m ./internal/connectors/conformance -run '^TestConformance/dockerhub$' -count=1` — pass.
- [x] `go test -timeout 20m -run '^TestEveryImplementedCommandPassesRuntimePreflight$' ./internal/connectors/commandrunner` — pass.
- [x] Built binary no-credential proof — all 41 implemented Docker Hub commands
  stop at `error: missing --credential`; reports are retained under the task
  temporary proof directory and no provider request was made.
- [x] `go run ./cmd/pm docs generate --dir docs/cli` and
  `go run ./cmd/pm docs validate --connectors-dir docs/connectors` — pass.
- [ ] `go run ./cmd/connectorgen boundary . --json` — required check remains
  unverified locally. Two detached capture attempts were killed by the worker
  process boundary before producing stdout, stderr, or an exit record; CI still
  gates this exact check. No check was weakened or silently skipped.

## Complete six-class map checkpoint (before certification)

- [x] Docker Hub `3ee815c01` retained as the accepted source-lock map template;
  all nine other connectors now have both
  `sources/<connector>-operation-crosswalk.json` and
  `sources/<connector>-declaration-disposition.json`.
- [x] Local map-integrity assertion: each crosswalk and disposition has exactly
  its pinned source count, every operation has one primary class and a
  foundation record, class counts sum to the source count, and both
  definition-owned transport gaps are present.
- [x] `go run ./cmd/connectorgen validate internal/connectors/defs/notion --json`
  — 0 findings.
- [x] `go run ./cmd/connectorgen validate internal/connectors/defs/<connector>
  --json` for Stripe, Bitbucket, GitLab, CircleCI, Sentry, Vercel, Asana, and
  Jira — 0 findings for each.
- [x] `go run ./cmd/connectorgen surface-sync --check` — 552 connectors
  scanned, 0 fields filled or corrected.
- [x] No certification sweep, live credential test, or provider call was run
  before the complete map existed. Live certification remains pending for all
  ten connectors.
- [x] After the map gate: `go run ./cmd/connectorgen certification-sweep .
  --connector <connector> --check` passed for Docker Hub, Notion, Stripe,
  Bitbucket, GitLab, CircleCI, Sentry, Vercel, Asana, and Jira. The respective
  current row/CLI counts were 43/41, 51/49, 10/8, 7/5, 5/4, 2/0, 1/0, 2/0,
  251/249, and 592/590.
- [x] `go test -timeout 20m ./internal/connectors/conformance -run
  'TestConformance/(dockerhub|gitlab|jira|vercel|notion|stripe|bitbucket|circleci|sentry|asana)$'
  -count=1` — pass (3.861s), fixture-backed only.
- [x] `go test -timeout 20m -run '^TestEveryImplementedCommandPassesRuntimePreflight$'
  ./internal/connectors/commandrunner -count=1` — pass (5.123s).
- [ ] `go run ./cmd/connectorgen boundary . --json` was retried detached at
  `.tmp/connector-boundary-map.JaA7HT` after fleet disk recovery. The child
  again vanished before it wrote `result.txt`, stdout or stderr, so no exit
  status exists to claim. This is the third observed worker-containment
  failure; CI remains the required gate and no check was weakened.
