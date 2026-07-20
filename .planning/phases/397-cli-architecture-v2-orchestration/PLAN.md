# Issue #397 Parent Orchestration Continuation Plan

Status: active
Parent branch: `feat/cli-architecture-v2`
Parent PR: #438 (`main` <- `feat/cli-architecture-v2`)
Starting HEAD: `56a7ecb08f755184af7b55318c3285582d5adfb7`
Orchestrator session: `4e6d3ae8-19fc-4c5a-8135-a3e6e4fa1cfc`
Planning/review model: `openai-codex/gpt-5.6-sol`, thinking `xhigh`
Implementation/correction model: `openai-codex/gpt-5.6-sol`, thinking `high`

## Workflow

- GSD: `scripts/gsd doctor`; `scripts/gsd list`; attempted `scripts/gsd prompt programming-loop init --phase 397-cli-architecture-v2 --dry-run`.
- Adapter result: `programming-loop` is absent from the 69-command registry. Use the documented manual fallback in `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md` and record it per unit.
- Contracts: parent orchestrator, issue agent, stacked workflow, automated review routing, and exact-head review.
- Skills: `gsd-core`, `caveman`, `golang-how-to`, plus issue-specific CLI/testing/error/security/safety/lint/documentation/Cobra/Viper/observability/performance/concurrency/context/database/design skills.
- CLI-visible work follows `.agents/agentic-delivery/references/cli-help-docs-website-parity.md`.

## Accepted baseline

Preserve parent commits through #410 at starting HEAD. Do not reimplement #398-#406, #410, #421-#423, or safety issue #453. Existing PRs #460 (#424) and #461 (#415) must be reconciled from their exact remote heads and CI state.

## Dependency-ordered queue

1. Correct and independently ratify PR #460 / issue #424; promote only its reviewed exact head.
2. Correct and independently ratify PR #461 / issue #415; promote only its reviewed exact head.
3. Complete serialized namespace chain #425 -> #426 -> #427 -> #428 -> #429 -> #430 -> #431 -> #432 -> #433 -> #434 -> #435 -> #436 -> #437; then ratify umbrella #407.
4. Complete #408, then #409.
5. Complete #413 and #414 after #407/#408 prerequisites.
6. After #409, complete #411, #412, #416, and #469 in isolated worktrees, serialized wherever
   central CLI/help/golden write scopes collide. #462/PR #468 is integrated as a design gate, but
   external review coverage remains absent and must be recorded as human/enabled-automation fallback. #416 owns
   the human-first bare reverse workspace plus its `reverse guide` alias; #469 owns credential/
   connection setup. #411 owns the human-first bare query workspace plus its `query grid` alias.
7. Complete #417 after #411/#412/#413/#414/#416/#469.
8. Complete #418 after #411/#412/#414/#416/#469 and #463 when the chart slice is included.
9. #419 is explicitly human-deferred from this parent campaign. Do not implement the optional OpenTelemetry beta log bridge and do not add its beta dependency.
10. Complete #420 after #415/#417/#418; the required #419 decision record is satisfied by the explicit human defer.

## Per-unit lifecycle

1. Create an isolated worktree and issue branch from the exact current parent HEAD.
2. Record worker model, thinking, session, starting HEAD, expected write scope, plan, TDD ledger, verification checklist, and run state before production edits.
3. Add a focused failing test before behavior changes.
4. Implement with Sol/high; run focused green tests, issue gates, safety checks, and parity checks.
5. Commit the coherent green unit. Do not open another child PR unless collision isolation requires it.
6. Run independent Sol/xhigh exact-head correctness/security/architecture/coverage/evidence review.
7. Treat findings as Sol/high correction units, then repeat exact-head review.
8. Promote only reviewed commits to the parent branch, verify continuity, update the parent ledger, commit, push, and continue.

## Recovery budgets

- Implementation/test/lint failure: 3 bounded root-cause/fix cycles per unit.
- Integration conflict: 2 rebase/cherry-pick conflict-resolution cycles; preserve both accepted sides.
- Review findings: 3 correction/re-review cycles per unit.
- CI failure: 2 in-scope correction cycles; infrastructure failures are recorded and retried once deliberately.
- External review unavailability: rely on the required local independent Sol/xhigh exact-head review and record GitHub review-route status; do not spam bots.

## Final campaign

At one exact parent HEAD run:

```bash
gofmt -w cmd internal
git diff --exit-code -- cmd internal
git diff --check
go vet ./...
go test ./...
go test -race ./...
go build ./cmd/pm
make verify
```

Also run module-boundary, dependency/ADR delta, generated docs/help/manual, website parity, security/secret-pattern, CLI help/bare-namespace/invalid-action, integration applicability, and repository hygiene checks. Runtime-backed/credentialed checks remain not run unless explicitly requested.

Then run independent Sol/xhigh correctness, security, architecture, issue-coverage, and evidence reviews against that exact HEAD. Any actionable finding becomes a Sol/high correction unit; repeat affected gates and exact-head review until clean.

## 2026-07-20 Pi active continuation

- GSD evidence: `scripts/gsd doctor`, `scripts/gsd list`, and `scripts/gsd prompt plan-phase 397 --skip-research`; `programming-loop` remains absent from the 69-command registry, so the recorded manual universal-loop fallback continues.
- Parent branch/PR reconciled at live head `c3d8a7573bfaf661bdcab737db84e3497929cdff`, then locally merged `origin/main` at safe checkpoint `19fe02ec900aba548a997165014624197b451a33`; PR #438 remains draft and targets `main`.
- #462 / PR #465 was provisionally integrated at `a5474bcb`. Accepted design/safety corrections were integrated through PR #467 at `93a11710`; PR #468 was merged into the parent at `c3d8a7573bfaf661bdcab737db84e3497929cdff`. Local sidecars are still not external review coverage.
- #437 / PR #466 is open at head `26f98a72419010b961b5b8378ef4a695b0c0a06f`. Full local gates and all current GitHub checks pass. The CI timing flake was replaced by a deterministic concurrency proof without weakening the parallelism gate. Human/parent fallback review remains pending.
- GitHub blocked-by metadata now directly encodes #462 for #408, #409, #411, #412, #414, #416, #418, and #463.
- Actual Pi worker runtime used project `pm-gsd-worker` with `openai-codex/gpt-5.5:xhigh`; the project agent currently declares `thinking: xhigh`, while the requested routing policy says implementation `high`. The subagent API has no per-call model override; this mismatch is recorded rather than hidden.
- Reviewer discovery attempt was blocked because the parent runtime did not expose `grep/find/ls`; bounded exact-file `read` reviews succeeded. This does not satisfy Claude/Copilot/human coverage.

### Ready queue after live reconciliation

| Issue | State | Dependencies | Decision |
|---|---|---|---|
| #437 / PR #466 | `sub_pr_green`; checks green; no reviews | #436 provisionally integrated | `not_spawned_review_blocked` until human or enabled automation review clears |
| #408 | source-ready, launch deferred | #405 and integrated #462 | `not_spawned_write_scope_collision` with open PR #466 central CLI/help/golden/docs/website files |
| #407 | dependency blocked | #437 integration | `not_spawned_dependency_blocked` |
| #413 | dependency blocked | #407 | `not_spawned_dependency_blocked` |
| #409 | dependency blocked | #408 | `not_spawned_dependency_blocked` |
| #416 | dependency blocked | #409; #462 integrated | `not_spawned_dependency_blocked`; reverse-only ownership remains |
| #469 | dependency blocked | #409; #462 integrated | `not_spawned_dependency_blocked`; setup remains separate from #416 |
| #419 | human-gated | #404/#410 complete | `not_spawned_human_gate` for optional beta dependency decision |

### Phase 437 pending intake preservation

- Untracked Phase 437 pending-request/research/debug files from the #466 worktree were copied to
  `.planning/traces/phase-437-pending-intake/` at parent checkpoint `c3d8a757` without changing
  PR #466's tested head `26f98a72`.
- These files are triage only. Do not implement them, regenerate docs, commit to PR #466, or edit
  their referenced GitHub issues until the human coordinator explicitly authorizes implementation
  or planning synchronization.
- Triage owners: #437 for native connectors/certify follow-ups, #411 for connector list/browser/
  progressive inspect, #412 as full-manual pager consumer, and #417 for hierarchical help/manuals.

### Automated review availability

- `.github/workflows/claude-review.yml` exists but GitHub reports workflow state
  `disabled_manually` (workflow id `310534134`); no current Claude review covers #438, #466, or
  the #467/#468 ranges.
- PR #466 and PR #438 have no GitHub review records and no requested reviewers.
- Copilot backup was probed non-mutatingly and is unavailable in this session (`@copilot` collaborator
  probe returned HTTP 404); no Copilot review was requested.

## Human gates

- Never merge PR #438 to `main`.
- No new dependencies beyond an explicit accepted ADR/approval record; #419 is explicitly deferred and grants no dependency approval.
- No secrets, auth-scope changes, credentialed connector checks, production operations, destructive actions, quality-gate reductions, or generic write tools.
- Reverse ETL execution remains plan -> preview -> approval -> execute and is not part of ordinary verification.

## 2026-07-20 #437 integration and post-merge queue

- Human fallback review coverage recorded for PR #466 exact head `26f98a72419010b961b5b8378ef4a695b0c0a06f`: https://github.com/polymetrics-ai/cli/pull/466#issuecomment-5026616557.
- Pre-integration gate passed: head unchanged, required checks green, active review threads empty.
- PR #466 merged only into `feat/cli-architecture-v2` at parent merge commit `1008f75ff8fe7d43a0a67a802ccf05ef296eae7f`; parent PR #438 remains draft and unmerged to `main`.
- #437 is provisionally integrated. #407 umbrella dependency state is complete on the parent branch because #421-#437 are now integrated.
- Rebuilt queue: #408 is the critical-path ready implementation issue; #413 is ready but deferred by write-scope collision with #408; #419 remains human-gated.
- Phase 437 pending intake remains planning-only under `.planning/traces/phase-437-pending-intake/`; no pending request is authorized for implementation.

- #408 dispatch: `pm-gsd-worker` in `/Users/karthiksivadas/Development/polymetrics-cli-agents/wt-408-flow-etl-dashboards`, branch `feat/408-flow-etl-dashboards`, from parent head `5b6037880eed78bc8bc276a3ced13302908cac53`; #413 deferred for write-scope collision.

## 2026-07-21 operator decision for #419

- Explicit human decision: SKIP/DEFER the optional OpenTelemetry beta log bridge from parent campaign #397.
- Do not add the beta dependency and do not dispatch an implementation worker for #419.
- This decision satisfies the parent campaign's required #419 inclusion-or-skip record. Issue #419 may remain open for future separately authorized work.
- This is not approval for any other dependency. All other dependency additions remain human-gated.

## 2026-07-21 #408 EXECUTE resume

- Live process ownership check found no worker using the preserved #408 cwd, so exactly one `pm-gsd-worker` was resumed there with Sol/high routing.
- Adopted branch `feat/408-flow-etl-dashboards` at plan head `361a6bec` plus 19 dirty entries without reset, clean, stash, rebase, overwrite, or worktree recreation.
- Worker pushed implementation/artifact commits `eb3c84cb` and `ff7be3bd`; worker tree and remote head now match and no sub-PR exists.
- Focused RED/GREEN/refactor, focused race, full non-race, vet, build, CLI parity, and `make verify` are recorded green. Full `go test -race ./...` timed out at 10m and the `internal/cli` retry timed out at 20m without a race finding.
- Correction remains before `SUB_PR_OPEN`: synchronize stale/contradictory phase evidence, preserve the race timeout as an unresolved verification gate, and record that `make verify` invoked its local temporary reverse smoke in the required plan → preview → approval → execute order despite the narrower worker boundary.
- #413 remains deferred for write-scope collision. No other mutating worker was dispatched.

## 2026-07-21 Pi model-routing correction

- User directive: route all active Pi/GSD roles through `openai-codex/gpt-5.6-sol`.
- Implementation roles (`pm-gsd-worker`, `pm-issue-worker`, and `pm-docs-writer`) use `thinking: high`.
- Orchestration, planning, research, issue creation, verification, review, review disposition, and
  Shepherd validation use `thinking: xhigh`.
- Raise the project parallel worker ceiling from three to the Pi runtime's safe cap of four;
  dependency order, disjoint write scopes, one issue per worker, and isolated worktrees remain hard
  prerequisites.
- Add a deterministic routing regression check before editing active model configuration. Update
  active Pi agents, prompts, Shepherd defaults, GSD config, and canonical runtime documentation;
  retain earlier `gpt-5.5` entries in phase/run artifacts as historical evidence.
- GSD adapter preflight: `scripts/gsd doctor` passes; `scripts/gsd prompt programming-loop init
  --phase 397 --dry-run` still reports `unknown GSD command: programming-loop`, so the documented
  manual universal-loop fallback applies for this bounded orchestration configuration correction.
- This correction changes future dispatch only. An already-running subagent cannot switch model
  mid-invocation; resume/re-dispatch only after its durable checkpoint.
- Shepherd hardening discovered during routing review: delete the previous validator verdict before
  each validation turn, discard any verdict from a validator process that exits nonzero, and honor
  `RUN.json.terminal` only after a fresh `PROCEED` from a successful validator process. Add a
  deterministic main-loop regression so validator failure or `RETRY` cannot create false
  human-readiness.
