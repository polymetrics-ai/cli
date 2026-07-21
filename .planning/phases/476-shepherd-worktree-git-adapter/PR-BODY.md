Refs #476
Refs #471

## Summary

- add a typed Git process adapter with sanitized environment, bounded output/deadlines, canonical
  repository/worktree/origin identity, and exact head evidence
- derive the only valid issue branch and trusted-root worktree path, persist hashed ownership, and
  fail closed on branch/path aliases or competing mutators
- preserve and report dirty, untracked, conflicted, stale, or unique state without exposing worktree
  removal, reset, clean, prune, force push, default-branch push, or arbitrary refspec capability
- emit handoff evidence bound to exact base/head SHAs, PR base, changed scope, verification state,
  repository identity, and worktree identity

## GSD / TDD

- GSD mode: `manual_gsd_fallback`
- Attempted command: `scripts/gsd prompt programming-loop init --phase 476-shepherd-worktree-git-adapter --dry-run`
- Adapter result: `programming-loop` is absent from the repo registry; full manual plan → RED →
  GREEN → refactor → verify → summary lifecycle completed.
- RED: both issue tests failed with `ERR_MODULE_NOT_FOUND` before either production adapter existed.
- GREEN/refactor: focused tests pass 16/16; full Shepherd suite passes 153/153.
- Skills: `gsd-programming-loop`, `gsd-workstreams`, `gsd-plan-phase`,
  `github-issue-first-delivery`, `architecture-patterns`, `javascript-testing-patterns`.
- `gitops-workflow` deviation: available skill is Kubernetes Argo/Flux-specific and inapplicable;
  issue #476 plus repository security/safety policy supplied the typed local-Git contract.

## Verification

- `node --test .pi/extensions/shepherd/workspace-adapter.test.ts .pi/extensions/shepherd/git-adapter.test.ts` — 16 passed, 0 failed
- `node --test .pi/extensions/shepherd/*.test.ts` — 153 passed, 0 failed
- strict no-emit TypeScript over both production adapters using installed Pi 0.80.6 Node types — pass
- exact `pi --list-extensions` — unsupported by Pi 0.80.6; documented offline RPC `get_commands`
  fallback found `pm-shepherd` from `extension` — pass
- `git diff --check` — pass
- Parent policy update: full Go/connectors verification is centralized at parent integration and CI.
  A complete `make verify` had already passed before that update (lint 0 issues; 547 connector
  definitions, 0 findings). Remaining standalone retries were stopped and recorded
  `cancelled_by_parent_policy`.

## Safety and scope

- no dependency, credential, model/network call, GitHub mutation adapter, controller wiring,
  destructive cleanup, default-branch action, or merge is included
- CLI help/docs/website parity is not applicable because no `pm` CLI surface changes
- production changes are limited to the two issue-owned adapter files

## Review route

Independent Codex 5.6 Sol xhigh review will cover the exact pushed head. Per parent direction, this
PR does not request Claude or Copilot review and must not be merged by this worker.
