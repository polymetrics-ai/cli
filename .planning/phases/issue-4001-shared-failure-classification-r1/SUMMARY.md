---
coverage:
  - id: D1
    description: Closed shared failure classification with non-retryable configuration semantics.
    verification:
      - kind: unit
        ref: internal/failures/classification_test.go
        status: pass
    human_judgment: false
  - id: D2
    description: Configuration, dispatch, and certification consumers use the common contract.
    verification:
      - kind: unit
        ref: internal/connectors/{configuration_validation_test.go,engine/configuration_validation_test.go,commandrunner/failure_contract_test.go,certify/failure_contract_test.go}
        status: pass
    human_judgment: false
---

# Summary — Issue 4001 shared failure classification

## Delivered

- Added dependency-free `internal/failures.Classification`: closed domains and dispatch kinds,
  stable snake-case reason codes, exact JSON-Pointer field paths, bounded identifier references,
  JSON validation, and private Go causes.
- Made retry eligibility structural: only `transient` is retryable.
- Changed declarative engine configuration validation to return shared typed classifications while
  keeping detailed diagnostics as internal causes.
- Added a dispatch-classification carrier to `commandrunner.BlockedCommandError` for #3991, and an
  optional common `untestable_reason` object to certification capability results.
- Added focused configuration, engine, dispatch, and certification tests without modifying the
  PostgreSQL driver or any provider surface.

## Verification

The full changed-package test suite, changed-package vet, PM build, and all non-suite `make verify`
gates listed in `VERIFICATION.md` passed. The pre-implementation Red output is retained in
`traces/red-run.txt`.
