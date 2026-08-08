# TDD ledger — seven connector extraction r1

## Plan checkpoint

- **GSD:** `discuss-phase --auto` and `plan-phase --tdd` prompts generated through the project
  adapter; execution is inline/manual for the documented single-worker reason in `PLAN.md`.
- **Green:** pending — import exactly the seven authored bundle deltas and regenerate derived
  surfaces/docs/data.
- **Refactor:** pending — review generated diff for allowlist compliance and generator drift.

## Required red command

```bash
go test -timeout 20m ./cmd/connectorgen -run 'Test(Chatwoot|Gmail|Greenhouse|HelpScout|Jira|LeverHiring|WorkdayRest)'
```

The red result must be recorded verbatim enough to distinguish an assertion failure from a missing
test, compilation failure, or unrelated baseline failure. No production bundle input may change
before that capture.

## Red — observed before any bundle input changed

The seven acceptance test files were imported from `c28bc75a3`; the bundle inputs remained the
current-main versions. The required command exited 1 with assertions (not a compile or unrelated
baseline failure):

```text
--- FAIL: TestChatwootAPISurfaceOperationLedger
    operation_ledger_version = 0, want 1
--- FAIL: TestGmailAPISurfaceOperationLedger
    operation_ledger_version = 0, want 1 (the v2 provenance ledger is required)
    34 legacy excluded row(s) remain, want 0
    covered(45)+blocked(0) = 45, want 79
--- FAIL: TestGreenhouseAPISurfaceOperationLedger
    operation_ledger_version = 0, want 1
    endpoints = 129, want 138 documented operations
    covered(126)+blocked(0) = 126, want 138
--- FAIL: TestHelpScoutAPISurfaceOperationLedger
    operation_ledger_version = 0, want 1
    endpoints = 8, want 144 documented operations
    covered(4)+blocked(0) = 4, want 144
--- FAIL: TestJiraDocumentedSurfaceIsComplete
    operation_ledger_version = 0, want 1
    documented endpoints = 15, want 617
    covered rows = 3, want 590
    blocked rows = 0, want 27
--- FAIL: TestLeverHiringAPISurfaceOperationLedger
    operation_ledger_version = 0, want 1
    endpoints = 11, want 106 documented operations
    covered(5)+blocked(0) = 5, want 106
--- FAIL: TestWorkdayRESTDocumentedSurfaceIsComplete
    operation_ledger_version = 0, want 1
    documented endpoints = 0, want 907
    covered(3)+blocked(0) = 3, want 4
FAIL	polymetrics.ai/cmd/connectorgen
```

**Red conclusion:** current-main bundle inputs cannot satisfy the source-complete operation ledgers;
the target-specific tests precisely fail on the intended missing surface/disposition work.
