# Phase 422 Summary

Status: planning checkpoint complete; red tests next.

## Current state

- Worker branch: `refactor/422-query-native-cobra`.
- Base branch: `feat/cli-architecture-v2`; branch rebased from dispatch base `e6faecfb` to parent ledger checkpoint `f12d573b` before edits.
- GSD adapter doctor passed; `programming-loop` prompt command missing, so manual GSD fallback recorded.
- Required reading and skills loaded. Repo-specific `.pi/skills/go-implementation/SKILL.md` is missing; global Go skills loaded.
- Scope limited to native `query` Cobra node/handler/tests, directly applicable query docs/help/generated artifacts, and issue-local phase artifacts.

## Planned delivery

- Promote `pm query` from legacy wrapper to native Cobra subtree.
- Add native `query run` with declared flags and legacy-compatible optional values / unknown flag tolerance.
- Preserve query output envelopes, agent-mode summary/stream behavior, late global flags, bare namespace help, invalid action usage mapping, and fresh-tree re-entrancy.
- Preserve SQL read-only guards in `App.QuerySQL`; no app/query-engine behavior change and no generic SQL write.
- Keep docs/goldens byte-identical unless an intentional reviewed change is required.

## Verification

Pending red/green/full gates.

## Safety

No secrets requested or printed. No credentialed checks. No runtime services started. No dependency changes. No parent/shared orchestration edits. No merge.
