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
