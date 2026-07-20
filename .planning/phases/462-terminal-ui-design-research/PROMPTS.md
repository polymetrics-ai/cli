# Phase 462 Prompts

## Planning invocation

```bash
scripts/gsd doctor
scripts/gsd prompt plan-phase 462 --skip-research
scripts/gsd sources plan-phase
```

The generated prompt was 10,704 bytes. Research was skipped only in the generator because the
primary-source browsing and hands-on isolated terminal lab had already been completed and are
recorded in the phase artifacts and design document.

## Implementation prompt

```text
Execute issue #462 as a documentation/design gate for parent #397. Do not implement production
Go or modify go.mod. Convert the completed primary-source and local interaction research into a
normative Bubble Tea v2 operator-workspace design, repo-local bubble-tea-tui-design skill, and
GSD/Pi/issue routing. Preserve plain/JSON safety, CLI parity, secret rules, reverse approval, and
all human dependency gates. Use a docs-only RED/GREEN/refactor ledger and verify exact scope.
```

## Downstream Pi prompt

The canonical paste-ready prompt for UI workers lives in
`.planning/traces/cli-architecture-v2-pi-prompts.md` under **TUI worker session**. It must be run
from that issue's isolated worktree after #462 and the issue's blocked-by dependencies integrate.

## Review correction invocation — 2026-07-20

```bash
scripts/gsd doctor
scripts/gsd prompt plan-phase 462 --skip-research > /tmp/gsd-plan-462.txt
scripts/gsd prompt programming-loop init --phase 462 --dry-run
```

Observed: `scripts/gsd prompt programming-loop ...` returns `scripts/gsd: unknown GSD command:
programming-loop`; manual universal-loop fallback is recorded in this phase. `/gsd-programming-loop`
remains unavailable through the shell adapter for this run.

```text
Execute accepted review corrections for issue #462 under parent #397 on branch
`docs/462-terminal-ui-design-review-fixes`. No production Go edits. Reopen the #462 phase artifacts
first, then minimally update delegated docs so bare namespaces render help instead of launching
TUI, approval tokens are never displayed, #462/D-TUI is directly encoded in all affected dependency
rows, status is review-blocked/provisional, and query export path safety is explicit. Validate with
RED/GREEN docs greps, skill validation, GSD doctor, diff/scope checks, and docs-check when feasible.
Open a new stacked PR to `feat/cli-architecture-v2`; do not merge.
```

Downstream artifact: branch `docs/462-terminal-ui-design-review-fixes`; planning and docs
correction commits pushed; terminal evidence recorded in this phase artifact update.
Verification result: docs-contract grep, dependency roster check, skill validation, JSON syntax,
scope check, `git diff --check`, `scripts/gsd doctor`, and `make docs-check` pass.
