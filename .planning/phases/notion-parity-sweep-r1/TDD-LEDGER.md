# Notion parity TDD ledger

> **Recorded retrospectively** — see the note in `PLAN.md`. The count-lock test was written and
> committed *before the PR opened* but *after* the bundle, so the honest label for this phase is
> test-after, not test-first. dynamodb onward is genuinely red-first.

## Baseline before production edits

- `connectorgen validate` passed on the pre-change bundle with 0 findings — which is exactly the
  point: validate checks internal consistency and cannot detect a missing operation surface.
- Baseline bundle: **6** `api_surface.json` endpoints (3 `covered_by` streams, 3 legacy `excluded`),
  `capabilities.write: false`, and no `cli_surface.json`, `operations.json` or `writes.json`.
  **0 of 51 documented operations were reachable as `pm notion <command>`.**
- Provider re-derivation: **51** operations from the official OpenAPI 3.1.0 document, vs the ledger's
  carried-forward **50**. Understated by 1.

## Tests / assertions

`cmd/connectorgen/notion_api_surface_test.go` asserts:

- 51 unique method+path actions across 54 rows, stripping the row qualifier before counting so rows
  and actions are never conflated;
- per-method split 21 GET / 18 POST / 8 PATCH / 4 DELETE;
- `covered_by` split of 6 stream, 18 direct_read, 24 write;
- `operation_ledger_version` set;
- exactly one disposition per row, none blank, no surviving legacy `excluded` rows;
- every blocked row `blocked_by_default` with a source citation and a `named_dependency=` note.

**Verified to bite, not merely to pass:** removing a single endpoint fails it three ways (row count,
unique-action count, covered+blocked total). Checked by temporarily mutating the bundle and restoring
it.

## Green evidence

- `go run ./cmd/connectorgen validate internal/connectors/defs`: 551 connectors, **0 findings**.
- `go test ./cmd/connectorgen -run TestNotionAPISurfaceOperationLedger`: passes.
- `go test ./internal/connectors/commandrunner -run TestEveryImplementedCommandPassesRuntimePreflight`:
  passes. Initially failed for all 18 direct reads until `connectorgen surface-sync` filled the
  runtime operation endpoint ledger; that diff was inspected and is confined to the `"notion"` block.
- `go run ./cmd/connectorgen surface-sync internal/connectors/defs --check`: clean.
- `go test ./internal/connectors/conformance`: passes.
- `make certify-timing`: **passes**, 89.1s of a 3m30s budget, 92 real CLI invocations at budget.
- `make docs-check`, `make connector-boundary`, `make agent-contract-check`, `make tidy-check`: pass.
- Binary built and every scope run: inspect, bare namespace (16 groups), direct_read, reverse_etl,
  destructive confirmation, and the blocked upload.

## Refactor / safety notes

- `TestGoldenTranscripts`: diff read **before** regenerating. Exactly one added line per root-help
  transcript — notion joining the connector command-surface list. Nothing else moved.
- Website catalog: 1,920 insertions compared **by object rather than by line** — 551 connectors
  before and after, none added, none removed, exactly one changed.
- Fixtures are synthetic and secret-free.
