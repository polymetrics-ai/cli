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
- Current implementation: Bitbucket definition bundle exists with executable reviewed streams, direct reads, approval-gated write actions, full 331-operation REST ledger, conformance fixtures, docs/catalog/website generated data, and blocked metadata for unsupported raw/local/binary/destructive/admin operations.

## Ready queue

| Issue | Lane | Dependencies | Write scope | Decision |
|---:|---|---|---|---|
| #90 | CLI surface metadata | parent plan + parent PR thread | `.planning/phases/issue-90-bitbucket-cli-surface/**`, `internal/connectors/defs/bitbucket/**`, focused tests | `local_critical_path` |
| #91 | Help renderer/docs | #90 metadata shape | docs/help renderer paths, Bitbucket defs/docs | `verified_green_local` |
| #92 | Stream runner | #90 + implemented stream definitions | commandrunner/Bitbucket defs | `verified_green_local` |
| #93 | Operation ledger | #90 seed bundle | Bitbucket `api_surface.json`, `operations.json`, tests | `verified_green_local` |
| #94 | Direct reads | #93 operation ledger + output policy design | commandrunner/direct-read policy + Bitbucket defs | `verified_green_local` |
| #95 | GraphQL/advanced engine | #93 classification proves need | Bitbucket docs/operations disposition | `verified_green_local` |
| #96 | Sensitive/admin policy | #93 write classification | operations/sensitive policy metadata, write confirmations | `verified_green_local` |

## Execution order

1. Commit parent planning artifacts as a deliberate seed diff.
2. Push `feat/79-bitbucket-cli-parity` and open a draft parent PR to `main` with `Refs #79`.
3. Execute #90 locally because this runtime does not expose the Pi `subagent` tool and #90 owns the seed bundle needed by later lanes. (Verified green slice complete.)
4. Commit/push the #90 verified slice. Complete: pushed at `0e359d76` on `feat/79-bitbucket-cli-parity`.
5. Continue ready queue after #90, preferring isolated workers/worktrees only if a runtime with mutating subagent isolation is available. Current Pi API session still has no subagent tool, so #91-#96 proceeded inline as a single verified local-critical-path implementation slice with issue-separated planning artifacts.
6. #91-#96 completed locally with red tests first, then green implementation, conformance fixtures, docs/catalog/website regeneration, and full local gates (`go vet ./...`, `go test ./...`, `go build ./cmd/pm`, `make verify`, `go run ./cmd/connectorgen validate internal/connectors/defs`).

## TDD policy

- Planning-only edits may precede production edits.
- Behavior or validation changes must add a failing test before production changes.
- For #90, first red test should assert the Bitbucket CLI surface/bundle is absent or incomplete, then green by adding validated metadata and the minimal safe seed bundle.

## CLI help/docs/website parity

Applies to connector surfaces and runtime help. #91 added connector-manual fallback help so `pm help bitbucket`, bare `pm bitbucket`, and `pm bitbucket --help` render contextual Bitbucket help successfully. Generated connector manuals/catalogs and website connector data were regenerated after the executable surface changed. Any code path that changes CLI behavior must verify:

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
