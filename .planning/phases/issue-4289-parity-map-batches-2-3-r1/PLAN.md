## Task Delivery Header

- Issue: Refs #4289 — chore(connectors): map parity batches 2 and 3
- Base branch: fm/cli-reverse-etl-destination-r1
- Merges into: fm/cli-reverse-etl-destination-r1 → main
- Delivery: PR #4300 remains open from `fm/cli-map-batch23-r1` after a non-rewriting merge of the temporary typed-destination foundation; its GitHub API base is `fm/cli-reverse-etl-destination-r1`. Historical local structural gates are green, while the captain pre-merge evidence gate is explicitly red until the recorded common foundations and generated evidence projections close.
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

## Captain Pre-Merge Foundation Evidence Gate — 2026-08-20

`FOUNDATION-GAP-LEDGER.json` is the machine-readable missing-foundation deliverable. It deduplicates each shared provider-neutral capability by stable id and fans it back out to exact source-locked provider operations with URL, document revision state, SHA-256, affected surface, runtime/validator refusal, owner lane, and closure commands. `OPERATION-EVIDENCE-LEDGER.json` adds the same operation-level merge-readiness state, so an open foundation gap is never credited as enabled or merge-ready. The current portfolio is deliberately red: seven gaps affect all 5,127 operations through required website-row and per-operation fixture/conformance projections, with further typed-destination, structured-body, bounded-binary-transfer, and scalar-union fan-out recorded separately. Batch-2 and batch-3 rollups are generated; an N/A surface is emitted only from the source-locked canonical parity class, never from risk, scope, tier, or destructive metadata.

## Batch 2 — Red / Green / Refactor

1. **Red:** the nine target bundles lack `sources/<connector>-operation-source-lock.json` and `sources/<connector>-declaration-disposition.json`; therefore no provenance or six-class complete-map assertion can pass.
2. **Green:** credential-free public descriptions are pinned, parsed into method/path inventories, and mechanically reconciled to the existing `api_surface.json`. Every mismatch remains an explicit disabled disposition; no request, response, pagination, or body schema is inferred.
3. **Refactor:** normalize each ledger to the corrected batch-1 row shape: `method`, `path`, `parity_class`, `api_surface`, `source`, `state`, `foundation`, `rejection`, and `declaration`. An absent connector-local command or contract is `declaration-pending`, not `foundation-gap`. Preserve real engine gaps only with concrete refusal file/line and smallest safe change.
4. Commit batch 2 after its local map integrity and connector validation gates are green.

## Batch 3 — Red / Green / Refactor

1. **Red:** the ten target bundles likewise lack complete connector-local source locks and corrected six-class disposition ledgers.
2. **Green:** repeat the public-source provenance and exact source/API-surface crosswalk for `gong`, `google-ads`, `facebook-marketing`, `linkedin-ads`, `aircall`, `xero`, `paypal-transaction`, `gocardless`, `amazon-seller-partner`, and `miro`; each provider DELETE is declared or explicitly disabled.
3. **Refactor:** apply the same vocabulary and transport assessment as batch 2. PR #4286 makes a declaration-owned ETL source possible, so an absent source `sync_transport.json` is connector-local `declaration-pending`. A typed write action can remain semantically eligible, but an operation with an open foundation gap is not enabled for that affected surface and cannot contribute to a merge-ready verdict. No `transport_binding` action, arbitrary action selector, or destination descriptor is invented.
4. Commit batch 3 after its local integrity and connector validation gates are green.

## Safety and Classification Rules

- Public API descriptions only; no credential, token, live provider request, provider write, or certification execution.
- `requires-elevated-scope` is enabled with source-backed runtime scope metadata, never a disabled reason.
- `unsafe-to-exercise` applies only to genuinely destructive/irreversible actions outside user intent. Documented deletes are required map rows.
- A `foundation-gap` requires engine refusal file/line plus minimal change. A missing command, operation contract, CLI surface, or transport declaration is `declaration-pending`.
- Endpoint classes are mutually exclusive: direct read, direct write (including delete), ETL, binary read, or binary write. `reverse_etl` is a connector-owned destination declaration and command surface, never an endpoint class. Until the updated #4304 foundation reaches this branch and the persisted App dispatch selects the generic destination, it is accurately reported as foundation-pending rather than deployable. ETL uses the merged source declaration contract and remains declaration-pending until connector-local source evidence exists.

## Captain Reconciliation — 2026-08-24

The current branch's machine-measured source inventory contains 5,127 locked
provider operations: 592 have an installed `implemented` command and are
already runnable to the missing-credential boundary; 4,535 are declarable now
(4,453 unauthored rows plus 82 source-cited scalar-union commands); and zero
currently have an evidenced missing engine capability. This is branch-local
accounting: the current branch has 2,644 installed reverse-ETL commands across
36 connectors, while merge base `060bb7864e` independently measures 2,191
across 27. The different totals reflect concurrent declaration work, not a
missing reverse-ETL foundation.

The first conversion investigation covers the 82 existing source-cited
`partial` reverse-ETL commands in Twilio (34), Xero (3), and GoCardless (45).
The GoCardless fields are genuine `string|integer` unions and pass the existing
declaration-owned JSON-field gate, so they remain a connector-local conversion
candidate. The Twilio and Xero fields are each only `string|null`: their public
sources are respectively
`https://raw.githubusercontent.com/twilio/twilio-oai/main/spec/json/twilio_api_v2010.json`
and `https://raw.githubusercontent.com/XeroAPI/Xero-OpenAPI/master/xero_accounting.yaml`.
The captain's diagnosis is confirmed: this is a shared materializer misrouting
defect, not a missing structured-JSON capability. The materializer rejects a
type list whose length is not one at
`cmd/connectorgen/batch_materialize.go:1568-1570`; meanwhile
`sourceProjectionFlagType` removes `null` and selects a normal scalar at
`cmd/connectorgen/sourceprojection.go:2063-2088`, and static command validation
admits a string when it is a declared schema arm at
`cmd/connectorgen/validate.go:2112-2129`. The structured JSON validator
correctly drops `null` and rejects the remaining scalar at
`internal/connectors/engine/record_schema_promotion.go:165-176`; it is reached
only because the materializer sends nullable scalars there. The 37 Twilio/Xero
operations are therefore in the genuinely-blocked column pending a shared,
provider-neutral materializer correction. This lane neither changes shared code
nor adds a connector special case. Red proof is the current partial command
block; the GoCardless green proof remains the same installed command reaching
exactly `error: missing --credential` after connector-local flags and
`availability: implemented` are declared. Every later operation remains
connector-local. A provider body that proves unbounded or user-defined is moved
to the genuinely-blocked ledger with its exact source citation and runtime
refusal; it is never guessed or routed to a shared workaround.

### Executed GoCardless Slice

All 45 GoCardless `string|integer` partial commands now have one exact named
`json` flag mapped to their required `record.<field>`, with the declared string
arm accepted as ordinary CLI text. A fresh built `pm` binary invoked every one
from a credential-free initialized project; all 45 returned exactly
`error: missing --credential`. The durable command transcript is
`evidence/gocardless-reverse-etl-command-proof-20260824.json`, and
`TestEveryImplementedCommandPassesRuntimePreflight` passes. Progress is
reported as **45 converted-and-proven**, **4,453 still declarable**, and **37
genuinely blocked** (Twilio 34, Xero 3); the latter does not include GoCardless.
Captain review also confirms that making `create a bank account holder
verification apply` partial corrects an unshipped branch-local over-claim: the
GoCardless CLI surface does not exist on `origin/main`, and this branch first
introduced the command as implemented. This is therefore a correction to the
truth, not a downgrade of a shipped credential-bound command.

### GoCardless Validation Attribution

The current GoCardless validator finding is
`sources/gocardless-operation-descriptor.json: canonical source descriptor is
missing`. It is **not** caused by the 45-command conversion: replacing the
current `cli_surface.json` with its pre-conversion `HEAD` content against the
same v3 source lock produces the identical finding. The branch merge base
`db289265354bcc7370bcc79a572f022a5668571c` validates clean because it retains
the legacy v2 lock, which did not require a canonical descriptor; it is not
evidence that this v3-source-lock finding was introduced by the conversion.

The authorized v3 mapping correction classified GoCardless's retained public
`openapi-schema-public.json` artifact as `openapi` version `3.0.0` and added it
to the aggregate OpenAPI inventory. Offline source import then reached the
provider's exact terminal grammar limit at
`/bank_account_holder_verifications` `POST`: its `application/json` request
schema has dynamic unbounded `additionalProperties`. No schema was invented,
no engine code was changed, and the descriptor remains absent until that
operation's precise bounded declaration disposition is resolved.

### Unbounded Dynamic-Body Pattern

GoCardless `POST /bank_account_holder_verifications` is a genuinely blocked
source operation under `declaration-bound-unbounded-dynamic-body`, not a
partial or invented typed-body declaration. Its public OpenAPI citation is
`paths["/bank_account_holder_verifications"].post` in
`https://developer.gocardless.com/openapi-schema-public.json`; offline import
refuses its unbounded dynamic `additionalProperties` request object and the
closed-body runtime refuses the same shape at
`internal/connectors/engine/structured_rest_body.go:1442-1443`. The existing
open-schema action's CLI command is explicitly `partial`, and the source row
carries that exact foundation evidence. This is the captain-owned common
pattern also observed for Asana, Zoom, and Docker Hub SCIM writes. Reuse this
named pattern and its citation discipline for later occurrences; do not invent
a schema, a tag workaround, or a raw/open request body.

### PayPal Zero-Input Direct-Read Slice

The retained PayPal OpenAPI bundle has four JSON GET operations with no
path-item or operation parameters. They are a separate connector-local read
slice: each can be declared without inventing a request schema, using the
direct-read executor's 1 MiB default response cap and `json_redacted` output.
The existing 80 PayPal source-projection findings are missing mutation actions
and must remain independently visible; they do not justify withholding these
faithfully representable read routes.

The slice is now **4 converted-and-proven**: every installed command passed
runtime preflight and reached exactly `error: missing --credential` from its
own initialized, credential-free project. Its source descriptor has no
remaining gap for those four source IDs; the separate PayPal mutation-action
hold remains explicit and unchanged.

### Grafana Access-Control Status Direct-Read Slice

The pinned Grafana OpenAPI source declaration records
`getAccessControlStatus`: a parameterless JSON `GET
/access-control/status` with a `200` response. The source-ledger row is already
bound to `/api/access-control/status` but remains `declaration-pending` solely
because Grafana has no connector-owned `rest_read` operation or executable CLI
command. This one-command slice adds only the bounded 1 MiB `json_redacted`
read contract and its exact source/API-surface/CLI bindings. It does not alter
the source lock, invent parameters or schemas, or change a write/ETL/binary
surface. Green proof is runtime preflight plus a fresh credential-free installed
binary result of exactly `error: missing --credential`.

### Grafana Parameterless JSON Direct-Read Slice

The same pinned Grafana OpenAPI artifact declares four additional exact,
parameterless JSON GET operations: `getHealth` at `/health`,
`RouteGetMuteTimings` at `/v1/provisioning/mute-timings`,
`RouteGetPolicyTree` at `/v1/provisioning/policies`, and `RouteGetTemplates` at
`/v1/provisioning/templates`. Their source rows and canonical API-surface
endpoints are present but declaration-pending. The connector-local slice adds
only bounded 1 MiB `json_redacted` direct-read contracts and source-bound
commands. It does not alter the stream-backed alert-rules ETL row, source
locks, schemas, pagination, writes, or shared engine code. Each installed
command must reach exactly `error: missing --credential` from a fresh
credential-free project.

### Elasticsearch Parameterless JSON Direct-Read Slice

The pinned Elasticsearch OpenAPI document declares five exact, parameterless
JSON `GET` operations: `info` at `/`, `cluster-remote-info` at `/_remote/info`,
`ilm-get-status` at `/_ilm/status`, `license-get-basic-status` at
`/_license/basic_status`, and `license-get-trial-status` at
`/_license/trial_status`. Their source-ledger/API-surface rows are
declaration-pending only because Elasticsearch has no connector-owned
`operations.json` or executable CLI surface. The connector-local slice adds
only bounded 1 MiB `json_redacted` direct-read contracts and source-bound
commands. It preserves the source document's required cluster privileges as
operation `auth_scopes` metadata (`monitor` or `read_ilm`); those scopes do not
disable the commands. It does not alter the source lock, invent request/body
schemas, alter stream pagination, or change the shared engine. Each installed
command must reach exactly `error: missing --credential` from a fresh,
credential-free project.

### Elasticsearch Administrative Metadata Direct-Read Slice

The same pinned Elasticsearch OpenAPI document declares five more exact,
parameterless JSON `GET` operations: `cat-help` at `/_cat`,
`dangling-indices-list-dangling-indices` at `/_dangling`,
`esql-list-queries` at `/_query/queries`, `get-script-context` at
`/_script_context`, and `get-script-languages` at `/_script_language`.
They require no request schema, body, path value, or pagination inference. The
provider declares `manage` for the dangling-index and script metadata reads and
`monitor_esql` for running-query metadata; those scopes are retained as
operation metadata and never suppress the command. The connector-local slice
adds only bounded 1 MiB `json_redacted` direct-read contracts, source/API
bindings, and commands proven by the credential-preflight boundary.

## Captain Delivery Discipline — 2026-08-24

The 4,535-operation declarable-now inventory is planning input, not delivered
surface. A conversion is credited only after its built `pm` binary invocation
reaches exactly `error: missing --credential` from an isolated project with no
credential configured. Each reviewable connector-local batch records
`converted-and-proven` separately from `still-declarable`; it must never report
their sum as executable reachability. If an action cannot be expressed from its
exact provider schema, record its source citation and precise runtime refusal
immediately, move it from declarable to engine-blocked, and stop rather than
adding shared code or a connector-specific engine branch.

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
