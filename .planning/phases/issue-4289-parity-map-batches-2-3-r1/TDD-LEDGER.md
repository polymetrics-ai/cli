# Issue #4289 — TDD Ledger

## Red

On `main`, all nineteen selected bundles lack their connector-local source lock and corrected six-class declaration-disposition ledger. The observable integrity assertion is therefore impossible: a source denominator cannot equal a nonexistent inventory or disposition map.

## Green

For each selected bundle, a source lock pins a credential-free public provider description and the disposition ledger contains exactly one row per documented operation. The local integrity check verifies:

- source inventory count equals `ledger_dispositions` count;
- every row has all required corrected batch-1 fields;
- every source operation has exactly one method/path API-surface binding;
- parity-class totals equal the pinned source denominator;
- every source lock has `counts.total`, per-kind/method counts, and non-self-referential `operations_found` with a coverage-confidence basis; a partial inventory is visible as a hold, not reported as 100% declared;
- `foundation-gap` records include a concrete engine file/line and minimal change; enabled typed `direct_write` rows carry reverse-ETL eligibility metadata using the actual persisted App/CLI dispatch refusal at `internal/app/transport_dispatch.go:53-67`, while the action ledger separately records the one-action-per-mode selection limit at `internal/connectors/sync_transport.go:388-415`;
- unauthored connector work is `declaration-pending`, and elevated scopes do not disable rows.

`connectorgen validate` and `surface-sync --check` remain the production structural green gates. No behavior changes are made, so no Go unit-test red phase is appropriate; the map-integrity assertion is the testable artifact behavior added by this issue.

## Refactor

Keep the generated ledgers connector-local and use the exact corrected batch-1 schema. Do not alter engine code, infer schemas, fabricate transport descriptors, request a credential, or produce a terminal-command/certification claim. Never promote a rendered index or a self-referential count to complete source coverage.

## Held-PR Repair — Red / Green

**Red:** the PayPal Transaction Search OpenAPI document exposed only two reporting routes, so a lock pinned solely to that file could not represent the provider's complete documented REST surface. The ten batch-3 locks also did not expose a root `counts.total`, preventing the fleet-wide source-accounting check the captain required.

**Green:** `generate-parity-maps.mjs paypal-transaction` now consumes all thirteen official `openapi/*.json` files from PayPal's published specification archive and maps 115 exact operation declarations. `verify-parity-maps.mjs` fails unless both root and REST total counts equal the immutable source-operation inventory. The fresh green run reports `verified 19 connectors / 5127 documented operations`.

## Reconciliation Relaunch — Red / Green / Refactor

**Red:** the source-ledger implementation proves only source accounting and the six original parity classifications. It does not prove that every faithfully representable documented operation has an exact typed direct action, a connector-owned typed-destination binding, a source transport declaration, or an installed-binary command artifact. The new seven-surface ledger assertion must therefore fail against the preserved source-map-only state.

`node .planning/phases/issue-4289-parity-map-batches-2-3-r1/traces/verify-seven-surfaces.mjs` → **Red**: exits 1 and writes `SEVEN-SURFACE-LEDGER.json`; for example, Grafana reports 139 missing direct reads, 171 missing direct writes, and four missing ETL bindings. The test derives this from each source-locked ledger, bundle operations, CLI commands, and transport declaration, rather than counting source rows alone.

**Green:** after merging `fm/cli-reverse-etl-destination-r1`, each connector-local definition carries only source-backed executable contracts: exact typed direct-read/write actions, distinct binary contracts where provider evidence represents a transfer rather than REST, ETL source declarations, eligible typed reverse-ETL destinations, and generated operation/CLI surfaces. The seven-surface ledger reports every connector and asserts no documented operation was silently omitted, disabled for privilege/destructiveness, or moved into a generic writer.

**Refactor:** retain the immutable source locks and existing disposition rows as inputs; normalize action and transport metadata mechanically through connector-local generation. Keep REST direct-write, binary transfer, and reverse-ETL destination contracts distinct. Unsupported remains a precise engine incapability with source and refusal evidence, never a proxy for missing authoring.

**Foundation hold:** the initial #4304 commit composes generic typed destinations but does not yet select them in the App/CLI persisted-dispatch path. Declarations can be structurally green, but an installed App/CLI reverse-ETL run is red until the updated foundation branch is merged and that path is exercised. No connector-local substitute is permitted.

## Typed Write Eligibility — Red / Green

**Red:** a source-row-only eligibility assertion leaves five existing typed actions without an exact action disposition because their legacy paths are base-relative or their published source inventory has no exact row. It also makes a one-action destination look like a connector-complete reverse-ETL declaration.

**Green:** `verify-parity-maps.mjs` requires every `writes.json` action to have exactly one action-level eligibility disposition. The 621 target actions are semantically eligible and individually representable as `declarative_api/declarative_typed_destination` action shapes; none is excluded for safety, privilege, or destructive behavior. The disposition records provider-operation provenance separately from the still-pending exact stream-to-required-input mapping and names `declarative-typed-destination-action-multiplicity` when the current one-action-per-mode descriptor cannot select all eligible actions. This is a foundation/declaration hold, never a connector-local selector or a completion claim.

## Installed Typed-Write Command Coverage — Red

**Red:** source and action evidence alone can report a reverse-ETL action as eligible while no installed CLI command selects it. The seven-surface assertion derives `implemented` and technically `partial` `reverse_etl` command-to-`write` coverage from every bundle's `cli_surface.json` and fails for every missing, duplicate, or orphan action. The connector-local generator creates 452 source-bound, fully representable action commands using the same closed record-schema projection as the batch materializer: scalar flags or a top-level object/array `json` flag, never a raw request body. Eighty-two source-bound actions have required scalar union values (Twilio string/null, Xero string/null, or GoCardless string/integer) that current closed flags cannot faithfully encode. They receive an explicit `partial` command with the exact technical reason; the ledger reports partial commands separately from executable command reachability. Five actions still lack exact provider-operation provenance and remain `declaration-pending`, not guessed CLI routes.

## REST Body Input — Red / Foundation Routed

Provider-backed nested REST bodies cannot be faithfully exposed from the current operation CLI surface: `commandrunner` accepts structured JSON only for fixed GraphQL operations and refuses an exact direct-write `body` mapping. Firstmate resolved `[key=rest-structured-body-cli]` by routing the bounded declaration-owned capability to a separate engine PR. No raw-body, unvalidated flag, or connector-local shared-engine workaround is allowed while this lane retains the exact action/source-binding evidence and waits to compose that foundation.

## Captain Foundation-Gap Deliverable — Red / Green

**Red:** source coverage, connector-level fixtures, and a structural command pass can conceal a missing shared engine or generator capability. A source operation with an open foundation gap was previously still capable of being counted as enabled or of disappearing into a generic disabled disposition.

**Green:** `verify-seven-surfaces.mjs` emits `FOUNDATION-GAP-LEDGER.json` alongside the seven-surface and per-operation evidence ledgers. Stable shared IDs fan out to every exact source-locked operation, preserving source URL, published-or-explicitly-absent document revision, SHA-256, affected surface, failure evidence, owner lane, status, and closure commands. The verifier emits batch-2, batch-3, and portfolio rollups; it fails while any gap is open. It records generated website rows and source-operation fixture/conformance bindings as routing-required common gaps rather than falsely crediting generic documentation or connector-wide tests. An unauthored connector contract is still `declaration-pending`, not a foundation gap.

## Source-lock v3 Migration — Red / Green / Terminal Causes (2026-08-24)

**Red:** after the multi-document and rendered-reference foundations landed, `GOFLAGS='-p=3' go run ./cmd/connectorgen validate internal/connectors/defs --json` rejected all nineteen batch locks before descriptor import because the legacy v2 wire shape retains `rest.format` or `rest.operation_counts` fields that the strict v3 decoder deliberately has no slot for.

**Green (structural):** a hash-manifested local copy of all nineteen pre-migration locks was preserved before edits. The authorized mechanical conversion uses v3 document kinds, retains every operation inventory and coverage basis, folds `inventory_basis` and Facebook's `graph-node-edge` evidence into that basis, preserves n8n's 63 YAML path fragments as independently hashed rendered-reference documents, and records Swagger 2.0 explicitly for Slack. The same validator then accepts fifteen source-lock shapes and exposes the four terminal representation defects: Amazon SQS's native AWS Query action paths, Google Ads duplicate route aliases, Miro's malformed trailing-`?` path, and Trello's duplicate source identity.

**Terminal import evidence:** source import then exercised each structurally accepted lock using only its pinned public artifact and a local content-addressed cache. The importer reached provider grammar bounds for Grafana (unbounded numeric request schema), Slack (non-object response schema), and Twilio (unbounded string request schema); it reported immutable-artifact drift for Amazon Seller Partner, Elasticsearch, GoCardless, Gong, Google Calendar, LinkedIn Ads, and n8n; and it generated descriptors that revealed missing executable actions for Aircall, Facebook Marketing, and Xero. PayPal Transaction imported successfully. Elasticsearch's retry reproduced its locked-artifact mismatch, so the worker stopped under the repeated-obstacle rule before refreshing any provider lock or inferring any replacement schema. No connector-local workaround was used.

## Captain Reconciliation — Scalar-Union Command Slice (2026-08-24)

**Red:** 82 source-cited reverse-ETL commands are present but `partial` solely
because the old batch materializer rejects a record field whose closed provider
schema has more than one scalar type. The current command therefore does not
reach the missing-credential boundary.

**Investigation result / Green boundary:**
`ValidateStructuredJSONRecordField` accepts the GoCardless closed
`string|integer` field unions, so its 45 commands are eligible for
connector-local JSON flag declarations and installed-binary missing-credential
proof. It rejects all Twilio (34) and Xero (3) `string|null` fields because it
correctly excludes `null` and requires two remaining non-null kinds (or an
object/array) at `internal/connectors/engine/record_schema_promotion.go:165-176`.
The actual defect is earlier and provider-neutral: the generated materializer
rejects all multi-entry type lists at
`cmd/connectorgen/batch_materialize.go:1568-1570`, although
`cmd/connectorgen/sourceprojection.go:2063-2088` already normalizes nullable
scalars to ordinary scalar flags and
`cmd/connectorgen/validate.go:2112-2129` validates them. Those 37 commands are
genuinely blocked pending the shared materializer correction; no generic body,
URL, operation, or connector special case is permitted here. Any green command
must still return exactly `error: missing --credential` from the built binary.

**GoCardless Green:** all 45 exact `string|integer` actions were promoted with
one declaration-owned `json` flag per required record field. The fresh-binary
credential-free probe records 45/45 exact `error: missing --credential`
responses in `evidence/gocardless-reverse-etl-command-proof-20260824.json`.
`go test -timeout 20m ./internal/connectors/commandrunner -run
'^TestEveryImplementedCommandPassesRuntimePreflight$' -count=1` passes.

## PayPal Zero-Input Direct Reads — Red

**Red:** PayPal Transaction's retained `bundle` descriptor proves 115 provider
operations but its four zero-input JSON GET routes have no `operations.json`
contract or installed command. The baseline validator reports 77 distinct
missing mutation actions; this is pre-existing write coverage and is kept
separate from the four read routes.

**Planned Green:** declare only the four archived OpenAPI GET operations whose
path-item and operation parameter lists are both empty:
`/v2/invoicing/accounting-sync/merchant/connections`,
`/v1/notifications/webhooks-lookup`,
`/v1/notifications/webhooks-event-types`, and
`/v1/payment-experience/web-profiles`. Each uses the direct-read executor's
standard 1 MiB response cap, `json_redacted` output policy, no request schema,
and a credential-free installed-binary probe that must stop exactly at
`error: missing --credential`.

**Green:** all four operations have exact `rest_read` contracts, source-bound
API-surface coverage, and installed commands. The regenerated descriptor clears
their former generic execution gaps; its validator has **zero** implemented
source-operation gaps and retains only the independent 77 missing mutation
actions. `TestEveryImplementedCommandPassesRuntimePreflight` passes, and a
fresh binary run from `pm init --root <isolated-project> --json` records 4/4
exact missing-credential results in
`evidence/paypal-zero-input-direct-read-command-proof-20260824.json`.

## Grafana Access-Control Status Direct Read — Red

**Red:** the hash-pinned Grafana OpenAPI artifact declares parameterless JSON
`getAccessControlStatus` at `GET /access-control/status`, while the corresponding
`/api/access-control/status` API-surface and source-disposition entries remain
`declaration-pending`. There is no Grafana `operations.json`, no `cli_surface`
command, and no generated operation-endpoint-ledger entry, so no installed
binary command can reach credential preflight.

**Green target:** declare only the exact bounded `rest_read` operation and its
`access control status get` command, bind both to the existing source row, and
regenerate only the connector/global definition artifacts the repository owns.
`TestEveryImplementedCommandPassesRuntimePreflight` must include the command,
and a built `pm` invoked from a new credential-free project must emit exactly
`error: missing --credential`; no provider request is made.

**Green:** `grafana.access_control.status.get` now binds the source-pinned GET
route at a 1 MiB `json_redacted` cap, `TestEveryImplementedCommandPassesRuntimePreflight`
passes, and the built binary exits 1 with no stdout and exactly `error: missing
--credential` on stderr. The evidence is
`evidence/grafana-access-control-status-direct-read-proof-20260825.json`; its
project was initialized without a credential and made no provider request.

## Source-Lock v3 Trace Reconciliation — Red / Green Boundary (2026-08-25)

**Red:** `verify-parity-maps.mjs` read only the legacy top-level
`*.operations` arrays, so every schema-version-3 lock reported an empty source
inventory; its legacy lock summary also emitted null source URL, byte, and hash
fields. `verify-seven-surfaces.mjs` found nested artifacts but could not attach
a `bundle` artifact to its source operation, incorrectly dropping the pinned
hash for every bundled Amazon Seller Partner, Facebook Marketing, and PayPal
source row.

**Green boundary:** both traces now associate each operation with its v3
`rest.source_documents` entry, preserve every document's immutable artifact
metadata in the source-lock report, and use that entry's artifact hash for the
operation trace. The three n8n workflow rows whose retained artifacts use
immutable commit URLs now cite those exact URLs rather than mutable `master`.
`node traces/verify-seven-surfaces.mjs` records 5,127 pinned source traces and
zero missing source hashes; it remains intentionally red for the still-unmapped
provider surfaces. `node traces/verify-parity-maps.mjs` now reaches the real
Google Ads action-disposition delta rather than failing at the obsolete v2
reader.

## Grafana Parameterless JSON Direct Reads — Red

**Red:** the pinned Grafana OpenAPI document declares `getHealth`,
`RouteGetMuteTimings`, `RouteGetPolicyTree`, and `RouteGetTemplates` as
parameterless JSON GET operations, but all four source rows remain
`declaration-pending`. They have no exact `rest_read` operation contract or
installed command, so none can reach credential preflight. The ETL-classified
`RouteGetAlertRules` operation is explicitly outside this direct-read slice.

**Green target:** add only the four source-bound 1 MiB `json_redacted`
direct-read contracts and commands, then prove each with the real binary from
its own initialized credential-free project. No request, response, pagination,
or body schema is inferred or added.

**Green:** all four contracts bind their exact source operation and canonical
API-surface endpoint. `TestEveryImplementedCommandPassesRuntimePreflight` and
`TestConformance/grafana` pass; `pm docs validate --connectors-dir
docs/connectors` passes after generated connector documentation. A freshly
built binary run from four separately initialized no-credential projects exits
1 with no stdout and exactly `error: missing --credential` for every command.
Evidence: `evidence/grafana-parameterless-direct-read-command-proof-20260825.json`.

## Elasticsearch Parameterless JSON Direct Reads — Red

**Red:** the source-pinned Elasticsearch OpenAPI artifact declares `info`,
`cluster-remote-info`, `ilm-get-status`, `license-get-basic-status`, and
`license-get-trial-status` as JSON `GET` operations with no path-item or
operation parameters. The five exact source rows are declaration-pending, and
Elasticsearch has neither an `operations.json` contract nor installed command
for any of them. The root route's previous `disallowed` classification was a
legacy sync-stream exclusion, not a runtime incapability.

**Planned Green:** declare only the five exact 1 MiB `json_redacted`
`rest_read` contracts and source-bound commands. Preserve the provider-stated
cluster privileges as operation `auth_scopes` metadata (`monitor` for all but
the ILM status's `read_ilm`) without withholding any surface. A fresh built
binary in a separately initialized project with no credential must emit exactly
`error: missing --credential` for each command; no provider request is made.

**Green:** the five exact contracts now bind their source rows and canonical
API-surface routes. `TestEveryImplementedCommandPassesRuntimePreflight` and
`TestConformance/elasticsearch` pass, as do generated connector documentation
validation and namespace/command help checks. A fresh binary in five separately
initialized no-credential projects exits 1 with no stdout and exactly
`error: missing --credential` for every command. Evidence:
`evidence/elasticsearch-parameterless-direct-read-command-proof-20260825.json`.

**Residual importer boundary:** `connectorgen source-import elasticsearch
--defs internal/connectors/defs --check` verifies the retained v3 artifact then
stops at `paths["/_alias"].get` because its first parameter has an ambiguous
request schema using `allOf`. The refusal is emitted by
`sourceReferenceResolver.preflightDocument` through
`sourcePrepareSourceDocument` in `cmd/connectorgen/sourceimport.go:3376-3377`.
It is independent of the five input-free reads above; no descriptor or schema
was hand-authored to bypass it.

## Elasticsearch Administrative Metadata Direct Reads — Red

**Red:** the pinned Elasticsearch OpenAPI artifact declares `cat-help`,
`dangling-indices-list-dangling-indices`, `esql-list-queries`,
`get-script-context`, and `get-script-languages` as parameterless JSON GET
operations. Their source/API-surface rows are declaration-pending, so no
connector-owned typed contracts or installed commands can reach credential
preflight. Provider-required privileges (`manage` and `monitor_esql`) are
metadata, not a technical reason to keep these routes disabled.

**Planned Green:** declare only the five exact source-bound 1 MiB
`json_redacted` rest reads and preserve each stated cluster privilege in
`auth_scopes`. A fresh built binary from five separate initialized projects
with no credential must emit exactly `error: missing --credential` for every
command, with no provider request.

**Green:** all five contracts bind their exact source/API routes, including
`manage` and `monitor_esql` as runtime metadata. The commandrunner preflight,
Elasticsearch conformance, generated docs validation, and namespace/command
help checks pass. A fresh binary from five separately initialized no-credential
projects exits 1 with no stdout and exactly `error: missing --credential` for
every command. Evidence:
`evidence/elasticsearch-administrative-direct-read-command-proof-20260825.json`.

## Elasticsearch Lifecycle and Platform Metadata Direct Reads — Red

**Red:** `indices-get-data-lifecycle-stats`, `inference-get-region-policy`,
`ingest-geo-ip-stats`, `ingest-processor-grok`, and
`migration-get-feature-upgrade-status` are parameterless JSON GET operations in
the pinned Elasticsearch source, but their source/API-surface entries remain
declaration-pending. No installed command can prove dispatch. The provider's
`monitor`, `monitor_inference`, and `manage` requirements must be operation
metadata, never a disabling disposition.

**Planned Green:** add only the five exact 1 MiB `json_redacted` rest-read
contracts and source-bound commands. The built binary must stop at exactly
`error: missing --credential` in a fresh isolated project for every command;
no request, body, pagination, or provider call is invented.

**Green:** the five source-bound contracts now retain all provider-stated
cluster privileges as `auth_scopes` and reach credential preflight as installed
commands. Commandrunner preflight, Elasticsearch conformance, generated docs
validation, and help checks pass. A fresh binary in five isolated projects with
no credential exits 1 with zero stdout and exactly `error: missing
--credential` for every command. Evidence:
`evidence/elasticsearch-platform-direct-read-command-proof-20260825.json`.

## Elasticsearch Security and ML Observation Reads — Red

**Red:** five parameterless JSON GETs in the pinned source—`ml-info`,
`security-authenticate`, `security-get-builtin-privileges`,
`security-get-stats`, and `security-get-user-privileges`—are
declaration-pending. The absence is connector-local authoring, not a privilege
or safety incapability. The slice explicitly excludes `/_security/enroll/*`,
whose token-generation semantics require their own action classification.

**Planned Green:** expose only the five bounded redacted direct reads, keeping
the provider-stated `monitor_ml`, `manage_security`, and `monitor` scopes in
`auth_scopes`, then prove every command reaches exactly the no-credential
boundary from a new project without making a provider request.

**Green:** the five direct reads now bind exact source/API routes; applicable
privileges remain runtime metadata and the user-information output remains
`json_redacted`. Commandrunner preflight, Elasticsearch conformance, generated
docs validation, and help checks pass. A fresh no-credential binary run from
five separate projects exits 1 with zero stdout and exactly `error: missing
--credential` for each command. Evidence:
`evidence/elasticsearch-security-observation-direct-read-command-proof-20260825.json`.
