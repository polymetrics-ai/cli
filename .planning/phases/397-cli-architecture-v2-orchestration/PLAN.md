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
6. After #409 and reviewed integration of #462/PR #468, complete #411, #412, #416, and #469 in
   isolated worktrees, serialized wherever central CLI/help/golden write scopes collide. #416 owns
   reverse guidance; #469 owns credential/connection setup.
7. Complete #417 after #411/#412/#413/#414/#416/#469.
8. Complete #418 after #411/#412/#414/#416/#469 and #463 when the chart slice is included.
9. #419 remains human-gated because its issue requires explicit inclusion of an optional beta dependency; if no approval exists, record an explicit skip rather than implementing it.
10. Complete #420 after #415/#417/#418 and the #419 decision record.

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
- Parent branch/PR confirmed at `93a117100c6421955262aa32794a91a158d267e1`; PR #438 remains draft and targets `main`.
- #462 / PR #465 was provisionally integrated at `a5474bcb`. Accepted design/safety corrections were integrated through PR #467 at parent commit `93a11710`. A later local review found one stdin/stdout TTY-gate contradiction; follow-up PR #468 is green at head `5092e115d4aa35ab4595a9b9537f64d3f63e6406`, with human review pending. Local sidecars are not external review coverage.
- #437 / PR #466 is open at head `26f98a72419010b961b5b8378ef4a695b0c0a06f`. Full local gates and all current GitHub checks pass. The CI timing flake was replaced by a deterministic concurrency proof without weakening the parallelism gate. Human/parent fallback review remains pending.
- GitHub blocked-by metadata now directly encodes #462 for #408, #409, #411, #412, #414, #416, #418, and #463.
- Actual Pi worker runtime used project `pm-gsd-worker` with `openai-codex/gpt-5.5:xhigh`; the project agent currently declares `thinking: xhigh`, while the requested routing policy says implementation `high`. The subagent API has no per-call model override; this mismatch is recorded rather than hidden.
- Reviewer discovery attempt was blocked because the parent runtime did not expose `grep/find/ls`; bounded exact-file `read` reviews succeeded. This does not satisfy Claude/Copilot/human coverage.

### Ready queue

| Issue | State | Dependencies | Decision |
|---|---|---|---|
| #437 | `sub_pr_green` / local clean | #436 provisionally integrated | `not_spawned_review_blocked` until human/parent review fallback clears |
| #407 | dependency blocked | #437 integration | `not_spawned_dependency_blocked` |
| #462 follow-up | PR #468 updated / local docs gates green | GitHub CI, human review, and parent integration | `not_spawned_review_blocked` |
| #408 | review blocked | #405 integrated; #462 direct blocker remains open | `not_spawned_review_blocked` |
| #413 | dependency blocked | #407 | `not_spawned_dependency_blocked` |
| #469 | dependency blocked | #409 plus reviewed integration of #462/PR #468 | `not_spawned_dependency_blocked`; implement setup separately from #416 |
| #419 | human-gated | explicit optional beta dependency decision | `not_spawned_human_gate` |

## Human gates

- Never merge PR #438 to `main`.
- No new dependencies beyond an explicit accepted ADR/approval record; #419 still requires its issue-specific explicit inclusion decision.
- No secrets, auth-scope changes, credentialed connector checks, production operations, destructive actions, quality-gate reductions, or generic write tools.
- Reverse ETL execution remains plan -> preview -> approval -> execute and is not part of ordinary verification.
