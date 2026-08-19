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
inventory rows only where the pinned OpenAPI supports their GET or mutating
method and documented parameters/content type. A connector-local crosswalk and
declaration-disposition ledger account for all 54 source rows. Every inventory
row binds to an exact existing API-surface method/path; all non-terminal rows
remain blocked in `api_surface.json`. The three HEAD rows are explicit
`foundation-gap` dispositions and the deprecated login row is explicit
`provider-does-not-expose`.

### Refactor

Keep the derivation connector-local and data-only. Do not add a generator,
engine executor, command, fixture, credential path, transport descriptor, or
provider request. Re-run surface synchronization only to prove it introduces
no unreviewed command-surface drift.

Observed green result: `operations.json` now holds 49 source-backed inventory
contracts (23 `rest_read`, 26 `rest_write`, including six `delete` contracts).
The exact 54-row crosswalk and disposition ledger retain four declared stream
bindings and 50 disabled terminal rows: 46 `requires-elevated-scope`, three
`foundation-gap` HEAD operations, and one `provider-does-not-expose` deprecated
login. Docker Hub-only validation, source-byte verification, certification
sweep check, conformance, runtime preflight, surface sync, docs, CLI golden
tests, boundary/canon checks, and the applicable repository gates pass without
a provider credential or provider request.
