# TDD ledger — Monday parity wave02 r1 (#82)

## Skills used

- gsd-core (manual fallback because `scripts/gsd prompt programming-loop` returned unknown command)
- golang-how-to
- golang-cli
- golang-testing
- golang-error-handling
- golang-security
- golang-safety
- golang-documentation
- golang-design-patterns
- golang-structs-interfaces
- golang-graphql
- context-mode

## Red slice

Added `cmd/connectorgen/monday_api_surface_test.go` before production edits.

Assertions:

- Monday metadata enables write capability and keeps read/check enabled.
- `api_surface.json` is operation-ledger mode with 254 docs-sourced operation rows.
- Ledger counts: 66 planned `graphql_query` operations and 188 `graphql_mutation` operations.
- Direct reads include stream coverage for legacy read streams; fixed query commands remain planned until shared duplicate `POST /` classifier validation and GraphQL `errors[]` direct-read semantics land.
- Mutations include typed scalar GraphQL writes and planned/blocked rows for unsupported complex/binary/admin/destructive operations.
- `delete_board` exists as a destructive write action with `confirm: "destructive"`, fixed GraphQL document, no secret literals, and typed record schema.
- `cli_surface.json` exposes read/query/reverse command metadata without raw GraphQL/HTTP escape hatches.

Initial red run failed on Monday `metadata.capabilities.write=false`, confirming the bundle had no write parity surface.

## Green slice

Implemented connector-local Monday parity metadata:

- Generated docs-sourced operation ledger (`operations.json`) for 254 official public-reference GraphQL operations.
- Expanded `api_surface.json` into a full operation ledger with 5 stream reads, 102 typed scalar GraphQL write actions, and 147 planned/blocked query/complex/binary/admin/destructive rows.
- Added `writes.json` with fixed GraphQL mutation documents, draft-07 record schemas, and `confirm: "destructive"` for delete/archive/clear/remove-style or admin/destructive actions.
- Added `cli_surface.json` command metadata for `stream`, `query`, and `reverse` groups.
- Updated Monday `metadata.json`, `docs.md`, Monday connector manual/skill, connector catalog, and CLI golden transcripts.
- Added fixture evidence for `update_board` and destructive `delete_board` under `internal/connectors/defs/monday/fixtures/writes/`.
- Appended an idempotent captain-policy addendum to parent #82 and subissues #111-#117 using `gh-axi`; verified one marker per issue.

## Guardrails preserved

- No live `https://api.monday.com/v2/get_schema` call; no monday.com credentials or provider writes.
- No generic GraphQL, generic HTTP, or generic SQL write command exposed.
- Reverse ETL remains plan -> preview -> explicit approval -> execute; destructive/admin writes add typed `destructive` confirmation.
- Complex GraphQL arrays/objects and file/binary mutations that cannot be safely represented by the current scalar variable contract or multipart upload handling are planned/blocked with source evidence.
