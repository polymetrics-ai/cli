# #4091 — TDD ledger

**Status:** red test authored; expected failure pending execution.

| Checkpoint | Evidence | Result |
| --- | --- | --- |
| Discuss | Generated `scripts/gsd prompt discuss-phase 4091` resolved; issue decisions are recorded in `CONTEXT.md` and `DISCUSSION-LOG.md`. | complete |
| Plan | Generated `scripts/gsd prompt plan-phase 4091 --tdd` resolved; scope, required skills, acceptance evidence, and safety boundaries are recorded before production edits. | complete |
| Foundation | Rebased onto `origin/integration/4015-mvp-flat-r1`; `internal/app/authorization.go` (#4132) and `internal/connectors/database/managed_target_delivery_ledger.go` (#4135) exist. | complete |
| Red: missing/disabled opt-in | `TestIssueLabelTransportNonAdditiveModesRequireExplicitConnectionConsent` configures `full_overwrite` and `incremental_upsert` on the persisted connection. Disabled paths assert zero POST and PUT sends; enabled paths require a persisted `set_issue_labels` plan. | reproduced: enabled paths fail at the existing `issues/full_append` admission before any provider send |
| Red: changed durable scope | Pending: recorder test must fail before any provider request when an authorization scope changes. | pending |
| Green: authorized modes | Pending: set-replace and keyed read-back results prove each authorized write completed. | pending |
| Refactor/verify/review | Pending. | pending |

## Red evidence

`go test -count=1 -timeout 20m ./internal/app -run '^TestIssueLabelTransportNonAdditiveModesRequireExplicitConnectionConsent$' -v` failed before production edits. Both enabled cases returned `closed issue-label transport connection ... must use issues/full_append`; the disabled cases passed with zero recorded POST and PUT requests. The failure proves the current one-path demonstrator cannot produce the requested non-additive plan.
