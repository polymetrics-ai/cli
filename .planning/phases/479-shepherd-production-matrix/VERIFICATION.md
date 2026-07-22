# Verification

## Verdict

**FUNCTIONAL PASS — 17/17 production-matrix rows.**

Production code was frozen at `9169241531dc0f88710bcfc637a272dd379dec99`. The later
future-safe test fixture and this documentation freeze do not change production code. One
consolidated Codex 5.6 Sol xhigh review returned three release blockers; one bounded correction
pass supplied behavior RED → GREEN evidence, and the same reviewer marked every finding **CLOSED**
with final disposition **VERIFIED** at the production code head.

The historical GSD/TDD process gap remains recorded separately: the original implementation did
not preserve behavior-level RED evidence for every row. The final correction blockers do have
genuine RED → GREEN commits. Functional verification must not be misreported as retroactive proof
of the missing original RED checkpoints.

## Final machine gates

| Gate | Result |
|---|---|
| Combined production release suite | **PASS — 808/808**, sequential, 0 failed, 0 skipped |
| Exact Git integration CAS | **PASS — 3/3**, including parent race, default alias, deterministic retry |
| Generation-2 recovery | **PASS** — runtime 10/10, recovery 83/83, controller 34/34 |
| Complete Shepherd inventory | Final-head rerun at test head `7e02778e`: 1,685 total, **1,620 pass, 64 fail, 1 skipped**, 108.1s. All 64 local failures enter the file/Git/worktree lease path and terminate at the managed sandbox's `spawn EPERM`; the nine expired test-fixture failures from the prior run were corrected and now pass. Fresh CI must rerun this gate outside the managed sandbox. |
| Strict production TypeScript | **PASS** — 46 non-test Shepherd modules, TS 5.9.3, strict/no-emit, ES2024 + NodeNext, pinned Pi type roots |
| Offline Pi registration | **PASS** — pinned Pi 0.80.6 RPC exposes `pm-shepherd` from the production extension |
| Real command RPC | **PASS** — bare/help, invalid action, and read-only `status --issue 479`; no model, auth, or network call |
| Plan/intake/docs | **PASS — 4/4** production plan/intake tests; README schema-2 sample remains valid |
| Diff/worktree hygiene | **PASS** — `git diff --check`; worktree clean before the docs freeze |
| Independent blocker disposition | **VERIFIED** — all three review findings closed at `91692415` |

The broad Go/connector gates are intentionally not part of this TypeScript/Pi sub-issue. They
remain parent-head gates and do not substitute for Shepherd verification.

## 17-row production matrix

| # | Verdict | Direct final-head evidence |
|---:|:---:|---|
| 1 | **PASS** | Genuine `index.ts` composition reaches reviewed schema-2 intake, parallel children, isolated workspaces, bounded verification, stacked PRs, exact-head review/correction, CAS integration, parent finalization, and exact human wait. Controller trajectory and host/runtime tests execute the path. |
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

## Security and authority result

- The model receives bounded scoped workspace capabilities, not generic shell, Git, GitHub, or
  merge-main tools.
- Verification uses configured executable/argv with `shell: false`, bounded cwd/environment,
  output, timeout, and cancellation.
- Child integration constructs a deterministic exact reviewed merge and updates only the
  non-default parent ref under `--force-with-lease=<parent>:<baseSha>`. GitHub never performs the
  merge; it observes exact evidence and publishes the receipt.
- GitHub authentication remains ambient host authority. No token value is requested, persisted,
  printed, summarized, or placed in a prompt.
- The parent PR to the default branch remains human-owned. Shepherd completes only after observing
  the exact authoritative human merge.

## Remaining external gates

1. Fresh CI must run the complete Shepherd suite where the process-identity/Git/worktree commands
   are permitted; the managed Codex sandbox blocks those 64 tests at `spawn EPERM` before their
   assertions execute.
2. Live GitHub mutation was not run because no designated sandbox repository was authorized. The
   typed transport and bare-remote CAS fixtures provide local evidence; a live sandbox smoke remains
   optional before parent integration.
3. No branch was pushed and no parent/default-branch merge was performed. Parent integration stays
   behind the repository's human gate.
