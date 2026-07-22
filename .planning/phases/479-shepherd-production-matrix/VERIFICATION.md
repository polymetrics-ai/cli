# Verification

## Verdict policy

`PASS` below means the required behavior has direct tests and is reachable through the production
composition being assessed. `OPEN` means a concrete mechanism is absent or incomplete; green helper
tests do not close that row. Machine counts and the exact implementation head are frozen only after
all shared code changes are committed.

The pre-correction baseline was commit `7fecbe86`. Its sequential focused audit passed 82/82 tests,
but that did not make the production extension executable: `index.ts` still selected the legacy
local autonomous controller. A later 15-file domain run passed 107/107 tests. These are supporting
regression results, not the final-head gate.

## Final-head machine gates

| Gate | Command/scope | Result |
|---|---|---|
| Focused production matrix | 15 production contract/intake/scheduler/state/effect/recovery/verification/workspace/review/human/parent/pipeline/controller/argument/extension test files, sequential | PENDING final code freeze |
| Complete Shepherd suite | `node --test --test-concurrency=1 .pi/extensions/shepherd/*.test.ts` | PENDING final code freeze |
| Strict TypeScript | TypeScript 5.9.3, strict/no-emit, ES2024 + NodeNext, Pi 0.80.6 type roots, every Shepherd `.ts` file | PENDING final code freeze |
| Offline Pi registration | pinned Pi 0.80.6 RPC `get_commands`, then command/help/status smoke without model/auth/network | PENDING final code freeze |
| README schema sample | extract JSON and call `validateProductionParentPlan(sample, 471)` | PASS — schema 2, three children |
| Documentation hygiene | stale-MVP grep, JSON parse, bounded secret-pattern scan, `git diff --check` | PENDING final docs freeze |
| Changed-path ownership | only `.pi/README.md` and `.planning/phases/479-shepherd-production-matrix/**` for this docs commit | PENDING final docs freeze |
| Live GitHub | explicitly designated sandbox plus healthy ambient `gh` authentication | NOT RUN — no live mutation is required for this docs lane |

The Go/connector gates in the repository root are intentionally not run for this TypeScript/Pi
sub-issue. They remain parent-head gates and do not substitute for Shepherd-specific verification.

## 17-row production matrix

The table records the current correction candidate. Rows marked PASS must still survive the final
machine gates above; rows marked OPEN are release blockers, not deferred hardening.

| # | Verdict | Evidence or precise open mechanism |
|---:|:---:|---|
| 1 | **OPEN** | Domain composition is covered by `production controller drives parallel disjoint children...`, `composes every child stage...`, and parent lifecycle tests. `production-runtime.test.ts` covers a fail-closed injected factory. However, the extension's `index.ts` still constructs the legacy `AutonomousShepherdController`; it does not construct `ProductionShepherdController` with concrete durable recovery, parent-ready authority/readiness, implementation/review AgentSessions, and genuine Git authority. `/pm-shepherd start/resume` therefore does not yet execute the production intake→PR→integration→parent-wait graph. |
| 2 | **PASS** | `selects the maximum deterministic disjoint set within capacity`, `explains capacity, write-scope collision, dependencies, and completion...`, plus controller idle-state persistence cover deterministic dependency/collision/capacity scheduling. |
| 3 | **PASS** | `cycle 8 admits bounded disjoint mutator leases...`, `join waits...releases exactly that session's lease once`, `abort joins only sessions for its run...`, and workspace lifecycle tests cover coexistence, collision denial, and own-lease release. |
| 4 | **PASS** | `correction and stale-parent refresh both force verification and a fresh exact-head review...` plus `real workspace refresh reclaims...and rebases...` cover parent advance, typed refresh/reclaim, and downstream invalidation. Planned-parent branch binding was corrected in `b31424b1`. |
| 5 | **PASS** | State tests persist bounded attempt/correction counters and exact intervention waits. New controller/pipeline cases `failed pre-publication verification runs one bounded correction...`, `corrects an exact failed verification before any pull request...`, and `requests an issue-bound retry decision...` cover pre-PR correction/exhaustion; review correction retains exact findings/dispositions. No score or prose path marks success. |
| 6 | **OPEN** | Journal/barrier tests cover phase monotonicity and ordered replay; pipeline publication tests cover several timeout-after-effect cases; runtime composition tests project one already-checkpointed observed effect without replay. The durable `ProductionEffectRecord` still lacks sufficient effect coordinates/results to reconstruct every fresh-process prepare/publish/observe/consume/apply/persist case, production recovery remains an injected authority with no extension construction, and parent human request/consume/merge observation is not fully journaled. The required all-effect, fresh-process, exactly-once trajectory is absent. |
| 7 | **PASS** | Broker tests accept only one exact unedited allowlisted command and reject bot, edited, disallowed, stale target/head/generation, duplicate, ambiguous, consumed, hostile, emoji, prose, CI, and silence inputs. Production human-gate tests bind the exact request. |
| 8 | **OPEN** | Controller tests cover stop at the six child pipeline stages and lower-level AgentSession/workspace/extension tests cover abort/join and unresolved initialization. `ProductionShepherdController.stop()` aborts/joins the child pipeline only; intake, correction substeps, parent finalization, parent request/poll, and broker backoff are not all run-owned and joined before a stopped checkpoint/lease release. |
| 9 | **PASS** | State CAS rejects stale generations; controller aborts a running sibling on failure; AgentSession and extension tests cover stop/shutdown/late-creation races. Persistence failure cannot be converted into sibling success. |
| 10 | **PASS** | `typed commit and push reconcile authoritative exact-head state after post-publication timeouts`, `reconciles a timed-out PR publication with one physical PR...`, and GitHub orchestrator integration timeout tests cover authoritative reconciliation before retry. |
| 11 | **PASS** | Production state/workspace tests preserve stable ownership and require an exact durable refresh receipt for base/claim replacement. Resume rejects plan and policy drift before recovery/mutation. The correction separates incrementing execution `generation` from immutable `resourceGeneration`, so retained exact PR/review/integration identities remain usable without admitting late execution results. |
| 12 | **OPEN** | Review adapter tests persist findings/dispositions and rerun on head movement. In the production controller, integration `head_moved` is still classified as `stale_parent`; the controller calls parent refresh, which rejects when only the child head moved. There is no distinct child-head reconciliation transition that invalidates and reruns verification/review on an unchanged parent. |
| 13 | **PASS** | Workspace handoff, child pipeline, GitHub evidence, and parent finalizer tests reject dirty/scope-escaped state, wrong branch/base/head, draft child PRs, incomplete/untrusted CI, and non-clean/prose-only review evidence. |
| 14 | **OPEN** | Tests prove approval alone is not completion, an exact authoritative merge is required, and no controller/extension parent-to-default merge action exists. A parent head move while waiting still throws without durably invalidating the gate and returning through finalization/review/new request. Child integration also identifies default branches only by `main|master` rather than authoritative repository default-branch evidence. |
| 15 | **PASS** | Contract/intake/tool-policy/Git/GitHub/verification tests cover proxy/accessor/cycle/oversize records, traversal, symlinks, controls, hostile text, fixed argv/cwd/environment, output/time bounds, and cancellation. Production runtime configures exactly `node → process.execPath`. |
| 16 | **PASS** | Argument and extension tests cover bare/help, invalid shapes, status with no AgentSession dispatch, one process-wide active run, and stop/shutdown during unresolved extension initialization. Offline RPC remains a final machine gate. |
| 17 | **PASS** | Production contract/intake/orchestration validation rejects top-level read-only children. AgentSession/tool-policy tests prove internal xhigh review roles are read-only and cannot enter integration or fabricate receipts. |

Current matrix total: **12 PASS, 5 OPEN** (rows 1, 6, 8, 12, and 14). Because row 1 is the
executable production route and the remaining OPEN rows are required contracts, issue #479 is not
ready for parent integration on this candidate even if all regression commands are green.

## Security and authority checks

- Plan task text and verification argv are secret-bearing input surfaces; operator docs prohibit
  secrets there. Production state validation rejects sensitive summaries and persists no task,
  prompt, reasoning, raw model output, credential, or unrestricted process output.
- GitHub credentials remain ambient host authentication. No docs or tests request, print, or store
  their values.
- Verification is a configured executable/argv runner with `shell: false`; the production factory
  exposes only the `node` mapping.
- Parent readiness is limited to a typed existing-draft→ready transition. No parent merge method is
  exposed. Child integration must still gain authoritative default-branch fencing (row 14).

## Human/external gates

- A fresh stable-head independent review of the final correction candidate is not recorded here.
- Live GitHub publication/decision/merge observation requires a designated sandbox and healthy
  ambient authentication; it was not authorized for this documentation lane.
- Parent/default-branch integration remains human-owned regardless of later code completion.
