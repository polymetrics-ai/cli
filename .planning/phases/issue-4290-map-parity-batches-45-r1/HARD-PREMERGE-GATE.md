# Issue #4290 hard pre-merge gate

Status: **blocked**. This is a fail-closed operation ledger, not a claim that the current declaration maps are deployable.

- Enumerable provider operations: 7301
- Connectors with dynamic or unavailable provider inventories: salesforce, tiktok-marketing, ebay-fulfillment
- Runtime reachability proven: 0
- Generated CLI commands proven: 0
- Generated website rows proven: 0
- Executable fixture/conformance proofs: 0
- Provider-deprecated/absent operations with direct source evidence: 31
- Operations blocked by an open foundation gap: 0

Every enumerable provider operation has its locked source URL, source-document locator/hash/bytes/version status, canonical API-surface mapping, generated-command disposition, website disposition, fixture/conformance disposition, runtime control metadata, and separate ETL/reverse-ETL/direct-read/direct-write/binary-download/binary-upload reconciliation in `HARD-PREMERGE-GATE.json`.

`not-asserted`, `pending`, `declared-unproven`, and a source version that was not materialized are all pending evidence—not N/A. The only provider-capability absence status is backed by the source record’s deprecation/absence metadata. Scope, tier, destructive, risk, and confirmation requirements remain typed runtime controls and never turn an otherwise supported operation into an exclusion.

Final push remains paused until the shared foundation publishes its App/CLI generic-destination dispatch integration, is merged locally, is proven as an ancestor, and passes the real installed App/CLI-path exercise. This gate also remains blocked until each operation has the missing reachability, generated website, and executable conformance evidence.

| Connector | Provider inventory state | Enumerable operations | Local bindings outside provider inventory |
| --- | --- | ---: | ---: |
| salesforce | browser-rendered | 0 | 10 |
| hubspot | fetched | 3118 | 0 |
| pipedrive | fetched | 213 | 5 |
| mailchimp | browser-rendered | 295 | 28 |
| zendesk-support | fetched | 629 | 6 |
| quickbooks | fetched | 129 | 5 |
| bamboo-hr | fetched | 319 | 26 |
| airtable | fetched | 103 | 2 |
| google-analytics-data-api | fetched | 23 | 5 |
| woocommerce | fetched | 140 | 0 |
| pinterest | fetched | 279 | 5 |
| tiktok-marketing | skipped | 0 | 7 |
| linear | browser-rendered | 539 | 4 |
| buildkite | fetched | 129 | 3 |
| sonar-cloud | fetched | 156 | 1 |
| launchdarkly | fetched | 397 | 2 |
| fastly | fetched | 732 | 0 |
| squarespace | fetched | 53 | 0 |
| ebay-fulfillment | skipped | 0 | 11 |
| shipstation | fetched | 47 | 0 |
