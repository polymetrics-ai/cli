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

**Red:** all ten selected bundles lack `sync_transport.json`; copying GitHub's declaration would name its exact evidence and issue-label destination actions without a source-derived contract.

**Green:** `REJECTION-LIST.json` has one `sync_transport`/`foundation-gap` entry with `recoverable: true` for every selected connector, and `TRANSPORT-GAP.md` cites the factory/evidence admission code and its smallest safe recovery. No invalid descriptor is introduced.

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

**Green:** each selected connector receives recoverable source/destination `sync_transport` foundation-gap records tied to #4093, with the evidence and smallest safe recovery already documented in `TRANSPORT-GAP.md`. No descriptor is copied from GitHub or invented from REST documentation.

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
`schema-incompatible`. The Docker Hub ledger reports both `declared_percent`
and `enabled_percent`, while live certification remains pending.

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
`surface-sync --check` are green; Docker Hub reports 54/54 declared (100.00%)
and 41/54 enabled (75.93%), with zero `unsafe-to-exercise` rows.

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
