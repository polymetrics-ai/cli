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
