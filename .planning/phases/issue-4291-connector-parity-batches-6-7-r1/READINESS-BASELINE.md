# Seven-surface readiness baseline — issue #4291

This is the credential-free before-state after merging the typed destination
foundation. It derives from the source-locked ledgers and current bundle files;
it deliberately distinguishes user reachability, auto-exercise safety, and
provider-live certification. The machine-readable source is
`READINESS-BASELINE.json`.

| Connector | Inventory | Documented | Direct read command / documented | Typed write / documented | ETL source / streams | Reverse destination | Binary CLI / ledger | Deletes | Implemented CLI commands |
| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| close-com | complete | 300 | 0 / 123 | 12 / 163 | 0 / 14 | 0 | 0 / 0 | 52 | 0 |
| outreach | complete | 259 | 0 / 0 | 163 / 163 | 0 / 96 | 0 | 0 / 0 | 36 | 0 |
| salesloft | complete | 211 | 0 / 115 | 0 / 91 | 0 / 5 | 0 | 0 / 0 | 18 | 0 |
| copper | complete | 89 | 0 / 32 | 0 / 52 | 0 / 5 | 0 | 0 / 0 | 11 | 0 |
| zoho-bigin | complete | 75 | 0 / 26 | 6 / 43 | 0 / 13 | 0 | 0 / 0 | 11 | 0 |
| klaviyo | complete | 345 | 0 / 198 | 0 / 141 | 0 / 6 | 0 | 0 / 0 | 30 | 0 |
| braze | unproven | 95 | 0 / 21 | 29 / 53 | 0 / 21 | 0 | 0 / 0 | 4 | 0 |
| customer-io | complete | 166 | 0 / 82 | 10 / 68 | 0 / 16 | 0 | 0 / 0 | 14 | 0 |
| intercom | complete | 231 | 0 / 103 | 0 / 123 | 0 / 5 | 0 | 0 / 0 | 31 | 0 |
| freshdesk | complete | 170 | 0 / 73 | 0 / 92 | 0 / 5 | 0 | 0 / 0 | 22 | 0 |
| segment | complete | 201 | 0 / 97 | 0 / 101 | 0 / 3 | 0 | 0 / 0 | 28 | 0 |
| activecampaign | complete | 296 | 0 / 128 | 0 / 157 | 0 / 11 | 0 | 0 / 0 | 50 | 0 |
| iterable | complete | 148 | 0 / 49 | 0 / 96 | 0 / 3 | 0 | 0 / 0 | 12 | 0 |
| help-scout | unproven | 144 | 50 / 55 | 65 / 65 | 0 / 24 | 0 | 1 / 0 | 18 | 139 |
| gorgias | complete | 114 | 41 / 42 | 61 / 68 | 1 / 4 | 1 (App dispatch pending) | 1 / 1 | 18 | 94 |
| service-now | dynamic templates | 6 | 0 / 1 | 2 / 4 | 0 / 3 | 0 | 0 / 0 | 1 | 0 |
| chatwoot | complete | 148 | 32 / 57 | 60 / 84 | 0 / 7 | 0 | 0 / 0 | 18 | 100 |
| chargebee | complete | 527 | 0 / 128 | 36 / 367 | 0 / 32 | 0 | 0 / 0 | 0 | 0 |
| square | complete | 334 | 0 / 117 | 0 / 213 | 0 / 4 | 0 | 0 / 0 | 27 | 0 |
| braintree | unproven | 73 | 0 / 18 | 0 / 45 | 0 / 10 | 0 | 0 / 0 | 6 | 0 |

Totals after the Gorgias increment: 3,932 documented operations; 1,465 direct reads with 123 exact
command bindings; 2,189 direct writes with 444 exact typed-action bindings; 287 streams, one source
transport, one typed-destination declaration pending #4304 persisted App/CLI dispatch, 424 documented
deletes, and zero provider-live certifications.

Help Scout still has a binary-ledger classification defect. Gorgias now maps its
implemented binary command as one binary source row. No operation may stay
unreachable because it is destructive, privileged, uncommon, binary, unsafe to
auto-exercise, or pending provider-live certification.

## Increment — Gorgias declarative destination proof

The `tickets → update_ticket` mapping carries only the exact `id` and `status` fields and all
three closed modes. It is the initial declaration proof, not connector completion: 59 ordinary
typed actions await #4304's closed exact-action selection, and `upload_file` has a named binary/
multipart semantic exclusion while remaining reachable as `pm gorgias files upload`. The installed
binary reaches both `tickets messages list` and `tickets update` before stopping at the expected
credential preflight; no provider credentials or provider calls were used.
