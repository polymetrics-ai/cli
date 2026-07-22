# Verification

## Verdict

**FUNCTIONAL PASS — 17/17 production-matrix rows.**

The original 17-row implementation was frozen at `9169241531dc0f88710bcfc637a272dd379dec99`.
The 2026-07-22 release-blocker correction is frozen at
`78708cbef64b33e54ed32078bf2a107d81126236`. The 2026-07-23 merge-readiness implementation
checkpoint is `307ea409648e2f293c8a48cc957ffc312cc44542`; deterministic Pi-family verification is
`a594be98`; parent reconciliation is `45c27b9d` and its child merge is `766709b3`. These
checkpoints add CI and evidence only and do not change production code. A bounded independent audit
returned **PASS** for the four requested release
paths after one correction pass and targeted cross-layer closure.

The historical GSD/TDD process gap remains recorded separately: the original implementation did
not preserve behavior-level RED evidence for every row. The release-blocker correction has genuine
behavior RED → GREEN checkpoints recorded in `TDD-LEDGER.md`. Functional verification must not be
misreported as retroactive proof of the missing original RED checkpoints.

## Final machine gates

| Gate | Result |
|---|---|
| Combined production release suite | **PASS — 808/808**, sequential, 0 failed, 0 skipped |
| Release-blocker affected suite | **PASS — 767/767**, sequential, 0 failed, 0 skipped, at code head `78708cbe` |
| Exact Git integration CAS | **PASS — 3/3**, including parent race, default alias, deterministic retry |
| Generation-2 recovery | **PASS** — runtime 10/10, recovery 83/83, controller 34/34 |
| Complete Shepherd inventory | Exact merge-readiness checkpoint rerun: 1,712 total, **1,647 pass, 64 fail, 1 skipped**, 93.6s. All 64 failures enter the file/Git/worktree lease path under the managed sandbox through `defaultProcessIdentity`; no independent product failure class appeared. Fresh CI must rerun this gate outside the managed sandbox. |
| Complete-suite CI gate | **CONFIGURED; REMOTE GREEN PENDING** — `.github/workflows/shepherd.yml` uses read-only contents permission, exact Node 24.13.1, exact Pi 0.80.6, no secrets, and the complete sequential inventory command. |
| Strict production TypeScript | **PASS** — 49 non-test Shepherd modules, TS 5.9.3, strict/no-emit, ES2024 + NodeNext, pinned Pi type roots |
| Offline Pi registration | **PASS** — pinned Pi 0.80.6 RPC exposes `pm-shepherd` from the production extension |
| Real command RPC | **PASS** — bare/help, invalid action, and read-only `status --issue 479`; no model, auth, or network call |
| Plan/intake/docs | **PASS** — 4/4 bootstrap tests, including real `host_inspect` tool-policy compilation and real GitHub orchestration-plan construction; README matches generated-plan and parent-draft behavior |
| Diff/worktree hygiene | **PASS** — branch-range and worktree `git diff --check`; six committed blank-EOF defects are repaired |
| Independent release-blocker disposition | **PASS** — bounded final audit confirmed all four requested paths and their cross-layer contracts at `78708cbe` |

The broad Go/connector gates are intentionally not part of this TypeScript/Pi sub-issue. They
remain parent-head gates and do not substitute for Shepherd verification.

## 2026-07-23 merge-readiness closure

- RED proved that no workflow ran the complete Shepherd inventory, the production-matrix summary
  was absent, and the child range failed diff hygiene in six files.
- GREEN adds a path-scoped pull-request/main-push/workflow-dispatch CI job on an ordinary
  GitHub-hosted runner. The only globally installed test runtime is the already-required exact Pi
  0.80.6 package; install scripts are disabled.
- Local exact-head checks pass for workflow structure, strict TypeScript, offline Pi registration,
  GSD/TDD evidence, and diff hygiene. The full local suite reproduces only the known managed-sandbox
  `/bin/ps` block.
- Codex 5.6 Sol xhigh review at `ca3f6c6f` confirmed the local failure classification and found two
  release blockers. The exact Pi-family blocker is closed at `a594be98`; the local parent-range
  reconciliation is closed at `45c27b9d`/`766709b3`. Exact-head follow-up, authoritative
  parent-first publication, policy review coverage, and the first remote CI result remain pending.

## 17-row production matrix

| # | Verdict | Direct final-head evidence |
|---:|:---:|---|
| 1 | **PASS** | Genuine `index.ts` composition reuses a valid schema-2 plan or bootstraps one from authoritative issue facts through an xhigh planning AgentSession and host-materialized child issues, then reaches parallel children, isolated workspaces, bounded verification, stacked PRs, exact-head review/correction, CAS integration, parent finalization, and exact human wait. |
| 2 | **PASS** | Scheduler/controller tests prove dependency ordering, deterministic disjoint selection, canonical write-scope collisions, concurrency capacity, and persisted idle reasons. |
| 3 | **PASS** | Workspace and AgentSession tests prove disjoint mutators coexist, overlapping scopes wait, and join/abort releases exactly the accepted child's lease. |
| 4 | **PASS** | Stale siblings refresh/rebase/reclaim and invalidate verification/review. Exact parent-base CAS preserves an unrelated parent advance and routes the child through refresh/reverify/rereview. |
| 5 | **PASS** | Attempts/corrections are bounded and durable; exhaustion creates an exact issue decision instead of prose success. Generation-2 recovery preserves one request and one authorized retry. |
| 6 | **PASS** | Four crash windows are exercised for all 14 external-effect kinds. Fresh-process tests also cover parent request/merge observations and Git CAS before GitHub receipt without duplicate mutation. |
| 7 | **PASS** | Broker/gate tests reject stale, unauthorized, bot, edited, duplicate, ambiguous, consumed, hostile, emoji, prose, CI, and silent replies; only one exact allowlisted command is accepted. |
| 8 | **PASS** | Stop tests cover intake, every child stage, correction, parent finalization, request, and poll; accepted work is aborted and joined before stopped state and lease release. |
| 9 | **PASS** | CAS generation fences, sibling abort, persistence failures, late initialization, shutdown deadlines, and stop/shutdown races fail closed. |
| 10 | **PASS** | Commit, push, PR, integration, parent-ready, and receipt timeout-after-effect paths reconcile authoritative state before retry and do not republish. |
| 11 | **PASS** | Stable ownership survives resume; changed plan/scope/base/policy fails before mutation. Execution generation is separated from stable resource generation. |
| 12 | **PASS** | Findings need exact dispositions plus a causally later clean review. Child/head or parent movement invalidates authorization; fresh verification/review is required before CAS integration. |
| 13 | **PASS** | Dirty/scope-escaped work, wrong branch/head/base, draft child PRs, untrusted/incomplete CI, forged receipts, and prose-only review are rejected. The final ref update is exact-old-SHA leased. |
| 14 | **PASS** | Start/resume re-inspect live remote-default authority; `main`, `master`, and `trunk` are protected integration targets. Git rechecks symbolic HEAD immediately before CAS, GitHub rechecks before receipt, parent approval is invalidated by head movement, and no operation merges the parent PR to the default branch. |
| 15 | **PASS** | Contract/intake/tool/Git/GitHub/verification tests cover traversal, symlinks, terminal controls, hostile/sparse/proxied payloads, bounded argv/output/time, cancellation, hooks/helpers/transports, and secret-bearing text rejection. |
| 16 | **PASS** | Argument/extension tests plus real offline RPC cover help, bare command, invalid command, status, one active run, initialization stop, and shutdown/join behavior. |
| 17 | **PASS** | Schema-2 intake rejects top-level read-only children. Internal review roles are read-only and cannot obtain mutation leases, integrate, or fabricate receipts. |

## Review correction disposition

| Finding | RED | GREEN | Disposition |
|---|---|---|---|
| Generation-2 intervention recovery used resource generation | `0fe22e9e` | `e2dedad7` | **CLOSED** |
| Start/resume and mutation boundary did not fully fence default targets | `06e50e21` | `a8104613` plus CAS boundary | **CLOSED** |
| Child merge lacked atomic exact parent-base CAS | `5ef7ba15` | `37dbc42c` | **CLOSED** |
| Correction-discovered crash after CAS before receipt | `32a0d50e` | `91692415` | **CLOSED** |

## 2026-07-22 release-blocker disposition

| Requested blocker | Final behavior | Disposition |
|---|---|---|
| Missing issue plan | `start --issue N` observes authoritative issue/repository facts, runs one xhigh planning AgentSession, validates the semantic proposal, creates/reconciles marker-bound child issues, inserts only returned issue numbers, and atomically publishes the ignored schema-2 plan. Existing invalid/conflicting plans are never overwritten. | **CLOSED** |
| Agent-run test feedback | Implementation/correction sessions can call ID-only `host_verify` for RED→GREEN; independent verification requests every immutable ID in exact order; the host executes closed command tuples and returns bounded redacted diagnostics that drive correction. | **CLOSED** |
| Useful autonomous authority | Mutating agents keep scoped read/edit/write authority inside their child worktree. They do not receive generic shell, raw argv/environment, secrets, GitHub mutation, or parent/default-branch merge authority. | **CLOSED** |
| Same-second ready transition | Exact `draft:false` evidence with `revision >= expectedRevision` is accepted; equal revision while still draft and every moved-head observation remain rejected and restart-safe. | **CLOSED** |

The bounded review also closed command-recipe injection, descendant process escape, AgentSession
exception bypass, plan/verifier bounds, real GitHub orchestration parity, topology/scope parity, and
real planning tool-schema compilation. Exact RED/GREEN counts are recorded in `TDD-LEDGER.md`.

## Security and authority result

- The model receives bounded scoped workspace capabilities, not generic shell, Git, GitHub, or
  merge-main tools.
- Verification uses closed Node/Go/Make quality-gate recipes, canonical host-owned executables,
  `shell: false`, bounded cwd/environment/output/time, cancellation, POSIX process-group termination,
  and a hard post-kill settlement bound.
- Child integration constructs a deterministic exact reviewed merge and updates only the
  non-default parent ref under `--force-with-lease=<parent>:<baseSha>`. GitHub never performs the
  merge; it observes exact evidence and publishes the receipt.
- GitHub authentication remains ambient host authority. No token value is requested, persisted,
  printed, summarized, or placed in a prompt.
- The parent PR to the default branch remains human-owned. Shepherd completes only after observing
  the exact authoritative human merge.

## Remaining external gates

1. Publish/fetch parent `45c27b9d` first, verify the authoritative remote head, then push the
   non-default child branch and run the new complete-suite CI gate where
   process-identity/Git/worktree commands are permitted; the managed Codex sandbox blocks those 64
   tests at `spawn EPERM` before their assertions execute.
2. Obtain one bounded Codex 5.6 Sol xhigh follow-up review of the exact child handoff head.
3. Obtain repository-policy `claude_auto` coverage or an allowed recorded fallback on the
   authoritative PR range.
4. Live GitHub mutation was not run because no designated sandbox repository was authorized. The
   typed transport and bare-remote CAS fixtures provide local evidence; a live sandbox smoke remains
   optional before parent integration.
5. No parent/default-branch merge was performed. Parent integration stays
   behind the repository's human gate.

## 2026-07-23 PR #489 CI repair checklist

Current verdict: **REPAIR IN PROGRESS**.

- [x] Diagnose the complete-suite failure as fixture cleanup after an unpinned CI Go toolchain
  auto-download; no Shepherd assertion failed independently.
- [x] Diagnose the security failure as inherited `golang.org/x/text v0.36.0`; current `main`
  already contains the fixed `v0.39.0` baseline.
- [ ] Parent branch contains the current `main` baseline without conflicts.
- [ ] Shepherd CI installs exact Go `1.25.12` before the complete inventory.
- [ ] Focused real-Go AgentSession verification passes.
- [ ] Complete Shepherd inventory passes outside any known managed-sandbox process restriction.
- [ ] `go mod verify` and `govulncheck ./...` pass on the synchronized child head.
- [ ] Parent and child exact heads are published in order; PR #489 CI is green.
- [ ] Fresh exact-head review is complete before parent integration.

Cleanup policy: do not mask `EACCES`, leave downloaded toolchains behind, weaken the security scan,
or merge the human-gated parent PR into `main`.
