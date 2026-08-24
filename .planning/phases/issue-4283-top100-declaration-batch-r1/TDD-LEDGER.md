# Issue #4283 — Increment 1 TDD Ledger

## Docker Hub, GitLab, Jira, Vercel, Notion, Stripe, Bitbucket, CircleCI, Sentry, Asana

### Red

`test -f internal/connectors/defs/<connector>/sources/<connector>-operation-source-lock.json` fails for every selected connector on the starting revision. The required source provenance is absent.

### Green

For every selected connector, the source lock exists and contains its public URL, SHA-256, exact byte count, retrieval timestamp, OpenAPI/Swagger version where published, per-method counts, and a method/path operation inventory. The inventory is mechanically compared to `api_surface.json`; mismatches are recorded as rejected/disabled rather than inferred.

Observed green result: `SOURCE-LOCK-VERIFICATION.json` passes 10/10 local raw-byte/SHA-256 comparisons. `PROGRESS-LEDGER.json` records 4,378 documented source operations and 4,378 corresponding API-surface declarations. `go run ./cmd/connectorgen validate` reports 552 connectors checked with zero findings.

### Refactor

Keep the source lock declarative and connector-local. Do not add a generator, dependency, or shared runtime rule in this lane.

### Non-live certification boundary

`connectorgen validate`, `surface-sync --check`, structural runtime preflight, and fixture replay are green evidence only for declarations. Provider credential and live cleanup evidence is deliberately absent and recorded as `pending`.

Observed non-live green result: each selected bundle's `certification-sweep.json` was generated then byte-checked with `connectorgen certification-sweep --connector <name> --check`. The sweep performs no provider request.

### Transport parity follow-up

**Red:** before PR #4286, no reusable declarative source factory could admit
connector-owned evidence, while the sole destination remained the closed
issue-label adapter.

**Green:** each selected bundle now has a source-only `sync_transport.json`
with a concrete stream allowlist and its own evidence reference. App open
registers the shared source adapter. `REJECTION-LIST.json` records only the
reverse leg as recoverable `generic-typed-destination-executor`; no invalid
destination is introduced.

## Increment 2 TDD Ledger

### Gitea, Grafana, Trello, Slack, n8n, Google Calendar, Gmail, Twilio, Amazon SQS, Elasticsearch

### Red

The starting branch contains no source lock for any of these ten public descriptions. A raw-byte/SHA-256 check therefore cannot establish reproducible source provenance, and their source method/path inventories are not yet mechanically reconciled to their API surfaces.

### Green

Each selected connector has a source lock with the public source URL, retrieval timestamp, exact bytes, SHA-256, format/version metadata, per-kind counts, and every documented operation. The recorded inventory is compared mechanically to `api_surface.json`, with every unmatched source operation declared disabled rather than inferred or omitted. `SOURCE-LOCK-VERIFICATION.json` passes its raw-byte and SHA-256 checks for the second increment.

### Refactor

Use connector-local JSON only. Preserve provider-specific source formats and existing action records; do not add an importer, engine executor, generic HTTP write, or guessed schemas in this declaration lane.

### Non-live certification boundary

The generated certification sweeps, structural validation, `surface-sync --check`, fixture-backed tests, and runtime-preflight checks are declaration evidence only. No credentialed provider request is made. Each selected connector records live certification as `pending`.

### Transport parity follow-up

**Red:** none of the ten selected bundles has a connector-specific registered declarative source and typed destination factory with its own evidence and acknowledgement contract.

**Green:** this later-superseded draft must use the current source-only
declaration rule: the source is declared through the shared adapter and any
reverse leg without a typed destination uses
`generic-typed-destination-executor`. No descriptor is copied from GitHub or
invented from REST documentation.

## Full-parity correction — Docker Hub source-contract inventory

### Red

`internal/connectors/defs/dockerhub/operations.json` contains an empty array,
so no pinned Docker Hub operation has a typed source-contract inventory. There
is also no connector-local source-to-API-surface crosswalk or per-operation
disposition artifact. The existing global rejection list proves broad API
surface blocking but does not prove that each pinned source row has a durable
declaration record.

### Green

The Docker Hub bundle contains source-derived `rest_read` and `rest_write`
contracts only where the pinned OpenAPI supports their GET or mutating method
and documented parameters/content type. A connector-local crosswalk and
disposition ledger account for all 54 source rows. Each enabled contract binds
to one existing API-surface method/path and then to a runnable command or typed
write action. The three HEAD rows are explicit `foundation-gap` dispositions;
credential/session routes remain `unsafe-to-exercise`.

### Refactor

Keep the derivation connector-local and data-only. Do not add a generator or
engine executor, request or use a credential, invent a transport descriptor,
or make a provider request. Re-run surface synchronization and the binary
preflight so a command claim cannot drift from its declared executor.

## Docker Hub elevated-scope correction

### Red

The first Docker Hub disposition treated 46 merely privileged operations as
`requires-elevated-scope` disabled declarations. That conflates a runtime
authorization result (`403`) with a missing executable foundation and reports a
false disabled count.

### Green

Every non-credential-management Docker Hub operation with a pinned
`bearerAuth`/administrator requirement is enabled; its source security metadata
stays visible in the crosswalk and disposition record. Repository creation and
the immutable-tag routes are normal provider-authorized actions, not paid-plan
gates. Exactly the session/token minting and access-token management routes
remain `unsafe-to-exercise`; three HEAD routes and two collection-paging routes
remain `foundation-gap`; the CSV export and two SCIM media-type writes remain
`schema-incompatible`. The Docker Hub ledger reports operations found with
source-input confidence and the separate `enabled_percent`, while live
certification remains pending.

### Refactor

Update only connector-local source declarations and issue evidence. Do not
invent scopes the source does not enumerate, a direct command, a binary
transport limit, or a runtime authorization bypass.

## Docker Hub runnable command/action correction

### Red

After the source-contract inventory was added, Docker Hub exposes only four ETL
commands and zero write actions. `surface-sync` cannot promote the 49 inventory
entries into commands, so the binary has no user-facing command for 33 enabled
source operations. The old “full parity” count is therefore invalid.

### Green

Each executable Docker Hub source row is bound to exactly one runnable command:
the four existing ETL streams retain their command bindings, 13 remaining JSON
GET operations get direct-read commands, and 16 normal JSON mutating operations
get reverse-ETL commands backed by typed write actions. All four ordinary
DELETE operations are typed delete actions with destructive confirmation. The
source-required SCIM mutation media type and the two operation-scoped
pagination shapes remain disabled because the executor cannot represent them
without foundation work. Docker Hub's inherited OpenAPI path-item parameters
are copied exactly from the pinned document while `params-import` is recorded
as a generator limitation. Surface synchronization derives command metadata
without inventing endpoints. Runtime preflight must prove direct reads reach
the credential boundary and reverse-ETL writes reach the shared lifecycle
boundary without a provider request.

### Refactor

Do not promote unsafe credential/session routes, response-less HEAD routes, or
the unbounded CSV export. Do not hand-author opaque pagination, schema fields,
or provider scopes. Leave live certification pending.

## Docker Hub credential-sensitive correction

### Red

The initial runnable-parity projection incorrectly places 13 credential and
token-management rows in `unsafe-to-exercise`. It exposes no terminal command
for the eight token metadata routes, and the five secret-bearing routes have no
declared `secret_sensitive`/`sensitive_policy` contract or source response
secret-field marker. `connectorgen validate` first reports the new declared
commands without their generated output bounds and reverse-ETL required flag
mappings.

### Green

The eight token list/detail/update/delete rows are runnable: four direct reads
and four typed reverse-ETL commands, including two destructive deletes. Each
is source-bound and reaches the normal no-credential runtime boundary. The two
token creation routes declare `token` as a redacted secret response; login,
2FA, and auth-token declare their exact x-secret request fields and redacted
response secrets. All five remain recoverable `foundation-gap` because
`internal/connectors/engine/bundle.go:2772-2776` explicitly withholds live
secret-write execution. After `surface-sync`, `connectorgen validate` and
`surface-sync --check` are green; Docker Hub reports 54 operations found from
a high-confidence machine-readable source and 41/54 enabled (75.93%), with
zero `unsafe-to-exercise` rows.

### Refactor

Retain the engine gap rather than inventing a response schema field or a secret
write executor in the connector bundle. For later connectors, classify a
secret hazard as a `foundation-gap` unless the live operation itself is truly
destructive or irreversible without user intent.

## Complete six-class source map

### Red

Before this checkpoint, batch-1 did not have a per-source-operation account of
one primary parity class, user-reachable command binding, or definition-owned
transport admission. Running certification at that point would certify a
changing partial surface.

### Green

The map integrity assertion requires, for each of Notion, Stripe, Bitbucket,
GitLab, CircleCI, Sentry, Vercel, Asana, and Jira: source disposition rows and
crosswalk rows equal the pinned source denominator; primary class counts sum to
that same denominator; each row has an enabled/disabled state plus foundation
record; and both #4286 source and destination declaration-pending records are
present.
The assertion reports 49, 589, 331, 1,755, 111, 223, 400, 249, and 617 rows,
respectively, with no missing class or transport record.

### Refactor

Keep the map connector-local and data-only. The absence of a command,
transport, executor, acknowledgement, or conformance run is a named,
recoverable gap; it is never filled from an unrelated connector or from an
unstated provider behavior.

## Vocabulary correction — declaration pending is not a foundation request

### Red

All 3,889 non-enabled rows in the nine new maps were labelled
`foundation-gap`, although their stated minimal change was only to derive a
connector-local operation contract, command or transport declaration. That
would incorrectly send declaration work to the foundation lane.

### Green

Those 3,889 rows now use `declaration-pending` with `foundation.state:
present`; their map evidence and minimal declaration change are retained. Docker
Hub now has the same six-class row fields. Exactly five rows across the ten
maps remain `foundation-gap`: three response-less HEAD operations refused by
`bundle.go:2676-2681` and two operation-scoped pagination operations not
expressible by `direct_read_paginate.go:126-130`.

### Refactor

Foundation dashboards consume only `foundation_gap` records and `gap_ids`.
`declaration_pending_ids` is an explicit separate backlog and cannot be used to
request shared engine work.

## Source-lock completeness correction

### Red

A declaration percentage can only compare a map to its own source-lock count.
It cannot establish that the source lock is the provider’s complete API, so it
is a self-referential and invalid coverage signal.

### Green

Batch 1’s ten locks each expose `counts.total`, per-method counts and a
same-sized operation inventory from a provider-published OpenAPI document:
Docker Hub 54, Notion 49, Stripe 589, Bitbucket 331, GitLab 1,755, CircleCI
111, Sentry 223, Vercel 400, Asana 249, and Jira 617. The maps and progress
ledger now report `operations_found` and `coverage_confidence: high` with their
machine-readable-spec basis; they no longer report `declared_percent` or a
percentage based on mapping a lock to itself.

### Refactor

Before a future connector is mapped, reject a lock without `counts.total` or
per-kind counts, and re-pin any implausibly small complete-surface count rather
than treating map parity as source completeness.

## Provider-surface reconciliation correction

### Red

Using an existing `api_surface.json` as the map boundary could preserve a
legacy undercount even after a correct source lock is available.

### Green

Each Batch-1 provider OpenAPI method/path set was compared directly with its
`api_surface.json`. All 4,378 provider operations are present; no surface is
understated and no regeneration is needed. The durable
`API-SURFACE-REALITY-AUDIT.json` records old count, provider count, unchanged
new count and basis per connector. Notion, Sentry and Vercel retain only their
explicitly described extra bounded/legacy entries.

### Refactor

For every later connector, calculate the provider operation set first and
generate/extend `api_surface.json` from that set. An old API surface is never
evidence of the provider’s complete documented surface.

## CI verify gap — shared declarative source evidence

### Red

`go test -timeout 20m ./internal/app -run
'^TestDefinitionTransportFactoriesSelectDeclaredEvidence$' -count=1` fails:
the source factory has Asana as the deterministic first evidence and GitHub in
`AcceptedSourceEvidences`, but the test demands that GitHub occupy the primary
field. The failing CI run is `32271368383`.

### Green

The test asserts that the GitHub declaration's exact evidence occurs in the
factory's accepted source-evidence set (primary or additional), which is the
contract enforced by `synctransport.RegisterDeclaredTransports`. It continues
to assert GitHub's destination evidence exactly. A direct rerun of the test
and the `internal/app` package demonstrate the fix.

### Refactor

Keep factory evidence ordered only as an implementation detail; no connector
name or registry ordering is introduced into production composition.

## Authorized downstream shared-test repair — 2026-08-24

### Red

```sh
go test -timeout 20m -count=1 -run '^(TestOperationEvidenceFixed100RejectsEveryRegression|TestOperationEvidenceCheckRunsFixed100Gate)$' ./cmd/connectorgen
go test -timeout 20m -count=1 -run '^TestDefinitionTransportFactoriesSelectDeclaredEvidence$' ./internal/app
```

The first command rejects the current fixed cohort because the test workspace
contains only GitHub while it loads a cross-connector reference beginning with
Asana. The second command rejects a correct shared factory because GitHub's
exact evidence is accepted but is not the first registry entry. The exact
outputs and selection counts are recorded in
`VERIFY-SHARED-TEST-FIXTURE-EVIDENCE-2026-08-24.md`.

### Green plan

- Preserve the checked-in 100-row fixed reference unchanged.
- Build the temporary operation-evidence workspace from the fixed reference's
  exact connector prefixes and retain only those generated website rows.
- Add a deliberate temporary GitHub source-row removal assertion proving the
  direct fixed validator and CLI `--check` both fail on a genuine loss.
- Assert GitHub's exact declared source evidence occurs in the factory's
  primary-or-accepted set; retain the exact GitHub destination assertion.

### Green

```text
ok  polymetrics.ai/cmd/connectorgen  16.514s
ok  polymetrics.ai/internal/app      2.896s
```

The changed fixed-cohort tests create a separate temporary workspace, remove
`github.rest.issues/list-for-repo` from its copied source lock, and assert the
specific ID appears in the direct validator failure and in the CLI `--check`
failure. The temporary workspace is discarded by `t.TempDir`; the checked-in
source lock and fixed reference are unchanged. The app test rejects any factory
whose primary-plus-accepted exact evidence records omit GitHub's declared
conformance reference.

### Refactor boundary

No production transport code, connector definition, source lock, generated
artifact, or test guard is removed or loosened. This is a test-fixture
completeness and implementation-order-independence repair only.

## Classification correction — direct write is not reverse ETL

### Red

The completed map placed 250 ordinary write endpoints in the `reverse_etl`
primary class. That made direct-write enabled counts appear as zero and claimed
118 enabled reverse-ETL operations despite no destination transport contract.

### Green

All 250 rows are now `direct_write`, yielding 2,370 direct-write endpoints and
118 enabled direct-write bindings. Each direct-write row carries a distinct
`reverse_etl_eligibility` attribute set to the recoverable
`generic-typed-destination-executor` gap, with zero eligible operations. ETL
remains a primary endpoint class where its source transport is declared.

### Refactor

Keep the map taxonomy independent of legacy command lifecycle intent: no typed
write action, create/update/upsert/delete kind, or CLI binding can imply a
reverse-ETL destination. Only the definition-owned destination contract can.

## PR #4297 Docker Hub repair loop — observed result

### Red

`connectorgen validate internal/connectors/defs` initially rejects a Docker
Hub `rest_status` declaration with `has unsupported kind "rest_status"`.
The executor itself is present, so the red result identifies an integration
gap in loader admission rather than a missing status executor.

### Green

The four contracts that PR #4297 can load without a further shared change now
pass `connectorgen validate`, `surface-sync --check`,
`TestEveryImplementedCommandPassesRuntimePreflight`, and the non-live sweep:
`GET /v2/auditlogs/{account}`, `GET /v2/scim/2.0/Users`,
`POST /v2/scim/2.0/Users`, and `PUT /v2/scim/2.0/Users/{id}`. Each reaches
`error: missing --credential` in an isolated initialized project, proving
dispatch without provider I/O. Docker Hub moves from 41 / 54 to 45 / 54.

### Refactor

The three HEAD checks and CSV export remain disabled as
`operation-kind-loader-registration`. The shared loader/validation path at
`internal/connectors/engine/bundle.go:2451,2676,2705,2733` omits
`rest_status` and `text_export`. The minimal shared fix is to register
`rest_status` and `text_export` in the bundle loader operation-kind switch and
validation so a definition can declare them, then add a loader regression test.
This definitions-only repair does not make that foundation edit, and it does
not repurpose the five secret-response rows as unsafe-operation refusals.

## Local verify recovery — Notion operation ledger

### Red

The detached local `make verify` run reached
`TestNotionAPISurfaceOperationLedger` and failed: the surface had 55 rows
where the checked provider ledger permits 54 (51 operations plus three
source-qualified response arms). It also found that the three OAuth rows and a
duplicate unqualified search row did not preserve the required
`named_dependency=` disposition evidence.

### Green plan

Remove only the redundant unqualified `POST /v1/search` presentation row; the
two existing source-qualified search rows remain the exhaustive operation
declaration. Restore the dependency prefix on the three recoverable
secret-policy rows, then re-run the exact ledger test and the full local gate.

### Green

`go test -timeout 20m ./cmd/connectorgen -run
'^TestNotionAPISurfaceOperationLedger$' -count=1` now passes. The surface has
54 rows and continues to represent all 51 provider actions through the two
qualified search arms; no provider operation was removed.

## Reverse-ETL preparation — Docker Hub typed actions

### Red

The completed map treated all 27 Docker Hub direct writes as equally available
for a future destination. That would silently include five credential/session
operations and would let a typed `json` action mislabel the two SCIM writes,
whose pinned request media type is `application/scim+json`.

### Green

`sources/dockerhub-reverse-etl-action-audit.json` exhaustively classifies all
27 source-backed writes: 20 existing typed actions are ready for a future
connector-neutral destination, two SCIM actions are held by the recoverable
`typed-action-content-type` gap, and five credential lifecycle/session
operations are not reverse-ETL targets. No destination or `transport_binding`
is declared before #4303 supplies the generic typed destination factory.

## Reconciliation relaunch — 2026-08-20

| Slice | Red | Green | Refactor / evidence |
| --- | --- | --- | --- |
| Seven-surface audit | Existing Batch-1 maps record reverse ETL as a shared missing-destination gap, so an audit requiring exact destination/action/input/acknowledgement evidence fails. | Merge #4304, then make every eligible typed action pass the destination-factory declaration and real command preflight path. | Persist the audited ten-row ledger and keep only exact refusing engine file/line gaps. |
| Connector-local command reachability | A documented row without a `cli_surface` binding produces `unknown command` or has no real command path. | Source-backed command declarations reach `missing --credential` (or their existing plan/preview approval boundary) without any provider request. | Regenerate command docs and retain operation-level source/crosswalk evidence. |
| Typed destination conformance | A direct-write action without exact source fields, strategies, acknowledgement, and delivery evidence is rejected by the #4303 factory/preflight. | Each eligible action carries the complete definition-owned destination contract and is admitted structurally; no raw HTTP destination is introduced. | Retain binary operations as binary-only and list only true executor/schema refusals as recoverable gaps. |

### Observed reconciliation slice — 2026-08-20

**Red:** the seven-surface shell audit exited 1 before this slice: Docker Hub
`54/45/20`, Stripe `589/8/3`, GitLab `1755/4/0`, CircleCI `111/0/7`, Sentry
`223/0/0`, Vercel `400/0/18`, Asana `249/249/73`, and Jira `617/590/292` in
`documented/CLI-command/typed-action` order; no connector had a declared
destination transport. The first Jira attempt also failed structural loading:
`projectIdOrKey` is not a concrete transport mapping identifier because its
upper-case characters are refused at `internal/connectors/sync_transport.go:673-690`.

**Green:** declared the fixture-backed update mappings that the closed contract
can represent now: Notion `views -> update_view`, Stripe
`customers -> update_customer`, CircleCI `schedules -> update_schedule`, and
Vercel `projects -> update_project`. Added the two missing CircleCI/Vercel
ETL and reverse-ETL command paths. The four installed-binary commands below
all returned exactly `error: missing --credential` from an isolated initialized
project, with exit 1 and no provider I/O:

```
pm circleci schedules list --json
pm circleci schedules update --id sched_fixture_1 --name nightly-build --preview --json
pm vercel projects list --json
pm vercel projects update --id prj_fixture_1 --name fixture-app --preview --json
```

Structural green commands:

```
go run ./cmd/connectorgen validate internal/connectors/defs/<each of dockerhub,notion,stripe,bitbucket,gitlab,circleci,sentry,vercel,asana,jira> --json
go test -count=1 -timeout 20m ./internal/connectors/commandrunner -run 'TestEveryImplementedCommandPassesRuntimePreflight'
go test -count=1 -timeout 20m ./internal/app -run '^TestDefinitionTransportFactoriesRunTypedDestinationFromDefinition$'
```

All passed. `SEVEN-SURFACE-RECONCILIATION.json` contains the stable
write-action-set SHA-256 selector and an explicit disposition for every typed
action. The remaining exact foundation gap is
`action-scoped-source-binding`; persisted App/CLI destination dispatch remains
the pending upstream #4304 dependency, not a completed deployment claim.

### Per-action eligibility clarification — 2026-08-20

**Red:** a connector-level action-set selector can prove which set was audited,
but it does not itself make the eligibility of each typed action inspectable.

**Green:** the seven-surface ledger now contains one named entry for every
action emitted by each connector's `writes.json`: 491 entries in total, with
four `eligible_bound_fixture_mapping` entries and 487 eligible pending entries.
The mechanical cross-check compares the sorted `writes.json` names against the
ledger names for all ten connectors and reports zero missing or extra actions.
Each pending entry is bound to its connector's precise source-identity,
nested-object, or `action-scoped-source-binding` foundation record; no entry
uses risk, privilege, deletion, cost, or missing live credentials as an
ineligibility reason.

### Documented-operation command reachability boundary — 2026-08-20

**Red:** joining every source-crosswalk operation to the current API-surface
coverage leaves 3,366 of 4,378 pinned operations without a declared command
binding. A representative Jira blocked body and a CircleCI unmaterialized
endpoint cannot be added as partial commands: `checkCLISurfaceEndpointCoverage`
rejects their unbound endpoints, and `resolvePreflightCommand` rejects the
operation-backed partial route before provider dispatch.

**Captain decision / green design:** the disabled command proposal is rejected:
`BlockedCommandError` is not API reachability. The new capability audit retains
an exact source identity, method, path, location, prior rejection, and next
executable capability for all 3,366 rows. Its allocation is 1,389 fixed REST
reads, 1,828 fixed REST writes, 120 bounded binary transfers, 10 status-kind
registrations, and 19 provider contracts without a bounded typed schema. The
separate typed-action audit identifies 224 nested object/array actions as
downstream #4305 structured-body consumers.

**Refactor / dispatch:** `EXECUTABLE-OPERATION-FOUNDATION-DESIGN.md` defines
the closed source-import, header, structured-body, binary/status/text,
destination-dispatch, and connector-batch slices. Each promoted command must
use an exact source operation, fixed method/path, declared input schemas,
bounded values, existing policy, and real provider I/O. No raw/generic
transport or caller-selected metadata is introduced.

## Main-base source-import compatibility — 2026-08-23

### Red

After the main-base merge, `go run ./cmd/connectorgen validate` rejects the
Batch-1 source-lock presentation and `go run ./cmd/connectorgen source-import
dockerhub` refuses the exact hash-pinned public YAML before emitting its
canonical descriptor. The reproduction needs no credential or provider write.

### Diagnosed boundary

Removing stale, derivable `operation_counts` and `source_completeness`
metadata lets the lock reach the importer. The exact public artifact then
fails at `cmd/connectorgen/sourceimport.go:1305-1318`: Docker Hub uses numeric
response-status YAML keys, and the importer accepts only string YAML scalar
keys. A connector-local source rewrite would sever the pin, so this lane
records `source-import-yaml-scalar-key-normalization` as a shared foundation
gap instead of guessing a descriptor.

### Separate Docker Hub SCIM validation diagnosis

`go run ./cmd/connectorgen validate internal/connectors/defs/dockerhub` also
reports an independent body-schema failure before its missing-descriptor
finding: commands 43 and 44 (`scim user create` and `scim user update`) expose
OpenAPI `example` annotations inside their canonical `body_schema`, while the
engine compiler rejects `example` as an unknown JSON Schema keyword. The exact
operation IDs are `dockerhub.post__v2_scim_2.0_users` and
`dockerhub.put__v2_scim_2.0_users__id_`. This is not the YAML numeric-key
importer gap; source projection must reconcile the annotation only from the
future canonical descriptor, rather than a hand-authored body mutation.

### Main-base nine-connector source-import sweep

**Red:** a public, credential-free source-import run produced one descriptor
and eight explicit refusals. Notion, Bitbucket, GitLab, CircleCI, Vercel, and
Jira returned `source-lock refresh required: fetched artifact does not match
locked bytes and SHA-256`; Stripe refused `reference cycle at
"#/components/schemas/file"` while preflighting `GET /v1/account` response
`200`; Asana reached its pinned YAML and hit the same non-string response-key
refusal as Docker Hub at `/paths/~1access_requests/get/responses`.

**Observed result:** Sentry imported all 223 locked operations without
credentials, but its canonical source projection exposed a separate reachability
gap: `connectorgen validate internal/connectors/defs/sentry` reports 34 source
operations with no executable action, including documented DELETE operations.
This confirms the saved 223/223 declaration inventory is not sufficient proof
of installed-binary reachability on current main.

**Green pending decision/foundation:** preserve all existing denominators and
source pins. Resume source refresh only under an explicit captain decision;
resume Stripe/Asana/Docker Hub after their precise importer gaps land; derive
Sentry operation actions only from its emitted canonical descriptor.

## e338cd301 source-lock refresh qualification — 2026-08-23

**Red (measured after merge):** current `origin/main` at `e338cd301` admits
the previously rejected cyclic schemas and `$ref` descriptive siblings, but no
checked-in pin was rewritten before the full public-source contract could be
verified. Each candidate artifact was fetched without credentials, measured,
and applied only to an isolated copied bundle before invoking the production
importer.

| Connector | Measured result | Refusing boundary |
| --- | --- | --- |
| Notion | `1304814` bytes, `dee576…2f258`; blocked at `GET /v1/blocks/meeting_notes` response schema depth | `cmd/connectorgen/sourceimport.go:65,4269-4272` |
| Bitbucket | `1359673` bytes, `3dbfe6…0dec3`; blocked at pull-request comments response schema depth | `cmd/connectorgen/sourceimport.go:65,4269-4272` |
| GitLab | `3576860` bytes, `6b6ad5…6cf82`; `epic_issue_id` is not a required source path parameter | `cmd/connectorgen/sourceimport.go:6036-6048` |
| CircleCI | `621321` bytes, `61c6ce…66d07`; imports 111, then 27 mutation actions are missing | `cmd/connectorgen/sourceprojection.go:137-143,210-211` |
| Vercel | `10463249` bytes, `74cb7f…3da28`; `POST /api-keys` response uses `patternProperties` | `cmd/connectorgen/sourceimport.go:4311-4315` |
| Jira | `2456011` bytes, `511d0b…7e5e8`; imports 617, then 16 mutation actions are missing | `cmd/connectorgen/sourceprojection.go:137-143,210-211` |

CircleCI's missing actions include the secret-bearing context environment
variable PUT. Its required `env_only` support remains PR #4334, verified open,
unmerged, and behind `main`; that caveat is real but does not explain the other
26 unmapped mutations. Vercel's read-only coverage caveat was not reached: the
importer refuses `patternProperties` first. The expected Notion, Bitbucket,
GitLab, and Jira clear result is not reproduced, so no pin, denominator, or
derived declaration is changed on this evidence.

## PR #4294 main-base typed-destination binding repair — 2026-08-24

**GSD/manual fallback:** generated the `discuss-phase`, `plan-phase --gaps`,
`execute-phase --gaps-only`, `verify-work`, and `code-review` prompts and
executed the single-worker gap plan/review inline because this dispatcher
cannot provide the adapter's isolated Pi worker runtime. Required skills loaded: `golang-how-to`,
`golang-troubleshooting`, `golang-testing`, `golang-error-handling`,
`golang-safety`, and `golang-continuous-integration`.

**Red:** at PR head `761c1bc8a`, both native target builds and the verify
certification harness reject the same loaded bundle contract. The local,
credential-free reproduction is:

```text
go test -timeout 20m -count=1 -run 'TestWarehouseMaterializesTablesAsParquet|TestQuerySQLAggregatesOverParquetTables|TestReverseETLReadsAParquetSourceTable' ./internal/app
```

All three tests fail before source or provider I/O with
`declarative typed destination requires an action-owned source binding`.
The declaration inventory identifies exactly four affected destination
transports: `circleci/update_schedule`, `notion/update_view`,
`stripe/update_customer`, and `vercel/update_project`. Each has one ordinary
`source_bindings` entry with the exact action's existing `input_fields` mapping
but no `action` identity. The runtime guard at
`internal/app/issue_label_warehouse_transport.go:907` correctly refuses that
ambiguous binding.

**First Green attempt (rejected by the next runtime gate):** add the
corresponding existing `writes.json` action name to each of those four
connector-local binding objects without altering input fields, routes, schemas,
source locks, engines, or shared transport code. The ensuing red proves that
action identity alone cannot establish safe delivery; no temporary binding or
batch declaration remains in the final change. No credentials or live provider
calls are authorized.

**Second Red / root cause:** action identity made the next validator boundary
observable: every affected binding lacks the required declaration-owned batch,
and, after that mechanical addition, the runtime correctly refuses
`update_schedule` because it has no provider idempotency key header. The same
guard would then require action-owned bounded read-back declarations. These
are not inert JSON defaults: `internal/app/issue_label_warehouse_transport.go`
requires provider idempotency at lines 971-980 and action-owned read-back at
982-991. The existing Notion action explicitly records that the provider has
no idempotency header in `internal/connectors/defs/notion/writes.json:914`.
The other three actions have no source-cited header or action-owned read-back
contract in their pinned declarations. Inventing either would turn a direct
write command into an unsafe replaying destination transport.

**Green:** removed the four invalid `destination_transport` claims rather
than inventing headers, acknowledgement units, or read-back routes. Their
source transports, source locks, writes.json actions, and installed direct
commands remain intact; generated manuals now say `Destination transport:
unsupported`. The credential-free focused app reproduction and
`TestSampleOutboxWriteLifecycleAgainstRealCLI` both pass after the removal.
The remaining reverse-ETL destination capability is recorded as a
source-cited declaration/foundation gap, not a disabled or omitted provider
operation.

**Final local gates:** `make tidy-check`, `make lint`, `go vet ./...`, `go
build ./cmd/pm`, focused `internal/app` and `internal/connectors/certify`
tests, `make docs-check-no-build`, `make smoke-no-build`,
`make connector-boundary`, `make connector-canon-check`, all four
`connectorgen-certification-*` checks, `make github-parity-artifacts-check`,
and `make release-workflow-check` pass. `certification-subject` first reded on
a stale generated fingerprint; running its documented generator changed only
`internal/connectors/certifications/current-subject.json`, and its `--check`
then passes. `operation-evidence --check` likewise reded on a stale generated
artifact; `go run ./cmd/connectorgen operation-evidence --write-fixed-100`
regenerated it to 5,903 rows and the fixed-100 check then passes.

**Deferred validator gates (pre-existing shared foundations, no mutation):**
`go run ./cmd/connectorgen validate internal/connectors/defs` reaches 49
findings: the batch's Asana, Bitbucket, CircleCI, Docker Hub, GitLab, Jira,
Notion, Stripe, and Vercel descriptors are absent; Docker Hub also retains the
source-derived SCIM `example` dialect refusal; Sentry retains 34 action
projection gaps. `go run ./cmd/connectorgen surface-sync
internal/connectors/defs --check` stops first at the same absent Asana
descriptor. The authoritative refusing code is
`cmd/connectorgen/sourceprojection.go:1810-1814` and
`cmd/connectorgen/surfacesync.go:227-232`; these are recorded gaps rather than
invented descriptors or altered source locks.

**Independent installed-binary audit:** `make connector-runtime-preflight`
reaches the pre-existing Docker Hub SCIM body-schema dialect gap and rejects
only `scim user create` and `scim user update`, both because the pinned-derived
schema retains an unsupported OpenAPI `example` keyword. The exact compiler
refusal is `internal/connectors/engine/schema.go:165-168`. The operations stay
declared and their direct command routes stay present; changing them to look
partial or deleting their schema would hide the missing provider-dialect
foundation rather than repair it.

**Credential-free command reachability:** an initialized disposable project
was used with no credential configured. The four repaired routes (`circleci
schedules update`, `notion view update`, `stripe customers update`, and
`vercel projects update`) were each invoked with their required identity flag
and stopped at `error: missing --credential`. Thus removal of the invalid
destination declarations did not unpublish or misroute any installed command;
no provider request was made.

**Current-main regression red:** after the required merge of main commit
`27664370c` (`#4334`), the full merged package run first fails the new
fixed-100 test because its temporary workspace copies only GitHub while the
cohort includes Asana, then fails the shared-source factory test because it
compares GitHub's evidence to the first registered (Asana) factory. The
production operation-evidence check passes. These are shared foundation/test
contracts, recorded with exact lines in `FOUNDATION-GAPS.md`; no connector
declaration is altered to conceal them.

## Captain hard pre-merge gate — 2026-08-20

### Red

The reconciliation aggregate can show a source count, a declared command
count, and action eligibility without proving that every provider operation is
runtime-enabled, generated-CLI reachable, website-derived from the same ledger,
and covered by the semantically applicable ETL, reverse-ETL, direct-read,
direct-write, binary-download, and binary-upload surfaces.

### Green design

`SEVEN-SURFACE-RECONCILIATION.json#pre_merge_gate` now makes completion a
per-operation two-way source/ledger/CLI/website trace. It requires fixture or
conformance evidence, output preservation, zero silent omissions, and exact
generated-surface drift checks. `N/A` is restricted to provider-evidenced lack
of semantic capability; risk, scope, cost, and absent credentials are runtime
policy concerns. Zoom, Twenty, and Gong retain their authorized provider-live
certification requirement; no other connector is called live-certified without
credentialed accepted evidence.

### Refactor / pause

No connector implementation or provider I/O is introduced. The branch stays
paused until F0, F2/F4, #4305, and final #4304 dispatch heads publish.

### Deferred executable enforcement

The JSON gate is not a test. After those foundations publish, implement the
fixed-100 CI validator on the integration branch with schema-backed per-
connector and aggregate verdicts. Its red suite must deliberately fail for a
missing source hash, missing or unreachable command, CLI/website drift, absent
fixture/conformance evidence, incorrect semantic surface, collapsed binary
direction, a disabled callable operation, and a dropped non-secret output. Its
green result is a passing repository command; the planning ledger alone is
never green enforcement.

### Missing-foundation fanout

**Red:** a shared engine deficit could be recorded as a connector-local
`disabled`/`N/A` decision, omit its source identity or exact closure evidence,
or vanish from an aggregate when one provider operation fans out to another.

**Green design:** `MISSING-FOUNDATION-DELIVERABLE.json` defines eight stable
shared gaps and machine-joins their owning lane, closure verification, and
fanout to 3,366 exact source operations and 491 typed actions. All rows remain
open and not enabled; 28 unresolved action identities are visible F0/F1 rows,
not fabricated source bindings. The batch rollup remains 10/100 and the
portfolio rollup marks 90 connectors unmapped.

## Generated skills synchronization and Docker Hub preflight comparison — 2026-08-24

### Red

`GOFLAGS=-p=3 make verify` reports that `pm skills generate` output differs
from tracked `docs/skills`. The same aggregate run reports Docker Hub `scim
user create` and `scim user update` as runtime-preflight failures because their
source-derived request schemas retain unsupported `example` keywords.

### Green plan

- Regenerate only `docs/skills` with `pm skills generate --dir docs/skills`,
  then run the exact tracked-output parity test.
- Run `TestEveryImplementedCommandPassesRuntimePreflight` in a clean,
  disposable current-main tree and compare the Docker Hub result to this
  branch before changing any SCIM declaration.

### Refactor boundary

The generator repair must not alter connector facts. A current-main failure
with the same operation IDs and schema refusal is required evidence before
treating it as pre-existing.

### Green result / branch regression

The generator changed only the ten Batch-1 `docs/skills` files, and
`TestSkillsGenerateMatchesTrackedSkills` passes. The clean current-main
comparison at `origin/main` `3c394a0e` passes
`TestEveryImplementedCommandPassesRuntimePreflight`; the same branch command
fails for exactly Docker Hub `scim user create` and `scim user update` on the
unsupported `example` keyword. This is a branch regression and requires the
non-rewriting current-main merge before re-verification.

### Post-merge Docker Hub open-object body boundary

**Red:** after merge commit `f528b806d`, the two same source-backed commands
reach `requireClosedBoundedStructuredRESTBody` and fail because their pinned
SCIM request schemas have object `properties` but no
`additionalProperties:false`.

**Candidate green rejected:** retaining the command/operation/API-surface
bindings as `availability: partial` makes runtime preflight pass, but the
immutable fixed-100 check rejects it as an execution-evidence regression for
`dockerhub.rest.post_/v2/scim/2.0/Users`. Restore the prior declarations: a
partial placeholder is not a full-parity correction.

**Refactor:** no source lock, source-derived `operations.json` body schema, or
shared engine code changes. The public document SHA-256 remains
`99d9d53c...53d0756`; the unresolved recovery is a typed bounded open-object
body capability.

## Authorized fixed-100 runtime-preflight correction — 2026-08-24

**Decision:** Firstmate's inbox item `010.msg` authorizes a focused correction
to operation-evidence runtime eligibility. It does not authorize an engine
change, a Docker Hub declaration/source-lock edit, or a regenerated checked-in
fixed-100 reference.

**Red (planned):** add a `cmd/connectorgen` regression test over the existing
Docker Hub SCIM create and update source IDs. Before production code changes,
it must fail because the evidence projector reports both rows runtime-enabled
and includes them in `buildOperationEvidenceFixed100`, while the production
commandrunner preflight refuses their source-declared open request bodies.

**Green (planned):** route each matching implemented command through
`commandrunner.Preflight` using the loaded declarative `engine.New` connector.
The projected rows must report a `runtime_reachability` gap and disabled runtime
for the two SCIM operations; the would-be cohort must omit them without any
hand selection.

**Refactor boundary:** do not duplicate the runtime's structured-body logic,
do not expose raw bodies, and do not change the actual cohort baseline. Skills:
`golang-how-to`, `golang-design-patterns`, `golang-structs-interfaces`,
`golang-error-handling`, `golang-security`, `golang-safety`, and
`golang-testing`.
