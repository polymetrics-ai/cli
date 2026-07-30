# GSD plan — issue #80 Linear parity wave01-r1

## Context

- Branch: `fm/cli-linear-parity-wave01-r1` in disposable worktree.
- GSD adapter: `scripts/gsd doctor` passed. `scripts/gsd prompt gsd-plan-phase issue-80-linear-parity --skip-research` generated `.planning/traces/issue-80-linear-parity-gsd-plan-phase-prompt.md`.
- Programming-loop command fallback: `scripts/gsd prompt programming-loop init --phase issue-80-linear-parity --dry-run` is unavailable in this adapter (`unknown GSD command: programming-loop`), so this phase uses the manual GSD/TDD loop required by AGENTS.md.
- Required skills loaded: `gsd-core`, `golang-how-to`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-testing`, `golang-graphql`, `golang-cli`, `golang-documentation`; also read CLI help/docs parity guidance and connector v2 migration conventions/design.

## Scope

Primary production paths are connector-local Linear artifacts: `internal/connectors/defs/linear/**` including schemas, fixtures, docs, operation ledger, command metadata, and certification metadata when needed. Minimal shared CLI help-topic rendering and golden/test updates are allowed only for provider CLI/help parity. No shared runtime, hook, native, provider-call, credential, merge, or default-branch work.

## Source inventory

- Parent/subissues: #80, #97-#103. Captain addendum appended idempotently with `gh-axi` to include destructive/delete operations with typed confirmation and preserve existing counts.
- Official pinned Linear schema blob from parent: `3934265499c95f1d6b8e4d5c695ad0b6f1d52fec` (`packages/sdk/src/schema.graphql`). Raw blob retrieval used `gh api` because `gh-axi api` truncates raw blobs after 4000 chars; all issue mutations used `gh-axi`.
- The connector-local generated evidence must not claim live certification or changed implemented counts unless exact local artifacts prove them.

## Implementation slices

1. Generate/refresh Linear operation inventory from the pinned GraphQL schema.
   - Parse root `Query`, `Mutation`, and `Subscription` fields.
   - Classify executable connector-local streams/writes only where the current declarative engine can express fixed GraphQL documents safely.
   - Classify unsupported direct/binary/CDC/remaining writes as operation-ledger blocked rows, not legacy unsafe exclusions.
2. Update Linear bundle files.
   - `metadata.json`: truthful read/write capabilities and risk text.
   - `streams.json`: fixed GraphQL documents for streamable Query list/connection fields; no raw GraphQL escape hatch.
   - `writes.json`: typed fixed GraphQL mutations for generated write actions, with `confirm: "destructive"` on delete/archive/destructive/admin-style actions.
   - `api_surface.json`: operation ledger mode with every inventoried official root operation partitioned exactly once.
   - `operations.json`/`cli_surface.json`/`certification.json`: bounded docs/command/cert evidence without enabling unsupported generic surfaces.
   - Schemas and sanitized fixtures for representative stream and write/delete conformance.
3. Validate connector-local artifacts.
   - `go run ./cmd/connectorgen validate internal/connectors/defs --json` (the validator expects a defs root; pointing it at `defs/linear` treats `schemas/` and `fixtures/` as connector dirs).
   - `go test ./internal/connectors/conformance -run 'TestConformance/linear' -count=1`
   - targeted engine/commandrunner tests if validation indicates connector metadata regressions.
4. Commit one coherent checkpoint. If shared foundations prevent complete parity, record the exact dependency and stop as blocked rather than claiming complete certification/full parity.

## Risks and constraints

- No live Linear credentials or provider calls.
- No generic GraphQL/method/path/body commands.
- Destructive operations remain in scope only as typed actions/operation rows with plan → preview → explicit approval → execute and `destructive` confirmation.
- Do not edit shared runtime files to close any missing foundation.
