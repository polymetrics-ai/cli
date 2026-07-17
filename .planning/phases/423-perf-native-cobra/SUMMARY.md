# Phase 423 Summary

Status: red tests captured; production code not touched yet.

## Current state

- Worker branch: `refactor/423-perf-native-cobra`.
- Base branch: `feat/cli-architecture-v2`; dispatch/planning head `6fbff849932e891a8184000fb677e1b6fca7f6d4`.
- GSD adapter doctor passed; `programming-loop` prompt command missing, so manual GSD fallback recorded.
- Required reading and skills loaded. Repo-specific `.pi/skills/go-implementation/SKILL.md` is missing; global Go skills loaded.
- Scope limited to native `perf` Cobra node/handler/tests, directly applicable perf docs/help/generated artifacts, and issue-local phase artifacts.

## Planned delivery

- Promote `pm perf` from legacy wrapper to native Cobra subtree.
- Add native `perf compare` and `perf sync-modes` with declared `StringArray` flags, `NoOptDefVal="true"`, unknown-flag whitelist, optional-value normalization, docs-map help/usage, and no-file completion seams.
- Remove `perf` namespace legacy wrapper and its `parseFlags` call sites.
- Preserve perf output envelopes, repeated flags, bare bool/value sentinels, unknown flag/extra arg tolerance, late global flags, bare namespace help, invalid action usage mapping, config-backed runtime endpoints, and fresh-tree re-entrancy.

## Verification state

Red test captured: `go test ./internal/cli/ -run 'Perf|CobraRouterShell' -count=1` fails because `perf` remains a legacy wrapper and native perf subcommands/flags are missing. Pending green implementation, full local gates, CLI parity checks, PR, and review-route recording.

## Safety

No secrets requested or printed. No credentialed checks. No runtime services started. No dependency changes. No parent/shared orchestration edits. No merge.
