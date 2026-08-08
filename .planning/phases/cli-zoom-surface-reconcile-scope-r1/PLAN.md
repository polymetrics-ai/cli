# Plan — scoped operation-surface reconciliation for Zoom slices, R1

## Delivery record

- Parent parity issue: [#3915](https://github.com/polymetrics-ai/cli/issues/3915); immediate consumer: the Healthcare slice [#3946](https://github.com/polymetrics-ai/cli/issues/3946).
- Required skills: `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-safety`, `golang-security`, `golang-design-patterns`, `golang-structs-interfaces`, and `no-mistakes`.
- GSD resolution: the provider slice already recorded `scripts/gsd doctor`, `scripts/gsd sources discuss-phase|plan-phase|execute-phase|verify-work|code-review`, and the generated prompts. `gsd-sdk query init.phase-op cli-zoom-surface-reconcile-scope-r1` has no initialized phase, so this small shared-tooling foundation uses the same documented inline/manual fallback. Role spawning is forbidden by the parent contract.

## Problem

`connectorgen surface-reconcile` can currently narrow reconciliation only with text in
`operation.reason`. Zoom's durable provider-category discriminator is instead in the ledger's
`operation.notes` (`provider_module=<name>`). A Healthcare-only run therefore attempted to rewrite
838 unrelated blocked direct-read reasons while deriving the two Healthcare coverages.

## TDD implementation

1. Add a failing command-level test for `--notes-contains`: an unmatched module leaves the row
   untouched; a matching `provider_module=healthcare` selects and reconciles it.
2. Run the focused test and capture its unknown-flag failure before production code changes.
3. Add the optional `--notes-contains` selector to `surface-reconcile`, combine it with the
   existing reason selector when both are supplied, and update the command help.
4. Re-run focused command-generator tests and use the selector for the Healthcare generated
   direct-read coverage.

## Scope and reuse

- Owned code: `cmd/connectorgen/surfacereconcile.go`, its focused tests, command help, and this
  evidence directory.
- No connector surface, provider artifact, credential, or request execution changes are in this
  foundation commit.
- It unblocks every remaining Zoom provider-module slice and any future connector ledger whose
  stable category provenance is stored in `operation.notes` rather than the human-readable reason.
