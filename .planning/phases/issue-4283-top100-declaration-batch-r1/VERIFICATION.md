# Issue #4283 — Verification Checklist

- [x] Source-lock file exists and parses for each increment-1 connector; `SOURCE-LOCK-VERIFICATION.json` confirms raw byte and SHA-256 agreement (10 / 10).
- [x] Source-lock operation inventory and `api_surface.json` method/path
  inventory are reconciled: 4,378 operations found and 4,378 mapped rows.
- [x] Every Batch-1 source lock records `counts.total`, per-method counts and
  an equal-sized operation inventory from a complete provider-published OpenAPI
  document; each map records high machine-readable-spec input confidence.
- [x] `go run ./cmd/connectorgen validate` passes: 552 connectors, zero findings.
- [x] `go run ./cmd/connectorgen surface-sync --check` passes: zero fields filled or corrected.
- [x] `make connector-runtime-preflight` and `make connector-canon-check` pass.
- [ ] `connector-boundary` is required and CI-verified: the detached local
  capture was terminated by the worker boundary before it produced an exit
  record, so it is not claimed as locally passed.
- [x] Fixture-backed conformance runs for the ten selected bundles pass via `go test -timeout 20m ./internal/connectors/conformance -run 'TestConformance/(dockerhub|gitlab|jira|vercel|notion|stripe|bitbucket|circleci|sentry|asana)$'`.
- [x] Generated non-live sweep artifacts were generated and byte-checked for every selected connector.
- [x] `make docs-check`, `make smoke-no-build`, `make agent-contract-check`, `make tidy-check`, `make lint`, and `go build ./cmd/pm` pass.
- [x] No provider credential is requested, read, printed, or stored.
- [x] Live certification is recorded `pending` for every connector.
- [x] All ten source-only `sync_transport.json` files validate and App opens
  through definition-owned source composition. The ten reverse legs are
  explicit, recoverable `generic-typed-destination-executor` gaps; no GitHub
  evidence, destination action, or generic writer was copied.

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
  foundation record, class counts sum to the source count, and the current
  source-declared/reverse-destination-gap transport disposition is present.
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
  current row/CLI counts after source-transport generation are 44/41, 52/49,
  11/8, 8/5, 6/4, 3/0, 2/0, 3/0, 252/249, and 593/590.
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

## Definition-owned ETL source retrofit (PR #4286)

- [x] Added source-only `sync_transport.json` for Docker Hub, Notion, Stripe,
  Bitbucket, GitLab, CircleCI, Sentry, Vercel, Asana, and Jira. Each has a
  concrete stream allowlist matching `streams.json`, the exact registered
  `declarative_stream_source` executor, and unique evidence reference.
- [x] `go run ./cmd/connectorgen validate internal/connectors/defs --json` —
  552 connectors, zero findings.
- [x] `go run ./cmd/connectorgen surface-sync --check` — 552 connectors, zero
  filled or corrected fields.
- [x] `go test -timeout 20m ./internal/app -run
  '^TestOpenRegistersDefinitionOwnedProductionTransports$' -count=1` — pass
  (2.487s), proving production App composition accepts the declarations.
- [x] `generic-typed-destination-executor` is recorded for all ten reverse
  legs with the exact `issue_label_destination` factory evidence and the
  definition-selected typed-factory recovery. Live certification remains
  pending; no provider request was made.

## Vocabulary correction

- [x] Map-integrity assertion confirms every row in all ten maps carries
  `method`, `path`, `parity_class`, `api_surface`, `source`, `state`,
  `foundation`, and `rejection`; class totals equal each pinned denominator and
  all six classes are present across the cohort.
- [x] The assertion finds exactly five `foundation-gap` rows: three Docker Hub
  HEAD rows and two Docker Hub operation-scoped pagination rows, each with its
  engine refusal file/line. The 3,889 disabled rows in the nine newly mapped
  connectors use `declaration-pending`, not `foundation-gap`; Docker Hub is
  normalized to the same six-class row shape.

## Classification correction — direct-write endpoint taxonomy

- [x] Map-integrity assertion confirms exactly five primary endpoint classes
  (`direct_read`, `direct_write`, `etl`, `binary_read`, `binary_write`) and a
  separate `reverse_etl_eligibility` attribute on every direct-write row.
- [x] The correction reclassifies 250 rows: the cohort now records 2,370
  direct-write endpoints and 118 enabled direct-write bindings; reverse-ETL
  eligibility is zero.
- [x] Every zero-eligible attribute cites the recoverable
  `generic-typed-destination-executor` gap and
  `internal/app/issue_label_warehouse_transport.go:85-95`; no destination
  descriptor, acknowledgement, apply strategy, or `transport_binding` action
  was invented.

## Website generated-data follow-up

- [x] PR #4294's `Website Data` CI check identified stale repository-owned
  output after the source-backed connector declarations changed.
- [x] `cd website && pnpm run gen:website-data` regenerated
  `website/data/connectors.generated.json`,
  `website/lib/connectors.catalog.data.generated.json`, and
  `website/lib/connectors.catalog.generated.ts`.
- [x] The generator was run a second time with no further generated-file drift.

## Provider surface reconciliation

- [x] Compared each Batch-1 `api_surface.json` method/path set to the complete
  pinned provider OpenAPI method/path set, rather than treating the old surface
  as the completeness boundary.
- [x] `API-SURFACE-REALITY-AUDIT.json` records the old count, provider count,
  new count and basis for every connector: 4,378 provider operations found;
  zero missing from `api_surface.json`; zero surfaces regenerated.
- [x] Notion (6), Sentry (1), and Vercel (22) retain explicitly described
  connector-specific or legacy entries beyond their current OpenAPI count; none
  masks a missing provider operation. No source is instance-dependent.

## CI verify regression — shared source-evidence assertion

- [x] Retrieved the failed GitHub `Verify` log for run `32271368383`; its sole
  test failure was `TestDefinitionTransportFactoriesSelectDeclaredEvidence`.
- [x] The assertion now accepts the GitHub declaration evidence in either
  `SourceEvidence` or `AcceptedSourceEvidences`, matching the production
  conformance verifier's contract and avoiding registry-order coupling.
- [x] `go test -timeout 20m ./internal/app -run
  '^TestDefinitionTransportFactoriesSelectDeclaredEvidence$' -count=1` — pass
  (2.428s).
- [x] `go test -timeout 20m ./internal/app -count=1` — pass (238.304s).
- [x] `go run ./cmd/connectorgen validate internal/connectors/defs --json` —
  552 connectors, zero findings.
- [x] `go run ./cmd/connectorgen surface-sync --check` — 552 connectors,
  zero fields filled or corrected.
- [x] `make docs-check` initially found stale generated connector artifacts;
  `go run ./cmd/pm docs generate --dir docs/cli` regenerated the ten affected
  connector manuals/skills and `docs/connectors/catalog/all-connectors.json`.
  The subsequent `make docs-check` passes.
- [x] `make tidy-check`, `make lint`, `make agent-contract-check`, `make
  smoke-no-build`, `make connector-runtime-preflight`, and `make
  connector-canon-check` pass. `make release-workflow-check` passes.
- [x] A read-only local reconciliation independently checks each pinned lock's
  total/per-method/inventory counts and its source method/path set against its
  `api_surface.json`: 4,378 found, zero missing.

## Reconciliation slice 1 — declaration-owned typed destination proof (2026-08-20)

- [x] Merged the then-current `fm/cli-reverse-etl-destination-r1` stack at
  `192180675`; PR #4294 is retargeted to that stack. Its later persisted
  App/CLI dispatch work remains an explicit pre-push dependency.
- [x] The ten-row machine-readable seven-surface ledger records a disposition
  for every typed action, including a stable action-set SHA-256 selector,
  source mapping, acknowledgement/delivery facts, and a semantic exclusion or
  exact foundation dependency where the current closed contract cannot bind it.
- [x] A mechanical per-connector comparison of `writes.json` action names to
  the ledger's named eligibility rows reports 491 / 491 action names, with
  zero missing or extra. Four are fixture-bound; the remaining 487 remain
  eligible and carry a specific mapping/foundation dependency rather than a
  safety, privilege, deletion, price, or credential exclusion.
- [x] Definition-owned static destination mappings validate for Notion
  `views -> update_view`, Stripe `customers -> update_customer`, CircleCI
  `schedules -> update_schedule`, and Vercel `projects -> update_project`.
  This is structural/fixture evidence only, not a provider-live deployment
  claim.
- [x] `go run ./cmd/connectorgen validate internal/connectors/defs/<each of
  the ten connectors> --json` — zero findings for each connector.
- [x] `go run ./cmd/connectorgen surface-sync internal/connectors/defs --check`
  — 552 connectors scanned; zero fields filled or corrected.
- [x] `go test -count=1 -timeout 20m ./internal/connectors/commandrunner -run
  'TestEveryImplementedCommandPassesRuntimePreflight'` — pass.
- [x] `go test -count=1 -timeout 20m ./internal/app -run
  '^TestDefinitionTransportFactoriesRunTypedDestinationFromDefinition$'` —
  pass.
- [x] Generated CircleCI and Vercel certification sweeps are current (six
  rows/two CLI commands each); both have no `certification.json`, so a
  fixture-backed certification candidate cannot be invoked and is recorded as
  not applicable rather than failed.
- [x] An isolated built binary reached credential preflight for the four new
  CircleCI/Vercel list/update commands, each returning exactly
  `error: missing --credential` without provider I/O.
- [x] `make docs-check` passes after generated manuals, skills, and the
  connector catalog were regenerated.
- [ ] Complete documented-operation command reachability, the latest #4304
  persisted App/CLI dispatch proof, `connector-boundary`, full `make verify`,
  and provider-live certification remain outstanding. Live certification is
  intentionally pending; no credentials are authorized.
- [ ] The source-crosswalk audit finds 3,366 documented endpoints without a
  declared command binding. The required closed disabled-operation command
  target is recorded as `declaration-bound-disabled-command-surface` with
  exact refusing code in the foundation-gap log. It requires a keyed shared
  foundation decision and is not implemented in this connector-local lane.
