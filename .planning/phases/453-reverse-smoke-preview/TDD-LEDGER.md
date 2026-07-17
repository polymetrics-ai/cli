# Phase 453 TDD Ledger

Issue: #453 — require reverse preview in smoke gate.

## Skills loaded

`gsd-core`, `caveman`, `golang-how-to`, `golang-testing`, `golang-cli`, `golang-security`, `golang-safety`, `golang-error-handling`.

Repo skill gap: `.pi/skills/go-implementation/SKILL.md` was required by worker instructions but is absent in this checkout (`ENOENT`); global Go skills above are loaded and used.

Rule anchors:

- `required-skills-routing`: Go work always loads `golang-how-to`; CLI/command behavior loads `golang-cli`, `golang-testing`, `golang-error-handling`, and `golang-security` for args/filesystem/external IO.
- `golang-how-to`: tests route to `golang-testing`; CLI/args/I/O route to `golang-cli`, `golang-security`, and `golang-safety`.
- `golang-testing`: #1 named table/subtests when table-driven, #3 independent tests, #5 observable behavior/public contract over fragile implementation details.
- `golang-cli`: preserve stdout/stderr discipline and machine-readable output; test CLI/scripted behavior without corrupting stdout.
- `golang-security`: trust-boundary questions #1-#3; no secrets in logs/artifacts; treat command args and filesystem paths as untrusted.
- `golang-safety`: avoid panic-prone assumptions and unsafe resource handling.
- `golang-error-handling`: #1 check errors, #2 context on propagated errors, #7 log-or-return not both, #9 no panic for expected errors.

## GSD command evidence

```bash
scripts/gsd doctor
scripts/gsd list
scripts/gsd prompt plan-phase 453 --skip-research >/tmp/gsd-plan-phase-453.prompt
scripts/gsd prompt programming-loop init --phase 453 --dry-run >/tmp/gsd-programming-loop-453.prompt
```

Result:

- `doctor`: pass.
- `list`: pass; 69 commands listed.
- `plan-phase`: prompt written to `/tmp/gsd-plan-phase-453.prompt`.
- `programming-loop`: blocked by adapter registry (`scripts/gsd: unknown GSD command: programming-loop`); manual GSD fallback active using `.pi/prompts/pm-gsd-loop.md` + universal runtime loop.

## Red / green / refactor log

| Step | Kind | Command / test | Result | Notes |
|---:|---|---|---|---|
| 0 | Planning | Create PLAN/TDD-LEDGER/VERIFICATION/SUMMARY/RUN-STATE/PROMPTS | Green | Pre-production artifact checkpoint; no production code touched. |
| 1 | Red | `go test ./internal/safety -run TestSmokeNoBuildReversePlanPreviewRunOrdering -count=1` | Fail | Fails before Makefile edit because `smoke-no-build` lacks `pm reverse preview`. |
| 2 | Green | `gofmt -w cmd internal`; `go test ./internal/safety -run TestSmokeNoBuildReversePlanPreviewRunOrdering -count=1` | Pass | Makefile preview inserted; focused ordering green before any `make smoke` or `make verify`. |
| 3 | Full gate | Issue-required gates | Pass | Standalone gates and `make verify` passed after focused green; smoke ran local temp outbox only and included preview before run. |

## Planned red test

- `TestSmokeNoBuildReversePlanPreviewRunOrdering`: reads the repository `Makefile`, extracts the `smoke-no-build` recipe, and asserts the safety ordering `reverse plan` output → `PLAN_ID=` extraction → `reverse preview` with same temp root and JSON sink → `APPROVAL=` extraction → `reverse run`.

## Exact red output

```bash
go test ./internal/safety -run TestSmokeNoBuildReversePlanPreviewRunOrdering -count=1
```

```text
--- FAIL: TestSmokeNoBuildReversePlanPreviewRunOrdering (0.00s)
    smoke_makefile_test.go:30: smoke-no-build recipe missing reverse preview step "./pm reverse preview \"$$PLAN_ID\" --root \"$$SMOKE_DIR\" --json >/dev/null;"
FAIL
FAIL\tpolymetrics.ai/internal/safety\t0.438s
FAIL
```

## Exact green output

```bash
gofmt -w cmd internal && go test ./internal/safety -run TestSmokeNoBuildReversePlanPreviewRunOrdering -count=1
```

```text
ok  \tpolymetrics.ai/internal/safety\t0.161s
```

## Final gate outputs

```bash
gofmt -w cmd internal
```

Result: pass; no output.

```bash
go test ./...
```

Result: pass. Slow package examples: `ok  \tpolymetrics.ai/internal/cli\t167.313s`, `ok  \tpolymetrics.ai/internal/connectors/certify\t338.571s`, `ok  \tpolymetrics.ai/internal/safety\t3.167s`.

```bash
go vet ./...
go build ./cmd/pm
```

Result: pass; no output.

```bash
make smoke
```

Result: pass; local temp smoke printed `smoke ok: <tmpdir>`. The recipe echo included the new `./pm reverse preview "$$PLAN_ID" --root "$$SMOKE_DIR" --json >/dev/null;` step before approval extraction and run.

```bash
make verify
```

Result: pass; covered `gofmt -w cmd internal`, `go mod tidy` + go.mod/go.sum diff check, `go vet ./...`, `go test -timeout 20m ./...`, `go build ./cmd/pm`, docs validation, local temp smoke, `golangci-lint`, and `connectorgen validate`. Terminal tail:

```text
0 issues.
go run ./cmd/connectorgen validate internal/connectors/defs
connectorgen validate: 547 connector(s) checked, 0 findings
```

```bash
git diff --check origin/feat/cli-architecture-v2...HEAD
git diff -- go.mod go.sum
```

Result: pass; no output.

## Review-fix TDD ledger — PR #454 MEDIUM finding (2026-07-17)

Accepted disposition: raw substring matching in `internal/safety/smoke_makefile_test.go` is too weak because comments, `echo`/`printf`/help text, unrelated targets, or prefixed shell statements can satisfy the current assertions without an executable preview command in `smoke-no-build`.

Skills loaded this run: `gsd-core`, `caveman`, `golang-how-to`, `golang-testing`, `golang-cli`, `golang-security`, `golang-safety`, `golang-error-handling`, `golang-lint`.

Repo skill gap persists: `.pi/skills/go-implementation/SKILL.md` is required by worker instructions but absent in this checkout (`ENOENT`); global Go skills above are loaded and used.

Rule anchors for review-fix handoff:

- `required-skills-routing`: Go work always loads `golang-how-to`; review/hardening loads `golang-security`, `golang-safety`, `golang-error-handling`, `golang-lint`, and `golang-testing`.
- `golang-how-to`: tests route to `golang-testing`; CLI/args/I/O route to `golang-cli`, `golang-security`, and `golang-safety`; lint/static checks route to `golang-lint`.
- `golang-testing`: #1 named table/subtests, #3 independent tests, #5 observable behavior over implementation details.
- `golang-cli`: stdout/stderr discipline and machine-readable output must remain intact; no CLI behavior change in this review-fix.
- `golang-security`: trust-boundary questions #1-#3; no secrets in logs/artifacts; treat command args and filesystem paths as untrusted.
- `golang-safety`: avoid panic-prone assumptions while parsing recipe text.
- `golang-error-handling`: #1 check errors, #2 add context, #7 log-or-return not both, #9 no panic for expected errors.
- `golang-lint`: run `go vet ./...` as a static quality gate; do not suppress or weaken findings.

GSD command evidence this run:

```bash
scripts/gsd doctor
scripts/gsd list >/tmp/gsd-list-453-reviewfix.out
scripts/gsd prompt plan-phase 453 --skip-research >/tmp/gsd-plan-phase-453-reviewfix.prompt
scripts/gsd prompt programming-loop init --phase 453 --dry-run >/tmp/gsd-programming-loop-453-reviewfix.prompt
```

Result: `doctor`, `list`, and `plan-phase` pass; `programming-loop` remains unavailable with `scripts/gsd: unknown GSD command: programming-loop`; manual GSD fallback remains active via `.pi/prompts/pm-gsd-loop.md` plus universal runtime loop.

Review-fix red:

```bash
go test ./internal/safety -run 'TestSmokeNoBuildReverse' -count=1
```

Result: fail before parser hardening on new negative synthetic cases because the current helper used raw `strings.Index` over the recipe body.

```text
--- FAIL: TestSmokeNoBuildReversePlanPreviewRunOrderingRejectsSyntheticFalsePositives (0.00s)
    --- FAIL: TestSmokeNoBuildReversePlanPreviewRunOrderingRejectsSyntheticFalsePositives/commented_markers_do_not_count (0.00s)
        smoke_makefile_test.go:69: accepted invalid smoke-no-build reverse recipe:
            \t# PLAN_OUTPUT=$$(./pm reverse plan customers_to_outbox --source-table sample_customers --destination outbox:outbox-local --map id:external_id --map name:full_name --map email:email --root "$$SMOKE_DIR")
            \t# PLAN_ID=$$(printf '%s\n' "$$PLAN_OUTPUT" | awk '/Created reverse plan/ {print $$4}')
            \t# ./pm reverse preview "$$PLAN_ID" --root "$$SMOKE_DIR" --json >/dev/null
            \t# APPROVAL=$$(printf '%s\n' "$$PLAN_OUTPUT" | awk '/Approval token:/ {print $$3}')
            \t# ./pm reverse run "$$PLAN_ID" --approve "$$APPROVAL" --root "$$SMOKE_DIR" --json >/dev/null

    --- FAIL: TestSmokeNoBuildReversePlanPreviewRunOrderingRejectsSyntheticFalsePositives/echo_text_does_not_count (0.00s)
        smoke_makefile_test.go:69: accepted invalid smoke-no-build reverse recipe:
            \techo 'PLAN_OUTPUT=$$(./pm reverse plan customers_to_outbox --source-table sample_customers --destination outbox:outbox-local --map id:external_id --map name:full_name --map email:email --root "$$SMOKE_DIR")'
            \techo 'PLAN_ID=$$(printf '%s\n' "$$PLAN_OUTPUT" | awk '/Created reverse plan/ {print $$4}')'
            \techo './pm reverse preview "$$PLAN_ID" --root "$$SMOKE_DIR" --json >/dev/null'
            \techo 'APPROVAL=$$(printf '%s\n' "$$PLAN_OUTPUT" | awk '/Approval token:/ {print $$3}')'
            \techo './pm reverse run "$$PLAN_ID" --approve "$$APPROVAL" --root "$$SMOKE_DIR" --json >/dev/null'

    --- FAIL: TestSmokeNoBuildReversePlanPreviewRunOrderingRejectsSyntheticFalsePositives/printf_help_text_does_not_count (0.00s)
        smoke_makefile_test.go:69: accepted invalid smoke-no-build reverse recipe:
            \tprintf '%s\n' usage:'PLAN_OUTPUT=$$(./pm reverse plan customers_to_outbox --source-table sample_customers --destination outbox:outbox-local --map id:external_id --map name:full_name --map email:email --root "$$SMOKE_DIR")'
            \tprintf '%s\n' usage:'PLAN_ID=$$(printf '%s\n' "$$PLAN_OUTPUT" | awk '/Created reverse plan/ {print $$4}')'
            \tprintf '%s\n' usage:'./pm reverse preview "$$PLAN_ID" --root "$$SMOKE_DIR" --json >/dev/null'
            \tprintf '%s\n' usage:'APPROVAL=$$(printf '%s\n' "$$PLAN_OUTPUT" | awk '/Approval token:/ {print $$3}')'
            \tprintf '%s\n' usage:'./pm reverse run "$$PLAN_ID" --approve "$$APPROVAL" --root "$$SMOKE_DIR" --json >/dev/null'

    --- FAIL: TestSmokeNoBuildReversePlanPreviewRunOrderingRejectsSyntheticFalsePositives/false-prefixed_commands_do_not_count (0.00s)
        smoke_makefile_test.go:69: accepted invalid smoke-no-build reverse recipe:
            \tfalse && PLAN_OUTPUT=$$(./pm reverse plan customers_to_outbox --source-table sample_customers --destination outbox:outbox-local --map id:external_id --map name:full_name --map email:email --root "$$SMOKE_DIR")
            \tfalse && PLAN_ID=$$(printf '%s\n' "$$PLAN_OUTPUT" | awk '/Created reverse plan/ {print $$4}')
            \tfalse && ./pm reverse preview "$$PLAN_ID" --root "$$SMOKE_DIR" --json >/dev/null
            \tfalse && APPROVAL=$$(printf '%s\n' "$$PLAN_OUTPUT" | awk '/Approval token:/ {print $$3}')
            \tfalse && ./pm reverse run "$$PLAN_ID" --approve "$$APPROVAL" --root "$$SMOKE_DIR" --json >/dev/null

FAIL
FAIL\tpolymetrics.ai/internal/safety\t0.452s
FAIL
```

Review-fix green:

```bash
gofmt -w internal/safety && go test ./internal/safety -run 'TestSmokeNoBuildReverse' -count=1
```

```text
ok  \tpolymetrics.ai/internal/safety\t0.430s
```

Review-fix full gates:

- `gofmt -w internal/safety` — pass; no output.
- `go test ./internal/safety -run 'TestSmokeNoBuildReverse' -count=1` — pass: `ok  \tpolymetrics.ai/internal/safety\t0.160s`.
- `go test ./...` — pass; package list ended with `ok  \tpolymetrics.ai/internal/worker\t(cached)`.
- `go vet ./...` — pass; no output.
- `go build ./cmd/pm` — pass; no output.
- `make smoke` — pass; local temp outbox only; Make echoed plan → preview → approval → run and ended `smoke ok: /var/folders/tk/bmp_tx0976s4rkh1phvrpjlw0000gn/T/tmp.DhRT7oIVIV`.
- `make verify` — pass; covered gofmt/tidy-check/vet/test/build/docs/smoke/lint/connectorgen and ended `connectorgen validate: 547 connector(s) checked, 0 findings`.
- `git diff --check origin/feat/cli-architecture-v2...HEAD` — pass; no output.
- `git diff -- go.mod go.sum` — empty.
