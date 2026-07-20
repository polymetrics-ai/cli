# Phase 462 Summary

Status: provisionally integrated / review blocked. PR #465 (`docs/462-terminal-ui-design-research`)
was merged into the parent branch from head `6853fee28e0208381b49931fb1f5dfec42ee50ef`, but Claude
review is disabled, Copilot backup exhausted quota, fallback is human review, and an accepted
correction PR from `docs/462-terminal-ui-design-review-fixes` is pending.

The requested reference applications have been exercised in an isolated local lab and distilled
into a Polymetrics-specific terminal design system. The chosen structural direction is a quiet
LazyGit-style operator workspace with fzf filter/list/preview behavior, bpytop exact metric/chart
density, Gum-focused wizard steps, and the existing pipeline rail. Explicit Normal/Filter/Edit
modes provide Vim navigation without hijacking text input.

The phase adds a repo-local Bubble Tea design skill, query chart/dashboard grammar, responsive
layout classes, accessibility and test matrices, and GSD/Pi routing. NTCharts v2 ran successfully
in isolation but remains unapproved for `go.mod`; a dedicated issue and human gate are required.

No production Go, CLI behavior, dependency, generated help, website, connector definition, or
credential data changes are part of this phase. Review correction work is reopening the docs-only
ledger to resolve accepted findings before the parent orchestrator finalizes the design gate.

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
   human, accepted correction PR pending.
5. Query export must document a typed read-only, path-confined, no-overwrite-by-default contract
   with explicit TTY confirmation/noninteractive `--force`, sanitized command echo, and exact
   `--no-input` guidance.
