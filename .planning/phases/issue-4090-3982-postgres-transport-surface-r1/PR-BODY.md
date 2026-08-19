## Intent

Refs #4090
Refs #3982

Keep the PostgreSQL transport surface honest: every advertised source/destination mode must be accepted by the production registry, while an unpaired mode must refuse before executor I/O.

## What changed

- Preserved `full_overwrite`, which PR #4167 had reduced without issue/plan authority and PR #4175 had already restored on the required base.
- Removed only source-only `incremental_dedupe_history`; no shipped destination admits it.
- Added production-composition coverage that preflights all five advertised PostgreSQL modes and observes a three-record full-overwrite fixture read.
- Regenerated the PostgreSQL certification shard with `connectorgen` and the connector catalog with `pm docs generate`; neither artifact was hand-edited.

## Test contract

- Happy: `TestOpenPreflightsEveryDeclaredPostgresDestinationMode` asserts the resolved production executor references for all five modes and exactly three produced full-overwrite records.
- Bad: `TestOpenRefusesPostgresUnpairedHistoryModeBeforeExecutorIO` asserts the exact source-mode refusal before executor I/O.
- Edge: `TestOpenPostgresTransportDeclarationsAreExactModeIntersection` asserts the finite, duplicate-free five-mode declaration on both sides.

## Red / Green

- Red: history mode was source-declared but production preflight refused it at the destination; exact output is recorded in the TDD ledger.
- Green: history is no longer declared by the source, all five declared modes preflight, and `pm connectors inspect postgres --json` reports the same sets.

## Verification

- Focused App, PostgreSQL native, transport, CLI, and `connectorgen` tests; targeted race test; `go vet`; built `pm`.
- `go run ./cmd/connectorgen certification-matrix --connector postgres` and `--check`.
- `pm help connectors`, bare `pm connectors`, `pm connectors inspect postgres --json`, generated docs, docs validation, smoke, lint, connector generation/boundary/canon, and agent-contract gates.

## GSD, skills, and safety

The adapter sources for `discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and `code-review` were resolved and executed inline under the direct-PR single-worker fallback. Required routing, Go testing, CLI parity, and runtime/PostgreSQL references were applied.

No dependency, credential, generic SQL/HTTP surface, reverse-ETL action, or `write:true` capability change. CLI docs parity is the regenerated connector catalog; no command/flag/manual/website text changed.

## Review

Inline review found no unresolved findings. Claude automatic review is expected after PR creation; Copilot was not requested.
