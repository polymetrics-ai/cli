# Verification — issue #4291

## Superseded verification record

The earlier complete-map validation is superseded by the 2026-08-19 source-lock completeness defect. PR #4296 is held. Its former `source_operation_count` / `declared_percent` figures were based on the legacy `api_surface.json` denominator and must not be used as provider-coverage evidence.

## Artifact-level red/green evidence

- **RED — batch 6:** `test ! -f` for every proposed batch-6 source lock and disposition ledger passed before implementation.
- **RED — batch 7:** `test ! -f` for every proposed batch-7 source lock and disposition ledger passed before implementation.
- **GREEN — initial map only:** the issue-local strict ledger-invariant check passed against the then-current `api_surface.json` denominator. It correctly checked parity classes, reachability, and reverse-ETL semantics, but it did not prove provider-source completeness. Its former total (**2,099 documented, 575 enabled, 348 commands, 448 writes, 211 deletes**) is therefore an initial crosswalk count, not a complete provider-surface claim.
- **GREEN — Salesloft comprehensive remap:** complete rendered-reference crawl passed: 315 public API-reference pages yielded **211 unique operations** (120 GET, 50 POST, 23 PUT, 18 DELETE), replacing the prior 12-operation/84,498-byte index-derived inventory. `counts.total`, `operations_found`, per-method counts, and `coverage_confidence` are recorded in the source lock; the regenerated API surface and disposition ledger each contain all 211 operations and omit `declared_percent`. The explicit map invariant passed at **211/211/211** (source lock/API surface/disposition rows), with 91 `direct_write` rows carrying the separate generic reverse-ETL foundation-gap attribute.
- **GREEN — official-spec remaps:** Iterable **148/148/148** (source lock/API surface/disposition rows), Klaviyo **345/345/345**, and Intercom **231/231/231**. Each source lock has `counts.total`, per-method counts, `operations_found`, and `complete_machine_readable_specification` evidence. Their former 4/9/10-row boundaries are not retained.
- **GREEN — Outreach and Customer.io authoritative-spec remaps:** Outreach **259/259/259** combines its 253-operation public OpenAPI 3.0.3 source with the provider’s six documented fixed generic custom-object routes; the per-account object schema is explicitly dynamic. Customer.io changes from 159 to **166/166/166** from its public OpenAPI 3.1.0 source. Every enabled row is an exact typed-write binding: Outreach 163 and Customer.io 10; unbound ETL stream rows remain disabled.
- **GREEN — Gorgias and Chatwoot complete-spec audits:** Gorgias **114/114/114** from its official OpenAPI 3.1.0 API registry and Chatwoot **148/148/148** from its official OpenAPI 3.1.0 `swagger/swagger.json`. They are explicit count-unchanged results. The enabled denominator reconciles to concrete bindings only: Gorgias 103 and Chatwoot 94.
- **GREEN — Freshdesk complete reference remap:** **170/170/170** after parsing all 171 endpoint sections in the provider’s single 3.2MB rendered reference (query examples normalize to 170 method/path operations). This replaces the legacy 10-row API-surface boundary without treating a partial crawl as complete.
- **GREEN — Copper complete reference remap:** **89/89/89** after completing the provider-published 637-document rendered MkDocs corpus (32 GET, 35 POST, 11 PUT, 11 DELETE), replacing five synthetic `HOOK` rows. The five native stream bindings are the provider-documented `POST /v1/<resource>/search` operations proven by `internal/connectors/native/copper/streams.go:5-25`; no command/action is bound, so enabled remains zero.
- **GREEN — Close and Chargebee complete-spec remaps:** Close changes from 297 to **300/300/300** from its provider-published OpenAPI 3.1.0 (137 GET, 64 POST, 47 PUT, 52 DELETE), retaining only its 12 exact typed-write bindings as enabled. Chargebee changes from 428 to **527/527/527** using the official API v2 SDK OpenAPI 3.1.0—the public contract Chargebee uses to generate its client libraries—rather than the incomplete rendered landing page; its 36 exact typed-write bindings remain enabled.
- **RED/GREEN — Segment provider-shape migration:** Segment’s complete provider-published OpenAPI 3.0.3 contains **201** operations and `GET /` (`getWorkspace`), but no `GET /workspaces`. The initial source remap made `connectorgen validate` fail `surface_incomplete` for the legacy stream. Per the captain decision, the connector now declares a non-paginated singleton `workspace` stream at `GET /` with `data.workspace` / `single_object=true`; its check, schema, fixtures, and documentation assert that exact response. The legacy `workspaces` / `GET /workspaces` declaration is visibly `REMOVED` in both the source lock and disposition ledger with reason `not present in the provider's authoritative OpenAPI` and replacement `workspace` / `GET /`. `go run ./cmd/connectorgen validate` then passed `552 connector(s) checked, 0 findings`; `surface-sync --check`, `go test -timeout 20m ./internal/connectors/engine`, and `go test -timeout 20m ./cmd/connectorgen` also passed.
- **GREEN — post-#4297 declaration reconciliation:** rebased onto `origin/main` at `51dd6d468` and ran `connectorgen surface-reconcile --check --json` for all 20 owned connectors. It found deterministic reason updates only in Help Scout (5), Gorgias (2), and Chatwoot (23), then applied and rechecked them cleanly. It covered **zero** rows: none of those 20 currently has an implemented, endpoint-matching runnable command whose runtime preflight proves the newly available executor capability. The audit therefore retains no false enabled claim; future mapped operations with such a binding will be promoted by this same runtime-backed pass.
- **RED/GREEN — Square source recovery:** static extraction initially found seven zero-card pages after 40/40 fetches, so it cleared `operations_found` and discarded the preliminary 255-row map. Browser rendering recovered the missing cards, but the rendered crawl remains `coverage_confidence: partial` and is not used as a source lock. The provider-published OpenAPI 3.0.0 specification is the settled source instead: **334/334/334** (lock/API/disposition), 121 GET, 150 POST, 36 PUT, 27 DELETE.

## Repository checks

| Command | Result |
| --- | --- |
| `go run ./cmd/connectorgen validate` | pass — `552 connector(s) checked, 0 finding(s)` |
| `go run ./cmd/connectorgen surface-sync --check` | pass — `552 connector(s) scanned, 0 field(s) need synchronization` |
| `go test -timeout 20m ./cmd/connectorgen` | pass |
| `go test -timeout 20m ./internal/connectors/engine` | pass |
| `go test -timeout 20m ./internal/cli` | pass |
| `go vet ./...` | pass |
| `go build ./cmd/pm` | pass |
| `gofmt -w cmd internal && go mod tidy && git diff --exit-code -- go.mod go.sum` | pass — formatting/mod-tidy introduced no module drift |
| `./pm docs validate --connectors-dir docs/connectors` | pass |
| `make smoke-no-build` | pass |
| `make lint` | pass — `golangci-lint` reported `0 issues` |
| `make agent-contract-check` | pass |
| `make connectorgen-validate` | pass |
| `make connectorgen-surface-sync` | pass |
| `node --test scripts/tests/github-combined-operation-ledger.test.mjs scripts/tests/gen-github-graphql-parity.test.mjs scripts/tests/github-source-drift.test.mjs` | pass — 15 tests, 0 failures |
| `node scripts/gen-github-graphql-parity.mjs --check` | pass |
| `node scripts/github-combined-operation-ledger.mjs --check` | pass |
| `go run ./cmd/connectorgen certification-matrix --check` | pass |
| `go run ./cmd/connectorgen certification-candidates --connector github --check` | pass |
| `go run ./cmd/connectorgen certification-sweep --connector github --check` | pass |
| `go run ./cmd/connectorgen boundary . --json` (captured to a temporary log and polled to completion) | pass |
| `bash scripts/tests/connector-canon.sh` | pass |
| `./scripts/tests/pinned-build-dependencies.sh` | pass |
| `./scripts/tests/homebrew-release-notify.sh` | pass |
| `./scripts/tests/release-target-parity.sh` | pass |

`go test -timeout 20m ./...` and aggregate `make verify` were deliberately not run as single commands: the repository AGENTS instruction says agents under a per-command timeout must run changed packages plus `internal/cli` separately and execute `make verify`'s non-suite gates individually, because the full 550+ connector suite is routinely cut off and indistinguishable from a hang. The targeted package tests and each applicable gate above were run; CI retains the full suite.

## Rendered-reference citation-contract reconciliation — 2026-08-24

- **Foundation merge:** merged `origin/main` at `3c66c33b8` (`#4332`) without a stash. Main deletes
  the lane source-lock paths, so all twenty local locks were copied and byte-compared under this
  issue phase before the merge and restored afterward.
- **RED → GREEN, source identity:** `go run ./cmd/connectorgen operation-evidence . --check` first
  refused Iterable's duplicate provider identity `iterable.rest.delete.delete`. The lock has nine
  duplicate provider IDs across 29 routes. Route-qualified source IDs and matching disposition
  crosswalk values make the same check pass after regeneration: `5,457` rows, `5` rollups, and the
  fixed-100 reference passes.
- **V3 parser increment:** `node .planning/phases/issue-4291-connector-parity-batches-6-7-r1/tools/migrate-source-lock-v3.mjs close-com salesloft copper zoho-bigin klaviyo` converts the first
  five locks from schema v2 to the landed v3 document shape. Each individual `connectorgen validate`
  reaches `canonical source descriptor is missing`, proving the source-lock parser and rendered
  citation validation have passed for that connector.
- **Next terminal refusal:** `go run ./cmd/connectorgen source-import close-com --out
  internal/connectors/defs/close-com/sources/close-com-operation-descriptor.json` reaches the
  hash-pinned public OpenAPI grammar and exits `1`: `/activity/` GET parameter 0 is a documented
  number without a `minimum`. The refusing generic code is
  `validateBoundedRequestSchemaType` (`cmd/connectorgen/sourceimport.go:6989-6991`) through
  `sourceValidateNumericBounds` (`:7277-7281`). No descriptor was written, no credentials were
  used, and no provider operation was called.
- **Second terminal refusal:** the valid v3 trial causes `go run ./cmd/connectorgen operation-evidence
  . --check` to report `source lock for "close-com" has no operations and no provider-evidenced
  absence`. `readOperationEvidenceSourceLock` has a schema-v3 branch at
  `cmd/connectorgen/operationevidence.go:460-462`, but reads only the legacy `rest.operations`
  block at `:483-498` and never iterates `rest.source_documents[].operations`. The five trial locks
  were restored from the byte-compared preservation copy; the mapper remains for the shared reader
  decision.

The remaining Outreach source migration is intentionally not fabricated: six custom-object
operations cite `developers.outreach.io`, while the only retained immutable capture is at
`api.outreach.io`. The landed same-origin rendered-reference contract correctly requires a second
captured developer-document artifact (bytes, SHA-256, and capture provenance) before those six
operations can be moved to a document-owned v3 lock.

## Main foundation reconciliation and three-way baseline — 2026-08-24

- Merged `origin/main` `060bb7864` without a stash at `e06b27835`. The nineteen pending source-lock
  migrations and the committed Iterable lock were first copied under this phase and then restored
  byte-for-byte after the merge.
- The current 20-connector source-locked inventory has **3,932** operations. The current
  `operation-evidence` projection classifies **288** as runtime-enabled with a generated CLI path,
  **3,644** as connector-definition-declarable (`runtime.enabled: false` and no declared
  foundation), and **0** as genuinely execution-foundation-blocked. This is a pre-conversion
  classification; provider-live certification is pending, and the installed-binary no-credential
  sweep is still required for every newly declared command.
- The five Help Scout v3 route rows are no longer a foundation gap: the definition uses
  `route: mailbox_v3`, and `TestHelpScoutV3DirectReadsUseTheirDeclaredRoute` passed against the
  real direct-read executor (`/v3`, never `/v2/v3`). `validateOperationRoutes` and
  `resolveOperationRoute` in `internal/connectors/engine/operation_route.go` now own the
  fail-closed route contract.
- The Help Scout `update_customer` action-specific binding is also closed. Its
  `sync_transport.json` has a second action-owned `customers(id) -> update_customer(customerId)`
  binding, and the focused `DestinationSourceBinding` tests passed. The shared closure is
  `DestinationTransportDescriptor.SourceBindingForAction` in
  `internal/connectors/sync_transport.go`, which selects the exact declared action and refuses a
  legacy fallback when action-owned bindings exist.
- **Remaining shared evidence gate:** a reversible Close v3 probe produced a valid
  `rest.source_documents[0]` OpenAPI inventory with 300 operations, but
  `go run ./cmd/connectorgen operation-evidence . --check` failed
  `source lock for "close-com" has no operations and no provider-evidenced absence`. The current
  `readOperationEvidenceSourceLock` (`cmd/connectorgen/operationevidence.go`) accepts
  `schema_version == 3` but iterates only legacy `rest.operations`; it does not read
  `rest.source_documents[].operations`. Restoring the preserved v2 lock made the check pass again
  at 5,457 rows. This must be repaired in the shared foundation before valid v3 locks can land;
  no connector-local shim is permitted.
- **Close importer correction:** `source-retain close-com` verified and retained the pre-pinned
  1,340,508-byte OpenAPI document at SHA-256
  `0dcf3303e9d7b875429c4247a4b6c6419a6e7676b3155c7596d15436c1d9aa94`.
  The post-`#4345` importer no longer stops at a missing numeric bound. It now correctly reaches
  `GET /activity/`'s optional `id__in` query parameter and refuses its documented
  `anyOf: [array<string>, null]` shape at
  `validateBoundedRequestSchemaWithinEnum` (`cmd/connectorgen/sourceimport.go:7175-7178`).
  This is a distinct shared source-contract gap: safely normalize only the nullable-null form or
  record a per-operation source gap; never choose an arbitrary `anyOf` arm in connector JSON.

## V3 source retention increments — 2026-08-24

The first ten migrated source locks reached the v3 document validator and each correctly advances
only to its missing canonical descriptor. `source-retain` was then run against exactly the locked
public document URL, without credentials and without a provider API operation. Close is the sole
byte-identical artifact: it retained one verified OpenAPI document and its offline import exposed
the nullable-union foundation gap above. The other nine writes were refused before artifact
publication: Salesloft, Copper, Zoho Bigin, Klaviyo, Customer.io, and Intercom have byte/SHA drift;
Braze redirects; Freshdesk returns HTTP 403; and Segment returns HTTP 404. No lock, URL, digest,
or byte count was rewritten.

The machine-readable state is [`SOURCE-LOCK-MIGRATION-READINESS.json`](SOURCE-LOCK-MIGRATION-READINESS.json).

The next retained-source imports found two more genuine shared contract refusals. Iterable's
`POST /api/auth/jwts/invalidate` parameter 0 and Gorgias's `POST /api/account/settings`
`application/json` body both use documented objects with dynamic `additionalProperties`.
`validateBoundedRequestSchemaWithinEnum` refuses them before either descriptor can be emitted.
The smallest safe foundation is a per-operation non-executable source-contract disposition for
open request shapes, preserving every other operation in the descriptor; a connector-local
closed object would invent a provider contract.

Chargebee supplies a separate import-scaling refusal: its retained 11 MB OpenAPI artifact reaches
the fixed global source-reference index byte limit before source descriptors are emitted
(`sourceReferenceIndex.checkAddition`, `cmd/connectorgen/sourceimport.go:2491-2497`). The safe
shared closure is a bounded, operation-scoped reference index or equivalent streaming traversal;
raising the package budget or discarding source positions would weaken the parser contract.

## Gorgias executor-boundary reconciliation — 2026-08-24

The retained Gorgias source supplies six non-declarable operations that the initial global
projection incorrectly counted as ordinary declaration work. Five are genuine engine foundation
gaps: the two bare-scalar writes `UpdateCustomerCustomFieldValue91` and
`UpdateTicketCustomField107` cannot use the default JSON writer because it always materializes an
object (`internal/connectors/engine/write.go:674-692`); `GetStatistic71` requires a recursive
50-arm filter-expression input and direct read deliberately allows raw input only for an exact
`text/plain` root-string contract (`direct_read.go:286-293`, `:785-786`);
`DownloadLegacyStatistic77` is a POST CSV export but the binary executor requires GET
(`binary_read.go:292-293`, `:333-334`); and `UpdateViewItems112` is a provider-defined PUT read
while the operation reader accepts GET or POST only (`direct_read.go:429-432`).

`DownloadFile18` is deliberately not a foundation gap: its provider source documents a signed
cross-host redirect but no exact allowed redirect host or final media contract. The engine already
has the safe capability and refuses only unbounded redirect metadata
(`operation_headers.go:455-491`); the necessary change is provider evidence, not an engine or
connector shim. The machine-readable split is therefore **288 runtime-enabled**, **3,638
connector-declaration-pending**, **5 execution-foundation-blocked**, and **1
provider-contract-unavailable** (3,932 total). No credentials, provider operation, or live
certification was used for this source-only audit.

A connector-local exception scan then reviewed every batch 6/7 `api_surface.json` row whose
blocked record names `named_dependency=`. It is not used as the provider operation denominator;
it is only a check that existing ledgers do not already identify another executor refusal. Gorgias
is the sole concrete dependency holder. Its deprecated ticket-messages entry is explicitly
`named_dependency=none` and remains reachable, so it is not an exception or an omission.

Outreach remains intentionally at schema v2. Current `source-retain` stops at
`parse source lock: json: unknown field "source_url"` before it can verify the existing OpenAPI
artifact, while its six custom-object operations also cite `developers.outreach.io` rather than the
captured `api.outreach.io` document. The queued retain-only reader in #4350 addresses only the
former legacy-reader refusal; a second immutable developer-document capture remains mandatory
before this lock can be migrated without false provenance.

## Relaunch baseline and CI repair — 2026-08-20

- **RED:** PR run `32283259925` failed `TestGorgiasAPISurfaceOperationLedger`: the recovered v2
  surface left seven blocked rows without their source citation and named dependency, while its
  connector-local test intentionally verifies the v1 ledger metadata contract. The same PR head
  failed `Website Data` because the generated connector catalog still described the former Segment
  `workspaces` stream.
- **GREEN:** Gorgias now keeps the source-locked ReadMe OpenAPI citation and a named
  connector-local dependency on every affected blocked row, and declares ledger version 1 as its
  existing test requires. `go test -timeout 20m ./cmd/connectorgen -run
  TestGorgiasAPISurfaceOperationLedger -count=1` passed.
- **GREEN:** `npm --prefix website run gen:website-data` regenerated only
  `website/data/connectors.generated.json` and
  `website/lib/connectors.catalog.data.generated.json`, carrying the already-correct Segment
  singleton stream into website data.
- **GREEN:** `go run ./cmd/connectorgen validate` passed (`552 connector(s) checked, 0 findings`)
  and `go run ./cmd/connectorgen surface-sync --check` passed (`552 connector(s) scanned, 0
  field(s) filled and 0 field(s) corrected across 0 connector(s)`). `jq empty` passed for the
  readiness baseline and Gorgias API surface.

## Gorgias connector-owned destination increment — 2026-08-20

- **RED:** Gorgias had two documented-but-disabled source rows despite existing command/operation
  support (`files download`) or a faithfully modelable deprecated provider route (`GET
  /api/tickets/{ticket_id}/messages`). It had no source or destination transport declaration, and
  every typed action still named the retired generic-destination gap. `connectorgen params-import`
  was also unable to import the otherwise hash-matching public OpenAPI because its `/api/jobs`
  parameter schema uses a valid union type that the importer cannot decode.
- **GREEN:** The legacy message route is an implemented bounded `rest_read` command with its exact
  required integer `ticket_id`; the binary route is explicitly a `binary_read` source row; and
  `sync_transport.json` declares all four streams plus the exact `tickets(id,status) → update_ticket`
  typed-destination proof. The endpoint ledger and website projections were regenerated. All 61
  typed actions have an explicit eligibility decision: one bound proof, 59 pending closed
  definition-owned action selection, and one multipart/binary semantic exclusion that remains
  CLI-reachable. Provider-live certification is still pending.
- **GREEN:** `go test -timeout 20m ./cmd/connectorgen -run TestGorgiasAPISurfaceOperationLedger -count=1`,
  `go test -timeout 20m ./internal/connectors/commandrunner -run
  TestEveryImplementedCommandPassesRuntimePreflight -count=1`, and `go test -timeout 20m
  ./internal/connectors/engine -run 'TestShippedOperationEndpointLedgerRejectsMissingProjection|TestLoadAll'
  -count=1` passed. A fresh installed binary in an initialized isolated project reached both
  `pm gorgias tickets messages list --ticket-id 1 --json` and `pm gorgias tickets update --id 1
  --status open --json`; each stopped at expected `error: missing --credential`, never an unknown
  command. No credential or provider request was used.
- **FOUNDATION RECONCILIATION:** foundation `609f23bb3861ba7bc2ef1f7bc5246f5751cf9e57` is now an
  ancestor and supplies generic persisted App/CLI dispatch plus exact action selection. The bound
  Gorgias action remains declaration/fixture proof until this connector's installed App/CLI fixture
  path is exercised; its 59 other typed actions are declaration-pending exact mappings/conformance.

## Chatwoot connector-owned destination increment — 2026-08-20

- **RED:** Chatwoot exposed seven fixture-backed streams and 60 source-bound typed actions but had
  neither a `sync_transport.json` declaration nor a per-action eligibility state beyond the superseded
  generic-destination gap.
- **GREEN:** `sync_transport.json` now declares all seven source streams and the concrete
  `contacts(id,blocked) → update_contact` proof for all closed modes, keyed delivery, durable
  acknowledgement, and fixture/dry conformance. The source disposition distinguishes the bound
  action from 59 eligible actions now declaration-pending exact source mappings and conformance. Its 60 existing approval-governed
  write commands remain user-reachable; no provider credential or call was used.
- **GREEN:** `go run ./cmd/connectorgen validate`, `go run ./cmd/connectorgen surface-sync --check`,
  `go test -timeout 20m ./internal/connectors/commandrunner -run
  TestEveryImplementedCommandPassesRuntimePreflight -count=1`, and `go test -timeout 20m
  ./internal/connectors/engine -run 'TestShippedOperationEndpointLedgerRejectsMissingProjection|TestLoadAll'
  -count=1` passed.
- **FOUNDATION RECONCILIATION:** foundation `609f23bb3861ba7bc2ef1f7bc5246f5751cf9e57` supplies
  generic persisted App/CLI dispatch and exact action selection. Chatwoot still needs its installed
  App/CLI fixture path; provider-live certification remains pending.

## Customer.io connector-owned destination increment — 2026-08-20

- **RED:** Customer.io had 16 source-locked streams and ten exact typed actions but no source or
  destination transport declaration. Its remaining 58 mutations lack a connector-owned typed
  action and had been incorrectly assigned a generic destination foundation gap.
- **GREEN:** `sync_transport.json` declares all 16 sources and the exact
  `snippets(name,value) → update_snippet` typed-destination proof, including keyed delivery,
  durable acknowledgement, all three mode strategies, and fixture/dry conformance. All ten typed
  actions explicitly carry eligibility: `update_snippet` is the bound proof; the other nine are
  declaration-pending exact source mappings and conformance. The remaining mutations are correctly `declaration-pending`,
  never safety-excluded. No credentials or provider calls were used.
- **GREEN:** `go run ./cmd/connectorgen validate`, `go run ./cmd/connectorgen surface-sync --check`,
  `go test -timeout 20m ./internal/connectors/commandrunner -run
  TestEveryImplementedCommandPassesRuntimePreflight -count=1`, and `go test -timeout 20m
  ./internal/connectors/engine -run 'TestShippedOperationEndpointLedgerRejectsMissingProjection|TestLoadAll'
  -count=1` passed.
- **FOUNDATION RECONCILIATION:** foundation `609f23bb3861ba7bc2ef1f7bc5246f5751cf9e57` supplies
  persisted App/CLI dispatch and exact action selection. Customer.io still needs its installed App/CLI
  fixture path; provider-live certification remains pending.

## Close connector-owned destination increment — 2026-08-20

- **RED:** Close had 14 source-locked streams and 12 exact typed actions but no source or
  destination transport declaration. Its remaining 151 mutations lack connector-owned typed actions
  and had been incorrectly assigned a generic destination foundation gap.
- **GREEN:** `sync_transport.json` declares all 14 sources and the exact
  `leads(id,name,description,url,status_id) → update_lead` typed-destination proof, including keyed
  delivery, durable acknowledgement, all three mode strategies, and fixture/dry conformance. All 12
  typed actions explicitly carry eligibility: `update_lead` is the bound proof and the other 11 are
  declaration-pending exact source mappings and conformance. The remaining mutations are correctly
  `declaration-pending`, never safety-excluded. No credentials or provider calls were used.
- **GREEN:** `go run ./cmd/connectorgen validate`, `go run ./cmd/connectorgen surface-sync --check`,
  `go test -timeout 20m ./internal/connectors/commandrunner -run
  TestEveryImplementedCommandPassesRuntimePreflight -count=1`, and `go test -timeout 20m
  ./internal/connectors/engine -run 'TestShippedOperationEndpointLedgerRejectsMissingProjection|TestLoadAll'
  -count=1` passed.
- **FOUNDATION RECONCILIATION:** foundation `609f23bb3861ba7bc2ef1f7bc5246f5751cf9e57` supplies
  persisted App/CLI dispatch and exact action selection. Close still needs its installed App/CLI
  fixture path; provider-live certification remains pending.

## Outreach connector-owned destination increment — 2026-08-20

- **RED:** Outreach had 96 source-locked streams and 163 exact typed actions but no source or
  destination transport declaration. All actions still named the superseded generic-destination
  foundation gap.
- **GREEN:** `sync_transport.json` declares all 96 sources and the exact
  `sequences(id) → activate_sequence` typed-destination proof, including keyed delivery, durable
  acknowledgement, all three mode strategies, and fixture/dry conformance. All 163 typed actions
  explicitly carry eligibility: `activate_sequence` is the bound proof and the other 162 are
  declaration-pending exact source mappings and conformance. No credentials or provider calls were used.
- **GREEN:** `go run ./cmd/connectorgen validate`, `go run ./cmd/connectorgen surface-sync --check`,
  `go test -timeout 20m ./internal/connectors/commandrunner -run
  TestEveryImplementedCommandPassesRuntimePreflight -count=1`, and `go test -timeout 20m
  ./internal/connectors/engine -run 'TestShippedOperationEndpointLedgerRejectsMissingProjection|TestLoadAll'
  -count=1` passed.
- **FOUNDATION RECONCILIATION:** foundation `609f23bb3861ba7bc2ef1f7bc5246f5751cf9e57` supplies
  persisted App/CLI dispatch and exact action selection. Outreach still needs its installed App/CLI
  fixture path; provider-live certification remains pending.

## Zoho Bigin connector-owned destination increment — 2026-08-20

- **RED:** Zoho Bigin had 13 source-locked streams and six exact typed actions but no source or
  destination transport declaration. Its other 37 mutations lack connector-owned typed actions and
  had been incorrectly assigned a generic destination foundation gap.
- **GREEN:** `sync_transport.json` declares all 13 sources and the exact
  `records(id) → delete_record` typed-destination proof, including keyed delivery, durable
  acknowledgement, all three mode strategies, and fixture/dry conformance. All six typed actions
  explicitly carry eligibility: `delete_record` is the destructive bound proof and the other five
  are declaration-pending exact source mappings and conformance. The remaining mutations are correctly
  `declaration-pending`; destructive safety remains confirmation/approval metadata. No credentials
  or provider calls were used.
- **GREEN:** `go run ./cmd/connectorgen validate`, `go run ./cmd/connectorgen surface-sync --check`,
  `go test -timeout 20m ./internal/connectors/commandrunner -run
  TestEveryImplementedCommandPassesRuntimePreflight -count=1`, and `go test -timeout 20m
  ./internal/connectors/engine -run 'TestShippedOperationEndpointLedgerRejectsMissingProjection|TestLoadAll'
  -count=1` passed.
- **FOUNDATION RECONCILIATION:** foundation `609f23bb3861ba7bc2ef1f7bc5246f5751cf9e57` supplies
  persisted App/CLI dispatch and exact action selection. Zoho Bigin still needs its installed App/CLI
  fixture path; provider-live certification remains pending.

## Chargebee connector-owned destination increment — 2026-08-20

- **RED:** Chargebee had 32 source-locked streams and 36 exact typed actions but no source or
  destination transport declaration. Its other 331 mutations lack connector-owned typed actions and
  had been incorrectly assigned a generic destination foundation gap.
- **GREEN:** `sync_transport.json` declares all 32 sources and the exact
  `customers(id) → update_customer` typed-destination proof, including keyed delivery, durable
  acknowledgement, all three mode strategies, and fixture/dry conformance. All 36 typed actions
  explicitly carry eligibility: `update_customer` is the bound proof and the other 35 are
  declaration-pending exact source mappings and conformance. The remaining mutations are correctly `declaration-pending`;
  finance and destructive safety remain approval metadata. No credentials or provider calls were used.
- **GREEN:** `go run ./cmd/connectorgen validate`, `go run ./cmd/connectorgen surface-sync --check`,
  `go test -timeout 20m ./internal/connectors/commandrunner -run
  TestEveryImplementedCommandPassesRuntimePreflight -count=1`, and `go test -timeout 20m
  ./internal/connectors/engine -run 'TestShippedOperationEndpointLedgerRejectsMissingProjection|TestLoadAll'
  -count=1` passed.
- **FOUNDATION RECONCILIATION:** foundation `609f23bb3861ba7bc2ef1f7bc5246f5751cf9e57` supplies
  persisted App/CLI dispatch and exact action selection. Chargebee still needs its installed App/CLI
  fixture path; provider-live certification remains pending.

## ServiceNow connector-owned destination increment — 2026-08-20

- **RED:** ServiceNow's public Table API provides six fixed templates but customer table/field
  schemas are instance-dependent. Expanding the template into finite instance operations would be
  invented provider evidence; only two existing source-action projections were visible in the map.
- **GREEN:** `sync_transport.json` declares all three fixture streams and the exact
  `incidents(sys_id) → update_incident` proof with keyed delivery, durable acknowledgement, closed
  modes, and fixture/dry conformance. `typed_action_eligibility` explicitly accounts for every
  source-backed typed action (one bound and one declaration-pending exact source mapping/conformance)
  without fabricating dynamic table source identities. No credentials or provider calls were used.
- **GREEN:** `go run ./cmd/connectorgen validate`, `go run ./cmd/connectorgen surface-sync --check`,
  `go test -timeout 20m ./internal/connectors/commandrunner -run
  TestEveryImplementedCommandPassesRuntimePreflight -count=1`, and `go test -timeout 20m
  ./internal/connectors/engine -run 'TestShippedOperationEndpointLedgerRejectsMissingProjection|TestLoadAll'
  -count=1` passed.
- **FOUNDATION RECONCILIATION:** foundation `609f23bb3861ba7bc2ef1f7bc5246f5751cf9e57` supplies
  persisted App/CLI dispatch and exact action selection. ServiceNow still needs its installed App/CLI
  fixture path; provider-live certification remains pending.

## Braze connector-owned destination increment — 2026-08-20

- **RED:** Braze had 21 fixture streams and 29 exact typed actions without source/destination
  declaration. Its provider inventory remains explicitly unproven, so this increment cannot claim
  complete provider coverage.
- **GREEN:** `sync_transport.json` declares the existing streams and exact
  `content_blocks(content_block_id) → update_content_block` proof with keyed delivery, durable
  acknowledgement, closed strategies, and fixture/dry conformance. All 29 typed actions carry an
  eligibility disposition (one bound; 28 declaration-pending exact source mapping/conformance). No credentials or
  provider calls were used.
- **GREEN:** `go run ./cmd/connectorgen validate`, `go run ./cmd/connectorgen surface-sync --check`,
  `go test -timeout 20m ./internal/connectors/commandrunner -run
  TestEveryImplementedCommandPassesRuntimePreflight -count=1`, and `go test -timeout 20m
  ./internal/connectors/engine -run 'TestShippedOperationEndpointLedgerRejectsMissingProjection|TestLoadAll'
  -count=1` passed.
- **FOUNDATION RECONCILIATION:** foundation `609f23bb3861ba7bc2ef1f7bc5246f5751cf9e57` supplies
  App/CLI dispatch and exact action selection. Braze still needs its installed App/CLI fixture path,
  and its provider-inventory recovery remains independently open; neither is a completeness claim.

## Help Scout foundation reconciliation — 2026-08-20

- **RED:** Before foundation `609f23bb3861ba7bc2ef1f7bc5246f5751cf9e57`, `go run
  ./cmd/connectorgen validate` refused the exact `conversationId` source binding. The old common
  validator rejected camelCase before checking the selected `writes.json` action schema, and the
  persisted App/CLI destination route was not available.
- **GREEN:** The merged #4304 head validates the closed-schema
  `conversations(id) → update_conversation(conversationId)` declaration. These focused checks passed:
  `go run ./cmd/connectorgen validate`; `go run ./cmd/connectorgen surface-sync --check`; `go test
  -timeout 20m ./internal/connectors/commandrunner -run
  TestEveryImplementedCommandPassesRuntimePreflight -count=1`; `go test -timeout 20m
  ./internal/connectors/engine -run 'TestShippedOperationEndpointLedgerRejectsMissingProjection|TestLoadAll'
  -count=1`; `go test -timeout 20m ./internal/app -run
  'Test.*(DeclarativeTypedDestination|TypedDestination)' -count=1`; `go test -timeout 20m
  ./internal/cli -run 'Test.*DeclarativeTypedDestination|TestETLTransport' -count=1`; `go test
  -timeout 20m ./internal/connectors -run 'Test.*Destination.*Action|Test.*Camel' -count=1`; and
  `npm --prefix website run gen:website-data`.
- **Installed binary evidence:** a fresh built `pm` accepted `pm etl transport
  declarative-typed-destination --help`, and `pm connectors inspect help-scout --json` reports the
  declared source, exact destination action, closed modes, keyed delivery, and durable acknowledgement.
  The real `pm connections create ... --destination-action update_conversation` route parsed the
  persisted destination selection and stopped at `credential "credential-free-proof" not found`, before
  provider I/O. No credentials, provider request, provider write, or live certification was used.
- **OPEN FOUNDATION GAP:** `PATCH /v2/customers/{customerId}` / `update_customer` remains not enabled.
  `internal/connectors/sync_transport.go:436-445` rejects a second binding for the same declarative
  source executor at line 443, while `internal/app/issue_label_warehouse_transport.go:349-359` resolves
  the only conversation binding and rejects its `conversationId` input for `update_customer`. The
  machine-readable `declarative-typed-destination-action-specific-source-bindings` row records the
  source URL/revision/hash, fan-out, owner #4304, and exact closure verification. No connector-specific
  workaround was added.
- **OPEN FOUNDATION GAP — assigned route override:** all five source-locked Help Scout Mailbox API v3
  direct reads remain not enabled under `declarative-operation-route-override`, owned by
  `cli-operation-route-override-foundation-r1`. `internal/connectors/engine/direct_read.go:701-717`
  preserves a path whose version differs from the configured v2 base, and
  `internal/connectors/engine/direct_read_paginate.go:499-504` joins it into `/v2/v3/...`.
  Rewriting the stored base would break v2 connections, so no Help Scout workaround is permitted.
  The five operation rows retain exact provider URL/v3 revision/source hash, direct-read/CLI surfaces,
  failure evidence, fan-out, owner, and closure tests in `FOUNDATION-GAPS.json`. The common foundation
  must close definition-owned per-operation route selection across direct read/write, binary
  download/upload, ETL, and reverse ETL, refusing arbitrary caller URLs, undeclared routes, and
  silent fallback before I/O.

## Captain zero-omission pre-merge gate — 2026-08-20

Before merge, the final evidence will carry an operation-level, twenty-connector matrix with source
URL/version/hash, canonical mapping, runtime reachability, generated CLI command, generated website row,
and executable fixture/conformance evidence. ETL, reverse ETL, direct read/write, binary download, and
binary upload are independent cells. `N/A` requires provider evidence that the capability is absent;
scope, tier, destructive, and safety controls are typed runtime/confirmation metadata, never omission
reasons.

`FOUNDATION-GAPS.json` is the required machine-readable complement: it retains five stable
provider-neutral gap IDs (three resolved by foundation `609f23bb3861ba7bc2ef1f7bc5246f5751cf9e57` and
two open) with one provenance-rich row per affected operation, explicit per-batch and portfolio
fan-out, owner/lane/status, and exact closure verification. Its portfolio rollup is currently
`merge_ready: false`; `update_customer` and all five Help Scout v3 direct-read rows with open
foundation gaps are not enabled and cannot contribute to a merge-ready verdict.

## Main-foundation reconciliation and citation-contract pause — 2026-08-23

- **Main reconciliation:** merged `origin/main` at `cf493b834` into this branch and retargeted PR
  #4296 to `main`; GitHub API read-back returned `main`.
- **Independent GREEN:** `go test -count=1 -timeout 20m ./internal/connectors/engine -run
  '^(TestHelpScoutV3DirectReadsUseTheirDeclaredRoute|TestOperationRoutes)'` passed. `go run
  ./cmd/pm connectors inspect help-scout --json` was asserted to expose both eligible actions,
  `conversations.id → conversationId`, and `customers.id → customerId`. This is a definition-only,
  credential-free proof; no provider request or certification was run.
- **Citation RED / pause:** `go run ./cmd/connectorgen validate` reports 20 source-projection failures:
  schema-v2 operations reject the preserved `source_url` citation field, while v3 requires a
  captured OpenAPI artifact and published source per document. The rendered references have no such
  artifact and must not receive invented form pins or digests. `cli-rendered-reference-citation-contract-r1`
  owns the discriminated rendered-reference document form. The 20 source-lock changes are deliberately
  left uncommitted and the branch is not pushed until that contract lands.
# Stale typed-destination gap reconciliation — increment 1 (2026-08-20)

- **RED:** The operation-evidence ledger exposed 1,111 direct-write rows that still named
  `generic-typed-destination-executor` even though foundation
  `609f23bb3861ba7bc2ef1f7bc5246f5751cf9e57` provides the declaration-driven destination
  factory. Salesloft, Copper, Klaviyo, Intercom, and Freshdesk have no typed `writes.json`
  action, so the old shared foundation gap was false evidence.
- **GREEN:** The five connector-owned disposition ledgers now give every direct write explicit
  `declaration-pending` reverse-ETL status, name the missing typed action/CLI contract, and
  declare the destination transport work as connector-owned. No safety, privilege, destructive,
  or certification condition was used to exclude an operation.
- **ASSERTION:** `jq` verifies no `generic-typed-destination-executor` record remains in those
  five ledgers and no direct-write reverse-ETL state differs from `declaration-pending`. The
  operation evidence retains exactly 612 stale rows in the five untouched ledgers as the next
  bounded increment, plus five Help Scout route and one action-specific source-binding open rows.

# Stale typed-destination gap reconciliation — increment 2 (2026-08-20)

- **RED:** The first increment deliberately retained 612 stale generic-gap rows in Segment,
  ActiveCampaign, Iterable, Square, and Braintree so the connector scope stayed bounded.
- **GREEN:** Those five ledgers now have the same explicit connector-owned declaration-pending
  state as increment 1. All 1,111 affected direct-write rows are preserved as direct writes; none
  is hidden as unsafe, privileged, destructive, uncommon, binary, or un-certified.
- **ASSERTION:** The 3,932-row operation evidence has complete source URL/version/SHA-256 trace
  and all seven cell keys for every row; it contains no `generic-typed-destination-executor` gap.
  Exactly six open foundation-gap rows remain: five Help Scout route-override rows and one Help
  Scout action-specific source-binding row. `FOUNDATION-GAPS.json` has the same two open stable IDs
  and portfolio `merge_ready: false`.
- **GATES (all pass):** the exact five-ledger and 3,932-row `jq -e` assertions described above;
  `go run ./cmd/connectorgen validate` (`552 connector(s) checked, 0 findings`);
  `go run ./cmd/connectorgen surface-sync --check` (`552 connector(s)` and `0` drift);
  `go test -timeout 20m ./internal/connectors/commandrunner -run
  TestEveryImplementedCommandPassesRuntimePreflight -count=1`;
  `go test -timeout 20m ./internal/connectors/engine -run
  'TestShippedOperationEndpointLedgerRejectsMissingProjection|TestLoadAll' -count=1`;
  `npm --prefix website run gen:website-data`; `make connector-boundary` (managed detached run,
  exit `0`, log `/tmp/cli-map-batch67-r1-connector-boundary.log`); and `git diff --check`.
- **STACKED FOUNDATION / INSTALLED BINARY:** fetched `origin/fm/cli-reverse-etl-destination-r1`,
  merged it with `Already up to date`, and proved
  `609f23bb3861ba7bc2ef1f7bc5246f5751cf9e57` is an ancestor of this branch. A fresh
  `go build -o /tmp/cli-map-batch67-r1-proof.PtA3NA/pm ./cmd/pm` then passed
  `pm etl transport declarative-typed-destination --help` and
  `pm connectors inspect help-scout --json`: the installed binary reports the exact
  `update_conversation` destination action, all three declared modes, and durable acknowledgement.
  In a fresh initialized temporary project, the real
  `pm connections create help-scout-route-proof --source help-scout:credential-free-proof
  --destination help-scout:credential-free-proof --stream conversations --destination-action
  update_conversation` path exited `1` with `error: resolve source: credential
  "credential-free-proof" not found`, before provider I/O. No credential was created or used.
