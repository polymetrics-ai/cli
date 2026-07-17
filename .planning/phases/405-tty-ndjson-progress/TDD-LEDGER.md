# TDD LEDGER — Issue 405 TTY gate and NDJSON progress

## Loaded skills

- `gsd-core` — repo-local GSD adapter workflow.
- `caveman` — compact final handoff only.
- `golang-how-to` — required Go skill router.
- `golang-cli` — CLI flags, stdout/stderr discipline, run options.
- `golang-spf13-cobra` — current router/help flag integration.
- `golang-testing` — red/green table tests and CLI contract tests.
- `golang-error-handling` — validation errors and existing exit taxonomy.
- `golang-security` — no secrets in progress/logs/docs; terminal sanitization preserved.
- `golang-safety` — deterministic defaults, nil writer/env handling, no panic-prone assumptions.
- `golang-documentation` — runtime help, docs/cli, website parity.
- UI/design source docs: `docs/design/tui-ux-design.md`, `docs/adr/0003-interactive-tui-layer.md`.

Stack implementation skill note: `.pi/skills/go-implementation/SKILL.md` was requested by worker instructions but is absent in this checkout (`ENOENT`); loaded `gsd-core` plus required global Go skills from `.agents/agentic-delivery/references/required-skills-routing.md` instead.

## GSD command evidence

```bash
scripts/gsd doctor
```

Result: pass.

```bash
scripts/gsd list
```

Result: pass; 69 commands listed.

```bash
scripts/gsd prompt plan-phase 405 --skip-research
```

Result: generated official `/gsd-plan-phase 405 --skip-research` prompt.

```bash
scripts/gsd prompt programming-loop init --phase 405 --dry-run
```

Result: fail, adapter gap: `scripts/gsd: unknown GSD command: programming-loop`.

Fallback: loaded `.pi/prompts/pm-gsd-loop.md` and running the universal programming loop inline/manual; decision `local_critical_path`.

## Red / Green ledger

| Slice | Test / validation | Red evidence | Green evidence | Refactor evidence |
|---|---|---|---|---|
| 1 ui detection/styles | `go test ./internal/ui/... -count=1` | pending — write failing tests before production code | pending | pending |
| 2 CLI run options/global flags | `go test ./internal/cli/... -run 'TestRunWithOptions|TestGlobalUI|TestProgress' -count=1` | pending — write failing tests before production code | pending | pending |
| 3 stderr-only NDJSON | `go test ./internal/cli/... -run TestProgressNDJSON -count=1` | pending — write failing test before production code | pending | pending |
| 4 docs/help parity | `go test ./internal/cli/... -run 'Test.*Help|TestGoldenDocsGenerateMatchesTrackedCLIManuals' -count=1` plus docs/website grep | pending — validation should fail until docs updated | pending | pending |
| final gates | `gofmt -w cmd internal`; `go vet ./...`; `go test ./...`; `go build ./cmd/pm`; `make verify` | pending | pending | pending |

## Red test capture rule

Before production edits, add focused failing tests only. Capture exact command and failure output here before implementing each slice.

## Review disposition ledger

No automated review findings yet. Open stacked PR after local gates; record Claude/Copilot/human route status here and in PR body.
