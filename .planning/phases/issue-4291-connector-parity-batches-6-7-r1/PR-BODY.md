Refs #4291

## Status

Reconciliation is paused; do not merge. GitHub currently reports PR #4296 targets
`main`. Nineteen source locks are committed in schema v3 using the rendered-reference
citation contract; Outreach remains deliberately unstaged until its second
`developers.outreach.io` document has an immutable retained artifact. The remaining
shared blocker is not the citation schema: `readOperationEvidenceSourceLock` in
`cmd/connectorgen/operationevidence.go` accepts schema v3 but reads only legacy
`rest.operations`, not `rest.source_documents[].operations`. Therefore the current
tree must not be represented as globally generator-green or pushed as complete.

PR #4350 is an additional retain-only source-lock-reader dependency for Outreach's
legacy-lock verification. It does not supply the schema-v3 operation-evidence
projection, and a second immutable Outreach developer-document capture is still
required before that lock can migrate faithfully.

## Twenty-connector seven-surface baseline

`read` and `typed write` are exact current bindings over the source-locked
denominator. `ETL` is declared source transport / bundle streams; `reverse` is
declared typed-destination bindings; `binary` is CLI commands / ledger rows.
An unproven inventory needs source recovery. Safety and provider-live
certification never make an operation unreachable.

| Connector | Inventory | Read | Typed write | Write CLI | ETL | Reverse | Binary | Deletes |
| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| close-com | complete | 0 / 123 | 12 / 163 | 0 | 1 / 14 | 1 (foundation merged; connector App proof pending) | 0 / 0 | 52 |
| outreach | complete | 0 / 0 | 163 / 163 | 0 | 1 / 96 | 1 (foundation merged; connector App proof pending) | 0 / 0 | 36 |
| salesloft | complete | 0 / 115 | 0 / 91 | 0 | 0 / 5 | 0 | 0 / 0 | 18 |
| copper | complete | 0 / 32 | 0 / 52 | 0 | 0 / 5 | 0 | 0 / 0 | 11 |
| zoho-bigin | complete | 0 / 26 | 6 / 43 | 0 | 1 / 13 | 1 (foundation merged; connector App proof pending) | 0 / 0 | 11 |
| klaviyo | complete | 0 / 198 | 0 / 141 | 0 | 0 / 6 | 0 | 0 / 0 | 30 |
| braze | unproven | 0 / 21 | 29 / 53 | 0 | 1 / 21 | 1 (foundation merged; connector App proof pending) | 0 / 0 | 4 |
| customer-io | complete | 0 / 82 | 10 / 68 | 10 | 1 / 16 | 1 (foundation merged; connector App proof pending) | 0 / 0 | 14 |
| intercom | complete | 0 / 103 | 0 / 123 | 0 | 0 / 5 | 0 | 0 / 0 | 31 |
| freshdesk | complete | 0 / 73 | 0 / 92 | 0 | 0 / 5 | 0 | 0 / 0 | 22 |
| segment | complete | 0 / 97 | 0 / 101 | 0 | 0 / 3 | 0 | 0 / 0 | 28 |
| activecampaign | complete | 0 / 128 | 0 / 157 | 0 | 0 / 11 | 0 | 0 / 0 | 50 |
| iterable | complete | 0 / 49 | 0 / 96 | 0 | 0 / 3 | 0 | 0 / 0 | 12 |
| help-scout | unproven | 55 / 55 | 65 / 65 | 65 | 1 / 24 | 1 (two exact source bindings; connector fixture/CLI proof pending) | 1 / 0 | 18 |
| gorgias | complete | 41 / 42 | 61 / 68 | 61 | 1 / 4 | 1 (foundation merged; connector App proof pending) | 1 / 1 | 18 |
| service-now | dynamic templates | 0 / 1 | 2 / 4 | 0 | 1 / 3 | 1 (foundation merged; connector App proof pending) | 0 / 0 | 1 |
| chatwoot | complete | 32 / 57 | 60 / 84 | 60 | 1 / 7 | 1 (foundation merged; connector App proof pending) | 0 / 0 | 18 |
| chargebee | complete | 0 / 128 | 36 / 367 | 0 | 1 / 32 | 1 (foundation merged; connector App proof pending) | 0 / 0 | 0 |
| square | complete | 0 / 117 | 0 / 213 | 0 | 0 / 4 | 0 | 0 / 0 | 27 |
| braintree | unproven | 0 / 18 | 0 / 45 | 0 | 0 / 10 | 0 | 0 / 0 | 6 |

## Current reconciliation and remaining gaps

- The authoritative 3,932-row machine ledger currently has **576** enabled generated CLI bindings,
  **3,350** connector-local `declaration-pending` operations, **5** execution-foundation-blocked
  operations, and **1** provider-contract-unavailable operation. These are seven-surface claims,
  not provider-live certification; every enabled command remains subject to its normal safety and
  authorization controls.
- The five genuine execution-foundation gaps are all Gorgias: two scalar JSON `PUT` write bodies,
  one recursive structured reporting-filter input, one POST text-export binary read, and one
  provider `PUT` direct read. `FOUNDATION-GAPS.json` names each refusing file/function and minimum
  provider-neutral hook. Gorgias file download remains a separate
  `provider-contract-unavailable` row because the provider supplies no exact stable final-media
  host contract for its signed cross-host redirect; it is not a foundation gap.
- The 3,350 unauthored operations are `declaration-pending`, rather than unsupported: they need
  exact connector-local operation contracts, generated command paths, typed request mappings,
  destination conformance, and/or fixtures. Gorgias already has 61 user-reachable
  approval-governed write commands; its native `direct_write` intent count remains zero, so the
  ledger keeps direct capability distinct from reverse-ETL deployment.
- Gorgias, Chatwoot, Customer.io, Close, Outreach, Zoho Bigin, Chargebee, ServiceNow, and Braze now declare their current ETL stream sets and
  one exact typed-destination proof each (`tickets → update_ticket`, `contacts → update_contact`,
  `snippets → update_snippet`, `leads → update_lead`, `sequences → activate_sequence`, and
  `records → delete_record`, `customers → update_customer`, `incidents → update_incident`, and
  `content_blocks → update_content_block`) with
  keyed delivery, durable acknowledgement, per-mode strategies, and fixture/dry conformance
  evidence. Foundation `609f23bb3861ba7bc2ef1f7bc5246f5751cf9e57` now supplies persisted App/CLI
  dispatch and exact action selection; each connector still needs an installed App/CLI fixture path
  for its declaration. The remaining typed actions are explicitly eligible but declaration-pending
  their exact source mappings and conformance evidence; un-authored direct writes remain
  declaration-pending, and Gorgias multipart `upload_file` has a named binary/multipart semantic
  exclusion while remaining CLI-reachable.
- Help Scout's former shared gaps are closed. The five Mailbox API v3 direct reads use the declared
  `mailbox_v3` route and reach `/v3`, and both `update_conversation` and `update_customer` have
  exact action-specific source bindings. The remaining connector-owned work is fixture/installed
  CLI conformance evidence, not an execution-foundation gap.
- Braze, Help Scout, and Braintree source inventories remain unproven; the rest use a complete
  provider specification/reference or ServiceNow's explicitly dynamic fixed-template basis.
- Help Scout still needs its binary command represented as a binary parity row; Gorgias now is.
- Every connector is provider-live-certification pending; no credentials or provider calls are
  authorized for this lane.
- `FOUNDATION-GAPS.json` records shared provider-neutral gaps once per stable ID, with exact
  source URL/version/hash rows and batch/portfolio fan-out. No open foundation-gap operation is
  counted as enabled or merge-ready.
- `OPERATION-SURFACE-EVIDENCE.json` is the captain-required machine record for all 3,932 provider
  operations: source URL/version/SHA-256, canonical mapping, seven surface cells, generated
  projection cells, fixture/conformance status, and merge readiness. Its first independent
  reconciliation removed all 1,111 stale pre-#4304 generic-destination gap rows in Salesloft,
  Copper, Klaviyo, Intercom, Freshdesk, Segment, ActiveCampaign, Iterable, Square, and Braintree;
  the connectors still lack typed actions, so those rows are honestly `declaration-pending`.
  Missing per-operation fixture/conformance evidence is explicit `not_recorded`, never N/A or
  provider-live success.

## Lifecycle and verification

- GSD executed inline because this issue has no roadmap phase and the canonical worker prohibits
  role spawning. Required skills: `golang-how-to`, `golang-design-patterns`,
  `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`,
  `golang-testing`, `golang-cli`, and `golang-documentation`.
- Red/green evidence and exact commands are in
  `.planning/phases/issue-4291-connector-parity-batches-6-7-r1/VERIFICATION.md`.
- Focused current checks: `go test -timeout 20m ./internal/connectors/engine -run
  '^TestHelpScoutV3DirectReadsUseTheirDeclaredRoute$' -count=1`, `go test -timeout 20m
  ./internal/connectors -run
  '^(TestDestinationTransportDescriptorRequiresExactActionClosure|TestDestinationSourceBindingJSONOmitsAbsentBatch)$'
  -count=1`, and `go test -timeout 20m ./internal/connectors/commandrunner -run
  '^TestEveryImplementedCommandPassesRuntimePreflight$' -count=1` all pass. The machine-ledger
  exact aggregate assertion also passes at 576/3,350/5/1.
- `go run ./cmd/connectorgen operation-evidence . --check`, full `connectorgen validate`,
  `surface-sync --check`, generated checks, `connector-boundary`, and `make verify` remain final
  gates. They must be rerun after the shared schema-v3 evidence reader and Outreach retained
  developer-document capture unblock the current source-lock tree. Earlier pre-v3 green runs are
  historical evidence only; they do not certify this worktree.
- Before the latest push, `origin/fm/cli-reverse-etl-destination-r1` was fetched and merged
  (already current); foundation `609f23bb3861ba7bc2ef1f7bc5246f5751cf9e57` is an ancestor. A
  fresh installed binary passed declarative-typed-destination help and Help Scout transport
  inspection, then sent the actual closed `connections create ... --destination-action
  update_conversation` path to the expected missing-credential preflight before provider I/O. This
  proves sealed App/CLI dispatch, not provider-live certification.
  The installed, fresh-built binary exposes the closed declarative transport command and Help Scout
  inspection; the connection path stops at missing credentials before provider I/O. Full repository
  gates will be recorded after the connector-owned increments complete.
