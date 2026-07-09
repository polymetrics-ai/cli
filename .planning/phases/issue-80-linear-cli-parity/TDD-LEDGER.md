# TDD Ledger — Issue #80 Linear CLI parity parent

Date: 2026-07-09

## GSD / fallback record

- Parent planning prompt: `scripts/gsd prompt plan-phase issue-80-linear-cli-parity --skip-research`.
- Programming loop prompt attempted: `scripts/gsd prompt programming-loop init --phase issue-80-linear-cli-parity --dry-run` and later `scripts/gsd prompt programming-loop issue-80-linear-complete-ops --skip-research`.
- Result: unavailable (`unknown GSD command: programming-loop`), so manual-GSD fallback was used for PLAN → RED → GREEN → REFACTOR → VERIFY.

## Sub-issue evidence index

| Issue | Lane | Red evidence | Green evidence | Refactor/verification |
|---:|---|---|---|---|
| #97 | CLI surface metadata | `TestLinearCLISurfaceMapsImplementedStreams` failed before `cli_surface.json` existed | Added Linear command metadata for stream-backed list commands | Prior #97 gates passed and parent PR #131 was opened |
| #98 | Help renderer/docs | `TestLinearConnectorHelpRendersCommandSurface` failed before `pm help linear`/bare `pm linear` rendered connector help | `internal/cli/cli.go` now falls back to connector command-surface manuals and bare namespace help | `./pm help linear`, `./pm linear --help`, generated `docs/connectors/linear/**`, website data regenerated |
| #99 | Stream runner | `TestLinearCommandSurfaceRunsGraphQLIssueList` failed before Linear streams had fixed GraphQL request bodies | Linear list streams now use fixed `POST /graphql` GraphQL documents and command runner executes them through the stream path | Linear conformance and focused CLI tests passed |
| #100 | Operation ledger | `TestLinearOperationLedgerInventoriesGraphQLOperations` failed before ledger v1 was present/loadable; refreshed tests failed while the SDK-document slice had only 144 query rows and 321 mutation rows | `api_surface.json` inventories all 531 live Linear GraphQL root fields; all 514 official non-deprecated fields are covered, with only raw arbitrary GraphQL and deprecated `integrationSettingsUpdate` blocked | `go run ./cmd/connectorgen validate internal/connectors/defs --json` and `make verify` passed |
| #101 | Direct reads | `TestLinearCommandSurfaceRunsStreamBackedDirectRead` failed before stream-backed direct reads were runnable | Added single-object `issue/team/project/user` GraphQL view streams and stream-backed `direct_read` runner support | Focused CLI/direct-read tests and `make verify` passed |
| #102 | GraphQL/write engine | Linear write and optional GraphQL variable tests failed before fixed write mutations could omit absent record fields/default base URLs safely; source array/object test covered generated all-op variables | Added GraphQL `source` variable support and 369 fixed Linear GraphQL reverse-ETL actions covering every non-deprecated live mutation root | `TestLinearWriteActionUsesFixedGraphQLMutation`, `TestWriteGraphQLVariableSourcePreservesArraysAndObjects`, `TestLinearCommandSurfacePlansReverseETLWritePreview`, `go test ./...`, `make verify` passed |
| #103 | Sensitive/admin policy | Raw/admin/sensitive operations were not fully inventoried with executable gates | Generated mutation actions include risk text and destructive/admin/sensitive-shaped actions carry `confirm: destructive`; raw GraphQL remains the single blocked row | Help/docs/ledger expose no generic raw GraphQL/write escape hatch; connectorgen validation and make verify passed |

## Safety notes

No credentialed Linear checks were run. No secrets were requested, printed, or stored. No live Linear reverse ETL execution was performed; implemented writes are fixed-document reverse-ETL actions and local tests use dry-run previews or `httptest` only.
