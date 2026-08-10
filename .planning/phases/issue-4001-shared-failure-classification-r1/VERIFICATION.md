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
