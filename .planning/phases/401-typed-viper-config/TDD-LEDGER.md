# Issue 401 TDD Ledger — Typed Viper Configuration

**Issue:** #401
**Parent:** #397
**Branch:** `feat/401-typed-viper-config`
**Sub-PR base:** `feat/cli-architecture-v2`

## Loaded skills

- `gsd-core` — `.pi/skills/gsd-core/SKILL.md`.
- `caveman` — `.agents/skills/caveman/SKILL.md`.
- `golang-how-to` — routing table: Viper config layering -> `golang-spf13-viper` + `golang-spf13-cobra`; CLI behavior -> `golang-cli`; tests -> `golang-testing`; env/filesystem/security -> `golang-security` + `golang-safety`.
- `golang-cli` — config layering, exit-code preservation, stdout/stderr discipline, injected writer tests.
- `golang-testing` — Best Practices #1 named subtests, #3 independent tests, #5 observable behavior.
- `golang-error-handling` — Best Practices #2 wrapping, #5 chain inspection, #7 single handling rule.
- `golang-security` — explicit env allowlist, no unbounded env ingestion, no secret values in docs/tests.
- `golang-documentation` — config docs/manual/website parity.
- `golang-spf13-viper` — `viper.New()` isolation, explicit `BindEnv`, optional config file, mapstructure tags, no `AutomaticEnv`/`WatchConfig`.
- `golang-spf13-cobra` — fresh command tree and current global flag binding.
- `golang-structs-interfaces` — typed config structs with explicit tags.
- `golang-safety` — no package singleton, no state leakage, safe zero/default values.
- Website-awareness only: `vercel-react-best-practices`, `vercel-composition-patterns`; no React component work planned.

Missing repo-local stack skill: `.pi/skills/go-implementation/SKILL.md` returned `ENOENT`; `.pi/skills/ts-website/SKILL.md` returned `ENOENT`. User-required Go skills above were loaded.

## GSD evidence

```bash
scripts/gsd doctor
```

Result: pass (`ok` for node, repo root, official docs, command registry, upstream lock, Pi settings/extension/skill/prompt, commands=69).

```bash
scripts/gsd prompt plan-phase 401 --skip-research >/tmp/gsd-plan-phase-401.prompt && wc -c /tmp/gsd-plan-phase-401.prompt
```

Result: pass (`10668 /tmp/gsd-plan-phase-401.prompt`).

```bash
scripts/gsd prompt programming-loop init --phase 401 --dry-run >/tmp/gsd-programming-loop-401.prompt 2>/tmp/gsd-programming-loop-401.err
```

Result: adapter gap, exit 1:

```text
scripts/gsd: unknown GSD command: programming-loop
```

Manual fallback: `.pi/prompts/pm-gsd-loop.md` loaded and followed.

## Dependency decision evidence

```bash
go list -m -versions github.com/spf13/viper
```

Result includes latest stable v1:

```text
github.com/spf13/viper ... v1.20.1 v1.21.0
```

Selected `github.com/spf13/viper v1.21.0` per ADR 0002 / task dependency gate.

## Cycle 0 — planning setup

Status: complete.

Artifacts created before production edits:

- `.planning/phases/401-typed-viper-config/PLAN.md`
- `.planning/phases/401-typed-viper-config/TDD-LEDGER.md`
- `.planning/phases/401-typed-viper-config/VERIFICATION.md`
- `.planning/phases/401-typed-viper-config/SUMMARY.md`
- `.planning/phases/401-typed-viper-config/PROMPTS.md`
- `.planning/phases/401-typed-viper-config/RUN-STATE.json`

## Cycle 1 — red tests before implementation

Planned red commands:

```bash
go test ./internal/config/... -count=1
go test ./internal/cli/ -run Config -count=1
```

Expected red:

- `internal/config` package/tests absent before implementation.
- CLI malformed-config validation tests fail until `cli.Run` loads config and maps load errors to validation exit 3.

Actual red evidence: pending.

## Cycle 2 — green implementation evidence

Pending.

## Cycle 3 — refactor / docs parity evidence

Pending.

## Cycle 4 — final verification evidence

Pending.
