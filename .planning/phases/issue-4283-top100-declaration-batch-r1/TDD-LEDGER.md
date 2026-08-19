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
