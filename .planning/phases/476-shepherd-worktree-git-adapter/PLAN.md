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
- Go skills and Go repository gates: not applicable under the parent-authorized correction policy.
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
3. Git metadata and immutable claim/binding records are authoritative for crash/retry
   reconciliation. The existing append-only Shepherd lease fences the one writable mutator;
   a branch or path owned elsewhere is reported and preserved. No removal, prune, reset, force
   push, default-branch push, or arbitrary refspec API exists.

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

- [x] `node --test .pi/extensions/shepherd/workspace-adapter.test.ts .pi/extensions/shepherd/git-adapter.test.ts`
- [x] `node --test --test-concurrency=1 .pi/extensions/shepherd/*.test.ts`
- [x] strict no-emit TypeScript against pinned Pi 0.80.6 types
- [x] documented offline Pi 0.80.6 RPC `get_commands` discovery
- [x] `git diff --check`
- [x] pushed implementation/refactor branch head matched local and PR head
- [x] ready sub-PR targets `feat/471-pi-agent-session-shepherd`

## Exact-head correction cycle

Independent xhigh review of `906a45c53ae1a19c9d2efe1c3f24a64e36ef4d63` found three
blocking contracts. This correction cycle keeps the original base and owned-file boundary:

1. Add RED parity tests against `resolveCanonicalGitWorktree` so the Git port emits the same
   repository/worktree identities already persisted in Shepherd controller state.
2. Add RED handoff-tamper tests, then atomically persist canonical scopes plus the exact worktree
   binding and make handoff reread that immutable claim.
3. Add RED same-owner race, release, and dead-owner resume cases, then reuse Shepherd's existing
   append-only `FileStateStore` lease/fencing primitive instead of introducing a second lock design.
4. Verify only the parent-authorized focused tests, serialized full Shepherd suite, strict no-emit
   TypeScript, offline Pi RPC discovery, and exact diff/scope checks. Serial test-file execution
   isolates existing SDK wall-clock deadline assertions from the Git fixture process load without
   changing production or test timeouts.

## Exact-head correction cycle 2

Independent xhigh re-review of `d5181cd25d108e7748309216b14d91313f112fcd` found three
remaining mutation-boundary blockers. This cycle preserves the immutable base and issue-owned
file boundary:

1. Add deterministic RED coverage proving a released claim cannot mutate after a replacement
   lease is active, and proving release must serialize behind an already accepted Git mutation.
   Then move mutation authority to an adapter-minted, identity-checked lease capability and require
   it for fetch, worktree creation, commit, and push.
2. Add RED handoff coverage for a clean commit containing both allowed and out-of-scope paths.
   Then inspect the complete canonical `baseHead..head` path set before validating immutable scopes;
   dirty paths receive the same validation and no successful handoff may omit a changed path.
3. Add a local bare-repository RED proving a late `remote.origin.pushurl` cannot receive refs or
   objects. Then bind Git's effective fetch/push endpoints during inspection, revalidate them before
   push, reject divergence, and execute plus verify against the exact validated endpoint.
4. Checkpoint and push RED before production edits, then checkpoint GREEN/refactor/evidence. Run
   only focused #476 tests, serialized Shepherd tests, strict cached Pi 0.80.6 TypeScript, offline
   Pi RPC, and exact diff/scope gates; do not run Go, connector, or `make` gates.

Correction 2 result: RED `e8d1a3d7f0d463ea6ea3acfd928ea17e2acdf026`; GREEN/refactor
`6a22aa789095da67c5b10f51476de41d3f5643ca`. Focused 29/29, serialized Shepherd
166/166, strict cached-Pi TypeScript, offline Pi RPC, and exact diff/scope gates pass. Refactor
also closes reproduced alternate-root issuer, literal-backslash path, and chained URL-rewrite
bypasses without widening the owned file boundary.

## Exact-head correction cycle 3

Independent xhigh review of `9728f9ed12e8e545eabd8b9b1b8028af80150427` found five
remaining authority, history, and Git-process boundary blockers. This cycle preserves immutable
base `e659d6f1b666f58748e2d8c86599ceb4bbc62ff8` and the existing issue-owned files.

1. Add a wrapper RED proving `WorkspaceAdapter` never passes mutation authority through an
   overridable `GitAdapter` method. Replace the public issuer API with a module-private registered
   lease-acquisition closure; caller wrappers never receive or capture the authority.
2. Add RED cases for unrelated and out-of-scope canonical heads. Commit and push must revalidate
   the immutable `baseHead` ancestry, and push must audit the complete committed path history
   immediately before transferring an exact SHA refspec.
3. Add an add/commit/remove/commit out-of-scope RED. Handoff and push audit the sorted union of
   every path touched by commits in `baseHead..head`, plus dirty paths for handoff, rather than
   trusting only the final tree diff.
4. Add bounded marker RED cases for repository hooks, filters, credential helpers, and transport
   commands. All typed mutations use a deterministic sanitized environment/config prefix and fail
   closed on executable local Git configuration before a marker can run.
5. Add a remote symbolic-HEAD mismatch RED. Inspection binds the canonical remote default branch;
   mutation binding and a pre-push remote `HEAD` query revalidate it, and caller input must match
   the bound value.
6. Commit/push a plan checkpoint, a genuine test-only RED checkpoint, the smallest GREEN, and any
   separate refactor/evidence checkpoint. Verification is limited to focused adapter tests, the
   serialized Shepherd suite, strict cached Pi 0.80.6 TypeScript, offline Pi RPC, and exact
   diff/path scope. Go, connector, certification, runtime-backed, and `make` gates remain forbidden.

### Correction 3 verification checklist

- [x] focused adapter tests pass after a recorded test-only RED
- [x] complete Shepherd suite passes with `--test-concurrency=1`
- [x] strict no-emit TypeScript passes against cached Pi 0.80.6 types
- [x] documented offline Pi 0.80.6 RPC returns `true`
- [x] immutable-base diff check and exact changed-path scope pass
- [x] final local, tracking, and remote branch heads match exactly after the evidence commit

Correction 3 result: plan `2e255372`, genuine test-only RED `fa607d31`, GREEN
`db6bdd675aaced17f0d709b08a647258dfb87f15`, and refactor
`f7cb0cab0d2fb0c2ef01edc516bd3cdf950b5113`. Focused 36/36, serialized Shepherd
173/173 in 107.5s, strict cached-Pi TypeScript, offline Pi RPC `true`, and immutable-base
diff/path-scope gates pass. Evidence checkpoint `0a3cdfa4ce1ac46f87fc31ed14e295d17a4bb62c`
matched the local, tracking, and remote branch refs exactly before this terminal attestation.

## Exact-head correction cycle 4

Independent xhigh review of `1fe994a68ec3286ee69f1be4fadf71416d601257` found one remaining
pre-transfer scope invariant: push audits committed history but does not include current staged,
tracked-dirty, or untracked paths immediately before changing the remote. This cycle preserves the
immutable base and existing issue-owned file boundary.

Production is frozen until the correction-4 RED checkpoint. Baseline hashes at the reviewed head:

- `git-adapter.ts`: `fc4f95aa0b701f8ca29b418954eed9f1b1066cecb4435240760c2137666b8c14`
- `workspace-adapter.ts`: `59711b18cf3b2fe1ae38994c5ea491d0890d068fa7d2122cd1208b038ec854f0`

1. Add a test-only RED matrix for an out-of-scope untracked path, tracked modification, staged
   addition, staged rename, and literal-backslash path. Every case must reject and prove the bare
   remote still has no canonical issue ref.
2. GREEN inside the existing queued mutation/lease boundary: immediately before transfer, collect
   canonical status evidence, union both rename endpoints with the already-audited committed
   history, validate the union against the immutable lease scopes, then retain the existing exact
   SHA/ref, endpoint, default-branch, and lease checks.
3. REFACTOR only after GREEN: centralize status-path extraction/pre-transfer scope validation so
   commit, handoff, and push use one canonical interpretation without widening capabilities or the
   race window.
4. Commit and push plan, genuine test-only RED, minimal GREEN, refactor, and terminal evidence
   checkpoints. Run only focused adapter tests, serialized Shepherd tests, strict cached Pi 0.80.6
   TypeScript, offline Pi RPC, and immutable-base diff/path checks. Go, connector, certification,
   runtime-backed, and `make` gates remain forbidden.

### Correction 4 verification checklist

- [ ] focused adapter tests pass after a recorded test-only RED
- [ ] complete Shepherd suite passes with `--test-concurrency=1`
- [ ] strict no-emit TypeScript passes against cached Pi 0.80.6 types
- [ ] documented offline Pi 0.80.6 RPC returns `true`
- [ ] immutable-base diff check and exact changed-path scope pass
- [ ] final local, tracking, and remote branch heads match exactly
