# Seven-surface readiness baseline — issue #4291

This is the credential-free before-state after merging the typed destination
foundation. It derives from the source-locked ledgers and current bundle files;
it deliberately distinguishes user reachability, auto-exercise safety, and
provider-live certification. The machine-readable source is
`READINESS-BASELINE.json`.

| Connector | Inventory | Documented | Direct read command / documented | Typed write / documented | ETL source / streams | Reverse destination | Binary CLI / ledger | Deletes | Implemented CLI commands |
| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| close-com | complete | 300 | 0 / 123 | 12 / 163 | 1 / 14 | 1 (foundation merged; App proof pending) | 0 / 0 | 52 | 0 |
| outreach | complete | 259 | 0 / 0 | 163 / 163 | 1 / 96 | 1 (foundation merged; App proof pending) | 0 / 0 | 36 | 0 |
| salesloft | complete | 211 | 0 / 115 | 0 / 91 | 0 / 5 | 0 | 0 / 0 | 18 | 0 |
| copper | complete | 89 | 0 / 32 | 0 / 52 | 0 / 5 | 0 | 0 / 0 | 11 | 0 |
| zoho-bigin | complete | 75 | 0 / 26 | 6 / 43 | 1 / 13 | 1 (foundation merged; App proof pending) | 0 / 0 | 11 | 0 |
| klaviyo | complete | 345 | 0 / 198 | 0 / 141 | 0 / 6 | 0 | 0 / 0 | 30 | 0 |
| braze | unproven | 95 | 0 / 21 | 29 / 53 | 1 / 21 | 1 (foundation merged; App proof pending) | 0 / 0 | 4 | 0 |
| customer-io | complete | 166 | 0 / 82 | 10 / 68 | 1 / 16 | 1 (foundation merged; App proof pending) | 0 / 0 | 14 | 0 |
| intercom | complete | 231 | 0 / 103 | 0 / 123 | 0 / 5 | 0 | 0 / 0 | 31 | 0 |
| freshdesk | complete | 170 | 0 / 73 | 0 / 92 | 0 / 5 | 0 | 0 / 0 | 22 | 0 |
| segment | complete | 201 | 0 / 97 | 0 / 101 | 0 / 3 | 0 | 0 / 0 | 28 | 0 |
| activecampaign | complete | 296 | 0 / 128 | 0 / 157 | 0 / 11 | 0 | 0 / 0 | 50 | 0 |
| iterable | complete | 148 | 0 / 49 | 0 / 96 | 0 / 3 | 0 | 0 / 0 | 12 | 0 |
| help-scout | unproven | 144 | 50 / 55 | 65 / 65 | 1 / 24 | 1 (fixture/preflight proven; action-specific gap open) | 1 / 0 | 18 | 139 |
| gorgias | complete | 114 | 41 / 42 | 61 / 68 | 1 / 4 | 1 (foundation merged; App proof pending) | 1 / 1 | 18 | 94 |
| service-now | dynamic templates | 6 | 0 / 1 | 2 / 4 | 1 / 3 | 1 (foundation merged; App proof pending) | 0 / 0 | 1 | 0 |
| chatwoot | complete | 148 | 32 / 57 | 60 / 84 | 1 / 7 | 1 (foundation merged; App proof pending) | 0 / 0 | 18 | 100 |
| chargebee | complete | 527 | 0 / 128 | 36 / 367 | 1 / 32 | 1 (foundation merged; App proof pending) | 0 / 0 | 0 | 0 |
| square | complete | 334 | 0 / 117 | 0 / 213 | 0 / 4 | 0 | 0 / 0 | 27 | 0 |
| braintree | unproven | 73 | 0 / 18 | 0 / 45 | 0 / 10 | 0 | 0 / 0 | 6 | 0 |

Totals after the Help Scout foundation reconciliation: 3,932 documented operations; 1,465 direct reads
with 123 exact command bindings; 2,189 direct writes with 444 exact typed-action bindings; 287 streams,
ten source transports, ten typed-destination declarations, 424 documented deletes, and zero provider-live
certifications. One Help Scout action remains not enabled behind the open action-specific source-binding
foundation gap; declarations and fixture proof are not provider-live certification.

Help Scout still has a binary-ledger classification defect. Gorgias now maps its
implemented binary command as one binary source row. No operation may stay
unreachable because it is destructive, privileged, uncommon, binary, unsafe to
auto-exercise, or pending provider-live certification.

## Increment — Gorgias declarative destination proof

The `tickets → update_ticket` mapping carries only the exact `id` and `status` fields and all
three closed modes. It is the initial declaration proof, not connector completion: foundation
`609f23bb3861ba7bc2ef1f7bc5246f5751cf9e57` supplies exact action selection, but 59 ordinary
typed actions remain declaration-pending their exact source mapping and conformance, and `upload_file` has a named binary/
multipart semantic exclusion while remaining reachable as `pm gorgias files upload`. The installed
binary reaches both `tickets messages list` and `tickets update` before stopping at the expected
credential preflight; no provider credentials or provider calls were used.

## Increment — Chatwoot declarative destination proof

The `contacts(id, blocked) → update_contact` mapping makes Chatwoot the second fixture/dry
declaration proof, with all seven streams declared as ETL sources. Its 60 typed actions now each
have an eligibility disposition: one bound proof and 59 actions declaration-pending exact source
mapping and conformance. Foundation `609f23bb3861ba7bc2ef1f7bc5246f5751cf9e57` supplies persisted
App/CLI destination dispatch; installed connector fixture proof and provider-live certification remain pending.

## Increment — Customer.io declarative destination proof

Customer.io declares all 16 streams and one fixture/dry `snippets(name,value) → update_snippet`
proof. Its ten typed actions now have explicit eligibility states (one bound and nine declaration-pending
exact source mapping and conformance), while its remaining 58 direct writes are correctly `declaration-pending` until
their exact typed operation and CLI contracts are authored. The destination does not claim
installed App/CLI fixture proof and provider-live certification remain pending.

## Increment — Close declarative destination proof

Close declares all 14 streams and one fixture/dry
`leads(id,name,description,url,status_id) → update_lead` proof. Its twelve typed actions now have
explicit eligibility states (one bound and eleven declaration-pending exact source mapping/conformance); its
other 151 direct writes are correctly `declaration-pending` until their exact connector-owned
contracts and CLI reachability are authored. Foundation `609f23bb3861ba7bc2ef1f7bc5246f5751cf9e57`
supplies persisted App/CLI dispatch; installed connector fixture proof remains pending.

## Increment — Outreach declarative destination proof

Outreach declares all 96 streams and one fixture/dry `sequences(id) → activate_sequence` proof.
Every one of its 163 typed actions has an explicit eligibility state (one bound and 162 declaration-pending
exact source mapping/conformance). The mutating action remains approval-governed; no credential or provider
operation was used, and foundation `609f23bb3861ba7bc2ef1f7bc5246f5751cf9e57` supplies persisted App/CLI dispatch.

## Increment — Zoho Bigin declarative destination proof

Zoho Bigin declares all 13 streams and a fixture/dry `records(id) → delete_record` proof. Its
six typed actions have explicit eligibility states (one bound and five declaration-pending exact source
mapping/conformance); the selected DELETE remains confirmation- and approval-governed, not omitted for its
destructive kind. Foundation `609f23bb3861ba7bc2ef1f7bc5246f5751cf9e57` supplies persisted App/CLI dispatch.

## Increment — Chargebee declarative destination proof

Chargebee declares all 32 streams and one fixture/dry `customers(id) → update_customer` proof.
All 36 typed actions have explicit eligibility states (one bound and 35 declaration-pending exact source
mapping/conformance); financial/destructive safety remains approval metadata. Foundation
`609f23bb3861ba7bc2ef1f7bc5246f5751cf9e57` supplies persisted App/CLI destination dispatch.

## Increment — ServiceNow declarative destination proof

ServiceNow preserves the public API's fixed dynamic-template boundary while declaring its three
fixture-backed sources and `incidents(sys_id) → update_incident` proof. A separate six-action
eligibility ledger accounts for every table-specific typed action without inventing customer table
schemas; one is bound and one is declaration-pending exact source mapping/conformance. Foundation
`609f23bb3861ba7bc2ef1f7bc5246f5751cf9e57` supplies persisted App/CLI dispatch.

## Increment — Braze declarative destination proof

Braze declares its current 21 fixture-backed streams and `content_blocks(content_block_id) →
update_content_block` proof, with all 29 typed actions explicitly eligible (one bound, 28
declaration-pending exact source mapping/conformance). Its source inventory remains unproven pending
public recovery; no completeness claim is made. Foundation `609f23bb3861ba7bc2ef1f7bc5246f5751cf9e57`
supplies App/CLI dispatch.

## Help Scout typed-destination reconciliation

Foundation `609f23bb3861ba7bc2ef1f7bc5246f5751cf9e57` accepts the exact closed-schema binding
`conversations(id) → update_conversation(conversationId)`. The definition, generated surface, runtime
preflight, and fixture App/CLI path are green; the installed binary reaches the closed transport command
and the connection route stops at missing credentials before provider I/O. `update_customer` remains not
enabled: its required `customers(id) → customerId` binding cannot coexist with the conversation binding
for one declarative source executor. That is the open provider-neutral
`declarative-typed-destination-action-specific-source-bindings` gap, not a connector workaround. The
inventory and binary-ledger defect remain open.
