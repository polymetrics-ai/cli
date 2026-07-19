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
