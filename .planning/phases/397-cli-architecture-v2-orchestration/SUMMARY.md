# Issue #397 Parent Orchestration Summary

## Progressive setup refinement — 2026-07-20

PR #468 now carries the GSD-verified Phase 18 interaction contract. #416 owns the human-first bare
reverse workspace and `reverse guide` alias; #411 owns the equivalent bare query workspace and
`query grid` alias. Child #469 owns TTY-progressive credential and connection setup. The live
GitHub graph records #409/#462 → #469 → #417/#418, and the missing #411 → #463 edge is restored.
PR #468 is now integrated into the parent branch at `c3d8a757`; production TUI work remains
sequenced behind the rebuilt dependency queue and the open #437/#408 write-scope collision.

Status: ACTIVE — not final
Starting HEAD: `56a7ecb08f755184af7b55318c3285582d5adfb7`
Parent PR: #438 (draft)

The continuation run reconciled the repository, issue hierarchy, parent/child PRs, worktrees, ROADMAP, prior run states, and remote CI before integration. Accepted implementation through #410 and namespace grandchildren #421-#423 remains preserved.

PR #460 / #424 was corrected at `323d4a91`, independently re-reviewed clean, and promoted with ancestry preserved. PR #461 / #415 was corrected at `6cf5c48f`, independently re-reviewed clean, integrated after regenerating the sole website-data conflict, and independently reviewed clean in combination. Parent integration head is `1f5bd80f77ab267901be730f855728cf00120874`.

#425 nativized the version namespace, corrected assigned global JSON boolean handling, passed exact-head independent review, and was promoted at parent merge `0c57ec39`.

#426 nativized the skills namespace, passed exact-head independent review, and was promoted at parent merge `bb12f265`.

#427 nativized the docs namespace while preserving legacy trailing-help and literal-`--` behavior, passed exact-head independent review, and was promoted at parent merge `e68ccdf7`.

#428 nativized the agent namespace with fake-only container runtime tests, closed action-discovery and clustered-help effect bypasses, passed final exact-head security review, and was promoted at parent merge `569536d1`.

#429 nativized credentials, preserved env/stdin-only intake and legacy identifiers, hardened parser ownership, and added rooted local-write effect confinement. After iterative exact-head corrections, final independent review was clean and parent merge `a490eeba` passed bounded integration suites.

#430 nativized ETL with bounded batch/sync/output semantics preserved and fail-closed status operand ownership, passed exact-head independent review and parent integration race checks.

#431 nativized reverse ETL while preserving ordered plan → preview → approval → execute, token nondisclosure, and fail-closed parser compatibility. Exact-head review and parent integration race checks passed.

#432 nativized flow while preserving manifest/directory, checkpoint, ledger, event, telemetry, cancellation, and output contracts. Exact-head review and parent integration race checks passed.

#433 nativized schedule with fake-only backend/runtime seams and preserved no-effect behavior for unsupported actions. Exact-head review and parent integration race checks passed.

#434 nativized RLM with fake-only analyzer seams, mode-gated request forwarding, and no runtime/model effects. Exact-head review and parent integration race checks passed.

#435 nativized the hidden typed worker namespace with invocation-local fake status/serve seams and accurate no-dial evidence. Exact-head review and parent integration race checks passed.

#436 nativized hidden extract, added project-rooted RLM warehouse scope, and preserved parser/help/output behavior with generated docs parity. Exact-head review and integration race checks passed.

#462 / PR #465 added the terminal interaction design system. Five accepted design/safety corrections were integrated through PR #467 at parent commit `93a11710`. A follow-up local review tightened activation to stdin+stdout TTY, made `--plain`/`--json`/`--no-input` unconditional prompt bypasses, and expanded fallback RED coverage. PR #468 was merged into the parent at `c3d8a7573bfaf661bdcab737db84e3497929cdff`; its exact-head checks were green before merge. GitHub blocked-by edges now directly link #462 to all affected TUI issues. Local sidecars were not external coverage at that checkpoint; Wave 1 later verified exact-head human approval at https://github.com/polymetrics-ai/cli/pull/468#issuecomment-5054325561.

#437 / PR #466 contains the final serialized Phase 9 connectors/certify migration. After safety/correctness correction cycles, remote CI exposed a wall-clock concurrency-test flake. The worker replaced it with a deterministic barrier/counter proof without raising or removing the gate. Head `26f98a72` passes focused/repeated/race tests, full CLI/certify packages, `go test ./...`, vet, build, `make verify`, help/docs/website parity, fixture-only sample certification, connectorgen 547/0, and all current GitHub checks. Exact-path local re-review found no remaining actionable runtime/code issue.

#437 is not integrated because human/parent fallback review coverage remains pending. #408 is source-ready after #462 integration but is not spawned because it would collide with open PR #466 central CLI/help/golden/docs/website files. #407/#413 remain dependency-blocked on #437; #409/#416/#469 remain downstream; #419 remains an explicit human dependency-decision gate. Final parent verification and final parent review have not run; `verificationPassed` remains false.

## Live reconciliation — 2026-07-20T19:28Z

The parent branch was reconciled against live GitHub: PR #467 and #468 are merged, live parent head was `c3d8a757`, PR #466 remains open at `26f98a72` with green checks and no reviews, Claude review is disabled manually, and Copilot backup is unavailable in this session. Read-only `pm-scout`/`pm-reviewer` sidecars validated stale artifact updates, Phase 437 pending-intake triage, and the #408/#437 write-scope collision.

Untracked Phase 437 pending-request/research/debug files were preserved under `.planning/traces/phase-437-pending-intake/` without changing PR #466's tested head. They are not implementation authorization and their referenced GitHub issues were not edited.

The parent branch was safely synced with `origin/main` using ordinary no-ff merge commit `19fe02ec900aba548a997165014624197b451a33`; no force push was used and parent PR #438 was not merged to `main`.


## #437 integration — 2026-07-20T20:06Z

The human coordinator explicitly reviewed and tested PR #466 at exact head `26f98a72419010b961b5b8378ef4a695b0c0a06f` and approved integration into `feat/cli-architecture-v2`. The orchestrator verified the head was unchanged, checks were green, and active review threads were empty, then recorded human fallback review coverage at https://github.com/polymetrics-ai/cli/pull/466#issuecomment-5026616557.

PR #466 was merged only into the parent branch at `1008f75ff8fe7d43a0a67a802ccf05ef296eae7f`. Parent PR #438 remains draft and unmerged to `main`. #437 is provisionally integrated and the #407 umbrella dependency is complete on the parent branch. The ready queue now selects #408 as the critical-path ready implementation issue; #413 is ready but deferred for write-scope collision with #408, downstream UI/help issues remain dependency-blocked, and #419 remains a human dependency gate. Phase 437 pending intake remains planning-only.

## #419 operator decision — 2026-07-21

Human explicitly chose SKIP/DEFER for the optional OpenTelemetry beta log bridge from parent campaign #397. No #419 implementation worker will run and no beta dependency is authorized. This satisfies the parent inclusion-or-skip acceptance criterion while granting no approval for any other dependency; issue #419 may remain open for future separately authorized work.

## #408 EXECUTE resume — 2026-07-21

No live process owned the preserved #408 worktree, so exactly one Sol/high `pm-gsd-worker` resumed its existing committed plan and 19 dirty entries. The worker preserved all work, delivered the flow/ETL dashboard slice, and pushed clean synchronized heads `eb3c84cb` and `ff7be3bd`. Focused RED/GREEN/refactor, focused race, full non-race, vet, build, parity, local dual-TTY fixture, and `make verify` evidence is recorded with no dependency delta.

The 10-minute full race and 20-minute CLI race retry timed out without race findings. The worker also left stale/contradictory phase evidence and reported that mandatory `make verify` ran its local temporary reverse smoke in plan → preview → approval → execute order despite the narrower no-execution dispatch boundary. These are correction/verification blockers; no PR was opened and #413 remains collision-deferred.

## Wave 1 parent synchronization — 2026-07-23

Current main `873cd7b251f70c4a35a607a0d4e86051ea0fbd15` has been merged normally into isolated branch `fm/cli-architecture-v2-wave1-parent-sync-r1` from parent/PR head `21d195aff0c7bd60b3bf54f14b1ce165cec9e03f`. Merge commit `c545c3740c71b889fd2f1f64cec5491003f7b654` preserves Gong parity and CLI Architecture v2 contracts; full credential-free verification is green at pre-evidence task head `2a2e964b17144939b0a42f297de0d2b1c87383e1`.

At this historical checkpoint PR #438 remained draft at the old parent head until the Wave 1 stacked PR was reviewed and human-integrated. The parent was subsequently fetched and verified at `0f8c964ba9cfbe1b1eec8e7998eacf4158ef0e20`; PR #438 remains draft/human-only. #462 has exact-head human approval at https://github.com/polymetrics-ai/cli/pull/468#issuecomment-5054325561. #419 is `deferred_by_human`; historical #425–#436 waiver/review work remains pending. #408 remains excluded at branch head `6c643f5c971d1fac4a83e4ffe653b83847c2fceb` until PR #493's canonical PM migration integrates.

### PR #493 canonical PM migration gate

Post-integration reconciliation recorded the actual parent/PR #438 head as `0f8c964ba9cfbe1b1eec8e7998eacf4158ef0e20`. PR #493 must merge that resulting parent normally, reconcile only its owned routing/skill/Makefile guidance to the canonical PM route, rerun its gates, and integrate before another CLI Architecture v2 implementation worker starts. Until PR #493 integrates, #408 remains `not_spawned_dependency_blocked`.

## Pi 5.6 Sol routing and Shepherd hardening — 2026-07-21

Active Pi and GSD routing now uses `openai-codex/gpt-5.6-sol` exclusively. Mutating implementation
roles (`pm-gsd-worker`, `pm-issue-worker`, `pm-docs-writer`) run with `high`; orchestration,
planning, research, issue creation, verification, review, review disposition, and Shepherd
validation run with `xhigh`. Project concurrency now matches Pi's safe four-process cap while
retaining dependency, write-scope, issue, branch, and worktree isolation rules.

The Shepherd driver now deletes the prior validator verdict before every validation turn, discards
any verdict from a validator process that exits nonzero, accepts a terminal run state only after a
fresh successful `PROCEED`, passes coordinator thinking explicitly, and uses a 90-minute stall
window suitable for the long CLI/certification gates. Deterministic model-routing, stale-verdict,
failed-validator, terminal-authority, and live-child stall regressions are wired into `make verify`.
The full repository verification gate passed. Earlier 5.5 runtime entries remain unchanged as
historical evidence.
