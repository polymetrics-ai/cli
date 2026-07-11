# Twenty S6 CLI surface + help/manual/website parity plan (#283)

Status: IMPLEMENTED_GREEN_LOCAL_GATES_PASSED_COMMITTED_GSD_VERIFIED. Manual GSD fallback active; no subagents spawned by instruction.
Parent: #277 / parent PR #285 draft + human-gated. Base branch `feat/277-twenty-connector-parity`; worker branch `feat/283-twenty-cli-surface`. Start/base head: `62b8b46cde85ae620591194467d3474fda2a1e8a`.

## Issue contract / acceptance

Issue #283 scope: `internal/connectors/defs/twenty/cli_surface.json`, `docs/cli/**`, `website/**`, generated help/manual artifacts, and targeted CLI/help tests. Acceptance from issue/user:

- Add Twenty `cli_surface.json` with 168 commands: 28 object prefixes × `list|get|create|update|batch|delete`.
- `list`: implemented ETL stream to snake_case stream and `GET /rest/<TwentyObject>`.
- `get`: planned direct_read with `--id -> path.id`, `GET /rest/<TwentyObject>/{id}`, note generic JSON direct_read/output policy outside S6.
- `create`: implemented reverse ETL, write `create_<snake>`, commandrunner-safe top-level writable scalar flags only (`record.<field>`), plan/preview/approval risk, `POST /rest/<TwentyObject>`.
- `update`: implemented reverse ETL, write `update_<snake>`, `--id -> record.id` plus safe writable fields, `PATCH /rest/<TwentyObject>/{id}`.
- `batch`: partial reverse ETL, write `batch_<snake>`, note execution needs reverse ETL records with `records` array; no generic JSON/raw flag.
- `delete`: implemented reverse ETL, write `delete_<snake>`, `--id -> record.id`, destructive risk and typed `--confirm destructive`, `DELETE /rest/<TwentyObject>/{id}`.
- Runtime help parity: `pm help twenty`, `pm twenty`, `pm twenty --help`, and `pm connectors` bare help exit 0 without credentials; invalid actions still usage/policy errors.
- Docs parity: regenerate `docs/cli/**`, `docs/connectors/twenty/MANUAL.md`, `docs/connectors/twenty/SKILL.md`, and website generated data; second website generation idempotent.
- Tests: CLI help behavior plus optional `cli_surface` validation/count test.

## Required reading / skills loaded

- Read: `AGENTS.md`, issue #283 body/acceptance, `.agents/agentic-delivery/contracts/issue-agent-contract.md`, `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md`, `.agents/agentic-delivery/references/required-skills-routing.md`, `.agents/agentic-delivery/references/gsd-pi-adapter.md`, `.agents/agentic-delivery/references/cli-help-docs-website-parity.md`, `.planning/{config.json,PROJECT.md,ROADMAP.md,STATE.md}`, `docs/plans/universal-programming-loop-prd.md`, `docs/prompts/universal-programming-loop-prompts.md`, `docs/migration/HANDOFF-CODEX.md`, `docs/migration/conventions.md`, `docs/architecture/connector-architecture-v2-design.md`, and prior Twenty S5 phase artifacts.
- GSD skill loaded: `.pi/skills/gsd-core/SKILL.md`.
- Repo-local stack skills required by routing were checked but missing in this worktree: `.pi/skills/go-implementation/SKILL.md` and `.pi/skills/ts-website/SKILL.md` (`ENOENT`). Fallback skills loaded from available skill set: `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-documentation`, `golang-spf13-cobra`, `frontend-design`, `web-design-guidelines`, `vercel-react-best-practices`, `caveman`.
- Skill rule anchors for handoff/PR: CLI stdout/stderr + help behavior; testing rules 1/3/5; security threat model questions 1-3 + hardcoded secret rule; safety rules 2/4/6/10; error rules 1/2/7/9; docs concision/no invented context; parity checklist from CLI help/docs/website reference.

## GSD adapter path

Attempted:

```bash
scripts/gsd doctor && scripts/gsd list && scripts/gsd prompt programming-loop init --phase twenty-s6-cli-surface --dry-run
```

Result: `scripts/gsd doctor` and `scripts/gsd list` OK; `scripts/gsd prompt programming-loop ...` failed with `scripts/gsd: unknown GSD command: programming-loop` (exit 1). Manual GSD loop fallback recorded.

## Red validation evidence before production edits

```text
pm_exists=no
cli_surface_exists=no

go run ./cmd/pm help docs -> status=0
go run ./cmd/pm help connectors -> status=0
go run ./cmd/pm help twenty -> status=1; stderr: error: help topic "twenty" not found
go run ./cmd/pm twenty -> status=1; stderr: error: unknown command "twenty"; exit status 2
go run ./cmd/pm twenty --help -> status=1; stderr: error: help topic "twenty" not found
go run ./cmd/pm connectors -> status=0; rendered connectors help
```

## Implemented slices

1. **Plan/checkpoint** — created phase artifacts and recorded manual GSD fallback, skills, red evidence, and spawn decision `local_critical_path`.
2. **CLI surface JSON** — added `internal/connectors/defs/twenty/cli_surface.json` with 168 commands generated from existing streams/writes/API metadata; list implemented, get planned, create/update/delete implemented, batch partial; no generic/raw JSON write flags.
3. **CLI/help tests + code** — added red/green tests for `pm help twenty`, `pm twenty`, `pm twenty --help`, `pm connectors`, and invalid Twenty actions; implemented dynamic connector manual routing without credential resolution.
4. **Docs/manual/website artifacts** — updated `docs/cli/connectors.md`, `docs/connectors/twenty/{MANUAL.md,SKILL.md}`, `internal/connectors/defs/twenty/docs.md`, and website generated connector/catalog data.
5. **Verification** — local gates passed; `make verify` intentionally skipped because Makefile `smoke-no-build` executes reverse ETL.

## Safety / non-goals

- No live credentials, no credentialed connector checks, no reverse ETL execution.
- No generic/raw HTTP write, arbitrary JSON write flag, generic shell, or generic SQL write.
- No new dependencies; no `go.mod`/`go.sum` edits.
- `streams.json`, `writes.json`, `schemas/**`, fixtures, and `api_surface.json` unchanged.
- Do not push to `main`; parent PR #285 remains draft/human-gated.

## Execution decisions

- plan: `local_critical_path` — user instructed do not spawn subagents; worker owns one branch/cwd.
- tdd-gate: `local_critical_path` — red validation captured before production edits.
- implementation/review: `local_critical_path` — one coherent green slice on branch `feat/283-twenty-cli-surface`; push/PR after final commit.
