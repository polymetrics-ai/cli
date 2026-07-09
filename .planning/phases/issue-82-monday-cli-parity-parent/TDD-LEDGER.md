# TDD Ledger — issue #82 Monday CLI parity parent

## GSD/TDD mode

Manual programming-loop fallback is active because `scripts/gsd prompt programming-loop init --phase issue-82-monday-cli-parity --dry-run` is unavailable in the repo-local command registry. Red → green → refactor evidence remains mandatory for each behavior-changing lane.

## Parent ledger

| Lane | Issue | Red evidence | Green evidence | Refactor/notes |
| --- | ---: | --- | --- | --- |
| CLI surface metadata | #111 | Captured: Monday `cli_surface.json` load tests failed before metadata existed. | Captured: added `cli_surface.json`; targeted engine/connectorgen tests and full `connectorgen validate internal/connectors/defs --json` pass. | First local critical-path lane green. |
| Help renderer/docs | #112 | Captured: Monday manual test failed because overlapping groups rendered `board list` twice. | Captured: regrouped command surface and docs; connector manual/render tests pass. | Runtime connector manual is `pm connectors inspect monday`; dynamic `pm help monday` is not a CLI route. |
| Stream runner | #113 | Honest test-only verification: runner was already enabled by #111 metadata. | Captured: `TestRunMondayBoardListCommand` passes against local GraphQL replay server. | No credentialed live checks. |
| Operation ledger | #114 | Captured: legacy `api_surface.json` failed `operation_ledger_version = 0, want 1`. | Captured: operation ledger + operations.json cover 367 operations; 87 query rows, 280 mutation rows; validation clean. | Implemented streams remain executable; other rows blocked metadata. |
| Direct reads | #115 | Captured: missing implemented `me view` direct-read command. | Captured: fixed `me view` and `account view` direct reads execute bundled GraphQL query docs through local replay tests. | No raw GraphQL/HTTP escape hatch. |
| GraphQL engine | #116 | Captured: GraphQL direct-read tests failed to compile before `DirectReadRequest.Operation`. | Captured: fixed-operation `graphql_query` direct-read path with `graphql_json` output policy; GraphQL errors fail closed. | Mutations rejected; no user-supplied GraphQL documents. |
| Sensitive/admin policy | #117 | Captured: policy test failed because docs lacked `Sensitive/admin mutation policy`. | Captured: mutation policy docs plus blocked metadata tests pass. | 280 mutations remain blocked by default with typed confirmation metadata for high/critical/admin/secret/delete classes. |

## Current red target

All #111-#117 red/green slices are captured; current target is parent PR verification/review readiness.
