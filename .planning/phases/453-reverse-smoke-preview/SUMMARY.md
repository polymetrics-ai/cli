# Phase 453 Summary

Status: PR #454 open against `feat/cli-architecture-v2`; implementation complete; full local verification passed.

## Current state

- Worker branch: `fix/453-reverse-smoke-preview`; sub-PR: https://github.com/polymetrics-ai/cli/pull/454.
- Base branch: `feat/cli-architecture-v2`; dispatch/base parent commit `5680debb`.
- GSD adapter doctor passed; `programming-loop` prompt command missing, so manual GSD fallback recorded.
- Scope limited to `Makefile` reverse smoke ordering, one durable regression check, and issue-local phase artifacts.

## Delivered

- Added durable regression `internal/safety/smoke_makefile_test.go`.
- Fixed `Makefile` `smoke-no-build` ordering: `reverse plan` → `PLAN_ID` extraction → `reverse preview --json` with same temp root → approval extraction → `reverse run`.
- Created issue-local GSD/TDD/verification artifacts under `.planning/phases/453-reverse-smoke-preview/`.
- CLI help/docs/website parity marked N/A because no CLI product behavior changed.

## Verification

Red test captured before Makefile edit:

```bash
go test ./internal/safety -run TestSmokeNoBuildReversePlanPreviewRunOrdering -count=1
```

Result: fail as expected; missing `reverse preview` in `smoke-no-build`.

Focused green before smoke:

```bash
gofmt -w cmd internal && go test ./internal/safety -run TestSmokeNoBuildReversePlanPreviewRunOrdering -count=1
```

Result: pass (`ok  	polymetrics.ai/internal/safety	0.161s`).

Full gates passed: `gofmt -w cmd internal`, `go test ./...`, `go vet ./...`, `go build ./cmd/pm`, `make smoke`, `make verify`, `git diff --check origin/feat/cli-architecture-v2...HEAD`, `git diff -- go.mod go.sum`.

`make verify` covered gofmt, tidy-check, vet, `go test -timeout 20m ./...`, build, docs validate, local temp smoke, lint, and connectorgen validate; tail ended with `connectorgen validate: 547 connector(s) checked, 0 findings`.

## Safety

No secrets requested or printed by this implementation. No credentialed checks. No runtime services started. No dependency changes. No parent/shared orchestration edits. Reverse execution only occurred through existing local temp smoke after preview step existed and focused ordering test was green. No merge.

## Review-fix update — PR #454 MEDIUM finding (2026-07-17)

Status: accepted finding fixed locally; commit prepared; push and PR body update pending.

Delivered:

- Hardened `internal/safety/smoke_makefile_test.go` so the smoke contract parses executable `smoke-no-build` shell statements instead of raw substrings.
- Added negative synthetic cases for commented markers, `echo` text, `printf`/help text, `false &&` prefixes, unrelated targets, and wrong-order markers.
- Preserved exact required order: `reverse plan` → `PLAN_ID` extraction → `reverse preview` → approval extraction → `reverse run`.
- No Makefile behavior, dependency, docs, website, connector, or CLI product changes in this review-fix.

Focused red before parser hardening:

```bash
go test ./internal/safety -run 'TestSmokeNoBuildReverse' -count=1
```

Result: fail; synthetic false-positive cases were accepted by raw substring matching.

Focused green and full gates passed after parser hardening:

```bash
gofmt -w internal/safety
go test ./internal/safety -run 'TestSmokeNoBuildReverse' -count=1
go test ./...
go vet ./...
go build ./cmd/pm
make smoke
make verify
git diff --check origin/feat/cli-architecture-v2...HEAD
git diff -- go.mod go.sum
```

Review route: no Claude/Copilot requested per review-fix instruction; human/parent review remains the route.
