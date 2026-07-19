# Phase 437 TDD Ledger

Issue #437; invocation `issue-437-pi-sol-high-20260719T095145Z`; Sol/high; start `6c038bb4ab4a5497fca28a0cab42d0a7fa4eb22b`.

GSD doctor/list and plan-phase prompt passed. Adapter has no `programming-loop`; manual universal-runtime-loop fallback is active. Loaded `gsd-core`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-context`, `golang-concurrency`, `golang-documentation`, `golang-spf13-cobra`.

| Step | Kind | Evidence | Status |
|---:|---|---|---|
| 0 | Planning | Six issue-local artifacts written before tests/production | Complete |
| 1 | RED | `go test ./internal/cli -run 'TestConnectorsCommandIsNativeCobraSubtree|TestNativeConnectors|TestNativeCertify' -count=1` | Failed as required before production edits: `newConnectorsCobraCommand` and `newConnectorsCobraCommandWithRuntime` are undefined |
| 2 | GREEN | Native connectors/certify subtree, typed flags/runtime seam, compatibility normalization, and namespace-only parser removal | Pass: focused `3.876s`; focused router/golden/certify/telemetry `111.507s`; repeated ×10 `34.915s`; race `40.844s`; certify package `336.422s` |
| 3 | Help-refactor RED | Focused trailing action/nested certify help contract | Failed before the final help production edit: list/catalog/inspect returned operation output, certify positional help was usage 2, and certify JSON trailing help returned a certification envelope |
| 4 | Full gate | connector validation, docs/website, gofmt/vet/test/build/make verify, certify smoke | Pending |
| 5 | Delivery | Finalize, commit/push, no PR/review | Pending |

## RED contract

- `connectors` is native Cobra, absent from legacy wrappers, with native `list`, `catalog`, `inspect`, hidden positional help and existing manual aliases.
- `certify` is a native nested command supporting single connector, `--all`, and `--sweep`, with all currently consumed flags represented as StringArray flags with `NoOptDefVal=true` to preserve parser semantics.
- Repeated flags are last-wins where handlers currently use `first`; bare/assigned/space forms, positional connector selection, ignored tails, literal `--`, malformed/legal unknowns, invalid actions/operands, no later action discovery, and globals preserve the exact contract.
- Bare/topic/long/short/positional/trailing help is canonical in text/JSON and side-effect-free. Invalid actions remain usage errors.
- Certify reports retain exits 0 pass, 1 internal, 2 certification failure, 3 leak dominance and one-envelope behavior.
- Fresh trees remain re-entrant. Batch context cancellation terminates workers; parallel limit, event sequence, and telemetry names/status remain intact.
- No credential value reaches stdout, stderr, report JSON, events, telemetry, or parser errors.
- Dynamic connector dispatch stays on legacy parsing exactly; no connector defs or live checks/writes.

## Evidence log

Append actual failing/passing commands and results only; do not backfill.
