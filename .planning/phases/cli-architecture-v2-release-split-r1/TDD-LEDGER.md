# TDD ledger — CLI Architecture v2 Cobra/Viper release split

## Red / baseline

Pending before production edits:

- Confirm exact latest-main has no `internal/config` package and no `pm help config` topic.
- Capture a deterministic current-main binary and 17-case credential-free CLI transcript baseline (exit code, stdout bytes, stderr bytes).
- Source PR heads retain the original fail-first history for #399/#400/#401/#402/#453; this reconstruction preserves that source-to-candidate provenance rather than inventing new historical red commits.

Expected reconstruction failures from the audited scout:

- `internal/cli/cli.go` textual conflict when applying the Cobra shell over current-main Gong.
- Compile failure until Cobra routing uses current `runHelp(..., jsonOut)`.
- Focused Cobra/golden failures until the obsolete GitHub dynamic-surface assumption and only affected golden outputs are updated.

## Green

Pending:

- All five authorized source patches reconstructed on latest `main`.
- Current Gong behavior retained.
- Typed config precedence, invocation isolation, explicit-env behavior, and config consumers pass focused tests.
- Reverse smoke structurally enforces plan -> preview -> approval -> execute.
- All 17 unchanged CLI transcript cases match latest-main byte-for-byte.
- Full local verification passes.

## Refactor / hardening

Pending:

- Regenerate only affected docs/golden outputs.
- Confirm no TUI/event/logging/telemetry packages or dependencies enter the diff.
- Confirm no global Viper, `AutomaticEnv`, watcher, credential config, or hidden worker enablement.
- Add truthful ADR/release-split records without importing historical bootstrap state.

## Review and correction rounds

Pending exact-version PM Codex packets, one PM synthesis, and independent Shepherd only after clean synthesis. Manual/Claude/Copilot substitution is explicitly disallowed for this candidate.

## Skills

`gsd-core`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-lint`, `golang-documentation`, `golang-design-patterns`, `golang-spf13-cobra`, `golang-spf13-viper`, `golang-dependency-management`, `golang-continuous-integration`.
