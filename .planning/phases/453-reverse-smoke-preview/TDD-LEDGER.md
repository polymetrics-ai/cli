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
FAIL	polymetrics.ai/internal/safety	0.438s
FAIL
```

## Exact green output

```bash
gofmt -w cmd internal && go test ./internal/safety -run TestSmokeNoBuildReversePlanPreviewRunOrdering -count=1
```

```text
ok  	polymetrics.ai/internal/safety	0.161s
```

## Final gate outputs

```bash
gofmt -w cmd internal
```

Result: pass; no output.

```bash
go test ./...
```

Result: pass. Slow package examples: `ok  	polymetrics.ai/internal/cli	167.313s`, `ok  	polymetrics.ai/internal/connectors/certify	338.571s`, `ok  	polymetrics.ai/internal/safety	3.167s`.

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
