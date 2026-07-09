# TDD Ledger — issue #82 Monday CLI parity parent

## GSD/TDD mode

Manual programming-loop fallback is active because `scripts/gsd prompt programming-loop init --phase issue-82-monday-cli-parity --dry-run` is unavailable in the repo-local command registry. Red → green → refactor evidence remains mandatory for each behavior-changing lane.

## Parent ledger

| Lane | Issue | Red evidence | Green evidence | Refactor/notes |
| --- | ---: | --- | --- | --- |
| CLI surface metadata | #111 | Captured: Monday `cli_surface.json` load tests failed before metadata existed. | Captured: added `cli_surface.json`; targeted engine/connectorgen tests and full `connectorgen validate internal/connectors/defs --json` pass. | First local critical-path lane green. |
| Help renderer/docs | #112 | Pending | Pending | CLI help/docs/website parity applies. |
| Stream runner | #113 | Pending | Pending | No credentialed live checks. |
| Operation ledger | #114 | Pending | Pending | Count assertions: 367 official operations; 87 queries, 280 mutations; implemented streams 5. |
| Direct reads | #115 | Pending | Pending | Bounded safe direct-read only; no raw GraphQL/HTTP escape hatch. |
| GraphQL engine | #116 | Pending | Pending | Existing fixed-document support may make this metadata/tests-only. |
| Sensitive/admin policy | #117 | Pending | Pending | Blocked-by-default + typed confirmation metadata. |

## Current red target

#111 will add tests before production metadata edits.
