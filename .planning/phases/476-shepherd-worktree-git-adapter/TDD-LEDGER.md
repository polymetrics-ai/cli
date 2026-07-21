# Issue 476 TDD Ledger

## Policy

Production code does not begin until deterministic tests fail for the intended missing behavior.
Tests use bounded temporary local Git repositories and no credentials or network services.

| Slice | RED command/evidence | GREEN command/evidence | Refactor/broad evidence | Status |
|---|---|---|---|---|
| Typed Git adapter | `node --test .pi/extensions/shepherd/workspace-adapter.test.ts .pi/extensions/shepherd/git-adapter.test.ts` → exit 1, `ERR_MODULE_NOT_FOUND` for absent `git-adapter.ts` | Same command → 16/16 pass | Strict no-emit TypeScript passed; raw Git errors reduced to bounded exit evidence | green/refactored |
| Isolated workspace ownership | Same command → exit 1 before collection because the required adapter modules do not exist | Genuine temporary bare-remote/worktree cases pass | Full Shepherd suite 153/153 pass | green/refactored |
| Crash/retry and collision safety | Deterministic genuine-repository cases were present but could not collect until the missing adapters existed | Exact retry, owner collision, concurrent owners, alias branch, path collision, stale base, dirty preservation, and unrelated-head cases pass | Superseded by the correction lease/binding contract below | green/refactored |
| Correction: state identity parity | Focused run: adapter identity differed from `target-evidence.ts` for the same checkout | v1 repository/worktree identity parity passes for coordinator and linked worktree | Local/HTTPS/SSH/SCP/file/no-origin parity; strict TypeScript and full suite pass | green/refactored |
| Correction: immutable handoff claim | Focused run: all three mutable handoff variants were accepted and persisted claim lacked `allowedScopes` | atomic claim/binding records reject workspace and persisted-field tampering | Direct live-object and on-disk mutation cases; full suite passes | green/refactored |
| Correction: exclusive writable lease | Focused run: both same-owner contenders fulfilled; returned workspace had no lease capability | same-owner race yields one lease; release/retry and dead-owner resume pass | Existing append-only `FileStateStore` fencing reused; full suite passes | green/refactored |
| Correction 2: mutation capability fencing | Focused run: released claim committed through a replacement lease; workspace had no capability-bound mutation API | Workspace-held nonforgeable issuer plus adapter WeakMap capability fence fetch/worktree/commit/push; release drains accepted mutations | Alternate-root issuer forgery, forged capability, replacement success, recovery, and in-flight release cases pass | green/refactored |
| Correction 2: complete handoff scope | Focused run: a passed handoff omitted a clean committed path outside immutable scopes | Unfiltered `--no-renames` diff audits the complete canonical set before scope validation | Mixed committed scope, exact returned scope, and committed/dirty literal-backslash alias cases pass | green/refactored |
| Correction 2: effective push endpoint | Focused run: alternate local bare `pushurl` received the issue ref and object before the eventual verification error | Inspection binds effective fetch/push endpoint identities; push revalidates and uses the exact stable endpoint | Late `pushurl` and chained `insteadOf` bare remotes remain ref- and object-free | green/refactored |
| Correction 3: private lease acquisition | Focused run: wrapper captured the workspace issuer and used it to acquire/release a lease under an alternate state root | GitAdapter privately registers a one-way acquisition closure; no issuer or public acquisition method crosses the caller adapter | Wrapper method remains unused, alternate root stays absent, and forged mutation capabilities fail | green/refactored |
| Correction 3: ancestry and history scope | Focused run: commit/push accepted unrelated or historically out-of-scope heads; handoff accepted an add-then-remove path | Commit/push revalidate immutable-base ancestry; history audit unions every touched commit path; push transfers an exact SHA refspec | Unrelated and add/remove heads reject before remote refs/objects change; handoff rejects erased net diffs | green/refactored |
| Correction 3: sanitized Git mutations | Focused run: worktree/add/push accepted executable hook/filter/helper/transport configuration and marker paths were reachable | Mutations use a deterministic config environment and reject executable/local transport configuration before Git mutation | Worktree, clean-filter, pre-push hook, credential-helper, and SSH-command markers remain absent | green/refactored |
| Correction 3: bound default branch | Focused run: binding omitted `defaultBranch`, caller/live remote mismatch was accepted, and push used the mutable branch ref | Inspection and schema-v4 claims bind local origin symbolic HEAD; pre-push `ls-remote --symref` revalidates live HEAD | Caller and live-remote mismatch both reject before the issue ref exists remotely | green/refactored |
| Correction 4: pre-transfer full workspace scope | pending test-only RED matrix for untracked, tracked-dirty, staged, rename, and literal-backslash paths | pending | pending | planned |

Correction RED command: `node --test .pi/extensions/shepherd/workspace-adapter.test.ts
.pi/extensions/shepherd/git-adapter.test.ts` → 21 tests, 16 passed, 5 failed. The five failures map
directly to the reviewed contracts; no production file was changed before this run.

Correction GREEN command: the same focused command → 21 tests passed, 0 failed. Strict no-emit
TypeScript over both production adapters and their imports also passed against the cached Pi 0.80.6
Node type surface.

Correction REFACTOR/broad evidence: focused 21/21; serialized full Shepherd 158/158; strict
TypeScript pass; offline Pi RPC returned `true`; exact diff/scope hygiene pass. Test-file
serialization addresses concurrent Git load without changing SDK deadline assertions.

Correction 2 is active from reviewed head `d5181cd25d108e7748309216b14d91313f112fcd`.
Production remained unchanged for the RED command: `node --test
.pi/extensions/shepherd/workspace-adapter.test.ts .pi/extensions/shepherd/git-adapter.test.ts` →
25 tests, 21 passed, 4 failed. Failures were: alternate `pushurl` received the issue branch;
released claim retained commit authority; no adapter-minted workspace mutation API existed for
release serialization; and handoff passed after omitting a committed out-of-scope path.

Correction 2 GREEN/refactor checkpoint `6a22aa789095da67c5b10f51476de41d3f5643ca`:
focused 29/29; serialized full Shepherd 166/166 in 95.9s; strict no-emit TypeScript passed using
the cached Pi 0.80.6 Node type surface; offline Pi RPC returned `true`; exact diff/scope hygiene
passed. Read-only adversarial probes found alternate-root issuer, literal-backslash path alias, and
chained URL-rewrite variants during refactor; each received a deterministic regression before the
final gate.

Correction 3 test-only RED checkpoint `fa607d31`: focused 36-test run produced 26 passes and ten
expected failures while both production adapters remained at `6a22aa78`. Failures covered missing
bound default evidence, mutable branch-ref push, commit/push ancestry gaps, pre-transfer history
scope gaps, executable worktree/add/push Git configuration, caller/live default mismatch, captured
issuer alternate-root authority, and add-then-remove handoff omission.

Correction 3 GREEN checkpoint `db6bdd675aaced17f0d709b08a647258dfb87f15` and refactor
checkpoint `f7cb0cab0d2fb0c2ef01edc516bd3cdf950b5113`: focused 36/36; serialized full
Shepherd 173/173 in 107.5s; strict no-emit TypeScript passed against cached Pi 0.80.6 types;
offline Pi RPC returned `true`; exact immutable-base diff and owned-path checks passed.

## Required safety cases

- canonical repository identity is stable across linked worktrees and rejects a mismatched binding
- canonical issue branch, parent base, path, remote, ref, and relative scope validation
- typed command construction cannot express force, reset, arbitrary refspec, or cleanup
- default-branch push is unavailable
- dirty, untracked, conflicted, and unique state remains present and is reported
- base/head SHA evidence is exact and validated as 40 lowercase hexadecimal characters
- exact existing workspace reconciliation is idempotent after crash/retry
- alias or duplicate branch/worktree ownership fails closed
- concurrent create attempts cannot yield two active mutators
- every Git mutation requires the exact active capability and release drains accepted operations
- effective fetch/push endpoints remain bound and URL-rewrite-stable before external mutation
- handoff audits every canonical committed and dirty path without separator aliasing

## GSD gate

- Mode: `manual_gsd_fallback`
- Adapter failure: `scripts/gsd: unknown GSD command or prompt: programming-loop`
- Execution decision, plan cycle: `local_critical_path` — this worker already owns one isolated
  sub-issue worktree and no further delegation is authorized or needed.
- Execution decision, TDD gate cycle: `local_critical_path` — RED evidence captured before either
  production adapter file exists.
- Execution decision, execute cycle: `local_critical_path` — minimal typed Git and workspace
  adapters implemented inside the assigned files.
- Execution decision, refactor cycle: `local_critical_path` — validation, canonical scope handling,
  root pinning, and secret-safe Git failure reduction completed; focused and full suites pass.
- Execution decision, verify cycle: `local_critical_path` — authoritative narrowed local gates pass;
  parent integration owns the full Go/connectors rerun under the updated policy.
- Execution decision, summary cycle: `local_critical_path` — durable evidence and stacked PR handoff
  finalized without automated-review or merge authority.
- Execution decision, correction cycles: `local_critical_path` — the identity, claim, and lease
  fixes share one owned adapter boundary; focused RED preceded GREEN, then refactor and broad gates.
- Execution decision, correction 2: `local_critical_path` — the remaining lease, handoff, and
  endpoint findings share the owned adapter boundary; RED preceded capability/endpoint GREEN,
  adversarial refactor probes, and the parent-authorized Shepherd-only verification campaign.
- Execution decision, correction 3 plan: `local_critical_path` — all five findings are coupled in
  the two issue-owned adapters. A read-only design sidecar was attempted but the runtime thread cap
  was already occupied; local planning continues without widening the write scope.
- Execution decision, correction 3 RED: `local_critical_path` — test-only checkpoint `fa607d31`
  produced ten deterministic failures across all five reviewed contracts before production edits.
- Execution decision, correction 3 GREEN: `local_critical_path` — private acquisition, immutable
  ancestry/history scope, sanitized Git mutation, and bound default-branch contracts pass 36/36.
- Execution decision, correction 3 refactor: `local_critical_path` — safe Git config is centralized,
  worktree config errors fail closed, and duplicate commit flags were removed with focused/strict gates green.
- Execution decision, correction 3 verify: `local_critical_path` — focused 36/36, Shepherd 173/173,
  strict Pi 0.80.6 TypeScript, offline RPC, and exact diff/scope gates pass.
- Execution decision, correction 3 summary: `local_critical_path` — exact checkpoint and gate
  evidence is durable; the parent owns the fresh exact-head xhigh review and merge decisions.
- Execution decision, correction 4 plan: `local_critical_path` — the single finding is confined to
  the owned Git push boundary. A read-only recon sidecar was attempted, but the runtime thread cap
  was occupied; production remains frozen while the local test-only RED matrix is prepared.
