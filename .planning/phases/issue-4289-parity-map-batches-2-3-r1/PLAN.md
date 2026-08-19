## Task Delivery Header

- Issue: Refs #4289 — chore(connectors): map parity batches 2 and 3
- Base branch: fm/cli-reverse-etl-destination-r1
- Merges into: fm/cli-reverse-etl-destination-r1 → main
- Delivery: PR #4300 remains open from `fm/cli-map-batch23-r1` after a non-rewriting merge of the temporary typed-destination foundation; its GitHub API base is `fm/cli-reverse-etl-destination-r1`, local and generated gates are green, and final merge remains human-gated.
- Working branch: fm/cli-map-batch23-repair-r1 (publishes fast-forward to `fm/cli-map-batch23-r1`)
- Task: Reconcile the preserved source-locked maps for batch 2 (`grafana`, `trello`, `slack`, `n8n`, `google-calendar`, `gmail`, `twilio`, `amazon-sqs`, `elasticsearch`) and batch 3 (`gong`, `google-ads`, `facebook-marketing`, `linkedin-ads`, `aircall`, `xero`, `paypal-transaction`, `gocardless`, `amazon-seller-partner`, `miro`) across binary read, binary write, direct read, direct write, ETL, reverse ETL, and installed-binary CLI command surfaces. Every documented operation remains connector-owned, source-pinned, API-surface-bound, and reachable when faithfully representable; Gitea remains excluded.
- Verification: seven-surface ledger assertion; source-lock/disposition integrity checks; `go run ./cmd/connectorgen validate`; `go run ./cmd/connectorgen surface-sync --check`; generated-artifact checks; targeted conformance and commandrunner tests; detached `connector-boundary`; and repository verification gates applicable to definition data.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Every selected connector has auditable public provenance | live | Its source lock has the official artifact URL, capture date, SHA-256, byte count, `counts.total`, per-kind/method counts, and explicit `coverage_confidence` with a basis. The local verifier checks that captured inventory against the ledger without refetching mutable provider pages or discovery documents. A partial source is a hold, not a complete-map claim. |
| Every documented operation is accounted for exactly once | live | The integrity check proves source-inventory count equals disposition count and API-surface bindings, while primary parity-class totals equal that denominator. |
| Disabled operations have an honest reason | live | The verifier rejects missing state/rejection fields, treats un-authored contracts as `declaration-pending`, and requires file/line plus a minimal change for every `foundation-gap`. |
| No credentialed or live provider execution is claimed | live | Every ledger reports `live_certification: pending`; all source artifacts are public descriptions and all validation is local structural/fixture evidence. |

## Scope and Ownership

This is a nineteen-connector reconciliation exception to the one-connector implementation lane. The issue owns the listed disjoint connector bundle trees and common planning evidence. Production edits remain connector-local under `internal/connectors/defs/<listed-connector>/`; no engine or shared-tool change is in scope. The working directory name is `paypal-transaction`, which is the bundle corresponding to the issue's PayPal Transactions label.

## GSD Lifecycle and Inline Fallback

The adapter was checked with `scripts/gsd doctor`; its command sources were resolved for `discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and `code-review`; generated prompts were reviewed. The canonical contract forbids role spawning and this task is running in a non-Pi inline worker, so the lifecycle is executed manually in this phase record. `go run ./cmd/agentcontractgen check` passed before planning.

Required skills loaded: `golang-how-to`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`, and `golang-testing`. These are used only to assess existing engine contracts and validation truth; no Go implementation is added.

## Foundation Coordination — 2026-08-20

The merged `c6f03c937` typed-destination foundation admits `declarative_api/declarative_typed_destination` declarations through the orchestrator, but the App/CLI persisted-dispatch path does not yet select that generic destination as it does GitHub's specialized adapter. Connector-local `sync_transport.json` declarations and their structural/preflight tests may proceed; they must record reverse ETL as foundation-pending and must not claim application-level deployment or add a connector-local dispatch workaround. Before the final push, fetch and merge the latest `origin/fm/cli-reverse-etl-destination-r1`, prove its SHA is an ancestor of this branch, and exercise the installed App/CLI path against the real selected foundation head.

Every existing schema-backed typed write action gets an explicit reverse-ETL eligibility disposition. `risk`, required scopes, confirmation, and destructive classification govern approval and execution; they are never semantic exclusions. The current 621 target actions are individually representable as closed destination action shapes, but no connector-owned stream-to-required-input mapping has been evidenced yet, and they are not all bindable in one destination declaration: `DestinationTransportDescriptor.Validate` admits only one named `apply_strategies` action per sync mode (`internal/connectors/sync_transport.go:388-415`), and `ApplyStrategyFor` resolves that one action (`:471-480`). The action ledger records both the pending exact `input_fields` source binding and `declarative-typed-destination-action-multiplicity` as foundation/declaration dependencies rather than inventing a selector, source fields, a connector-local workaround, or a one-action-as-complete claim. A one-action destination may prove the App/CLI route after the dispatch repair lands, but it cannot close any connector's reverse-ETL reconciliation.

The connector-local command projection now accounts for every existing action without weakening that destination hold. `generate-installed-write-commands.mjs` projects only the action's existing closed `record_schema`: 534 actions have installed `implemented` commands with required scalar flags or declaration-bound top-level object/array JSON flags; 82 actions have installed `partial` commands explaining the technical scalar-union hold; five actions (`n8n` archive/unarchive/transfer workflow, Aircall update contact, and GoCardless customer-bank-account-token) remain `declaration-pending` because their preserved 5,127-operation source inventory has no exact provider-operation row. The partial set is source-bound and directly CLI-reachable but does not count as executable proof: its values include documented string/null or string/integer unions that cannot be represented by the current closed flag vocabulary. `internal/connectors/commandrunner/runner.go:1737-1816` accepts only concrete scalar flag types, while the materialized projection rejects non-single-type fields at `cmd/connectorgen/batch_materialize.go:1552-1581`. The required future change is a declaration-bound tagged scalar-union input that validates against the exact record field and never opens a raw body, URL, method, or action selector. This lane neither implements nor works around it.

## Foundation Coordination — REST Structured Direct-Write Bodies [key=rest-structured-body-cli]

Complete documented REST write reachability also requires a declaration-bound structured-body input for exact nested request schemas. Grafana's public OpenAPI `POST /api/access-control/roles` (`createRole`) requires `application/json` `CreateRoleForm`; the pinned public source supplies that schema. The current command runner refuses the only structured JSON flag shape outside a fixed GraphQL operation (`internal/connectors/commandrunner/runner.go:462-502`) and rejects a direct-write raw `body` mapping (`:1425-1455`). A connector-local raw body or schema-less flag would violate the closed command contract.

The decision is resolved: firstmate authorized the bounded capability and routed it to a separate engine PR. That foundation must permit a declaration-owned REST structured JSON body only when it is compiled and validated against that operation's exact `rest.body_schema`, retains declared redaction and size limits, and cannot choose a URL, method, operation, or arbitrary raw request. This is not a generic HTTP writer. This connector lane will neither implement shared engine code nor add a raw-body workaround; affected operations remain `declaration-pending` until the engine head is composed and exact bindings are authored.

## Batch 2 — Red / Green / Refactor

1. **Red:** the nine target bundles lack `sources/<connector>-operation-source-lock.json` and `sources/<connector>-declaration-disposition.json`; therefore no provenance or six-class complete-map assertion can pass.
2. **Green:** credential-free public descriptions are pinned, parsed into method/path inventories, and mechanically reconciled to the existing `api_surface.json`. Every mismatch remains an explicit disabled disposition; no request, response, pagination, or body schema is inferred.
3. **Refactor:** normalize each ledger to the corrected batch-1 row shape: `method`, `path`, `parity_class`, `api_surface`, `source`, `state`, `foundation`, `rejection`, and `declaration`. An absent connector-local command or contract is `declaration-pending`, not `foundation-gap`. Preserve real engine gaps only with concrete refusal file/line and smallest safe change.
4. Commit batch 2 after its local map integrity and connector validation gates are green.

## Batch 3 — Red / Green / Refactor

1. **Red:** the ten target bundles likewise lack complete connector-local source locks and corrected six-class disposition ledgers.
2. **Green:** repeat the public-source provenance and exact source/API-surface crosswalk for `gong`, `google-ads`, `facebook-marketing`, `linkedin-ads`, `aircall`, `xero`, `paypal-transaction`, `gocardless`, `amazon-seller-partner`, and `miro`; each provider DELETE is declared or explicitly disabled.
3. **Refactor:** apply the same vocabulary and transport assessment as batch 2. PR #4286 makes a declaration-owned ETL source possible, so an absent source `sync_transport.json` is connector-local `declaration-pending`. A typed write action remains an enabled `direct_write`; its nested and action-level reverse-ETL eligibility records distinguish semantic eligibility from the pending generic App/CLI dispatch and action-selection multiplicity foundations. No `transport_binding` action, arbitrary action selector, or destination descriptor is invented.
4. Commit batch 3 after its local integrity and connector validation gates are green.

## Safety and Classification Rules

- Public API descriptions only; no credential, token, live provider request, provider write, or certification execution.
- `requires-elevated-scope` is enabled with source-backed runtime scope metadata, never a disabled reason.
- `unsafe-to-exercise` applies only to genuinely destructive/irreversible actions outside user intent. Documented deletes are required map rows.
- A `foundation-gap` requires engine refusal file/line plus minimal change. A missing command, operation contract, CLI surface, or transport declaration is `declaration-pending`.
- Endpoint classes are mutually exclusive: direct read, direct write (including delete), ETL, binary read, or binary write. `reverse_etl` is a connector-owned destination declaration and command surface, never an endpoint class. Until the updated #4304 foundation reaches this branch and the persisted App dispatch selects the generic destination, it is accurately reported as foundation-pending rather than deployable. ETL uses the merged source declaration contract and remains declaration-pending until connector-local source evidence exists.

## Planned Gates and Checkpoints

1. Source/artifact capture and local source-lock/disposition integrity assertion for batch 2.
2. Batch-2 `connectorgen validate` and `surface-sync --check`, then commit and push checkpoint.
3. Repeat capture/integrity assertion and validation for batch 3, then commit and push checkpoint.
4. Run targeted fixture/conformance and commandrunner preflight tests, generated/check gates, detached `connector-boundary`, `verify-work`, and data-focused code review.
5. Rebase on `main`, push without force, open a Conventional Commit PR using `Refs #4289`, and read back its API-reported base.

## Source-Lock Completeness Remediation — 2026-08-19

Captain review found a fleet-wide defect: a source lock without `counts.total`, or a landing-page inventory that divides its own count by itself, cannot prove coverage. This phase replaces `declared_percent` with `operations_found` and `coverage_confidence`/basis everywhere, and makes `counts.total` plus per-kind/method counts mandatory. Every owned bundle is now re-derived from its public provider surface rather than bounded by the prior `api_surface.json`: machine documents where available, complete rendered references for Aircall and LinkedIn, the GoCardless public OpenAPI payload, all Amazon models, and all Facebook Business SDK code-generation declarations. Facebook's Graph paths are explicitly recorded as finite owner-type/method/node-edge declarations because the runtime object identifier is instance-dependent. `SOURCE-LOCK-VERIFICATION.json` records the old/new API-surface count and source basis for every connector; all nineteen locks are `complete`.

## Held-PR PayPal and Root-Count Repair — 2026-08-19

**Red:** the initial `paypal-transaction` pin was the two-operation Transaction Search document, not the complete official PayPal REST specification corpus; batch-3 locks also omitted the root `counts.total` that makes a captured denominator auditable. **Green:** the generator now pins the official `paypal-rest-api-specifications` archive and parses all thirteen `openapi/*.json` documents, yielding 115 exact source operations and eight documented deletes. Its generated lock and every batch-3 lock carry the Batch-2-shaped root counts object (`rest`, `graphql_query`, `graphql_mutation`, `total`), while retaining detailed REST method counts. The integrity verifier rejects a missing root total as well as a malformed nested REST count. This repair changes no transport binding or reverse-ETL declaration.
