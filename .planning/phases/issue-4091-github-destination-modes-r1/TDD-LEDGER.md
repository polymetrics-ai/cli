# #4091 — TDD ledger

**Status:** green; requested package and definition gates pass.

| Checkpoint | Evidence | Result |
| --- | --- | --- |
| Discuss | Generated `scripts/gsd prompt discuss-phase 4091` resolved; issue decisions are recorded in `CONTEXT.md` and `DISCUSSION-LOG.md`. | complete |
| Plan | Generated `scripts/gsd prompt plan-phase 4091 --tdd` resolved; scope, required skills, acceptance evidence, and safety boundaries are recorded before production edits. | complete |
| Foundation | Rebased onto `origin/integration/4015-mvp-flat-r1` at `1823a9169`; `internal/app/authorization.go` (#4132) and `internal/connectors/database/managed_target_delivery_ledger.go` (#4135) exist. | complete |
| Red: missing/disabled opt-in | `TestIssueLabelTransportNonAdditiveModesRequireExplicitConnectionConsent` configures `full_overwrite` and `incremental_upsert` on the persisted connection. Disabled paths assert zero POST and PUT sends; enabled paths require a persisted `set_issue_labels` plan. | reproduced: enabled paths fail at the existing `issues/full_append` admission before any provider send |
| Red: changed durable scope | `TestIssueLabelTransportNonAdditiveModesRequirePerConnectionAuthorizationBeforeProviderWrite` changes the configured label after an initial approved write. | reproduced and green: the later no-token apply fails with the recorder still at one PUT |
| Green: authorized modes | The same test executes `full_overwrite` and `incremental_upsert`, then reads the destination labels through the actual declarative reader. | complete: each first and identical-scope unattended run performs exactly one replacement PUT; read-back rejects the prior `legacy` label |
| Green: durable record | The one-time approval creates the landed #4132 `AuthorizationRecord`; the original token is replay-rejected, a no-token identical scope executes, and `RevokeAuthorization` stops the next run. | complete: recorder remains at two PUTs for replay and revocation refusals |
| Green: disabled consent | A plan minted while consent is enabled is applied after the exact per-connection switch is disabled. | complete: Apply rejects before the recorder receives a PUT or POST |
| Refactor/verify/review | Requested tests, vet, generated surface checks, manual review. | complete: see `VERIFICATION.md` and `REVIEW.md` |

## Red evidence

`go test -count=1 -timeout 20m ./internal/app -run '^TestIssueLabelTransportNonAdditiveModesRequireExplicitConnectionConsent$' -v` failed before production edits. Both enabled cases returned `closed issue-label transport connection ... must use issues/full_append`; the disabled cases passed with zero recorded POST and PUT requests. The failure proves the current one-path demonstrator cannot produce the requested non-additive plan.

## Green evidence

`go test -count=1 -timeout 20m ./internal/app -run '^TestIssueLabelTransportNonAdditiveModesRequirePerConnectionAuthorizationBeforeProviderWrite$' -v` passed for set-replace and keyed paths. Its stateful GitHub HTTP recorder starts with `transport-demo` and `legacy`, observes the PUT replacing that set with only `transport-demo`, and serves that exact state to the read-back request. It asserts zero additional PUTs for token replay, revocation, scope drift, and a subsequently disabled per-connection switch.

The inline/manual GSD fallback is deliberate: the generated canonical contract forbids role spawning for this isolated issue lane. The worker executed the generated discuss, plan, verify, and review prompts and recorded the red/green commands and review evidence in this phase directory.

## Generated CLI transcript audit

CI correctly required the `connectors_inspect_github_json` golden output to reflect the changed GitHub definition. It was regenerated only by the sanctioned command:

`POLYMETRICS_UPDATE_GOLDEN_TRANSCRIPTS=1 go test -count=1 -timeout 20m ./internal/cli -run '^TestGoldenTranscripts$' -v`

After the final rebase, `pm` was rebuilt because connector definitions are embedded at build time. The decoded GitHub JSON diff has one line: `"confirm": "destructive"` was added to the existing `set_issue_labels` action. That line is intentional because #4091 makes set-replace destructive and approval-gated. The SHA-256 of every non-GitHub transcript before and after fresh-binary regeneration was identical (`f66292f04a9487bc586aa381c80b156f087fbafd3f9ef6cc38c54124cc202100`), proving sample, unknown, and unsafe entries did not move.

The same fresh binary ran `./pm docs generate --dir docs/cli`. Its only catalog diff is GitHub's #4091 declaration: `set_issue_labels` gains destructive confirmation and its `issue_label/replace` binding for `full_overwrite` and `incremental_upsert`; the existing add and cleanup bindings each declare `full_append`. The corrected auth/rate-limit source and catalog text remained identical.
