---
phase: issue-4072-budget-lifecycle-residual-r1
status: complete
coverage:
  - id: D1
    description: Granted GitHub App token mint runs one Decide and one Finish.
    verification:
      - kind: unit
        ref: internal/connectors/hooks/github/hooks_test.go:TestGitHubAppAuthBudgetLifecycleGrantFinishesExactlyOnce
        status: pass
    human_judgment: false
  - id: D2
    description: Refused GitHub App token mint does not finish or send.
    verification:
      - kind: unit
        ref: internal/connectors/hooks/github/hooks_test.go:TestGitHubAppAuthBudgetLifecycleRefusalDoesNotFinishOrSend
        status: pass
    human_judgment: false
---

# #4072 budget lifecycle residual summary

The engine now consumes `connsdk.BudgetCoordinator` only at the narrow,
declaration-owned custom-auth request boundary. It derives a secret-free batch
from resolved policy identity, opaque scope, and declared budgets; a grant is
finished once after the no-retry token request. A refusal has no lease, hence
it deliberately makes no `Finish` call and sends no provider request.

The prior required-shared and no-retry tests remain unchanged and pass. See
`TDD-LEDGER.md` and `VERIFICATION.md` for red/green and local gate evidence.
