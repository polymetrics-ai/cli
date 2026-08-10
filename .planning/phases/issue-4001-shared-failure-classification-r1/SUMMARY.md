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

## Stacked-delivery replay

The exact seven-patch #4001 series was replayed conflict-free onto current campaign base
`5996a8a2a5e99c8aa8eb5a8603ecb1f6bba21f12` for child branch
`feat/4001-stack-shared-certification-failures`. Range-diff and stable patch IDs matched every
source patch. The current-base focused/full tests, public Save/Load safety proof, changed-path
build/vet/format checks, lint, docs, generator/boundary, agent-contract, and release checks passed.
No-mistakes run `01KZPSZDQ0VSQZ0Q8RV8K4MJ77` then passed its local intent, rebase, review, test,
documentation, and lint gates with zero findings. The correction accounting is 4/5 inherited from
PR #4013 plus 1 new typed-nil cause correction: 5/5 consumed with 0 remaining. PR #4013 remains
unchanged as the historical audit record.
