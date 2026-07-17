# Phase 423 Summary

Status: PR #458 open against `feat/cli-architecture-v2`; review-fix implementation and local verification complete on `refactor/423-perf-native-cobra`; commit/PR-body update/push pending.

## Current state

- Worker branch: `refactor/423-perf-native-cobra`; sub-PR: https://github.com/polymetrics-ai/cli/pull/458.
- Base branch: `feat/cli-architecture-v2`; dispatch/planning head `6fbff849932e891a8184000fb677e1b6fca7f6d4`.
- GSD adapter doctor passed; `programming-loop` prompt command missing, so manual GSD fallback recorded.
- Required reading and skills loaded. Repo-specific `.pi/skills/go-implementation/SKILL.md` is missing; global Go skills loaded.
- Scope limited to native `perf` Cobra node/handler/tests, directly applicable perf docs/help/generated artifacts, and issue-local phase artifacts.

## Delivered

- Promoted `pm perf` from legacy wrapper to native Cobra subtree.
- Added native `perf compare` and `perf sync-modes` with declared `StringArray` flags, `NoOptDefVal="true"`, unknown-flag whitelist, optional-value normalization, docs-map help/usage, and no-file completion seams.
- Removed `perf` namespace legacy wrapper and its `parseFlags` call sites.
- Preserved perf output envelopes, repeated flags, bare bool/value sentinels, unknown flag/extra arg tolerance, late global flags, bare namespace help, invalid action usage mapping, config-backed runtime endpoints, and fresh-tree re-entrancy.

## Verification state

Red test captured: `go test ./internal/cli/ -run 'Perf|CobraRouterShell' -count=1` failed because `perf` remained a legacy wrapper and native perf subcommands/flags were missing.

Focused green gates passed: `go test ./internal/cli/ -run 'Perf|CobraRouterShell' -count=1`, `go test ./internal/cli/... -run 'Perf|CobraRouterShell|Golden' -count=1`, `go test ./internal/cli/ -run Certify -count=1`, `gofmt -w cmd internal`, `go vet ./...`, and `go build ./cmd/pm`.

Full gates passed: `go test ./...`, `make verify`, and an explicit final `go build ./cmd/pm`.

Runtime help parity checked: `./pm help perf`, `./pm perf`, `./pm perf --help`, `./pm perf --json`, invalid action JSON usage error, `perf compare` JSON, and `perf sync-modes` JSON.

Docs/website/generated checks passed: `./pm docs generate` diff against `docs/cli`, `./pm docs validate`, and `npm --prefix website run gen:docs`; no tracked docs/website/golden changes.

Diff guards passed: `git diff --check origin/feat/cli-architecture-v2...HEAD`; `git diff -- go.mod go.sum` empty.

Sub-PR #458 opened non-draft against `feat/cli-architecture-v2`; remote checks queued at PR creation. Claude/Copilot not manually requested; review coverage pending per stacked PR policy.

## Review-fix state

Accepted review findings recorded in `PLAN.md`: perf exit status parity, runtime metadata/security text, bounded workload flags, runtime error redaction, Dragonfly/Temporal topology metadata docs, and Redis diagnostic routing. TDD red plan and verification checklist were updated before production edits.

Delivered review-fix slice:

- Added bounded perf workload constants and CLI validation: `--iterations` 1..1000 and `--records` 1..100000; invalid/oversized values return exit 3 JSON `validation_error`.
- Added internal perf API max guards, runtime-backed error redaction through `internal/logging`, and cleanup for temp benchmark roots.
- Routed go-redis diagnostics through the redacting logger at quiet level; loopback runtime compare produced no raw `redis:` stderr.
- Updated canonical perf manual, generated `docs/cli/perf.md`, golden transcripts, website source docs, and `website/lib/docs.generated.ts` for exit 3 and runtime metadata/topology wording.

Current decision: `local_critical_path` — same isolated worker branch/cwd, no subagent tool, no reset/recreate. Next required action: commit, update PR #458 body with dispositions/verification, push same branch.

## Safety

No secrets requested or printed. No credentialed checks. No runtime services started. No dependency changes. No parent/shared orchestration edits. No merge.
