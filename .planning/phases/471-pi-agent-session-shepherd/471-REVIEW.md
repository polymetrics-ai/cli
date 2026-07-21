---
phase: 471-pi-agent-session-shepherd
reviewed: 2026-07-21T09:36:01Z
depth: deep
files_reviewed: 10
files_reviewed_list:
  - .pi/README.md
  - .pi/extensions/shepherd/arguments.ts
  - .pi/extensions/shepherd/controller.ts
  - .pi/extensions/shepherd/domain.ts
  - .pi/extensions/shepherd/extension.ts
  - .pi/extensions/shepherd/index.ts
  - .pi/extensions/shepherd/runner.ts
  - .pi/extensions/shepherd/sdk-runner.ts
  - .pi/extensions/shepherd/state-store.ts
  - .pi/extensions/shepherd/target-evidence.ts
findings:
  critical: 10
  warning: 6
  info: 0
  total: 16
status: issues_found
---

# Phase 471: Historical Foundation Code Review

> This review covers the earlier read-only control-plane checkpoint only. Issue #471 was later
> expanded into the complete autonomous replacement described by #472 and #473-#481. Findings and
> fixes remain regression evidence; this document is not final release approval.

**Reviewed:** 2026-07-21T09:36:01Z
**Depth:** deep
**Files Reviewed:** 10
**Status:** issues_found

## Summary

The exact pushed implementation at `7f745427d38995940b8f57517d0241d1e10d3f64` was reviewed against issue #471 and the Pi 0.80.6 SDK contract. The 49 focused Node tests, strict production TypeScript check, offline command-registration probe, and `git diff --check` all pass, but executable adversarial probes exposed ten release-blocking correctness, isolation, cancellation, persistence, and lease failures. Several of these permit overlapping child sessions after ownership has been released, accept failed Pi turns as successful validation, or weaken the persisted target identity on resume.

No production source was modified by this review.

## Narrative Findings (AI reviewer)

## Critical Issues

### CR-01: Pi error and abort terminals are accepted as successful evidence

**File:** `.pi/extensions/shepherd/sdk-runner.ts:259-289`

**Issue:** `prompt()` resolving is treated as success. The event subscriber only counts bytes/events, and the runner then parses `getLastAssistantText()` without checking the terminal assistant message's `stopReason`. Pi resolves prompts whose terminal reason is `error` or `aborted`, and partial assistant text remains available. A fake Pi error terminal containing schema-valid evidence was accepted; with two such lanes, the controller persisted `status: "completed"`, no hard gates, and two succeeded lanes. The same missing post-prompt cancellation check lets direct `runner.abort(runId)` return late evidence.

**Fix:** Track typed `message_end`/`agent_end` events, require one final assistant terminal with `stopReason === "stop"`, and reject `error`, `aborted`, `length`, and `toolUse`. Parse the verified terminal message rather than `getLastAssistantText()`, and call `#assertRunActive(request)` immediately after `withTimeout()`. Add controller-level tests for every non-`stop` terminal reason and for abort racing with prompt resolution.

### CR-02: Resume silently drops a persisted PR binding

**File:** `.pi/extensions/shepherd/controller.ts:238-257`

**Also affects:** `.pi/extensions/shepherd/controller.ts:303-315`, `.pi/extensions/shepherd/arguments.ts:115-166`, `.pi/README.md:99-110`

**Issue:** `resume()` loads the previous run but passes the new command directly to `execute()`. Because the documented resume syntax permits omitting `--pr`, a PR-bound interrupted run is resumed as issue-only validation: the new target capture does not call `gh pr view`, and the replacement state omits `pr`. A reproduction starting from persisted `issue=397, pr=438` produced `capturePRs:[null,null]`, `resultPR:null`, `persistedPR:null`, and `status:"completed"`. Supplying a different PR is likewise not rejected as an identity mismatch.

**Fix:** Derive an effective resume command from persisted identity: inherit `existing.pr` when the flag is absent and reject any supplied PR that differs. Use that effective command for both target captures and persistence. Add regressions for omitted, matching, and mismatched `--pr`.

### CR-03: Lexical cwd identity splits leases and can reuse a controller for the wrong target

**File:** `.pi/extensions/shepherd/index.ts:25-26`

**Also affects:** `.pi/extensions/shepherd/index.ts:47-59`, `.pi/extensions/shepherd/extension.ts:93-103`, `.pi/README.md:102-105`

**Issue:** The state/lease key is a hash of `resolve(cwd)`, not a canonical Git worktree identity. The repository root, a subdirectory, and a symlink alias therefore receive different `active.lock` files while targeting the same checkout; root and subdirectory were reproduced with distinct fingerprints. Conversely, the extension caches controllers only by issue. A status in cwd A followed by a start for the same issue in cwd B reused A's controller and captured A. Both behaviors break the claimed repository/worktree-global lease and can validate or overlap the wrong target.

**Fix:** Resolve and verify one canonical target before controller creation (for example, `realpath(await git rev-parse --show-toplevel)` plus Git common-dir/worktree identity). Use it for evidence capture and the state fingerprint, key the controller cache by `(canonicalWorktree, issue)`, and reject context changes. Test root/subdirectory/symlink aliases and two distinct worktrees with the same issue.

### CR-04: Stale takeover can evict a newly acquired live lease

**File:** `.pi/extensions/shepherd/state-store.ts:423-450`

**Also affects:** `.pi/extensions/shepherd/state-store.ts:477-510`

**Issue:** `removeStaleLease()` checks the expected inode, awaits process liveness, then renames whatever currently occupies `active.lock`. During that await, another process can remove the stale lease and acquire a live lease. The original contender then renames that new owner's lock; while the pathname is absent, a third process can acquire it. A deterministic three-contender probe ended with B losing ownership and C owning the active lock while B's child could still run.

**Fix:** Serialize the entire stale-check/takeover/acquisition transaction with an OS-backed advisory lock or a CAS-capable claim scheme. Never rename the shared pathname based on a pre-await identity check. Add a deterministic interleaving test that proves a live replacement can never be displaced.

### CR-05: A lane persistence failure releases the lease while a sibling session is still running

**File:** `.pi/extensions/shepherd/controller.ts:157-173`

**Also affects:** `.pi/extensions/shepherd/controller.ts:228-234`, `.pi/extensions/shepherd/controller.ts:333-337`, `.pi/extensions/shepherd/controller.ts:377-415`, `.pi/extensions/shepherd/controller.ts:435-443`

**Issue:** `mapWithLimit()` uses fail-fast `Promise.all`. If one lane's pre-dispatch persistence rejects, `start()` immediately releases the lease, removes lifecycle ownership, and resolves `done`; sibling consumers are neither cancelled nor joined. An injected save failure reproduced `released:true`, `running:1`, and `stopResult:"not-owned"`. A new process can therefore acquire the lease while the orphaned AgentSession continues.

**Fix:** Implement structured task-group semantics: on the first worker failure, mark the lifecycle cancelled, abort the runner, await every consumer with `Promise.allSettled`, and only then release the lease/remove ownership. Test that `start()` cannot settle or release its lease until a blocked sibling has exited.

### CR-06: Stop races with terminal lease release and cannot persist the accepted cancellation

**File:** `.pi/extensions/shepherd/controller.ts:268-280`

**Also affects:** `.pi/extensions/shepherd/controller.ts:228-234`, `.pi/extensions/shepherd/controller.ts:371-374`

**Issue:** After the final cancellation check, `execute()` returns a terminal state while its lifecycle remains in `activeRuns` during asynchronous lease release. `stop()` in that window accepts ownership, sets `cancelReason`, and waits, but no code remains to persist `stopped`. A release-gated reproduction returned a completed start, then threw `did not persist a stopped state`, with persisted status still `completed`. Shutdown has the same terminal window.

**Fix:** Linearize terminal commit and lifecycle ownership. Before awaiting lease release, atomically transition the lifecycle to a terminal/non-stoppable phase so late stop rejects as unowned (or returns the terminal result); if cancellation wins first, persist it before terminal commit. Cover the lease-release interleaving with a deterministic test.

### CR-07: Session setup is outside every timeout and cannot be cleaned up

**File:** `.pi/extensions/shepherd/sdk-runner.ts:209-255`

**Also affects:** `.pi/extensions/shepherd/sdk-runner.ts:276-285`, `.pi/extensions/shepherd/sdk-runner.ts:306-329`, `.pi/extensions/shepherd/extension.ts:185-191`

**Issue:** The request deadline wraps only `session.prompt()`. `resourceLoader.reload()` and `createSession()` can hang indefinitely before an `OwnedChild` is registered, so abort/close has nothing to clean up. With `createSession()` never resolving, a 10 ms request timeout remained pending and `close()` only timed out waiting for idle. This pins controller ownership/lease indefinitely in a live process, while the extension eventually suppresses its 45-second shutdown failure.

**Fix:** Apply one deadline/abort signal across reload, session creation, prompt, and teardown. Track setup promises as owned operations; if creation resolves after cancellation, immediately abort/dispose the late session. Make close join all setup and child tasks before reporting completion. Add hanging-reload, hanging-create, and late-create tests.

### CR-08: Symlink-following persistence permits out-of-root writes and spoofed state

**File:** `.pi/extensions/shepherd/state-store.ts:311-343`

**Also affects:** `.pi/extensions/shepherd/state-store.ts:515-535`, `.pi/extensions/shepherd/state-store.ts:538-565`

**Issue:** Root preparation follows a symlink and chmods its target. State loading uses pathname `stat()` plus `readFile()`, which follows a symlinked `issue-N.json`, and the post-rename pathname `chmod()` introduces another swap-follow race. Probes caused a symlinked root to receive an external state write and be changed to mode 0700, and caused a symlinked issue file to be trusted as the persisted run. This violates state integrity and the bounded path contract.

**Fix:** Anchor the store beneath the canonical trusted Pi agent directory, `lstat`/validate every root component, and reject symlink or broad-root targets. Open state files with `O_NOFOLLOW`, then bound and validate the same handle with `fstat`/`readFile`; rely on the already-0600 temp inode after same-directory rename rather than pathname `chmod`. Add root, state-file, and swap-race tests.

### CR-09: Regex redaction persists common credential formats unchanged

**File:** `.pi/extensions/shepherd/state-store.ts:159-188`

**Also affects:** `.pi/extensions/shepherd/controller.ts:349`, `.pi/extensions/shepherd/controller.ts:423-430`, `.pi/extensions/shepherd/sdk-runner.ts:542-568`, `.pi/README.md:101-104`

**Issue:** Model summaries and arbitrary runner/provider error messages are persisted after a narrow regex pass. Tests with JSON (`{"token":"..."}`), `OPENAI_API_KEY=...`, `AWS_SECRET_ACCESS_KEY=...`, and credential-bearing `DATABASE_URL` values all survived. This directly contradicts the issue's hard stop and the README promise that credentials are not persisted.

**Fix:** Do not persist arbitrary provider errors or model-authored free text as a secrecy boundary. Store host-derived structured categories and fixed error codes. If summaries remain, handle quoted JSON, quoted/unquoted environment assignments, authorization headers, and credential-bearing URLs, then reject secret-like output before saving. Add a broad marker corpus and assert that no marker reaches disk.

### CR-10: Absolute-path validation rejects every native Windows checkout

**File:** `.pi/extensions/shepherd/sdk-runner.ts:439-441`

**Issue:** A cwd is accepted only when it starts with `/` and is split on `/`. Native Windows paths such as `C:\repo` and UNC paths therefore fail before session creation, even though Polymetrics publishes native Windows builds. The command is unusable on that supported platform.

**Fix:** Use platform-native `path.isAbsolute()` and platform-aware segment/canonicalization checks rather than POSIX string rules. Add Windows CI coverage for drive-letter and UNC worktree paths.

## Warnings

### WR-01: A crash during lease publication creates an unrecoverable lock

**File:** `.pi/extensions/shepherd/state-store.ts:274-301`

**Also affects:** `.pi/extensions/shepherd/state-store.ts:363-395`, `.pi/extensions/shepherd/state-store.ts:477-510`

**Issue:** `open("wx")` exposes an empty `active.lock` before metadata is written. A crash in that window leaves an empty or partial file, and `readLeaseRecord()` rejects before resume can reach stale-owner recovery. A zero-length mode-0600 lock reproduced `invalid ... lock is not a bounded regular file`. PID-only liveness can also mistake a reused PID for the old owner.

**Fix:** Fully write and fsync a unique mode-0600 owner file, then atomically publish it with a hard link/advisory lock. Include process-start/boot identity rather than PID alone, and add crash-point recovery tests.

### WR-02: Persisted-state validation accepts contradictory state machines

**File:** `.pi/extensions/shepherd/state-store.ts:102-156`

**Issue:** Validation checks individual fields but not cross-field invariants. It accepted a `completed` run containing a `running` lane, non-date timestamps, and a dependency on a missing lane. `status()` trusts the contradictory record while `resume()` refuses it because the outer status says completed, making recovery misleading or impossible.

**Fix:** Require canonical timestamps; validate dependency existence and acyclicity; and enforce run/lane coherence (terminal runs contain only compatible terminal lanes, with completed runs containing successful lanes). Add corrupted-state fixtures for every invariant.

### WR-03: The controller contract permits stores with no lease enforcement

**File:** `.pi/extensions/shepherd/controller.ts:37-45`

**Also affects:** `.pi/extensions/shepherd/controller.ts:217-221`, `.pi/extensions/shepherd/controller.ts:241-245`

**Issue:** `StateStore.acquireLease` and both calls are optional, so a valid exported controller dependency can silently disable the ownership fence. The production entry point currently supplies `FileStateStore`, but the public port makes an unsafe adapter type-correct and easy to introduce in future wiring.

**Fix:** Make lease acquisition mandatory in `StateStore`; supply an actually exclusive in-memory lease implementation to tests rather than using optional chaining.

### WR-04: Repeated close calls report success before cleanup finishes

**File:** `.pi/extensions/shepherd/sdk-runner.ts:314-333`

**Issue:** `#closed` is set before cleanup completes, so concurrent later `close()` calls return immediately. A second caller observed success while a child remained undisposed; after the first close fails, all later closes also resolve successfully rather than reporting the same failure.

**Fix:** Store one `#closePromise` and return it to every caller. Mark the runner successfully closed only after that promise completes, and preserve/rethrow its failure consistently. Test concurrent and retry-after-failure calls.

### WR-05: Extension shutdown suppresses every cleanup failure

**File:** `.pi/extensions/shepherd/extension.ts:185-194`

**Issue:** Controller shutdown results are discarded with `Promise.allSettled`, and the outer deadline rejection is swallowed with `.catch(() => undefined)`. A controller that threw `cleanup failed` still produced a successful extension shutdown, after which controller ownership was cleared. Operators receive no indication that sessions or leases may remain.

**Fix:** Inspect and aggregate settled failures, report them through Pi's extension error path, and propagate the shutdown deadline failure. Clear controller ownership only after successful cleanup or retain explicit failed-shutdown state.

### WR-06: Documentation points users to a nonexistent standalone Go Shepherd

**File:** `.pi/README.md:107-111`

**Issue:** The README says the standalone Go Shepherd “remains the durable option,” but this repository contains no implemented Go Shepherd command; the referenced durable controller remains planned work. Users cannot follow the stated recovery/operating alternative.

**Fix:** Describe it as planned design work until an executable command exists, or link to the exact implemented command and its usage if one is added.

---

_Reviewed: 2026-07-21T09:36:01Z_
_Reviewer: the agent (gsd-code-reviewer)_
_Depth: deep_
