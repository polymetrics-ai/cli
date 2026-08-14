# #4091 — TDD ledger

**Status:** plan checkpoint complete; red test design pending code discovery.

| Checkpoint | Evidence | Result |
| --- | --- | --- |
| Discuss | Generated `scripts/gsd prompt discuss-phase 4091` resolved; issue decisions are recorded in `CONTEXT.md` and `DISCUSSION-LOG.md`. | complete |
| Plan | Generated `scripts/gsd prompt plan-phase 4091 --tdd` resolved; scope, required skills, acceptance evidence, and safety boundaries are recorded before production edits. | complete |
| Foundation | Rebased onto `origin/integration/4015-mvp-flat-r1`; `internal/app/authorization.go` (#4132) and `internal/connectors/database/managed_target_delivery_ledger.go` (#4135) exist. | complete |
| Red: missing/disabled opt-in | Pending: recorder test must fail with a non-additive request and prove zero provider sends. | pending |
| Red: changed durable scope | Pending: recorder test must fail before any provider request when an authorization scope changes. | pending |
| Green: authorized modes | Pending: set-replace and keyed read-back results prove each authorized write completed. | pending |
| Refactor/verify/review | Pending. | pending |
