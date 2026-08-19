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
- **PENDING FOUNDATION:** This is declaration/fixture proof only. Per the captain update, #4304
  still needs persisted App/CLI generic-destination dispatch. Do not mark application-level
  reverse-ETL deployable until the final foundation merge is an ancestor and the real App/CLI path
  exercises the declared destination.

## Chatwoot connector-owned destination increment — 2026-08-20

- **RED:** Chatwoot exposed seven fixture-backed streams and 60 source-bound typed actions but had
  neither a `sync_transport.json` declaration nor a per-action eligibility state beyond the superseded
  generic-destination gap.
- **GREEN:** `sync_transport.json` now declares all seven source streams and the concrete
  `contacts(id,blocked) → update_contact` proof for all closed modes, keyed delivery, durable
  acknowledgement, and fixture/dry conformance. The source disposition distinguishes the bound
  action from 59 eligible actions awaiting #4304 exact selection. Its 60 existing approval-governed
  write commands remain user-reachable; no provider credential or call was used.
- **GREEN:** `go run ./cmd/connectorgen validate`, `go run ./cmd/connectorgen surface-sync --check`,
  `go test -timeout 20m ./internal/connectors/commandrunner -run
  TestEveryImplementedCommandPassesRuntimePreflight -count=1`, and `go test -timeout 20m
  ./internal/connectors/engine -run 'TestShippedOperationEndpointLedgerRejectsMissingProjection|TestLoadAll'
  -count=1` passed.
- **PENDING FOUNDATION:** persisted App/CLI generic-destination dispatch remains #4304 work; this
  connector declares a fixture/dry route only and does not claim application-level deployment.
