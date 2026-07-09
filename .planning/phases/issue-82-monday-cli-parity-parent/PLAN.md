# Plan — issue #82 Monday CLI parity parent

## Scope

Parent issue: #82 (`feat/82-monday-cli-parity`). Sub-issues: #111-#117. Connector definition scope: `internal/connectors/defs/monday/`; shared runner/schema changes only when required by Monday safety-gated parity.

## GSD mode

- Adapter health passed: `scripts/gsd doctor`, `scripts/gsd verify-pi`, `scripts/gsd list --json`.
- Planning prompt generated with `scripts/gsd prompt plan-phase issue-82-monday-cli-parity --skip-research`.
- Programming-loop command is unavailable in the current registry (`scripts/gsd prompt programming-loop ...` exits `unknown GSD command: programming-loop`). Manual GSD programming-loop fallback is active and recorded here; TDD/red-first evidence is still required before production edits.

## Skills

Loaded: `gsd-core`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-context`, `golang-concurrency`, `golang-graphql`, `golang-documentation`, `golang-spf13-cobra`, `golang-lint`.

## Parent orchestration state

- Parent branch: `feat/82-monday-cli-parity` (current worktree).
- Parent PR: draft PR #130 (`https://github.com/polymetrics-ai/cli/pull/130`).
- Worker spawning: not spawned in this harness because no Pi `subagent` tool is available in the active tool set. Decision for this cycle: `local_critical_path`; #111-#117 were completed sequentially in this worktree with issue-specific artifacts.
- Human gates: parent PR merge to `main`, auth-scope changes, secrets, new dependencies, destructive external actions, production deploys, quality-gate reductions, generic write tools, credentialed connector checks, and reverse ETL execution.

## Lane order and dependencies

1. **#111 CLI surface metadata** — add Monday `cli_surface.json`, red tests for metadata load/validation and implemented stream commands.
2. **#112 help/docs parity** — update Monday connector docs and help-rendered metadata; verify runtime help surfaces.
3. **#113 stream runner** — prove implemented stream commands run through the generic connector command runner using fixture/mock runtime, no credentials.
4. **#114 operation ledger** — replace legacy 8-row `api_surface.json` with operation-ledger mode covering the 367 official GraphQL operations (87 query, 280 mutation) from the canonical monday reference pages.
5. **#115 direct reads** — expose bounded safe `me view` and `account view` direct reads through fixed bundled GraphQL query documents. No raw GraphQL/HTTP escape hatches.
6. **#116 GraphQL/advanced engine** — fixed-document `graphql_query` direct-read support only; mutations rejected and GraphQL errors fail closed.
7. **#117 sensitive/admin policy** — classify mutation risks, redaction/input modes, typed confirmation requirements, and blocked-by-default policy for sensitive/admin/destructive operations.

## Safety rules

- No secrets in prompts, logs, artifacts, fixtures, or docs.
- No credentialed connector checks unless explicitly requested.
- No generic raw HTTP write, arbitrary GraphQL mutation, generic shell write, or generic SQL write tools.
- Binary/file operations stay metadata-only unless a bounded executor and output policy exist.
- Reverse ETL execution remains plan → preview → approval → execute with typed confirmation for destructive/sensitive operations.

## Commit checkpoints

1. Parent GSD plan seed.
2. #111 red tests.
3. #111 green metadata/docs slice.
4. Subsequent lane red/green slices, one coherent commit each.
5. Final verification and parent PR readiness update.
