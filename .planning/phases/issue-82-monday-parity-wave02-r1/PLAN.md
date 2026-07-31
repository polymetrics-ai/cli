# Plan — Monday connector parity wave02 r1 (#82)

## Scope

Implement documented Monday connector parity inside connector-owned surfaces only: `internal/connectors/defs/monday`, Monday fixtures, tests, generated/manual docs, CLI metadata, and planning artifacts. No shared runtime/foundation edits, live provider calls, credentials, writes, certification, pushes, merges, or PRs.

## Required workflow notes

- Worktree isolation verified before branch creation; branch `fm/cli-monday-parity-wave02-r1` created.
- `no-mistakes doctor` passed; no init required.
- Repo-local GSD adapter is present, but `scripts/gsd prompt programming-loop ...` returned `unknown GSD command`; manual GSD/TDD fallback used and recorded.
- Required skills loaded: gsd-core, Go how-to/CLI/testing/error/security/safety/docs/design/structs/GraphQL, and context-mode.

## Implementation plan

1. Inventory issue graph (#82, #111-#117) and official Monday public docs without live monday.com provider calls.
2. Add issue-body captain-policy addendum clarifying destructive/admin operations are in scope only behind typed schemas, reverse ETL approval, and destructive confirmation.
3. Add red test for Monday operation-ledger parity, destructive typed write, and CLI metadata.
4. Generate/author Monday connector-local metadata:
   - `metadata.json` write capability and safety notes.
   - `api_surface.json` full docs operation ledger.
   - `operations.json` read/mutation operation metadata.
   - `writes.json` fixed GraphQL reverse-ETL actions with typed record schemas.
   - `cli_surface.json` stream/query/reverse command metadata.
   - fixture evidence for representative update and destructive delete.
   - Monday docs/manual/skill/catalog/golden help updates.
5. Verify connector validation, conformance, CLI docs/help/golden tests, vet/build, connector boundary, diff check.
6. Commit implementation locally, append final status line, and stop for firstmate/no-mistakes/PR handoff.

## Results

- Public docs inventory found 254 official GraphQL operations: 66 queries and 188 mutations.
- Implemented ledger representation for all 254 docs operations.
- Executable with current safe scalar contracts: 5 existing stream reads, 51 bounded scalar direct-read query commands, 102 scalar-input GraphQL write actions.
- Planned/blocked with evidence: 10 query operations and 86 mutation operations requiring complex arrays/objects, binary/file handling, or live-schema/source work outside this worker scope.
- Destructive/admin executable writes (for example `delete_board`) require the existing reverse ETL safety path plus `confirm: "destructive"`.
- Parent issue's r3 count of 292 live-schema operations is preserved in docs/planning as source-blocked reconciliation; this worker did not call `/v2/get_schema`.

## Stop conditions

- Do not run no-mistakes validation loop, push, open PR, merge, certify, or run live provider checks.
- Stop after local implementation commit and status append.
