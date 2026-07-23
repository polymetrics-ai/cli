# Issue #397 PM Orchestrator Extension Prompt Snapshot

## Source contract

Captain-approved PR #495 extension: make the existing PM/Pi parent orchestrator the canonical replacement when repo-local GSD has no `programming-loop`; require exact-head local Codex review plus independent Shepherd; remove Claude/Copilot as required/fallback PM coverage; preserve historical evidence and keep PR #493 disjoint.

## Runtime path

- `scripts/gsd doctor`
- `scripts/gsd list`
- `scripts/gsd sources plan-phase`
- `scripts/gsd sources code-review`
- `programming-loop` absent; `/pm-orchestrate` owns the manual lifecycle

## Downstream artifact

`.planning/phases/397-pm-orchestrator-extension/`

## Verification result

Focused and full credential-free gates passed at implementation head `d72a93018933541d390884f96b285856e269a1ab`; final evidence-head local Codex review, Shepherd validation, and PR checks remain pending.
