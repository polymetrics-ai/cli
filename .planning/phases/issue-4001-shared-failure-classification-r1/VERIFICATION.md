# Verification — Issue 4001

## Automated checks

- [x] Red command recorded before production code at `traces/red-run.txt`.
- [x] Focused failure/configuration/dispatch/certification command passes.
- [x] `gofmt` and `git diff --check` pass.
- [x] Changed-package tests pass with `-timeout 20m`.
- [x] `go vet` passes for changed packages.
- [x] `go build ./cmd/pm` passes.
- [x] Non-suite `make verify` gates pass individually: `tidy-check`, `lint`, `docs-check`,
      `smoke-no-build`, `agent-contract-check`, `connectorgen-validate`,
      `connectorgen-surface-sync`, `connector-boundary`, and `release-workflow-check`.

## Requirement checks

- [x] Stable JSON-safe domain and all five dispatch codes are closed and tested.
- [x] Empty/unknown code, malformed field pointer, invalid dispatch/domain pairing, and unsafe
      references are rejected.
- [x] Configuration is non-retryable by method contract.
- [x] User message and private cause remain separate; the cause does not serialize.
- [x] Generic configuration, engine dispatch, and certification have focused consumer tests.
- [x] `synccontract.RecoveryOutcome` remains unchanged and documented as CDC-only.

## Manual review focus

- [x] No provider, driver, or call-graph implementation leaked into this foundation.
- [x] No serialized error exposes a cause, credentials, request values, or provider response bodies.
- [x] JSON schema extension is optional/backward-compatible for certification reports.

## Environment note

`make smoke-no-build` passed. Its `mktemp` invocation ignored the supplied temporary-directory
environment and created one ephemeral system scratch project. It was moved to the operating
system Trash immediately after the check; no project data or credential was used.

## Stacked-delivery current-base verification — 2026-08-11

This section supplements the preserved original verification record without changing it.

- [x] Remote parent base confirmed as
      `origin/docs/4015-connector-release-certification` at
      `5996a8a2a5e99c8aa8eb5a8603ecb1f6bba21f12`.
- [x] Source head confirmed as `origin/fm/cli-cert-shared-foundations-r1` at
      `d517756f0f9bc0cc41d90b4cf717325328352a3f`.
- [x] The seven requested source commits were replayed in order with no cherry-pick conflict.
- [x] `git range-diff` and patch-id replay comparison against the source range.
- [x] Focused failure classification, engine, commandrunner, certification, and public Save/Load
      proof on the replay head.
- [x] Formatting, changed-path build/vet, lint, connector-boundary, docs, agent-contract, and
      relevant repository gates on the replay head.
- [x] Manual-inline GSD `verify-work` and deep `code-review`, recorded without spawning roles.

PR #4013 remains the preserved direct-to-main audit record until the replacement child is green;
this stacked delivery neither retargets, closes, nor mutates it.

### Current-base results

All commands below exited successfully on the replay before the delivery-evidence commit:

- `git range-diff origin/main..origin/fm/cli-cert-shared-foundations-r1
  origin/docs/4015-connector-release-certification..HEAD`, plus a stable patch-id comparison of
  all seven source/replay pairs.
- Focused and full changed-package tests:
  `./internal/failures`, `./internal/connectors`, `./internal/connectors/engine`,
  `./internal/connectors/commandrunner`, `./internal/connectors/certify`, and `./internal/app`;
  `./internal/cli` was also run separately with `-timeout 20m`.
- The public `certify.Report.Save` / `certify.LoadReport` proof persisted an `untestable_reason`,
  confirmed `domain=system` and `dispatch_kind=declared_but_unroutable_command`, confirmed the
  private cause was absent from JSON, and confirmed the loaded cause was nil. Its temporary
  source harness was removed immediately afterward.
- `gofmt` on every changed Go path, changed-package `go vet`, `go build ./cmd/pm`, and
  `git diff --check`.
- `make tidy-check`, `make lint`, `make docs-check`, `make smoke-no-build`,
  `make agent-contract-check`, `make connectorgen-validate`,
  `make connectorgen-surface-sync`, `make connector-boundary`, and
  `make release-workflow-check`.

The smoke project used only local sample connectors and was moved to the system Trash after the
check. No credentials, provider calls, PostgreSQL work, or reverse-ETL provider execution occurred.
