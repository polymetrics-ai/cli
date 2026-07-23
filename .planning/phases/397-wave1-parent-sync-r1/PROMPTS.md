# Issue #397 Wave 1 Prompt Snapshot

## Source contract

Synchronize current `main` into the published CLI Architecture v2 parent through an ordinary no-ff merge on branch `fm/cli-architecture-v2-wave1-parent-sync-r1`, preserve current Gong and Architecture v2 behavior, refresh truthful parent synchronization evidence, validate exact head, and open a draft stacked PR to `feat/cli-architecture-v2`. Do not implement #408.

## Runtime path

- `scripts/gsd doctor`
- `scripts/gsd list`
- `scripts/gsd sources <command>`
- `scripts/gsd prompt plan-phase 397 --skip-research`
- `programming-loop` absent; manual PLAN → RED → GREEN → REFACTOR → VERIFY → REVIEW → INTEGRATE fallback

## Downstream artifact

`.planning/phases/397-wave1-parent-sync-r1/`

## Verification result

Original synchronization review, Shepherd, and required branch-specific workflows passed at `3fd63fbe0f526873fa3adb8a75fa5f20342d52a6`. Captain-approved PM-orchestrator extension verification passed at implementation head `d72a93018933541d390884f96b285856e269a1ab`; final extension evidence-head review/Shepherd/PR checks remain pending.
