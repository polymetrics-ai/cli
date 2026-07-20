# Phase 462 Summary

Status: provisionally integrated / review blocked. PR #465 (`docs/462-terminal-ui-design-research`)
was merged into the parent branch from head `6853fee28e0208381b49931fb1f5dfec42ee50ef`, but Claude
review is disabled, Copilot backup exhausted quota, and fallback is human review. Correction PR #467
merged into the parent branch at parent commit `93a117100c6421955262aa32794a91a158d267e1` from old
head `e8286ea83a76ac2c6f6257c6e2d40fd21af81640`. Follow-up correction PR #468 is open at starting
head `fd122c52458a6ef0db12f60f303c261ed2e63d4c`; human review is pending. Git/GitHub remain the
current source of truth after these starting snapshots. Local sidecar review is local evidence only,
not external review coverage.

The requested reference applications have been exercised in an isolated local lab and distilled
into a Polymetrics-specific terminal design system. The chosen structural direction is a quiet
LazyGit-style operator workspace with fzf filter/list/preview behavior, bpytop exact metric/chart
density, Gum-focused wizard steps, and the existing pipeline rail. Explicit Normal/Filter/Edit
modes provide Vim navigation without hijacking text input.

The phase adds a repo-local Bubble Tea design skill, query chart/dashboard grammar, responsive
layout classes, accessibility and test matrices, and GSD/Pi routing. NTCharts v2 ran successfully
in isolation but remains unapproved for `go.mod`; a dedicated issue and human gate are required.

No production Go, CLI behavior, dependency, generated help, website, connector definition, or
credential data changes are part of this phase. Review correction work reopened the docs-only
ledger and resolved accepted findings before the parent orchestrator finalizes the design gate.

Pre-delivery verification passes: GSD doctor/provenance, skill validation, `git diff --check`,
exact no-production/dependency scope, and `make docs-check`. Live markers are present on all seven
affected UI issues; #462 is nested under #397 and #463 is nested under #411. The isolated tmux lab
was stopped after screenshots so no monitoring/audio processes remain running.

Original delivery: branch `docs/462-terminal-ui-design-research`, stacked PR #465 targeting
`feat/cli-architecture-v2`. PR #464 was closed as superseded because its original `codex/` branch
prefix failed the repository's explicit `<type>/<description>` naming gate; the corrected `docs/`
branch preserves the verified content. Parent integration is provisional because external review
coverage is blocked.

## Accepted review corrections — 2026-07-20

Planned correction branch: `docs/462-terminal-ui-design-review-fixes`.

Accepted findings to fix across delegated docs:

1. Bare namespaces must not launch TUI surfaces. Bare `pm query` and bare `pm reverse` render
   contextual help/subcommand summaries and exit 0; invalid actions remain usage errors. Explicit
   interactive subcommands will be named consistently (`pm query grid`, `pm reverse guide`).
2. Approval tokens are sensitive one-time authorization values. Guided reverse may relay them only
   ephemerally in memory through the existing plan → preview → approval → execute path; never print
   them in final frames, transcripts, logs, screenshots, accessibility output, JSON, or
   shell-equivalent command text. Typed approval stays required.
3. #462/D-TUI must appear directly in each affected TUI dependency row, not only in prose.
4. Evidence status must remain provisionally integrated / review blocked: PR #465 head
   `6853fee28e0208381b49931fb1f5dfec42ee50ef`, Claude disabled, Copilot quota exhausted, fallback
   human; correction PR #467 started open at `e8286ea83a76ac2c6f6257c6e2d40fd21af81640` with CI
   green and human/parent review pending.
5. Query export must document a typed read-only, path-confined, no-overwrite-by-default contract
   with confirmation only when stdin/stdout are TTYs, noninteractive `--force`, sanitized command
   echo, and exact `--no-input` guidance.

Correction verification passes: docs-contract contradiction grep, direct #462/D-TUI roster checks,
approval-token/query-export/status marker checks, skill validation (`Skill is valid!`), JSON syntax,
exact scope diff, `git diff --check`, `scripts/gsd doctor`, and `make docs-check`. Full `make verify`
was not run because this issue is docs-only and the task requested `make docs-check` when feasible.

## Accepted review findings on correction PR #467 — 2026-07-20

This bounded docs correction aligns the normative TUI gate after review found stdout-only and
ambiguous TTY wording. Bubble Tea/Huh/prompt activation must require **both stdin and stdout TTYs**
plus the existing disables (`--json`, `--plain`, `--no-input`, `PM_NO_TUI`, `CI`, `TERM=dumb`).
With piped or non-TTY stdin, Polymetrics must fall back to deterministic plain/noninteractive
behavior, never consume scripted stdin unexpectedly, never hang, and never open `/dev/tty` to bypass
the gate.

Future production TUI issues must record RED tests for `stdin-piped+stdout-TTY`, `stdout-piped`,
`CI`, `--json`, `--plain`, and `--no-input`. The explicit `pm query grid`, `pm reverse guide`,
read-only query export, approval-token secrecy, and accessibility/plain contracts remain preserved.
Next gates are local finding disposition, human review, and parent integration; no Claude/Copilot
retry is requested in this blocker window.

## Accepted local review findings on follow-up PR #468 — 2026-07-20

This bounded docs correction fixes the remaining prompt-bypass and evidence-state findings.
`--plain`, `--json`, and `--no-input` must always bypass Bubble Tea, Huh, and all prompts; those
paths produce deterministic table/summary output when required flags are present, or exact
required-flag errors only. Sequential prompting is allowed only in explicit accessible mode when
both stdin and stdout are TTYs and none of those bypass flags are set.

The execution prompt must carry a shared TUI fallback RED matrix for all TUI phases, and Stage 16
must name the same matrix explicitly: stdin+stdout TTY activation, `stdin-piped+stdout-TTY`,
`stdout-piped`, `CI`, `--json`, `--plain`, and `--no-input`. Prior corrections remain preserved:
explicit `pm query grid` and `pm reverse guide`, bare namespace help exit 0, approval-token
nondisclosure, direct dependencies, path-safe typed export, and no `/dev/tty` bypass.

Planning/RED and green evidence are captured on branch `docs/462-terminal-ui-tty-gate-follow-up`.
Verification passes: contradiction grep, marker matrix, JSON parse, skill validation, direct
state/token/export/accessibility marker checks, exact scope, `git diff --check`, `scripts/gsd
doctor`, and `make docs-check`. PR #468 body update is part of the final handoff slice. No bot
retry or merge is requested.
