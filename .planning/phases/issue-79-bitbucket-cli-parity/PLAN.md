# Plan: Bitbucket CLI Parity Parent Orchestration

Parent issue: #79
Parent branch: `feat/79-bitbucket-cli-parity`
Default branch: `main`
Parent PR: https://github.com/polymetrics-ai/cli/pull/128 (draft). Initial checks/review pending.

## GSD command path

- `scripts/gsd doctor` — passed.
- `scripts/gsd verify-pi` — passed.
- `scripts/gsd list --json` — passed; registry contains 69 commands.
- `scripts/gsd prompt plan-phase issue-79-bitbucket-cli-parity --skip-research --tdd` — prompt generated and followed.
- `scripts/gsd prompt programming-loop init --phase issue-79-bitbucket-cli-parity --dry-run` — unavailable (`scripts/gsd: unknown GSD command: programming-loop`).
- Manual fallback: use `.pi/prompts/pm-gsd-loop.md` plus `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md`; record fallback in this phase and each sub-issue phase. TDD remains mandatory.

## Required skills loaded

- `gsd-core`
- `golang-how-to`
- `golang-cli`
- `golang-testing`
- `golang-error-handling`
- `golang-security`
- `golang-safety`
- `golang-design-patterns`
- `golang-structs-interfaces`
- `golang-documentation`
- `golang-context`
- `golang-concurrency`
- `golang-graphql`
- `golang-lint`
- `golang-spf13-cobra`
- `caveman` for compact orchestration notes

## Parent objective

Deliver Bitbucket connector CLI parity across metadata, help/docs, stream-backed command dispatch, operation ledger, direct reads, GraphQL/advanced body support if needed, and sensitive/admin policy. Preserve connector safety: no secrets in prompts/logs/files; no generic raw HTTP write, shell write, or SQL write; reverse ETL remains plan → preview → approval → execute; destructive/admin actions remain blocked by default.

## Source of truth

- Parent issue #79 and sub-issues #90-#96.
- User instruction overrides issue-body branch typo: use `feat/79-bitbucket-cli-parity` as the parent integration branch.
- Official Bitbucket Swagger: `https://api.bitbucket.org/swagger.json`.
- Expected official operation count: 331 operations (GET 179, POST 50, PUT 48, DELETE 54).
- Definition scope: `internal/connectors/defs/bitbucket/`.
- Current baseline: no Bitbucket definition bundle exists.

## Ready queue

| Issue | Lane | Dependencies | Write scope | Decision |
|---:|---|---|---|---|
| #90 | CLI surface metadata | parent plan + parent PR thread | `.planning/phases/issue-90-bitbucket-cli-surface/**`, `internal/connectors/defs/bitbucket/**`, focused tests | `local_critical_path` |
| #91 | Help renderer/docs | #90 metadata shape | docs/help renderer paths, Bitbucket defs/docs | `not_spawned_dependency_blocked` |
| #92 | Stream runner | #90 + implemented stream definitions | commandrunner/Bitbucket defs | `not_spawned_dependency_blocked` |
| #93 | Operation ledger | can run after parent PR; overlaps `api_surface.json` with #90 | Bitbucket `api_surface.json`, optional tests | `not_spawned_write_scope_collision` while #90 creates the seed bundle |
| #94 | Direct reads | #93 operation ledger + output policy design | commandrunner/Bitbucket defs | `not_spawned_dependency_blocked` |
| #95 | GraphQL/advanced engine | #93 classification proves need | engine/Bitbucket defs | `not_spawned_dependency_blocked` |
| #96 | Sensitive/admin policy | #93 write classification | operations/sensitive policy metadata, validation | `not_spawned_dependency_blocked` |

## Execution order

1. Commit parent planning artifacts as a deliberate seed diff.
2. Push `feat/79-bitbucket-cli-parity` and open a draft parent PR to `main` with `Refs #79`.
3. Execute #90 locally because this runtime does not expose the Pi `subagent` tool and #90 owns the seed bundle needed by later lanes. (Verified green slice complete.)
4. Commit/push the #90 verified slice. Complete: pushed at `0e359d76` on `feat/79-bitbucket-cli-parity`.
5. Continue ready queue after #90, preferring isolated workers/worktrees only if a runtime with mutating subagent isolation is available.

## TDD policy

- Planning-only edits may precede production edits.
- Behavior or validation changes must add a failing test before production changes.
- For #90, first red test should assert the Bitbucket CLI surface/bundle is absent or incomplete, then green by adding validated metadata and the minimal safe seed bundle.

## CLI help/docs/website parity

Applies to connector surfaces and future runtime help. #90 is metadata-only and records runtime command help as not-yet-executable, while adding generated connector manual/catalog and website data for the new bundle. Later #91 owns rendered Bitbucket runtime help/docs. Any code path that changes CLI behavior must verify:

- `pm help <topic>` where applicable.
- `pm <namespace>` bare behavior where applicable.
- `pm <command> --help` where applicable.
- `docs/cli/**` and `website/**` parity or explicit exemption.

## Human gates

- Parent PR merge to `main`.
- Auth-scope changes or `gh auth refresh`.
- Secrets or credentialed connector checks.
- New dependencies.
- Destructive/admin external actions.
- Production deploys.
- Quality gate reductions.
- Generic shell/HTTP/SQL write tooling.
- Reverse ETL execution outside plan → preview → approval → execute.
