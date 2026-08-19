Refs #4291

## Status

Reconciliation is in progress; do not merge. This PR now depends on
`fm/cli-reverse-etl-destination-r1` / PR #4304, merged into this branch as
`3c50ae3cda9bb04944a660f9c8793fac6ae3ef16`, including foundation
`609f23bb3861ba7bc2ef1f7bc5246f5751cf9e57`. The base has been verified through
the GitHub API as `fm/cli-reverse-etl-destination-r1`.

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
| help-scout | unproven | 50 / 55 | 65 / 65 | 65 | 1 / 24 | 1 (fixture/preflight proven; action-specific gap open) | 1 / 0 | 18 |
| gorgias | complete | 41 / 42 | 61 / 68 | 61 | 1 / 4 | 1 (foundation merged; connector App proof pending) | 1 / 1 | 18 |
| service-now | dynamic templates | 0 / 1 | 2 / 4 | 0 | 1 / 3 | 1 (foundation merged; connector App proof pending) | 0 / 0 | 1 |
| chatwoot | complete | 32 / 57 | 60 / 84 | 60 | 1 / 7 | 1 (foundation merged; connector App proof pending) | 0 / 0 | 18 |
| chargebee | complete | 0 / 128 | 36 / 367 | 0 | 1 / 32 | 1 (foundation merged; connector App proof pending) | 0 / 0 | 0 |
| square | complete | 0 / 117 | 0 / 213 | 0 | 0 / 4 | 0 | 0 / 0 | 27 |
| braintree | unproven | 0 / 18 | 0 / 45 | 0 | 0 / 10 | 0 | 0 / 0 | 6 |

## Remaining gaps

- 1,342 direct reads need provider-evidenced bounded operation contracts and generated CLI
  bindings; 1,745 direct writes need exact typed actions. Gorgias already has 61 user-reachable
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
- Help Scout now declares and fixture/preflight proves
  `conversations(id) → update_conversation(conversationId)`. The common camelCase rule and persisted
  App/CLI route are present at `609f23bb3861ba7bc2ef1f7bc5246f5751cf9e57`; the installed connection
  route correctly stops at missing credentials before provider I/O. Its `update_customer` operation
  is not enabled because the current closed model permits only one source binding per executor.
  `FOUNDATION-GAPS.json` records the exact source-traced
  `declarative-typed-destination-action-specific-source-bindings` dependency; no connector-specific
  workaround was added.
- Five source-locked Help Scout Mailbox API v3 reads are explicitly not enabled behind the assigned
  `declarative-operation-route-override` foundation lane
  `cli-operation-route-override-foundation-r1`. The current `/v2` bundle base would construct
  `/v2/v3/...`; no base rewrite, caller URL, generic HTTP path, or silent fallback is permitted.
  Their source URL/version/hash, exact runtime evidence, owner, fan-out, and closure checks are in
  `FOUNDATION-GAPS.json`; portfolio `merge_ready` remains false.
- Braze, Help Scout, and Braintree source inventories remain unproven; the rest use a complete
  provider specification/reference or ServiceNow's explicitly dynamic fixed-template basis.
- Help Scout still needs its binary command represented as a binary parity row; Gorgias now is.
- Every connector is provider-live-certification pending; no credentials or provider calls are
  authorized for this lane.
- `FOUNDATION-GAPS.json` records shared provider-neutral gaps once per stable ID, with exact
  source URL/version/hash rows and batch/portfolio fan-out. Its current `merge_ready: false` is
  authoritative: no open foundation-gap operation is counted as enabled or merge-ready.
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
- Current foundation-reconciliation checks all pass: `go run ./cmd/connectorgen validate`,
  `go run ./cmd/connectorgen surface-sync --check`, Help Scout commandrunner/engine/App/CLI/
  connectors focused tests, `npm --prefix website run gen:website-data`, and `git diff --check`.
  The first operation-evidence increment also passed exact five-ledger and 3,932-row `jq`
  assertions before its connector validation and surface-sync checks.
  The installed, fresh-built binary exposes the closed declarative transport command and Help Scout
  inspection; the connection path stops at missing credentials before provider I/O. Full repository
  gates will be recorded after the connector-owned increments complete.
