# Issue 4166 Verification Checklist

**Status:** all three gaps are verified, including the real-GitHub fresh-binary flow and zero-residue provider cleanup.

- [x] Gap 1 exact negative-control tests fail the sabotaged action and pass the intact definition.
- [x] Gap 1 report records exact operation counts and non-live categories/reasons.
- [x] Gap 2 exact tests observe declared transport source/stage/destination/read-back execution.
- [x] Gap 2 missing/unregistered declaration fails with zero false-positive execution.
- [x] Gap 3 one fresh binary runs real ETL and reverse action as one composed flow against the faithful provider control.
- [x] Gap 3 independently reopens durable Parquet and observes an advanced checkpoint in the faithful provider control.
- [x] Gap 3 independently reads the exact mutation from GitHub.
- [x] Gap 3 replay, expired/unapproved, and authentication refusals are typed and leave provider state unchanged in the faithful provider control.
- [x] Gap 3 authentication refusal leaves its durable checkpoint unchanged; observed `internal/internal_error` classification is tracked separately by #4169.
- [x] Disposable GitHub repository and all created resources leave zero residue.
- [x] Generic definition/flow path and every GitHub-specific hop are documented.
- [x] Focused package tests, vet, formatting, drift, agent-contract, and GSD workflow checks pass.
- [x] Changed artifacts contain no credential, approval token, response body, or rendered rate scope.
- [x] Derived artifacts regenerated once; final committed worktree cleanliness remains a delivery gate.
- [x] PR #4170 is open with per-gap evidence; issue #4166 has the original evidence comment and receives a live-proof correction comment at final delivery.
- [x] API-reported PR base equals `integration/4015-mvp-flat-r1`.

## Definition-Path Audit

The Gap 3 primary route is generic after registry construction:

`flow.Engine` → `connectorFlowActionRunner` → `app.ExecuteAuthorizedFlowAction` → registered `engine.Connector` → bundle-declared `issues` read / `comment_issue` validate, preview, write / `issue_comments` readback.

`comment_issue` is not handled by `internal/connectors/hooks/github.Hooks.ExecuteWrite`; it takes the declarative fallback from `writes.json`. Token bearer authentication and REST request construction are engine-declared behavior. The registry does import the generated hookset and attaches the GitHub hook object by connector name, but this chosen action calls none of its GitHub-specific write branches. A connector with equivalent definitions can reuse the same flow/action/app/engine path.

The non-generalizable GitHub-specific production code observed is limited to:

- Gap 2's `issueLabelTransportDefinitionFactories`, `isIssueLabelTransportConnector`, fixed issue-label configuration keys, and same-definition dispatch guard in `internal/app`.
- `internal/connectors/hooks/github` for GitHub App token exchange and eight compound/normalizing write actions. Those hooks are present on the connector but are not invoked by the chosen `comment_issue` loop.
- CLI HTTP-error classification remains generic after the declarative connector returns a verified authentication failure: the current 401 envelope is `internal/internal_error`. Issue #4169 owns the provider-neutral product correction; this validation PR does not alter it.
- GitHub definition paths, schemas, pagination, and rate-limit declarations. These are connector data, not flow-engine special cases.

The live proof targets only a uniquely named repository under `Polymetrics-Cert`; the pre-existing `pm-cert-3993-20260810-wz0fru` repository is never referenced or modified. The harness hard-refuses creation or cleanup outside the exact certification owner and `pm-cert-flow-4166-<10 hex>` namespace.

The certification runbook's resource-owner labels `polymetrics-cli-cert` and `polymetrics-cli-cert-1` are stale. The live API proves the actual organization is `Polymetrics-Cert`. The line-277 credential is a labeled 116-character line containing a 93-character token value; extracting the entire line produces 401, while extracting only the token value authenticates. This is an operational runbook finding, not a product change.

## Coverage Numbers

- GitHub source-locked REST inventory: 1,220 operations (`github-combined-operation-ledger --check`).
- Declarative write-action inventory: 607 actions.
- Definition preparation after this change: 607/607 actions, each through `engine.DryRunWrite`; a corrupt schema is terminal.
- Selected live lifecycle per full certification run: 2 provider mutations (`create_label` and `delete_label`) plus independent label readback. The other 605 actions are definition-prepared but not provider-mutated by that run, and retain explicit non-live reasons.
- Curated reversible lifecycle choices available in `certification.json`: 3 (`create_label/delete_label`, `create_issue/close_issue`, and `create_milestone/delete_milestone`). This PR does not broaden live mutation selection.
- Gap 3 faithful flow control: 1 issue extracted, 1 warehouse row read after reopen, 1 flow action acknowledged, provider comments 1→2, 1 committed checkpoint, 1 flow receipt, zero writes for each of replay/unapproved/auth refusal, and no checkpoint advancement on the auth refusal.
- Gap 3 live GitHub counts: 1 issue extracted, 1 warehouse row independently queried, 1 flow action acknowledged, 2 comments observed after the flow, 1 committed checkpoint, 1 persisted flow receipt, zero writes for replay/unapproved/auth refusal, unchanged auth-refusal checkpoint, and zero provider residue after repository deletion plus independent 404 read-back.
