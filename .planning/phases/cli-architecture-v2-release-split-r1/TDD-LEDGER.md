# TDD ledger — CLI Architecture v2 Cobra/Viper release split

## Red / baseline

Captured before production edits against product tree `873cd7b251f70c4a35a607a0d4e86051ea0fbd15` (the only branch delta was the committed phase plan):

- `test -f internal/config/config.go` exited 1: exact latest-main has no typed config package.
- `pm --root <empty-project> help config` exited 1 with zero stdout bytes and `error: help topic "config" not found` on stderr.
- Built the exact latest-main product tree and captured 17 credential-free CLI cases under an empty `HOME`, `CI=1`, and `TERM=dumb`. The manifest records exit code and byte/hash evidence for root/help/JSON help, bare namespaces, connector inspect, agent plan, version, invalid command/help, Gong dynamic JSON help, and worker status.
- Source PR heads retain the original fail-first history for #399/#400/#401/#402/#453; this reconstruction preserves that source-to-candidate provenance rather than inventing new historical red commits.

Observed reconstruction red evidence:

- Applying `8900db14` produced the expected single `internal/cli/cli.go` conflict; resolution selected the Cobra entrypoint only and retained current-main Gong code outside that hunk.
- The first focused build failed at `internal/cli/cobra_router.go:175`: `runHelp` needed the current `jsonOut` argument.
- After that API adaptation, focused tests failed only at the audited latest-main drift seams: the old test expected GitHub dynamic help to fail, `connectors_inspect_github_json` had a stale golden, and bare `github --json` now correctly rendered dynamic help instead of exit 2.

## Green

- All five authorized source patches were reconstructed on latest `main` in the required order.
- Cobra now calls the current `runHelp(..., jsonOut)` signature.
- The obsolete GitHub no-dynamic-surface expectation was replaced with a positive dynamic-help assertion.
- Golden regeneration changed only the two affected cases (`connectors_inspect_github_json` and the renamed `dynamic_connector_bare_json`).
- `go test ./internal/cli -run 'Test(Cobra|Golden|Config)' -count=1` passed after adaptation.

Pending:

- Typed config precedence, invocation isolation, explicit-env behavior, and config consumers pass all focused tests.
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
