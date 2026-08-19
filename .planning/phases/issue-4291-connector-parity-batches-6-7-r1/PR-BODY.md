Refs #4291

## Status

Reconciliation is in progress; do not merge. This PR now depends on
`fm/cli-reverse-etl-destination-r1` / PR #4304, merged into this branch as
`d27d4bb64`. The base has been verified through the GitHub API as
`fm/cli-reverse-etl-destination-r1`.

## Twenty-connector seven-surface baseline

`read` and `typed write` are exact current bindings over the source-locked
denominator. `ETL` is declared source transport / bundle streams; `reverse` is
declared typed-destination bindings; `binary` is CLI commands / ledger rows.
An unproven inventory needs source recovery. Safety and provider-live
certification never make an operation unreachable.

| Connector | Inventory | Read | Typed write | Write CLI | ETL | Reverse | Binary | Deletes |
| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| close-com | complete | 0 / 123 | 12 / 163 | 0 | 1 / 14 | 1 (App dispatch pending) | 0 / 0 | 52 |
| outreach | complete | 0 / 0 | 163 / 163 | 0 | 1 / 96 | 1 (App dispatch pending) | 0 / 0 | 36 |
| salesloft | complete | 0 / 115 | 0 / 91 | 0 | 0 / 5 | 0 | 0 / 0 | 18 |
| copper | complete | 0 / 32 | 0 / 52 | 0 | 0 / 5 | 0 | 0 / 0 | 11 |
| zoho-bigin | complete | 0 / 26 | 6 / 43 | 0 | 1 / 13 | 1 (App dispatch pending) | 0 / 0 | 11 |
| klaviyo | complete | 0 / 198 | 0 / 141 | 0 | 0 / 6 | 0 | 0 / 0 | 30 |
| braze | unproven | 0 / 21 | 29 / 53 | 0 | 1 / 21 | 1 (App dispatch pending) | 0 / 0 | 4 |
| customer-io | complete | 0 / 82 | 10 / 68 | 10 | 1 / 16 | 1 (App dispatch pending) | 0 / 0 | 14 |
| intercom | complete | 0 / 103 | 0 / 123 | 0 | 0 / 5 | 0 | 0 / 0 | 31 |
| freshdesk | complete | 0 / 73 | 0 / 92 | 0 | 0 / 5 | 0 | 0 / 0 | 22 |
| segment | complete | 0 / 97 | 0 / 101 | 0 | 0 / 3 | 0 | 0 / 0 | 28 |
| activecampaign | complete | 0 / 128 | 0 / 157 | 0 | 0 / 11 | 0 | 0 / 0 | 50 |
| iterable | complete | 0 / 49 | 0 / 96 | 0 | 0 / 3 | 0 | 0 / 0 | 12 |
| help-scout | unproven | 50 / 55 | 65 / 65 | 65 | 0 / 24 | 0 (camelCase foundation pending) | 1 / 0 | 18 |
| gorgias | complete | 41 / 42 | 61 / 68 | 61 | 1 / 4 | 1 (App dispatch pending) | 1 / 1 | 18 |
| service-now | dynamic templates | 0 / 1 | 2 / 4 | 0 | 1 / 3 | 1 (App dispatch pending) | 0 / 0 | 1 |
| chatwoot | complete | 32 / 57 | 60 / 84 | 60 | 1 / 7 | 1 (App dispatch pending) | 0 / 0 | 18 |
| chargebee | complete | 0 / 128 | 36 / 367 | 0 | 1 / 32 | 1 (App dispatch pending) | 0 / 0 | 0 |
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
  evidence. None is application-level deployable yet: #4304 must land persisted App/CLI
  generic-destination dispatch integration. The remaining typed actions are explicitly eligible but
  await closed exact-action selection; un-authored direct writes are correctly declaration-pending,
  and Gorgias multipart `upload_file` has a named binary/multipart semantic exclusion while
  remaining CLI-reachable.
- Help Scout has an exact candidate `conversations(id) → update_conversation(conversationId)`, but
  the unmerged common #4304 camelCase source-binding validation refuses it before I/O. It is not
  counted as declared until foundation integration and focused/generated gates pass.
- Braze, Help Scout, and Braintree source inventories remain unproven; the rest use a complete
  provider specification/reference or ServiceNow's explicitly dynamic fixed-template basis.
- Help Scout still needs its binary command represented as a binary parity row; Gorgias now is.
- Every connector is provider-live-certification pending; no credentials or provider calls are
  authorized for this lane.
- `FOUNDATION-GAPS.json` records shared provider-neutral gaps once per stable ID, with exact
  source URL/version/hash rows and batch/portfolio fan-out. Its current `merge_ready: false` is
  authoritative: no open foundation-gap operation is counted as enabled or merge-ready.

## Lifecycle and verification

- GSD executed inline because this issue has no roadmap phase and the canonical worker prohibits
  role spawning. Required skills: `golang-how-to`, `golang-design-patterns`,
  `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`,
  `golang-testing`, `golang-cli`, and `golang-documentation`.
- Red/green evidence and exact commands are in
  `.planning/phases/issue-4291-connector-parity-batches-6-7-r1/VERIFICATION.md`.
- Current checks: Gorgias ledger test, `connectorgen validate`, `surface-sync --check`, and
  website data generation pass. Full repository gates will be recorded after the connector-owned
  increments complete.
