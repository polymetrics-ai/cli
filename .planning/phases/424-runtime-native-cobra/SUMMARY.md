# Phase 424 Summary

Status: PR #460 positional-help correction verified and pushed on `refactor/424-runtime-native-cobra`; no Claude/Copilot requested.

## Current state

- Worker branch: `refactor/424-runtime-native-cobra`; sub-PR: https://github.com/polymetrics-ai/cli/pull/460.
- Base branch: `feat/cli-architecture-v2`; dispatch/planning head `56a7ecb08f755184af7b55318c3285582d5adfb7`.
- Parent PR: https://github.com/polymetrics-ai/cli/pull/438 (draft).
- GSD adapter doctor passed; `programming-loop` prompt command missing, so manual GSD fallback recorded.
- Required reading and skills loaded. Repo-specific `.pi/skills/go-implementation/SKILL.md` is missing; global Go skills loaded.
- Scope limited to native `runtime` Cobra node/handler/tests, directly applicable runtime docs/help/generated artifacts, and issue-local phase artifacts.

## Delivered

- Promoted `pm runtime` from legacy wrapper to native Cobra subtree.
- Added native `runtime doctor` subcommand with unknown-flag compatibility and no-file completion seam.
- Removed `runtime` namespace legacy wrapper and replaced legacy runtime handler with `runRuntimeDoctor`.
- Preserved runtime doctor output, config endpoint use, redaction, bare namespace help, JSON/stderr contract, and runtime service optionality on focused tests.

## Verification state

Red test captured: `go test ./internal/cli/ -run 'Runtime|CobraRouterShell' -count=1` failed because `runtime` remained a legacy wrapper and native `doctor` subcommand was missing.

Focused green gates passed: `go test ./internal/cli/ -run 'Runtime|CobraRouterShell' -count=1`, `go test ./internal/cli/... -run 'Runtime|CobraRouterShell|Golden' -count=1`, `go test ./internal/cli/...`, `go vet ./...`, and `go build ./cmd/pm`.

Full gates passed: `gofmt -w cmd internal`, `go vet ./...`, `go test ./...`, `go build ./cmd/pm`, and `make verify`.

Runtime help parity checked: `./pm help runtime`, `./pm runtime`, `./pm runtime --help`, `./pm runtime --json`, invalid action JSON usage error, and loopback-only `runtime doctor` JSON with unknown flag/extra arg tolerance.

Docs/website/generated checks passed: `./pm docs generate` temp diff against `docs/cli`, `./pm docs validate`, and `npm --prefix website run gen:docs`; no tracked docs/website/golden changes.

Diff guards passed: `git diff --check origin/feat/cli-architecture-v2...HEAD`; `git diff -- go.mod go.sum` empty.

Sub-PR #460 opened non-draft against `feat/cli-architecture-v2`; remote checks and Claude auto-review coverage pending. Claude/Copilot not manually requested.

## Review-fix status

- Fixed in scope: Cobra/pflag parse-error usage mapping, runtime optional-service docs, DragonflyDB/Temporal endpoint sanitization.
- Red/green/full-gate evidence tracked in `TDD-LEDGER.md` and `VERIFICATION.md`.
- High security finding against `internal/worker/podman_cmd.go` / `internal/worker/submit.go` checked against PR diff: not changed by PR #460; disposition recorded as out-of-scope follow-up, no code fix here.
- User disallowed Claude/Copilot; review route is human/parent-orchestrator only.

## Positional-help correction

Independent review found that native Cobra routing had dropped the legacy positional aliases `pm runtime help` and `pm runtime help --json`. Session `7050f706-72d2-47df-ac13-0b08979cc1ae` (`openai-codex/gpt-5.6-sol`, thinking `high`) started from exact HEAD `8d696cd4c27fad6840e905917e7658e785fa5436`, captured focused RED failures for both aliases, and added one hidden native `runtime help` command that delegates to the canonical runtime manual.

Focused runtime/router/golden and runtimecheck tests passed. Built-binary checks confirmed positional text help matches `pm help runtime`, positional JSON help returns `CommandManual/runtime`, and `runtime bogus --json` still exits 2 with usage and no manual envelope. Full gofmt, vet, repository tests, build, and `make verify` passed. Canonical help, docs, website, goldens, and dependencies have no delta. Correction commit `345399166711e6e733d8f0c84e17db55a2d90a2a` was pushed to the existing PR #460 branch; no new PR or external review request was made.

## Safety

No secrets requested or printed. No credentialed checks. No runtime services started. No dependency changes. No parent/shared orchestration edits. No merge.
