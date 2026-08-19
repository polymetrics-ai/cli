# GitLab provider-inventory parity — G0 + G1 plan

## Scope and ownership

- Branch: `fm/cli-gitlab-parity-implementation-r1` in the isolated treehouse worktree.
- Production scope: `internal/connectors/defs/gitlab/**`, its connector-owned generated/manual documentation, and website catalog output generated from that bundle.
- No shared engine, command runner, schema, dependency, credential, live-provider, or other-connector changes.
- Authoritative inventory: the GitLab OpenAPI v3 `19.3.0-pre` artifact, retrieved 2026-08-05, SHA-256 `7aab00db0124f7b6addcb1608c2c84fe76215cfedaf6036a2d705696d837bde0`; its checked-in provider-owned ledger is `internal/connectors/defs/gitlab/api_surface.json`.

## GSD delivery trace

- `scripts/gsd doctor`, `scripts/gsd list`, all required `scripts/gsd sources` commands, and `go run ./cmd/agentcontractgen check` passed.
- `scripts/gsd prompt discuss-phase cli-gitlab-parity-implementation-r1 --auto` was rendered and followed inline.
- Manual inline fallback: this connector-only task is not a registered phase in `.planning/ROADMAP.md`; registering it would mutate shared roadmap/state outside the assigned connector scope. The supplied report locks the decisions, so this directory records the equivalent plan, TDD ledger, verification, and review evidence without changing shared planning state.

## Required skills and references

- `golang-how-to`, `golang-cli`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-testing`, `golang-documentation`, and `golang-lint`.
- `docs/migration/HANDOFF-CODEX.md`, `docs/migration/conventions.md`, `docs/architecture/connector-architecture-v2-design.md`, and `.agents/agentic-delivery/references/cli-help-docs-website-parity.md`.

## Locked inventory decisions

- Land the corrected 1,745-row provider-owned `api_surface.json` first as a commit containing only that file.
- Preserve report classifications exactly: 4 executable stream reads, 1,618 implementable-now rows blocked pending GitLab-owned declarations, 64 provider-restriction rows, 45 rows blocked on the named operation-level multipart/file-upload executor, and 14 deprecated justified exclusions.
- G1 adds definition-owned CLI/help metadata only for existing `projects`, `groups`, `users`, and `issues` streams. It must not add provider operations, writes, generic request surfaces, credentials, or live calls.
- `capabilities.write` remains `false` until a GitLab write action is executable; metadata records the provider mutation surface in `risk.write` without advertising an executable write.
- No output redaction declaration is added merely to satisfy a validator. There are no direct-read or write commands in G1.

## TDD slices

1. **G0: provider ledger**
   - Replace only `api_surface.json` from the report's fenced canonical JSON and commit it separately.
   - Parse/count the file locally and validate the GitLab bundle after metadata is updated in the next slice.
2. **G1: four ETL commands**
   - Record the red baseline: built `pm gitlab projects list --help` reports an unknown command because the bundle has no `cli_surface.json`.
   - Add one implemented `etl` command for each existing stream, each with exactly one matching provider endpoint, bounded `--limit`, and safe credential/config references only.
   - Update GitLab metadata/docs and regenerate connector manual/website catalog output through the repository generator.
   - Build `./cmd/pm` and run all four command help paths plus the GitLab namespace help; inspect command preflight through the real built binary.

## Expected commits

1. `chore(gitlab): rebase provider operation ledger to OpenAPI v3` — `api_surface.json` only.
2. `docs(gitlab): publish existing stream commands` — G1 metadata, CLI surface, docs, generated documentation/catalog, and planning evidence.

## Completion boundary

Stop after G1. The next implementation wave is R1, a collaboration-read slice of at most 20 public GET/HEAD operations from Issues, Merge requests, Discussions, Notes, Award emoji, Labels, and Milestones.

## Execution record

- G0 landed first in `62c115ce5` (`chore(gitlab): rebase provider operation ledger to OpenAPI v3`), with `api_surface.json` as its only changed file.
- `fm-ensure-agents-md.sh .` was run as required. It reported that both `AGENTS.md` and `CLAUDE.md` are real files and must be reconciled manually; no durable cross-project instruction was added for this connector-only change.
- G1 adds `cli_surface.json` only for the four pre-existing streams and regenerates connector manuals and the website catalog from the bundle. It deliberately stops before R1.
