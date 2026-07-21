# Prompts — Issue #397 parent orchestration

## Delegated #408 Shepherd correction snapshot — 2026-07-20

Role: exactly one isolated Sol/high correction worker for issue #408; no subagent; no VERIFY, REVIEW, INTEGRATE, sub-PR, parent-ready, or merge action.

Command/evidence path:

```bash
scripts/gsd doctor
scripts/gsd prompt programming-loop init --phase 408-flow-etl-dashboards --dry-run
git ls-remote origin refs/heads/feat/cli-architecture-v2 refs/heads/feat/408-flow-etl-dashboards
```

Result:

- GSD doctor passed; `programming-loop` remains unavailable, so the recorded manual universal-loop fallback continues.
- Live parent head: `21d195aff0c7bd60b3bf54f14b1ce165cec9e03f`.
- Live graph: 44 nodes, 43 subissue edges, 65 dependency edges, 0 errors.
- #408 correction complete at implementation commit `c70ecf64`; synchronized evidence pushed at `64f1a920`; execute completion false pending Shepherd handoff and independent VERIFY.
- #413: `not_spawned_write_scope_collision`.
- #419: human-deferred; no beta dependency and no other dependency approval.
- Phase 437 pending intake: planning-only.

Downstream artifact:

- `.planning/phases/408-flow-etl-dashboards/{PLAN.md,TDD-LEDGER.md,VERIFICATION.md,SUMMARY.md,RUN-STATE.json,PROMPTS.md}`
- `.planning/phases/397-cli-architecture-v2-orchestration/{PLAN.md,TDD-LEDGER.md,VERIFICATION.md,SUMMARY.md,RUN-STATE.json,PROMPTS.md}`
- `.planning/traces/cli-architecture-v2-orchestration-state.yaml`

Verification result: #408 strict RED captured (`go test ./internal/ui/run -run '^TestBubbleTeaV2ModelAndTeatestProgram$' -count=1` -> missing Bubble Tea v2 required module before dependency/production edits); GREEN/full non-race/focused-race/module gates pass at `c70ecf64`, with synchronized evidence pushed through `64f1a920`. Full race and `make verify` were not rerun in CORRECT. Parent verification and review not run or claimed.
