---
phase: 471-pi-agent-session-shepherd
reviewed: 2026-07-21T10:28:01Z
depth: deep
base: 74ab381eb8236305170ffd44d5aed74f8d0d2936
head: d434f6727441f17047185658f538612fcaf6187e
files_reviewed: 38
files_reviewed_list:
  - .pi/README.md
  - .pi/extensions/shepherd/arguments.test.ts
  - .pi/extensions/shepherd/arguments.ts
  - .pi/extensions/shepherd/controller.test.ts
  - .pi/extensions/shepherd/controller.ts
  - .pi/extensions/shepherd/domain.test.ts
  - .pi/extensions/shepherd/domain.ts
  - .pi/extensions/shepherd/extension.test.ts
  - .pi/extensions/shepherd/extension.ts
  - .pi/extensions/shepherd/index.ts
  - .pi/extensions/shepherd/runner.ts
  - .pi/extensions/shepherd/sdk-runner.test.ts
  - .pi/extensions/shepherd/sdk-runner.ts
  - .pi/extensions/shepherd/state-store.test.ts
  - .pi/extensions/shepherd/state-store.ts
  - .pi/extensions/shepherd/target-evidence.test.ts
  - .pi/extensions/shepherd/target-evidence.ts
  - .planning/phases/471-pi-agent-session-shepherd/471-REVIEW.md
  - .planning/phases/471-pi-agent-session-shepherd/AGENTS.md
  - .planning/phases/471-pi-agent-session-shepherd/CONTEXT.md
  - .planning/phases/471-pi-agent-session-shepherd/PLAN.md
  - .planning/phases/471-pi-agent-session-shepherd/PRD-COVERAGE.md
  - .planning/phases/471-pi-agent-session-shepherd/PROMPTS.md
  - .planning/phases/471-pi-agent-session-shepherd/RUN-STATE.json
  - .planning/phases/471-pi-agent-session-shepherd/SUMMARY.md
  - .planning/phases/471-pi-agent-session-shepherd/TDD-LEDGER.md
  - .planning/phases/471-pi-agent-session-shepherd/VERIFICATION.md
  - .planning/phases/471-pi-agent-session-shepherd/agents/coordinator.md
  - .planning/phases/471-pi-agent-session-shepherd/agents/pi-sdk-scout.md
  - .planning/phases/471-pi-agent-session-shepherd/agents/repo-scout.md
  - .planning/phases/471-pi-agent-session-shepherd/traces/deep-review-remediation-trace.md
  - .planning/phases/471-pi-agent-session-shepherd/traces/github-topology-scout-trace.md
  - .planning/phases/471-pi-agent-session-shepherd/traces/green-workers-trace.md
  - .planning/phases/471-pi-agent-session-shepherd/traces/ownership-shutdown-review-trace.md
  - .planning/phases/471-pi-agent-session-shepherd/traces/pi-infrastructure-scout-trace.md
  - .planning/phases/471-pi-agent-session-shepherd/traces/pi-sdk-scout-trace.md
  - .planning/phases/471-pi-agent-session-shepherd/traces/pr-438-canary-trace.md
  - .planning/phases/471-pi-agent-session-shepherd/traces/preintegration-review-trace.md
findings:
  critical: 7
  warning: 3
  info: 0
  total: 10
status: issues_found
---

# Phase 471: Historical Foundation Exact-Head Review

> Historical only: this exact-head review predates the autonomous replacement scope. It validates
> the read-only foundation at its recorded checkpoint, not parent PR #472 or child issues #473-#481.
> A new final review is required after #481.

**Reviewed:** 2026-07-21T10:28:01Z
**Depth:** deep
**Base:** `74ab381eb8236305170ffd44d5aed74f8d0d2936`
**Exact pushed head:** `d434f6727441f17047185658f538612fcaf6187e`
**Status:** `issues_found` — release blocked

## Narrative Findings (AI reviewer)

### Summary

The exact pushed head and its tracking ref match. The production and test tree is unchanged from
the recorded implementation checkpoint `c1c5e9e9`; `d434f672` changes four phase-evidence files
only. The focused suite still passes 82/82, offline Pi 0.80.6 command discovery returns
`pm-shepherd`, and `git diff --check` passes. Those green gates do not cover the adversarial
interleavings and lifetime boundaries below.

Seven release blockers remain. Executable probes show that an accepted stop can make both start and
stop reject while disk stays `completed`; a direct `AgentRunner.abort()` can still return successful
evidence; Git configuration can hide an untracked file from the alleged clean target; the lease
journal deterministically bricks on run 257; malformed-lock recovery accepts a different issue
without proving owner death; filesystem-path reuse aliases a different Git repository to the same
persisted target; and the state/lease implementation still depends on Unix-only permission and
open-flag semantics despite the claimed Windows remediation.

Three additional warnings cover an unjoined pre-controller launch during shutdown, incomplete
persisted-state invariants, and root-directory check/use races that remain outside the opened
directory descriptor.

### Verification and adversarial evidence

- `node --test .pi/extensions/shepherd/*.test.ts`: **82 passed, 0 failed** at `d434f672`.
- Offline Pi RPC `get_commands`: **pass**, `pm-shepherd` discovered from the extension.
- `git diff --check 74ab381e..d434f672`: **pass**.
- Stop/final-save probe: start rejected with `invalid Shepherd state: stopped run has incompatible
  lanes`; stop rejected with `did not persist a stopped state`; disk remained `completed` with two
  `succeeded` lanes.
- Teardown-abort probe: `runner.abort("run-1")` completed while `waitForIdle()` was blocked, then the
  blocked cleanup was released; `run()` fulfilled with schema-valid evidence.
- Clean-target probe: a repository with `status.showUntrackedFiles=no` and `untracked.txt` present
  was accepted with `clean: true`.
- Lease-lifetime probe: acquire/assert/release cycles failed deterministically on iteration **257**
  with `successor chain exceeds its bound`; the state root contained 513 journal files.
- Crash-recovery probe: malformed `active.lock` text containing issue 471 was successfully taken
  over by `resume` for issue **999**.
- Worktree-identity probe: replacing one Git repository with a new repository at the same canonical
  path produced the exact same `CanonicalWorktree.identity`.
- State-invariant probe: a `completed` run with score `0.1`, a `read_only_violation` hard gate, and a
  `succeeded` lane carrying that same hard gate round-tripped successfully.
- Shutdown probe: `session_shutdown` resolved while an injected worktree resolver and its command
  handler remained unsettled.

The previously recorded root Go/build gates and generation-3 PR #438 canary were run at
`c1c5e9e9`. Their production code is byte-identical to this head, but neither exercised the
failures above.

### Prior finding dispositions

| Prior finding | Disposition at `d434f672` |
| --- | --- |
| CR-01 Pi error/abort terminals accepted | **Partial.** Typed terminal `stopReason` checks work, but abort during teardown still returns successful evidence (new CR-02). |
| CR-02 Resume drops persisted PR | **Resolved.** Omitted PR is inherited and a mismatched PR is rejected before capture. |
| CR-03 Lexical cwd splits/reuses ownership | **Partial.** Root/subdirectory/symlink aliases and cache keys are corrected, but the identity is still only a path and aliases a replacement repository (new CR-06). |
| CR-04 Stale takeover evicts live replacement | **Resolved for valid records.** Successor publication provides CAS behavior under the tested three-contender race. Malformed recovery remains unsafe (new CR-05). |
| CR-05 Lane failure releases before sibling exits | **Resolved.** The task group aborts and joins consumers before release. |
| CR-06 Stop loses terminal cancellation | **Partial.** Lease-release linearization is fixed, but the preceding terminal state-file write retains a cancellable await window (new CR-01). |
| CR-07 Setup outside timeout/ownership | **Resolved inside `SdkAgentRunner`.** Reload, creation, late creation, prompt, and cleanup are tracked; the extension's worktree-resolution setup is still untracked (WR-01). |
| CR-08 Symlink-following persistence | **Partial.** Final-file descriptor reads and POSIX symlink cases are improved. Root-relative operations still re-resolve pathnames (WR-03), and Windows lacks the relied-on flags (CR-07). |
| CR-09 Secret regex persisted credentials | **Resolved for persistence.** Disk summaries are fixed host-derived categories and arbitrary runner/model text is projected out. |
| CR-10 Windows cwd rejected | **Partial.** The SDK path parser accepts Windows-shaped strings, but the actual store/lease path remains Unix-specific (new CR-07). |
| WR-01 Crash publication leaves unrecoverable lock | **Partial.** Hard-link publication is atomic and process identity is recorded; malformed recovery bypasses issue/liveness proof (new CR-05). |
| WR-02 Contradictory persisted state accepted | **Partial.** Timestamp/dependency/basic lane-status checks exist, but completed/gate/score coherence remains unchecked (WR-02). |
| WR-03 Lease optional in controller port | **Resolved.** `acquireLease` is mandatory. |
| WR-04 Repeated close settles early | **Resolved.** All callers share one close promise and failure. |
| WR-05 Extension suppresses shutdown failure | **Resolved for registered controllers/active runs.** Failures and deadlines propagate and ownership is retained. Pre-controller launch work is not joined (WR-01). |
| WR-06 README names nonexistent Go Shepherd | **Resolved.** The README now calls it planned work. |

## Critical Issues

### CR-01: Stop still races the final state-file commit and produces an invalid state machine

**Classification:** BLOCKER
**File:** `.pi/extensions/shepherd/controller.ts:378-412`
**Also affects:** `.pi/extensions/shepherd/controller.ts:505-519`,
`.pi/extensions/shepherd/state-store.ts:167-192`

**Issue:** The controller converts lane outcomes to terminal lane statuses, computes the terminal
run status, and then awaits `persist()` while the lifecycle is still `active`. A stop arriving
during that await is accepted. After the first terminal snapshot is saved, line 411 calls
`persistCancelled()`, which only rewrites `running` or `pending` lanes. For a completed outcome the
lanes therefore remain `succeeded`, while the run becomes `stopped`; the validating file store
requires at least one `stopped` lane and rejects the write. Failed or halted lane outcomes are also
incompatible with both stopped and interrupted persistence.

This is a remaining form of prior CR-06. The existing regression gates lease release, after
`lifecycle.phase` has already become terminal; it never gates the final state save.

**Fix:** Choose the terminal linearization point before the final state save. After the last
cancellation check and after constructing the terminal snapshot, synchronously set
`lifecycle.phase = "terminal"`, then persist it. A stop that loses that race must reject as unowned;
a cancellation that wins must enter `persistCancelled()` before terminal lane statuses are
committed. Add real-`FileStateStore` tests that gate the final save for completed, failed, and halted
outcomes, then race both `stop()` and `shutdown()`.

### CR-02: Abort during teardown still returns successful AgentSession evidence

**Classification:** BLOCKER
**File:** `.pi/extensions/shepherd/sdk-runner.ts:382-407`
**Also affects:** `.pi/extensions/shepherd/sdk-runner.ts:411-418`

**Issue:** The post-prompt activity check closes the earlier prompt-resolution race, but `run()`
returns its parsed result only after an awaited `OwnedChild.cleanup()` in `finally`. If
`abort(runId)` or `close()` fires while cleanup is waiting for idle, the scope is cancelled and the
child is aborted, yet no activity check runs after cleanup. Once cleanup is released, `run()`
fulfills with the already parsed evidence. The direct probe returned `{status:"fulfilled"}` after
`abort("run-1")` had settled.

This leaves prior CR-01 only partially corrected and breaks the exported `AgentRunner.abort`
contract. Controller cancellation currently masks most user-visible cases, but callers and future
task-group paths cannot rely on an aborted lane rejecting.

**Fix:** Hold the parsed result until teardown completes and, on the success path, call
`#assertRunActive(request)` after cleanup and before `scope.finish()`, `#release()`, and promise
fulfilment. An abort/close/deadline that occurs anywhere before teardown completes must reject the
same run. Add abort-during-`waitForIdle`, close-during-`waitForIdle`, and deadline-during-dispose
regressions.

### CR-03: Repository Git configuration can hide a dirty target from both captures

**Classification:** BLOCKER
**File:** `.pi/extensions/shepherd/target-evidence.ts:155-178`

**Issue:** Cleanliness is inferred from `git status --porcelain` without overriding repository or
global configuration. `status.showUntrackedFiles=no` makes that command return an empty string even
when a non-ignored untracked file exists. The executable probe created `untracked.txt`, configured
that setting, and `captureTargetEvidence()` returned `clean: true`. Both initial and final captures
use the same command, so the dirtiness remains invisible and no `target_changed` gate is produced.

This violates the exact-clean-target contract and makes the recorded canary predicate dependent on
ambient Git configuration.

**Fix:** Invoke an explicit, configuration-resistant status form, for example
`git status --porcelain=v1 --untracked-files=all --ignore-submodules=none`, and override relevant
config keys with `-c` where necessary. Add real-repository tests for hidden untracked files and
submodule-ignore configuration. Consider explicitly detecting assume-unchanged/skip-worktree index
flags if those must also invalidate a canary.

### CR-04: The append-only lease journal permanently bricks the worktree on run 257

**Classification:** BLOCKER
**File:** `.pi/extensions/shepherd/state-store.ts:9-12`
**Also affects:** `.pi/extensions/shepherd/state-store.ts:591-617`,
`.pi/extensions/shepherd/state-store.ts:619-750`

**Issue:** Every acquisition and release appends one immutable successor, while `resolveLease()`
hard-fails after `MAX_LEASE_CHAIN = 512`. There is no epoch rotation, compaction, or garbage
collection. The 257th acquire can publish record 513, after which `assertOwned()`/`release()` fail
with `successor chain exceeds its bound`; all later lease operations are unusable. The probe
reproduced this exact failure with 513 files.

This is correctness and liveness failure, not merely a performance concern. Failed starts and
canaries also consume journal generations, so the bound is a repository/worktree lifetime limit.

**Fix:** Replace the unbounded historical chain with a bounded CAS head or a safely compactable
epoch design. Compaction/rotation must itself be serialized and must not permit an observer of an
old tail to overwrite a new live owner. Add sequential tests well beyond 512 transitions, crash
tests at every rotation publication point, and concurrent acquire/release/rotation tests. Clean up
or safely orphan old epoch files only after the new head is durably authoritative.

### CR-05: Malformed-lock recovery bypasses both same-issue and stale-owner proof

**Classification:** BLOCKER
**File:** `.pi/extensions/shepherd/state-store.ts:558-599`
**Also affects:** `.pi/extensions/shepherd/state-store.ts:708-745`, `.pi/README.md:105-107`

**Issue:** With `allowCrashPartial`, every JSON parse or metadata-validation failure is collapsed to
`{kind:"crash-partial"}`. `acquireLease()` then lets any `resume` claim publish
`.active.recovery`; it cannot compare the prior issue and performs no liveness or process-identity
check. A malformed record visibly containing issue 471 was successfully acquired by a resume claim
for issue 999. Corruption or a partial legacy/live publication can therefore bypass the global
ownership fence, contradicting the README claim that takeover follows proof of a stale same-issue
owner.

The new hard-link publisher writes complete metadata before publication, so empty/malformed
`active.lock` is not a legitimate crash point for the current format. Treating arbitrary malformed
data as safely recoverable weakens the atomic design.

**Fix:** Fail closed on malformed current-format records. If backwards recovery is required, use a
separately and atomically published recovery envelope that retains the prior issue, owner token,
PID, and process identity, then require the same issue and prove that owner stale before takeover.
Add wrong-issue, live-owner partial-publication, arbitrary-invalid-JSON, and stale recovery-epoch
tests.

### CR-06: Canonical worktree identity is only a pathname and can rebind persisted PR state

**Classification:** BLOCKER
**File:** `.pi/extensions/shepherd/extension.ts:79-97`
**Also affects:** `.pi/extensions/shepherd/index.ts:26-55`,
`.pi/extensions/shepherd/controller.ts:262-291`, `.pi/extensions/shepherd/domain.ts:48-62`

**Issue:** `canonicalizeGitWorktree()` returns `realpath(show-toplevel)` as both cwd and identity,
and the state-root fingerprint hashes only that string. Replacing a repository with an unrelated
Git repository at the same path produces the same controller cache key, state root, and lease. A
resume then inherits only the numeric PR and may bind it to the same-number PR in the replacement
repository. Conversely, moving the original worktree changes its identity and makes its state
unreachable.

The executable probe replaced a Git repository at one canonical path and received identical
identities before and after. Neither persisted state nor target evidence records a repository
owner/name, Git common-dir/worktree identity, or PR repository identity, so later target capture
cannot detect the rebind.

**Fix:** Derive and persist a stable repository plus worktree identity, not just a location. Include
canonical Git common-dir/worktree metadata (with device/inode or another replacement-resistant
identifier) and canonical repository identity such as the normalized remote/name-with-owner. Keep
linked worktrees distinct. Persist and verify the PR repository/URL on resume before inheriting its
number. Add move, path-reuse, new-clone-at-old-path, remote-change, and same-PR-number-across-repos
tests.

### CR-07: The Windows remediation stops at string parsing; persistence uses unavailable Unix primitives

**Classification:** BLOCKER
**File:** `.pi/extensions/shepherd/state-store.ts:3-5`
**Also affects:** `.pi/extensions/shepherd/state-store.ts:481-516`,
`.pi/extensions/shepherd/state-store.ts:558-648`,
`.pi/extensions/shepherd/state-store.ts:753-818`,
`.pi/extensions/shepherd/sdk-runner.test.ts:516-529`,
`.planning/phases/471-pi-agent-session-shepherd/VERIFICATION.md:10`

**Issue:** The added Windows test only passes Windows-shaped strings through the SDK request
validator while running on the current POSIX host. The real state path unconditionally ORs
`O_DIRECTORY` and `O_NOFOLLOW`, opens and syncs directories, calls `chmod(0700/0600)`, and requires
exact Unix owner/group/other mode bits on every root, state, and lease record. Node documents that
`O_DIRECTORY` and `O_NOFOLLOW` are not available on Windows and that Windows chmod can manipulate
only write permission without Unix owner/group/other distinctions. The implementation therefore
either fails its exact-mode checks or runs without the no-follow guarantees it claims. No Windows
store, controller, or extension test executes these paths.

See the official [Node file-system documentation](https://nodejs.org/api/fs.html#file-open-constants)
for the platform limits.

**Fix:** Either fail early and document the extension as unsupported on Windows, or implement a
Windows-specific secure state adapter using Windows ACL/private-directory and reparse-point-safe
handle semantics rather than Unix mode bits and unavailable flags. Run actual Windows CI covering
save/load, symlink/junction rejection, lease acquire/release/takeover, controller start/resume/stop,
and shutdown. A POSIX test accepting `C:\\...` text is not Windows coverage.

## Warnings

### WR-01: Shutdown reports success while pre-controller launch setup remains pending

**Classification:** WARNING
**File:** `.pi/extensions/shepherd/extension.ts:79-87`
**Also affects:** `.pi/extensions/shepherd/extension.ts:126-165`,
`.pi/extensions/shepherd/extension.ts:232-249`

**Issue:** `launchingIssue` records only a number. The promise returned by asynchronous worktree
resolution is neither stored nor cancellable, and the `git rev-parse` child has no timeout or abort
signal. `session_shutdown` waits only registered controllers and `activeRun`. The probe used a
never-settling resolver: shutdown resolved successfully while the command handler remained pending.
A stuck Git/filesystem operation can therefore outlive a reportedly successful shutdown and keep
the process alive.

**Fix:** Track a launch/setup promise and AbortController before awaiting worktree resolution. Give
the Git child a bounded timeout/signal, cancel it on shutdown, and include every launch/setup promise
in the shutdown settlement set. Add shutdown-during-resolution and late-controller-creation tests.

### WR-02: Persisted-state validation still accepts completed runs with hard gates and failing scores

**Classification:** WARNING
**File:** `.pi/extensions/shepherd/state-store.ts:167-236`

**Issue:** The new validator checks basic run/lane status compatibility, but it does not enforce the
decision invariants that give those statuses meaning. A `completed` run with score `0.1`, a
`read_only_violation` hard gate, and a `succeeded` lane carrying the same gate and score was accepted
and reloaded. `status()` presents the contradictory record, while `resume()` refuses recovery solely
because the outer status is completed.

**Fix:** Require completed runs and succeeded lanes to have empty hard gates and scores meeting the
proceed threshold; require failed/corrected and halted lanes to have compatible score/gate shapes;
and validate or recompute the aggregate run score from lane scores. Add corrupted-disk and direct
`save()` fixtures for every status/gate/score contradiction.

### WR-03: RootGuard does not make child operations descriptor-relative

**Classification:** WARNING
**File:** `.pi/extensions/shepherd/state-store.ts:547-648`
**Also affects:** `.pi/extensions/shepherd/state-store.ts:753-818`

**Issue:** `prepareRoot()` opens the canonical root and records its inode, but state and lease files
are still opened, linked, renamed, and removed through `join(root.path, name)`. `assertRootGuard()`
is a separate pathname check before those operations. If the leaf root is renamed and replaced
between the check and child open, `O_NOFOLLOW` protects only the final component; an intermediate
replacement symlink can redirect the operation. The opened directory handle is used only for
`sync()`, not for child resolution. The existing swap test replaces the final state pathname after
its descriptor is open and does not exercise replacement of the root directory itself.

**Fix:** Use descriptor-relative child operations where the platform permits them. Otherwise,
revalidate the root immediately after each child descriptor/open and before reading or writing,
compare parent/root identity again after publication, and fail closed on any change. Add
deterministic whole-root swap hooks around state open, lease open, link, and rename. Do not describe
the design as descriptor-anchored until child path resolution is actually anchored.

---

_Reviewed: 2026-07-21T10:28:01Z_
_Reviewer: the agent (gsd-code-reviewer)_
_Depth: deep_
_Release status: blocked — 7 blockers, 3 warnings_
