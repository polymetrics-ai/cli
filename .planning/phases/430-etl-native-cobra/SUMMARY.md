# Phase 430 Summary

Status: bounded correction active from exact head `9b0020abde7cd7e007412a0294db6e4cb576f5f3`; focused differential RED and private-capture GREEN complete, broad verification pending and non-terminal.

## Identity

- Session: `issue-430-pi-openai-codex-gpt-5.6-sol-high-20260718T225346Z`
- Model/thinking: `openai-codex/gpt-5.6-sol`, `high`
- Branch: `refactor/430-etl-native-cobra`
- Parent: #397; umbrella: #407; draft parent PR #438

## Plan

Nativize only the ETL namespace and current action/flag surface, preserving direct fixture operations, configured ETL runs/status, bounded batches, sync validation, cancellation, events/telemetry, stdout/stderr and JSON envelope behavior, and legacy help/unknown/literal compatibility. Remove only ETL parser calls; retain the dynamic connector parser.

## Workflow

GSD doctor/list passed and plan-phase generated. The adapter lacks `programming-loop`, so the manual universal-loop fallback is active. All six issue-local artifacts were created with exact identity/start before test or production edits. Execution decision is `local_critical_path` for this serialized isolated unit; no subagent tool is exposed.

## Safety

Fixture/local temporary connectors only. No secrets, credentialed external checks, optional services, standalone reverse execution, dependencies, unrelated writes, PR, or review. The required `make verify` used only its existing temporary-root local approval-gated smoke.

## Focused delivery

Strict focused test compilation failed as required on the missing `newETLCobraCommand` constructor before production edits. Native Cobra now owns ETL check/catalog/read/run/status/help and every current typed flag. ETL-only normalization preserves repeated/bare/assigned, action-tail help, literal separator, and unknown tolerance; only ETL legacy parser calls were removed. Focused GREEN passed in `13.396s`; broader ETL/router focused tests passed in `27.999s`.

## Local review correction

Local review found that `etl bogus --help|-h` rendered the namespace manual and exited 0 instead of retaining an invalid-action usage error. A focused correction test failed as required before correction production edits. Unrecognized actions are now bounded behind Cobra's literal separator before flag parsing. Focused, repeated ×5, race, and router/golden/help preservation gates pass.

## Verification

Focused/repeated/race/router/golden/full CLI/app/repository gates pass. Exact start-vs-head differential matched 20/20 preserved cases. Runtime help topic/bare/long-help is byte-identical; JSON/manual/invalid-action checks pass. Generated CLI docs and website generation are clean with no tracked docs delta. Gofmt, vet, build, dependency/scope guards, and final `make verify` pass (CLI `356.154s`, certify `335.400s`, lint 0, connector validation 547/0).

No `go.mod`, `go.sum`, connector definition, docs/website, golden, or unrelated namespace delta exists. No PR or external review was created.

## Bounded correction

`/tmp/pm-397-review-430.log` found that shared normalization can discard the first `etl status` operand when it looks like help, literal separation, or an unknown flag, then query a later valid ID. Exact legacy behavior owns the first post-action token and fails closed. The focused differential test failed as required for all four requested operand classes plus an internal-carrier-shaped adversarial operand. GREEN now captures only that operand before shared normalization into invocation-private context state with no argv carrier; exact differential/adversarial tests and all ETL tests pass. No completion claim is made until broad verification, `make verify`, commit, and push succeed.
