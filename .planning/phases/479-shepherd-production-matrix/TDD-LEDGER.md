# TDD ledger

## Evidence policy

The phase required behavior-level RED → GREEN → REFACTOR evidence for all 17 preflight rows. The
history does not prove that process completely. Commit `9848410e` added
`production-controller.test.ts` while `production-controller.ts` did not exist, so it established a
missing-module/compile RED only. The program's own stable-head convergence contract says that a
missing-module or compile-only failure is not behavior-level RED. No saved failing output proves
that all 17 assertions first executed against a compiling scaffold and failed for their intended
reason.

Later focused and full-suite GREEN runs are valid regression evidence for the final implementation,
but they cannot retroactively manufacture the missing behavior-level RED checkpoint. This process
gap remains open in the phase record even if every final functional gate passes.

## Slice ledger

| Slice | RED evidence | GREEN/regression evidence | Refactor/correction evidence | Process status |
|---|---|---|---|---|
| Plan and production contract | No separate behavior RED; `2e488a43` added contract and tests together. | Contract tests validate exact schema-2 input, bounded values, mutating-only top-level work, and hostile-shape rejection. | Subsequent validator and state corrections reuse the frozen DTOs. | GREEN regression; RED not evidenced. |
| Intake, scheduler, and bounded verification | No separate behavior RED; commits `7bfe115b`, `78034a8d`, and `a2cf90be` co-located tests and production. | Intake cancellation/non-symlink tests, deterministic dependency/collision/capacity tests, and fixed executable/argv/cwd/output/timeout tests. | Missing-intake error corrected in `bec16415`. | GREEN regression; RED not evidenced. |
| Durable state, effects, and recovery barrier | No separate behavior RED; commits `35787518`, `616a3429`, and `6fea793d` co-located tests and production. | CAS/binding/budget/generation tests plus prepared → observed → applied journal and ordered recovery tests. | Retry, lock, ownership, checkpoint, and exact merge evidence corrections in `236260dc` through `cf955711`. | GREEN regression; RED not evidenced. |
| Controller trajectory | `9848410e` is a test-only missing-module/compile RED, explicitly insufficient as behavior RED. | `81d3062d` supplied the controller; controller tests exercise dependency-ready parallel work, correction, stale-parent refresh, budget waits, stop/join, sibling abort, binding rejection, durable effect acknowledgement, and observed merge. | `d84e4740` added race coverage; later checkpoint and merge-evidence corrections refined state transitions. | Functional GREEN evidence; behavior RED gap remains. |
| Workspace/Git lifecycle | No separate behavior RED for the initial slice; production and tests arrived together. | Lifecycle tests cover stable ownership, isolated claims, verification-before-commit, refresh/rebase/reverify, per-run join/release, and commit/push reconciliation. | Lease, orphan-lock, resume ownership, refreshed-base, and planned-parent binding fixes include `86238846`, `16994d1c`, `2fba70be`, `d2f71aa0`, and `b31424b1`. | GREEN regression; RED not evidenced. |
| GitHub, review, and human decisions | No separate behavior RED; `3628cc90` and later parent-lifecycle commits co-located tests and production. | Deterministic transport, exact-head review/dispositions, fail-closed GitHub evidence, exact human commands, draft→ready, and observed human merge tests. | Exact parent lifecycle hardened in `d138fb99`; final-head fixes are recorded by the final verification head. | GREEN regression; RED not evidenced. |
| Production child pipeline and runtime composition | No separate behavior RED for the original slice; `7fecbe86` added pipeline and tests together. | Final combined production suite passes 808/808 and reaches the genuine Pi host composition. | Planned-parent binding fixed in `b31424b1`; exact effect recovery, parent lifecycle, and integration-CAS corrections end at code head `91692415`. | Functional GREEN; original behavior RED gap remains. |
| Documentation | Docs-only; behavior RED is not applicable. | Schema-2 README sample is parsed and validated; command RPC, documentation, JSON, and diff gates are recorded in `VERIFICATION.md`. | Stale schema-1/local-MVP claims removed; exact CAS/recovery behavior documented. | PASS. |

## Final correction RED → GREEN evidence

| Review blocker | Behavior RED | GREEN correction | Final evidence |
|---|---|---|---|
| Generation-2 child intervention recovery | `0fe22e9e`: fresh-process generation-2 request failed because recovery reconstructed generation 1. | `e2dedad7`: recovery binds the child decision to current execution generation while resource identities retain resource generation. | Runtime 10/10; recovery 83/83; controller 34/34. |
| Start/resume default-branch authority | `06e50e21`: resume accepted remote-default drift and a `trunk` → `main` parent target. | `a8104613`: start/resume re-inspect exact repository/remote authority and reject default aliases. | Host/extension 25/25; real offline Pi RPC passed. |
| Atomic parent-base integration | `5ef7ba15`: GitHub's head-only merge path accepted a parent race and did not provide an exact base CAS. | `37dbc42c`: deterministic exact-base/head merge plus `--force-with-lease` parent ref update; GitHub becomes receipt-only. | Exact CAS 3/3; orchestrator 515/515; pipeline 17/17. |
| Crash after Git CAS before receipt | `32a0d50e`: fresh-process retry stopped at `stale_parent` after one physical CAS and no receipt. | `91692415`: retry proves/reuses the exact deterministic merge, publishes one receipt, and still rejects unrelated parent movement. | Pipeline/orchestrator/evidence 666/666; combined final suite 808/808. |

## TDD disposition

- Functional result: **PASS — all 17 rows are implemented and verified** at production code head
  `91692415`; the later test/docs freeze does not change production code.
- Historical process result: **OPEN — behavior-level RED evidence is incomplete**.
- No future test-only commit should be described as the original RED for already implemented
  behavior. Any additional correction must add a genuinely failing behavior test first, capture its
  failure, then implement the smallest GREEN change and rerun the stable-head gates.

## 2026-07-22 correction ledger

| Slice | RED evidence | GREEN evidence | Status |
|---|---|---|---|
| Issue-driven plan bootstrap | Missing-module RED, then compiling throwing scaffolds: bootstrap 4/4 intended failures, GitHub facts 3/3 intended failures, and host start 2 intended failures/1 pass because no plan session ran. | Bootstrap/intake 7/7, GitHub facts 3/3, and host entrypoint/bootstrap focused cases all pass. Returned GitHub issue numbers, not planner-invented numbers, enter the canonical schema-2 plan. | GREEN |
| AgentSession trusted-local verification | Missing-module RED, then compiling scaffold 4/4 intended failures. A separate tool-policy behavior RED proved host-only mutation still exposed workspace edit/write authority. | ID-only ordered verification, omitted/out-of-order fail-closed behavior, bounded diagnostics, implementation/correction RED→GREEN reruns, real `go test ./...`, isolated environment, lifecycle binding, and recovery binding pass. | GREEN |
| Same-second ready reconciliation | Focused host/lifecycle RED: 24 pass and 3 intended equal-second failures. | Focused host/lifecycle GREEN: 27/27; non-draft equality succeeds, draft equality and moved head fail, restart/duplicate remain idempotent. | GREEN |
| One-pass independent correction review | Focused review RED: 17 pass and 4 intended failures proved arbitrary recipes were accepted, bootstrap could authorize `git push`, a descendant retained the verification process beyond timeout, and a cleanup exception after a failed test did not reject. | Focused GREEN: 21/21; only closed Node/Go/Make recipes pass, bootstrap rejects mutating Git, POSIX process groups are terminated within the hard bound, and AgentSession exceptions always propagate through protocol validation. | GREEN |
| Plan-to-runtime identifier and output parity | Focused RED: 6/8; plan/proposal validation accepted a space-bearing verification ID and a one-byte output limit that the AgentSession/host verifier rejected later. | Focused GREEN: 8/8; safe IDs and the 1 KiB output floor are enforced before bootstrap publication and represented in the proposal tool schema. | GREEN |
| Real GitHub orchestration parity | Focused RED: 8/11; real orchestration rejected intentionally empty child human gates, while plan intake accepted an empty required-skill list. | Focused GREEN: 11/11; empty exceptional gates map through real orchestration and every child declares at least one implementation skill. | GREEN |
| Unified plan topology/scope parity | Focused RED: 11/13; durable plan intake accepted inline-unsafe fields, duplicate verification IDs/slugs, and scheduler-invalid scopes. | Focused GREEN: 13/13; plan intake uses inline-safe GitHub fields and the shared dependency-graph validator, and rejects duplicate command IDs/slugs before journaling. | GREEN |
| Real planning tool-policy compilation | Focused RED: 3/4; the actual `host_inspect` proposal schema used unsupported `const`/`pattern` keywords and could not be compiled by the closed Pi tool policy. | Focused GREEN: 4/4; the complete real capability compiles using supported enum/length constraints and semantic proposal validation still enforces exact values. | GREEN |

The correction retains one broad historical caveat: missing-module RED alone is not sufficient. For
each new module, the test was rerun against a compiling throwing scaffold and failed for the intended
behavior before the implementation replaced it. Command output was captured in the active session;
the durable ledger records exact counts because those RED states were intentionally not committed.

## 2026-07-23 merge-readiness closure ledger

| Slice | RED evidence | GREEN target | Status |
|---|---|---|---|
| Complete-suite CI gate | `rg` over `.github/workflows` returned no complete Shepherd inventory command and exited 1 with `RED: no workflow runs the complete Shepherd test inventory`. | A pull-request/main-push workflow runs the exact sequential inventory on ordinary GitHub-hosted infrastructure, with read-only contents permission and no secrets. | IMPLEMENTED at `307ea409`; YAML/command/permission/pin checks pass, remote CI GREEN pending. |
| Diff hygiene | `git diff --check 69a1a988..HEAD` exited 2 on one added blank EOF line in each of six production-matrix artifacts. | The same range check and working-tree `git diff --check` pass. | GREEN at `307ea409`. |
| Phase summary | File existence check exited 1 because `479-shepherd-production-matrix/SUMMARY.md` was absent. | A concise summary records delivered scope, exact verification limits, remaining CI/review/integration gates, and the no-main-merge boundary. | GREEN at `307ea409`. |

The CI workflow is the only behavior-changing slice in this closure. The whitespace and summary
changes are documentation-only, so a runtime behavior RED is not applicable to them.

## Exact-head review correction

| Finding | RED evidence | GREEN target | Status |
|---|---|---|---|
| Complete Pi-family determinism | The `ca3f6c6f` review found that the top-level Pi package manifest uses caret ranges. File/step checks both exited 1 because no committed post-install assertion existed. | CI verifies the published `npm-shrinkwrap.json` and installed nested `pi-coding-agent`, `pi-ai`, `pi-agent-core`, and `pi-tui` are all exactly 0.80.6 before running tests. | GREEN at `a594be98`; real local installed family and exact tarball resolutions pass. |
| Published parent base | The review found local parent `69a1a988` 175 commits ahead of cached origin `2a89142e`; a child PR against that cached remote would expose the wrong range. | Reconcile local parent artifacts, publish parent first, fetch/verify authoritative remote head, then publish/open #479 and verify the PR range. | Local GREEN: parent reconciliation `45c27b9d` (now 176 commits ahead of cached origin) is merged into the child at `766709b3` without history rewrite. Remote GREEN remains blocked by DNS/auth. |

## 2026-07-23 remote CI repair ledger

| Slice | RED evidence | GREEN target | Status |
|---|---|---|---|
| Deterministic Go fixture environment | PR #489 complete-suite run `29959846371`, job `89057976671`: 1,712 total, 1,710 pass, one fail, one skip. The only failure is the real-Go fixture's cleanup hook attempting to unlink a read-only `go1.25.0` auto-download beneath its isolated `GOPATH`. | Install exact Go `1.25.12` before the suite so the fixture uses the repository toolchain and does not download a toolchain into the disposable root; retain strict cleanup. | RED captured; implementation pending. |
| Reachable x/text vulnerability | PR #489 security run `29959846280`, job `89057975920`: `govulncheck` reports `GO-2026-5970`, found in `golang.org/x/text v0.36.0`, fixed in `v0.39.0`, reachable through the PostgreSQL connector and `pgxpool.New`. | Merge current `origin/main`, which already carries `v0.39.0`, into the non-default parent; inherit that parent head in the child and rerun `govulncheck`. | RED captured; baseline synchronization pending. |

No new product behavior is introduced. The remote failing gates are the valid RED. The smallest
GREEN must correct the CI toolchain and stacked-branch baseline; suppressing cleanup, weakening
`govulncheck`, or adding a duplicate child-only dependency override is out of scope.
