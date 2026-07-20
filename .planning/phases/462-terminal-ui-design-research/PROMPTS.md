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

## Correction PR #467 TTY-gate/state invocation — 2026-07-20

```bash
scripts/gsd doctor
scripts/gsd prompt plan-phase 462 --skip-research > /tmp/gsd-plan-462-correction-467.txt
scripts/gsd prompt programming-loop init --phase 462 --dry-run
```

Observed: `scripts/gsd prompt programming-loop ...` returns `scripts/gsd: unknown GSD command:
programming-loop`; manual universal-loop fallback remains recorded. `/gsd-programming-loop` is not
available through the shell adapter for this run.

```text
Execute accepted review findings on correction PR #467 for issue #462 under parent #397 on branch
`docs/462-terminal-ui-design-review-fixes`, starting from
`e8286ea83a76ac2c6f6257c6e2d40fd21af81640`. No production Go edits and no merge. Reopen the #462
phase artifacts first, then minimally update delegated docs/skill/prompt sources so every TUI/Huh
prompt activation requires both stdin and stdout TTYs; piped/non-TTY stdin falls back to deterministic
plain/noninteractive behavior without consuming scripted stdin, without hanging, and without using `/dev/tty`;
future TUI workers record RED tests for `stdin-piped+stdout-TTY`, `stdout-piped`, `CI`,
`--json`, `--plain`, and `--no-input`. Update RUN-STATE/SUMMARY to record PR #467 open at the
starting head with CI green and human/parent review pending, then validate grep, JSON, skill, scope,
GSD, and docs gates. Do not retry Claude/Copilot.
```

Downstream artifact: local finding disposition recorded on branch `docs/462-terminal-ui-design-review-fixes`;
human review and parent integration gates remain. Git/GitHub remain the current source of truth for
branch/PR head after this starting snapshot.
Verification result: docs-contract contradiction grep, direct dependency/token/export marker check,
skill validation, JSON syntax, exact scope check, `git diff --check`, `scripts/gsd doctor`, and
`make docs-check` pass for this correction slice.

## Follow-up PR #468 invocation — 2026-07-20

```bash
scripts/gsd doctor
scripts/gsd prompt plan-phase 462 --skip-research > /tmp/gsd-plan-462-pr468.txt
scripts/gsd prompt programming-loop init --phase 462 --dry-run
```

Observed: `scripts/gsd prompt programming-loop ...` returns `scripts/gsd: unknown GSD command:
programming-loop`; manual universal-loop fallback remains recorded. `/gsd-programming-loop` is not
available through the shell adapter for this run.

```text
Execute accepted local review findings on follow-up PR #468 for issue #462 under parent #397 on
branch `docs/462-terminal-ui-tty-gate-follow-up`, starting from
`fd122c52458a6ef0db12f60f303c261ed2e63d4c` and base parent commit
`93a117100c6421955262aa32794a91a158d267e1`. No production Go edits and no merge. Reopen the #462
phase artifacts first, then minimally update delegated docs/skill/prompt sources so `--plain`,
`--json`, and `--no-input` always bypass Bubble Tea, Huh, and all prompts; bypass paths produce
only deterministic table/summary output or exact required-flag errors; sequential prompting is
allowed only in explicit accessible mode when both stdin and stdout are TTYs and no bypass flags are
set. Add the shared and Stage 16-specific TTY fallback RED matrix: stdin+stdout TTY activation,
`stdin-piped+stdout-TTY`, `stdout-piped`, `CI`, `--json`, `--plain`, and `--no-input`. Record PR
#467 merged at parent commit `93a117100c6421955262aa32794a91a158d267e1` from old head
`e8286ea83a76ac2c6f6257c6e2d40fd21af81640`; record PR #468 open at starting head
`fd122c52458a6ef0db12f60f303c261ed2e63d4c` with human review pending; GitHub is live source and
local sidecars are not external coverage. Preserve explicit query grid/reverse guide, bare help
exit 0, approval token nondisclosure, direct dependencies, path-safe typed export, and no
`/dev/tty`. Validate contradiction grep, marker matrix, JSON parse, skill validation, scope,
`git diff --check`, GSD doctor, and `make docs-check`. Commit/push planning/docs/evidence and
update PR #468 body. No bots and no merge.
```

Downstream artifact: planning/RED, delegated docs correction, and final evidence recorded on branch
`docs/462-terminal-ui-tty-gate-follow-up`; PR #468 remains open for human review. Git/GitHub are the
live source for branch/PR head after the starting snapshot.
Verification result: contradiction grep, marker matrix, JSON parse, skill validation, direct
state/token/export/accessibility marker checks, exact scope, `git diff --check`, `scripts/gsd
doctor`, and `make docs-check` pass.
