# Phase 453 Verification

## Safety interlock

- [x] Do not run `make smoke` until focused ordering regression is green and the Makefile preview step exists.
- [x] Do not run `make verify` until focused ordering regression is green and the Makefile preview step exists.
- [x] No external connector/runtime/credentialed checks.
- [x] Reverse execution, when smoke finally runs, remains limited to existing temporary local outbox fixture and ordered plan → preview → approval → execute.

## Required gate checklist

- [x] `go test ./internal/safety -run TestSmokeNoBuildReversePlanPreviewRunOrdering -count=1` red before Makefile edit.
- [x] `gofmt -w cmd internal`
- [x] `go test ./internal/safety -run TestSmokeNoBuildReversePlanPreviewRunOrdering -count=1` green before smoke.
- [x] `go test ./...`
- [x] `go vet ./...`
- [x] `go build ./cmd/pm`
- [x] `make smoke`
- [x] `make verify`
- [x] `git diff --check origin/feat/cli-architecture-v2...HEAD`
- [x] `git diff -- go.mod go.sum`

## CLI parity checklist

- [x] Applies: no.
- [x] Runtime help checked: not applicable; no command/flag/help/output change.
- [x] Bare namespace behavior checked: not applicable; no CLI behavior change.
- [x] `docs/cli/**` updated: not applicable; docs unchanged by issue scope.
- [x] `website/**` updated: not applicable; website unchanged by issue scope.
- [x] Generated help/manual artifacts updated: not applicable; no help/manual change.
- [x] Parity exemption reason: Makefile smoke ordering and a static regression test only.

## Results

```bash
go test ./internal/safety -run TestSmokeNoBuildReversePlanPreviewRunOrdering -count=1
```

Result: red as expected before Makefile edit; missing `reverse preview` in `smoke-no-build`.

```text
--- FAIL: TestSmokeNoBuildReversePlanPreviewRunOrdering (0.00s)
    smoke_makefile_test.go:30: smoke-no-build recipe missing reverse preview step "./pm reverse preview \"$$PLAN_ID\" --root \"$$SMOKE_DIR\" --json >/dev/null;"
FAIL
FAIL	polymetrics.ai/internal/safety	0.438s
FAIL
```

```bash
gofmt -w cmd internal && go test ./internal/safety -run TestSmokeNoBuildReversePlanPreviewRunOrdering -count=1
```

Result: pass; safety interlock cleared before `make smoke` / `make verify`.

```text
ok  	polymetrics.ai/internal/safety	0.161s
```

```bash
go test ./...
```

Result: pass. Full package output emitted in terminal run; slow packages included `ok  	polymetrics.ai/internal/cli	167.313s`, `ok  	polymetrics.ai/internal/connectors/certify	338.571s`, and `ok  	polymetrics.ai/internal/safety	3.167s`.

```bash
go vet ./...
go build ./cmd/pm
```

Result: pass; both exited 0 with no output.

```bash
make smoke
```

Result: pass. Local temp smoke completed after Makefile executed `reverse plan` → `reverse preview` → approval extraction → `reverse run`; terminal printed `smoke ok: <tmpdir>`.

```bash
make verify
```

Result: pass. Covered gofmt, tidy-check, vet, `go test -timeout 20m ./...`, build, docs validate, local temp smoke, lint, and connectorgen validate. Terminal tail:

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
