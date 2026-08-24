# Issue #4283 — Verification Checklist

> Current reachability supersession (2026-08-20): earlier “disabled” counts in
> this historical verification record are not an endpoint disposition. The
> authoritative status is `EXECUTABLE-OPERATION-CAPABILITY-AUDIT.json`: each
> of the 3,366 unbound source operations has an exact typed execution
> capability, and a command counts only after installed-CLI dispatch to real
> provider I/O. `BlockedCommandError` placeholders, scope, privilege, cost,
> destructive risk, and unavailable live credentials are not reachability
> exclusions.

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
- [x] The source-crosswalk audit reclassifies all 3,366 source operations
  without a binding by exact executable capability, source identity, method,
  path, source location, and existing rejection evidence. Its totals are 1,389
  fixed REST reads, 1,828 fixed REST writes, 120 bounded binary transfers, 10
  status registrations, and 19 provider contracts without bounded schemas.
- [ ] The closed executable-operation foundation slices in
  `EXECUTABLE-OPERATION-FOUNDATION-DESIGN.md` are implementation work: source
  artifact hash rehydration/import, typed headers, #4305 structured bodies,
  bounded binary/status/text, #4304 persisted dispatch, and connector command
  materialization. A disabled-command placeholder is expressly rejected.

## Executable-operation audit/design correction (2026-08-20)

- [x] JSON parsing, source-key uniqueness, source-lock/identity presence, and
  the two-way audit/rejection-ledger join pass: 4,378 total source operations,
  1,012 already bound, and 3,366 exact reclassifications, with no missing or
  duplicate record on either side.
- [x] The active rejection vocabulary contains zero `requires-elevated-scope`,
  `requires-paid-tier`, or `unsafe-to-exercise` exclusions. Their prior
  evidence, where any, remains historical context only.
- [x] `go run ./cmd/agentcontractgen check` — canonical delivery contract and
  registered projections are current.
- [x] `go run ./cmd/connectorgen validate internal/connectors/defs/<each of
  dockerhub,notion,stripe,bitbucket,gitlab,circleci,sentry,vercel,asana,jira>
  --json` — ten zero-finding runs.
- [x] `go run ./cmd/connectorgen surface-sync internal/connectors/defs --check`
  — 552 connectors scanned; zero fields filled or corrected.
- [x] `git diff --check` — clean. This correction changes planning/audit
  artifacts only; it deliberately makes no engine, command, generated-artifact,
  credential, or provider-I/O claim.

## Captain hard pre-merge gate (2026-08-20)

- [x] The canonical execution ledger records `merge_ready: false` and a
  blocked per-operation two-way source/ledger/generated-CLI/generated-website
  proof contract for ETL, reverse ETL, direct read, direct write, binary
  download, and binary upload.
- [ ] Every provider operation has the required source lock, semantic-surface
  mapping or provider-evidenced `N/A`, runtime fixture/conformance proof,
  output-preservation proof, and exact generated CLI/website drift proof.
- [ ] F0 source import, F2/F4 operation runtime, #4305 structured bodies, and
  the final #4304 persisted App/CLI dispatch/action-selection heads are
  published and integrated.
- [ ] Zoom, Twenty, and Gong have their separately authorized provider-live
  certification evidence. No other connector is claimed provider-live
  certified without credentials and accepted evidence.
- [ ] The integration branch has an executable CI validator for the fixed 100
  that emits schema-backed per-connector and aggregate verdicts. Its negative
  suite rejects the required source-hash, command-reachability, CLI/website
  drift, runtime-evidence, surface, binary-direction, disabled-operation, and
  non-secret-output failures. Planning JSON is not evidence that this check
  exists or passes.

## Missing-foundation mapping deliverable (2026-08-20)

- [x] `MISSING-FOUNDATION-DELIVERABLE.json` has eight deduplicated stable gap
  definitions and 3,366 source-operation plus 491 typed-action membership rows
  with source identity, surface, evidence, owner/lane, closure, and fanout
  join keys.
- [x] Every current-batch membership row is
  `open-foundation-gap-not-enabled`; it is neither a disabled callable
  operation nor an `N/A` surface, and it contributes zero merge-ready credit.
- [x] Twenty-eight unresolved typed-action source identities are explicit
  F0/F1 rows; no provider operation identity is inferred.
- [x] `jq empty MISSING-FOUNDATION-DELIVERABLE.json` and its structural
  membership/fanout query — eight unique IDs, zero dangling references, zero
  non-open rows, exact root-fanout agreement, and zero missing source/surface/
  evidence fields for the resolved source-operation/action records.
- [ ] The remaining 90 fixed-100 connectors emit equivalent rows and an
  integration validator produces the final portfolio verdict after all
  foundations publish.

## Main-base reconciliation verification — 2026-08-23

- [x] `git merge origin/main` completed at `8f2e6f298`; conflicts retained
  `origin/main` for shared foundation and generated catalog files, with the
  Batch-1 declaration cohort preserved.
- [x] PR #4294 was retargeted and API-read back with base `main` at
  `6410fe59c`.
- [x] **Red:** `go run ./cmd/connectorgen validate` reproduces the new
  source-lock/descriptor incompatibility without credentials.
- [x] **Diagnosed:** after the minimal source-lock schema migration,
  `go run ./cmd/connectorgen source-import dockerhub` reaches the exact
  pinned artifact and refuses its numeric YAML response key at
  `cmd/connectorgen/sourceimport.go:1305-1318`. The foundation gap and its
  recoverability are recorded as `source-import-yaml-scalar-key-normalization`.
- [x] **Separate Docker Hub diagnosis:**
  `go run ./cmd/connectorgen validate internal/connectors/defs/dockerhub`
  independently rejects SCIM operation IDs
  `dockerhub.post__v2_scim_2.0_users` and
  `dockerhub.put__v2_scim_2.0_users__id_`: their derived body schemas retain
  OpenAPI `example` annotations, which the engine compiler reports as
  `unknown keyword "example"` at `name.familyName`, `name.givenName`, and
  `schemas.items`. This is a schema-dialect/projection incompatibility, not
  the YAML mapping-key import failure; no derived body contract is hand-edited.
- [ ] **Green, remaining nine:** source-import, source projection, surface
  synchronization, sweeps, and focused validation continue for Notion,
  Stripe, Bitbucket, GitLab, CircleCI, Sentry, Vercel, Asana, and Jira.
- [ ] **Docker Hub pending shared foundation:** rerun its source import and
  derived projection only after the importer accepts the existing pinned
  artifact without changing its URL, SHA-256, byte count, or operation
  inventory; then reconcile the separate SCIM schema-dialect finding.

## Nine-connector public source-import sweep — 2026-08-23

- [x] **Sentry import:** `go run ./cmd/connectorgen source-import sentry
  --defs internal/connectors/defs --out
  internal/connectors/defs/sentry/sources/sentry-operation-descriptor.json`
  imported 223 locked operations without credentials.
- [x] **Sentry reachability red:**
  `go run ./cmd/connectorgen validate internal/connectors/defs/sentry` reports
  34 source operations with no executable action. The retained 223/223
  source/declaration denominator therefore does not establish current-main
  installed-binary reachability.
- [x] **Pinned-source drift preserved:** Notion, Bitbucket, GitLab, CircleCI,
  Vercel, and Jira each return `source-lock refresh required: fetched artifact
  does not match locked bytes and SHA-256`; no lock or denominator was changed.
- [x] **Independent parser reds:** Stripe rejects its public reference cycle at
  `#/components/schemas/file` from `GET /v1/account` response `200`; Asana
  rejects its numeric YAML response key at `/paths/~1access_requests/get/responses`.
- [ ] `make verify` is intentionally not started while these repeated
  source-import blockers make its `connectorgen validate` gate deterministically
  fail. No push is attempted before captain decides the source-lock refresh and
  shared-importer recovery path.

## e338cd301 source-lock refresh verification — 2026-08-23

- [x] Merged `origin/main` without rebasing at `c8cd27cb9`; fetched tip
  `e338cd30160be6d3e1b4eba21f74df1f580094ba` (`#4327`).
- [x] Fetched six cited public artifacts without credentials and verified the
  measured bytes/SHA-256 by applying them only in isolated copied bundles.
- [x] Ran production `connectorgen source-import` for Notion, Bitbucket,
  GitLab, CircleCI, Vercel, and Jira against those isolated refresh candidates.
  No candidate reaches a clean source/declaration projection. Exact per-
  connector evidence, artifact measurements, and refusing file/lines are in
  `PROGRESS-LEDGER.json#source_lock_refresh_e338cd301` and `TDD-LEDGER.md`.
- [x] Confirmed PR #4334 is still open, unmerged, and behind `main`; CircleCI
  has the expected secret-bearing context env-var action among its 27 missing
  actions, but the generic projection gap is broader.
- [x] Vercel fails before the known read-only coverage work: its public
  `/api-keys` POST response uses `patternProperties`, refused by the source
  importer at `cmd/connectorgen/sourceimport.go:4311-4315`.

## Current-main typed-destination declaration repair — 2026-08-24

- [x] **Red reproduced without credentials:** focused `internal/app` Parquet
  transport tests fail on the prior CircleCI, Notion, Stripe, and Vercel
  destination declarations before provider I/O with the exact action-owned
  source-binding refusal.
- [x] **Root cause established:** adding only the action identity reaches the
  next guard, a missing per-record batch declaration, then the provider
  idempotency/read-back admission. No provider header, delivery unit, or
  read-back operation was invented.
- [x] **Green:** the four invalid destination declarations were removed while
  their source transports, typed write actions, and installed direct commands
  remain declared. `go test -timeout 20m -count=1 -run
  'TestWarehouseMaterializesTablesAsParquet|TestQuerySQLAggregatesOverParquetTables|TestReverseETLReadsAParquetSourceTable'
  ./internal/app` passes.
- [x] `go test -timeout 20m -count=1 -run
  TestSampleOutboxWriteLifecycleAgainstRealCLI ./internal/connectors/certify`
  passes without credentials or provider I/O.
- [x] Credential-free installed-binary reachability is retained for the four
  affected direct write routes: in an initialized disposable project,
  `pm circleci schedules update --id sched_fixture_1`, `pm notion view update
  --view-id view_fixture_1`, `pm stripe customers update --id cus_fixture_1`,
  and `pm vercel projects update --id prj_fixture_1` each reaches the binary's
  credential boundary with `error: missing --credential`, never an unknown
  command or provider request.
- [x] `go build -o ./pm ./cmd/pm`, `./pm help docs`, `./pm docs generate --dir
  docs/cli --connectors-dir docs/connectors`, and `./pm docs validate
  --connectors-dir docs/connectors` pass; generated manuals and catalog match
  the source-only transport declarations.
- [x] Generated-artifact and local gates pass: `make tidy-check`, `make lint`,
  `go vet ./...`, `go build ./cmd/pm`, `make docs-check-no-build`, `make
  smoke-no-build`, `make connector-boundary`, `make connector-canon-check`,
  `make github-parity-artifacts-check`, all four
  `connectorgen-certification-*` checks, and `make release-workflow-check`.
  `certification-subject` and 5,903-row `operation-evidence.json` were
  regenerated and then passed their `--check` gates.
- [ ] `connectorgen validate` and `surface-sync --check` remain blocked by
  already-recorded source-projection foundations, not the repaired destination
  declarations: all ten batch connectors reach either a missing canonical
  descriptor or an existing source-derived gap; `surface-sync` stops at
  `asana`'s missing descriptor. No lock, descriptor, request schema, or
  provider contract was invented to make these gates appear green.
- [ ] The independent installed-binary preflight is correctly blocked only by
  Docker Hub's two SCIM user writes: their pinned-derived schemas retain the
  unsupported `example` OpenAPI keyword (`internal/connectors/engine/schema.go:165-168`).
  This is the existing body-schema dialect gap; their declarations and routes
  remain present pending its foundation repair.
- [ ] After merging `origin/main` at `27664370c` (`#4334`), the full merged
  package run exposes two additional shared-test failures. The fixed-100 test
  fixture copies only GitHub while its cohort now includes Asana; the
  declarative source-factory test compares GitHub's evidence to the first
  registered (Asana) shared factory. The production operation-evidence check
  itself passes. Exact recovery and refusing lines are logged in
  `FOUNDATION-GAPS.md`; neither is repaired by modifying locks, schemas, or
  connector-local transport facts.
- [ ] The merged CircleCI env-only-secret foundation does not by itself clear
  CircleCI source projection: its current validator still stops at the missing
  canonical descriptor. This is recorded as a wait for source-import/projection
  recovery, not a regression of #4334.

## Authorized shared-test repair — 2026-08-24

- [x] Firstmate authorized the bounded downstream repair after the fixed-100
  evidence proved an addition/selection change, not a GitHub surface removal.
  `internal/connectors/operation-evidence-fixed-100.json` remains unchanged.
- [x] `go test -timeout 20m -count=1 -run
  '^(TestOperationEvidenceFixed100RejectsEveryRegression|TestOperationEvidenceCheckRunsFixed100Gate)$'
  ./cmd/connectorgen` passes. The test removes
  `github.rest.issues/list-for-repo` from a disposable copied source lock and
  proves both direct fixed-cohort validation and CLI `--check` reject that
  exact source ID.
- [x] `go test -timeout 20m -count=1 -run
  '^TestDefinitionTransportFactoriesSelectDeclaredEvidence$' ./internal/app`
  passes. It requires GitHub's exact declared source conformance record in the
  shared factory's primary-plus-accepted set and retains the exact destination
  assertion.
- [ ] `go test -timeout 20m -count=1 ./cmd/connectorgen` was attempted but
  stopped in unrelated Freshservice fixture setup with `no space left on
  device`; the shared data volume reported 6.5 GiB free. No external temporary
  data was deleted. Full package and repository gates remain pending a usable
  filesystem.

## Generated skills and current-main SCIM comparison — 2026-08-24

- [x] `GOFLAGS=-p=3 go run ./cmd/pm skills generate --dir docs/skills --json`
  regenerated exactly the ten Batch-1 connector skills. The generator is
  metadata-only; it reads no credential or provider API.
- [x] `GOFLAGS=-p=3 go test -timeout 20m -count=1 -run
  '^TestSkillsGenerateMatchesTrackedSkills$' ./internal/cli` passes.
- [x] The exact command
  `GOFLAGS=-p=3 go test -timeout 20m -count=1 -run
  '^TestEveryImplementedCommandPassesRuntimePreflight$'
  ./internal/connectors/commandrunner` passes from a clean
  `origin/main` `3c394a0e` tree and fails from this branch for only Docker Hub
`scim user create` and `scim user update`, each on the unsupported
source-derived `example` schema keyword. The failure is therefore branch
specific and must be repaired by the required current-main merge, not marked
pre-existing.

## Post-main Docker Hub SCIM open-object disposition — 2026-08-24

- [x] Merge commit `f528b806d` contains current `origin/main` and removes the
  obsolete `example` keyword rejection. The exact no-credential preflight now
  identifies the next guard: both SCIM operations fail only because
  `requireClosedBoundedStructuredRESTBody` refuses a source-declared object
  without explicit `additionalProperties:false` at
  `internal/connectors/engine/structured_rest_body.go:1436-1444`.
- [x] Public-source confirmation: the locked Docker Hub artifact hashes to
  `99d9d53c2d93656a3c66d604885abd153dc5df285abc0ecb13802a3bc53d0756`, and
  its `scim_user_name`/`scim_user` object declarations at latest.yaml
  lines 3921-3959 have properties but no close-the-object keyword. No provider
  request or credential was used.
- [x] The source-faithful partial candidate was deliberately rejected, not
  shipped: it makes the structural preflight pass, but
  `connectorgen operation-evidence --check` then refuses immutable fixed-100
  execution evidence with `dockerhub.rest.post_/v2/scim/2.0/Users execution
  evidence regressed`. The unchanged implemented declarations still fail the
  structural preflight; this is an unresolved foundation, not a passing local
  repair.

## Authorized fixed-100 runtime-preflight correction — 2026-08-24

- [x] **Red:** `go test -timeout 20m -count=1 -run
  '^TestOperationEvidenceFixed100UsesRuntimePreflightForDockerHubSCIMWrites$'
  ./cmd/connectorgen` proves the metadata-only selector currently admits Docker
  Hub SCIM create/update despite production runtime preflight refusing both.
- [x] **Green:** the focused test passes after operation evidence invokes
  `commandrunner.Preflight`; both rows are runtime-disabled with
  `runtime_reachability` and the prospective cohort contains no SCIM row. It
  would contain Asana 33, Bitbucket 1, CircleCI 1, Docker Hub 23, GitHub 39,
  and Jira 3 rows.
- [x] **Focused static verification:** `go vet ./cmd/connectorgen` passes after
  the selector correction and its lazy connector-construction refactor.
- [ ] **Captain-gated regression:** `go test -timeout 20m -count=1 -run
  '^TestOperationEvidenceFixed100' ./cmd/connectorgen` correctly stops at the
  checked-in branch cohort's stale `dockerhub.rest.post_/v2/scim/2.0/Users`
  row. Generated artifact checks, connector boundary, and `make verify` remain
  held: do not write or commit a regenerated
  `operation-evidence-fixed-100.json` until the shipped-baseline decision.

## Authorized shipped-baseline restoration — 2026-08-24

- [x] Firstmate inbox `012.msg` restored
  `operation-evidence-fixed-100.json` byte-for-byte from `origin/main` in
  `4ad21d771`. Both bytes hash to
  `c0d600d323e7effb15c1e092dce6fb590193f23613b17a51917af79e0d74812f`;
  `git diff --exit-code origin/main -- internal/connectors/operation-evidence-fixed-100.json`
  is clean.
- [x] Current main was merged without rebase in `8b6abbf7b`; its binary-upload
  classification correction is retained alongside the local runtime-preflight
  selector correction.
- [x] Regenerated only `operation-evidence.json` with the ordinary projector,
  never `--write-fixed-100`. `go test -timeout 20m -count=1 -run
  '^TestOperationEvidence(Fixed100|ClassForCommand)' ./cmd/connectorgen`
  passes (28.264s), as do `go run ./cmd/connectorgen operation-evidence --check
  .`, `go vet ./cmd/connectorgen`, and `git diff --check`.
- [ ] A broader fixed cohort remains an additive, separately reviewed proposal;
  it must not replace the shipped all-GitHub evidence reference.
- [ ] The remaining local reachability gap is unchanged and foundation-owned:
  `go test -timeout 20m -count=1 -run
  '^TestEveryImplementedCommandPassesRuntimePreflight$'
  ./internal/connectors/commandrunner` fails only for Docker Hub `scim user
  create` and `scim user update`. Their pinned open request objects are refused
  by `internal/connectors/engine/structured_rest_body.go:1441-1444`, which
  requires `additionalProperties: false`. No connector-local schema was
  narrowed or invented; the earlier source-import refresh wait remains in
  effect until Firstmate directs a new refresh.
