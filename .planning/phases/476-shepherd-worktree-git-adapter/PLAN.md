# Issue 476 Plan — Shepherd Worktree and Git Adapter

## Contract

- Issue: `#476`
- Parent issue: `#471`
- Parent PR: `#472`
- Immutable base: `e659d6f1b666f58748e2d8c86599ceb4bbc62ff8`
- Branch: `feat/476-shepherd-worktree-git-adapter`
- PR base: `feat/471-pi-agent-session-shepherd`
- Production scope: `.pi/extensions/shepherd/workspace-adapter.ts`, `.pi/extensions/shepherd/git-adapter.ts`
- Test scope: matching issue-owned tests and bounded temporary-repository fixtures

## GSD mode and skills

- GSD mode: `manual_gsd_fallback`
- Attempted command: `scripts/gsd prompt programming-loop init --phase 476-shepherd-worktree-git-adapter --dry-run`
- Evidence: adapter health passed, then `scripts/gsd: unknown GSD command or prompt: programming-loop`.
- Loaded: `gsd-programming-loop`, `gsd-workstreams`, `gsd-plan-phase`,
  `github-issue-first-delivery`, `architecture-patterns`, `javascript-testing-patterns`.
- `gitops-workflow` is intentionally not applied: the available skill targets Kubernetes
  Argo/Flux GitOps, not local typed Git process boundaries. This plan instead applies issue #476,
  repository safety rules, and the Shepherd runtime contract.
- Go skills: not applicable to the TypeScript production slice. Go repository gates remain required.
- CLI help/docs/website parity: not applicable; this slice adds no `pm` CLI surface.

## Architecture

1. `git-adapter.ts` is the outer Git process adapter. It exposes only typed operations required by
   the issue: repository inspection, status, fetch, worktree inventory/addition, exact ref/head
   resolution, bounded diff/commit, and branch push. It owns sanitized Git environment and bounded
   output/timeout behavior.
2. `workspace-adapter.ts` is policy/application logic. It derives the canonical issue branch and
   worktree path, validates parent base and repository identity, reconciles an existing exact
   workspace, and fails closed on aliases, duplicate ownership, collisions, dirty state, or stale
   base evidence.
3. Git metadata is authoritative for crash/retry reconciliation. Existing exact ownership is
   adopted idempotently; a branch or path owned elsewhere is reported and preserved. No removal,
   prune, reset, force push, default-branch push, or arbitrary refspec API exists.

## TDD slices

### Slice 1 — Typed Git safety boundary

- RED: tests import the absent adapter and specify canonical repository identity, validation of
  branch/base/path/remote inputs, typed argv construction, status preservation, exact head/base
  evidence, bounded diff/commit, and protected push behavior.
- GREEN: implement the minimal typed adapter and pass focused tests.
- REFACTOR: centralize untrusted-input validation, sanitized execution, and evidence parsing.
- Checkpoint: commit and push a coherent green Git adapter slice.

### Slice 2 — Isolated workspace ownership

- RED: temporary repositories specify canonical issue naming, coordinator exclusion, trusted-root
  containment, duplicate/alias collision rejection, idempotent retries, existing dirty/untracked
  preservation, exact base verification, and concurrent two-mutator prevention.
- GREEN: implement derived paths plus create-or-reconcile logic using only the Git port.
- REFACTOR: make collision reports deterministic and keep policy independent of process details.
- Checkpoint: commit and push a coherent green workspace slice.

### Slice 3 — Full verification and handoff

- Run focused tests, all Shepherd tests, strict no-emit TypeScript using pinned Pi 0.80.6 types,
  Pi extension discovery, repository-wide Go/build/verify gates, and diff hygiene.
- Update `TDD-LEDGER.md`, `VERIFICATION.md`, `SUMMARY.md`, `PROMPTS.md`, traces, and
  `RUN-STATE.json` with exact evidence.
- Push exact head and open a ready PR against the parent branch. Do not merge or request automated
  review; independent Codex 5.6 Sol xhigh review is owned by the parent orchestrator.

## Safety invariants

- Never read, print, persist, or pass credentials.
- All user-controlled text is bounded and rejects control characters.
- Worktree paths are derived under a canonical trusted root and outside the coordinator checkout.
- Only canonical issue branches may be created or pushed; `main`, default-branch aliases, HEAD,
  options, reflog syntax, revision expressions, and arbitrary refspecs are rejected.
- Dirty, untracked, conflicted, or unique state is evidence to preserve and report, never a cleanup
  trigger.
- Repository and worktree identities bind operations to the originally inspected Git repository.
- Commit scope is an explicit bounded relative-path allowlist; no whole-repository staging API.

## Verification checklist

- [ ] `node --test .pi/extensions/shepherd/workspace-adapter.test.ts .pi/extensions/shepherd/git-adapter.test.ts`
- [ ] `node --test .pi/extensions/shepherd/*.test.ts`
- [ ] strict no-emit TypeScript against pinned Pi 0.80.6 types
- [ ] `pi --list-extensions`
- [ ] `git diff --check`
- [ ] `go vet ./...`
- [ ] `go test ./...`
- [ ] `go build ./cmd/pm`
- [ ] `make verify`
- [ ] pushed branch exact head matches local head
- [ ] ready sub-PR targets `feat/471-pi-agent-session-shepherd`
