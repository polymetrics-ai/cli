# Lever Hiring connector parity plan (#3223)

## Scope

Implement connector-local Lever Hiring official documentation parity on branch `fm/cli-lever-hiring-parity-wave05-r1`.

Allowed paths for this slice:

- `internal/connectors/defs/lever-hiring/**`
- `docs/connectors/lever-hiring/**`
- generated connector docs/catalog surfaces touched by `pm docs generate` when needed
- `.planning/phases/issue-3223-lever-hiring-parity-wave05-r1/**`

Non-goals and hard gates:

- No live Lever calls, credentials, provider writes, or certification claims.
- No new dependencies.
- No shared runtime/engine/CLI Go edits; unsupported behavior is recorded as a connector-local blocked/planned row with official-source evidence.
- No generic HTTP method/path/body/query passthroughs.
- Reverse ETL remains plan -> preview -> explicit approval -> execute.

## GSD command path and fallback

- Ran `scripts/gsd doctor`: passed.
- Ran `scripts/gsd list`: passed.
- Attempted `scripts/gsd prompt programming-loop init --phase issue-3223-lever-hiring-parity-wave05-r1 --dry-run`: failed with `scripts/gsd: unknown GSD command: programming-loop`.
- Manual GSD fallback is active using `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md`, `.opencode/commands/gsd-programming-loop.md`, `.pi/prompts/gsd.md`, and `docs/prompts/universal-programming-loop-prompts.md`.

## Required skills loaded

- `gsd-core`
- `golang-how-to`
- `golang-design-patterns`
- `golang-structs-interfaces`
- `golang-error-handling`
- `golang-security`
- `golang-safety`
- `golang-testing`
- `golang-cli`
- `golang-documentation`
- `golang-context`
- `golang-concurrency`
- `golang-lint`
- `context-mode`
- CLI help/manual/docs parity reference

## Source inventory

Official source for this slice:

- Lever Developer documentation: `https://hire.lever.co/developer/documentation`.
- Current fetched evidence: 107 documented HTTP operation rows after excluding webhook receiver examples, plus 10 documented webhook trigger/event names, for 117 official operations total.

Parent and child issue counts are dispatch inputs; post-change counts must come from the regenerated `api_surface.json` ledger.

## Implementation slices

1. **Operation ledger parity**
   - Regenerate `api_surface.json` from the current official docs.
   - Include each documented HTTP operation and webhook trigger exactly once.
   - Mark implemented rows with `covered_by`; mark unsupported rows with `operation` blocked/planned evidence instead of exclusions unless a row is a genuine duplicate/deprecation/non-data endpoint.

2. **Read and direct-read expansion**
   - Keep existing fixture-backed streams green.
   - Add safe stream definitions only when the current declarative contract can represent the operation without path/body/query passthroughs.
   - Add bounded direct-read operations/commands for fixed-target detail/query reads where the existing direct-read contract supports path variables, closed command flags, byte caps, and JSON redaction.
   - Block scalar-list, binary, webhook/CDC, or unsupported shapes with exact shared-runtime dependency evidence.

3. **Typed write expansion**
   - Add named Lever write actions for documented product-safe mutations supported by the generic write executor.
   - Use closed `record_schema` definitions, path fields, risk text, and destructive confirmation/idempotency evidence for delete/destructive operations.
   - Block multipart/file upload and webhook/CDC registration lifecycle when the current contract lacks a bounded transfer/CDC foundation or provider-supported idempotency evidence.

4. **CLI/docs/generated surfaces**
   - Add Lever-owned `cli_surface.json` for implemented stream/direct-read/write commands.
   - Regenerate connector manual/skill docs if command or capability surfaces change.
   - Update `docs.md` with truthful implemented/blocked/certification counts.

5. **Verification and issue update**
   - Run focused inventory/validation, conformance, CLI/golden/docs checks, build, connector-boundary, `git diff --check`, and `make verify`.
   - Update #3223 and #3224-#3230 once through `gh-axi` with final counts and verification evidence.
   - Commit the clean locally tested result; do not push or open/update a PR.

## Orchestration decision

`local_critical_path`: this worker owns a single connector-local tree in one isolated worktree. Mutating subagents were not spawned because all subissues share `internal/connectors/defs/lever-hiring/**`, so parallel mutation would collide. Read-only recon is performed inline with context-mode.
