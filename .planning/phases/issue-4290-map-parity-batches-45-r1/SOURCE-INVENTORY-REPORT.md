# Issue #4290 source-inventory report

Each provider total is sourced from the locked public description, not from the pre-existing bundle. `local bindings` are retained connector execution identities and are deliberately excluded from `operations found`.

| Connector | Old API-surface rows | Operations found | New API-surface rows | Local bindings | Retrieval | Coverage confidence and basis |
| --- | ---: | ---: | ---: | ---: | --- | --- |
| salesforce | 10 | unknown | 10 | 10 | browser-rendered | dynamic-instance-surface: Salesforce REST object resources and actions vary by tenant configuration. The public resource index is a stable generic reference but cannot settle a tenant-independent total. |
| hubspot | 3118 | 3118 | 3118 | 0 | fetched | machine-readable-spec: Official HubSpot public API-spec collection at the pinned provider commit. |
| pipedrive | 218 | 213 | 218 | 5 | fetched | machine-readable-spec: Official Pipedrive v1 OpenAPI document; count is HTTP method plus path. |
| mailchimp | 298 | 295 | 323 | 28 | browser-rendered | machine-readable-spec: Mailchimp Swagger root plus all 181 provider-owned path-item JSON documents were retrieved through chrome-devtools-axi; source bytes and SHA-256 are pinned per retrieved document. |
| zendesk-support | 631 | 629 | 635 | 6 | fetched | machine-readable-spec: Official Zendesk Support OpenAPI document; count is HTTP method plus path. |
| quickbooks | 11 | 129 | 134 | 5 | fetched | machine-readable-spec: The public Intuit API Explorer entity document enumerates 74 QuickBooks Online entities and 129 unique normalized HTTP method/path operations. |
| bamboo-hr | 340 | 319 | 345 | 26 | fetched | complete-rendered-reference: Official BambooHR rendered API reference embeds its complete 3.1 schema; 319 HTTP method/path operations are extracted from it. |
| airtable | 30 | 103 | 105 | 2 | fetched | complete-rendered-reference: Official Airtable rendered Web API reference embeds its complete 3.1 schema with 103 HTTP method/path operations. |
| google-analytics-data-api | 5 | 23 | 28 | 5 | fetched | machine-readable-spec: Official v1, v1alpha, and v1beta Discovery documents, deduped by HTTP method and path (23 unique provider requests). |
| woocommerce | 10 | 140 | 140 | 0 | fetched | complete-rendered-reference: Official WooCommerce REST v3 reference has 140 unique normalized method/path request examples after the duplicated root query variant is deduplicated. |
| pinterest | 12 | 279 | 284 | 5 | fetched | complete-rendered-reference: Official Pinterest v5 rendered API reference embeds 297 navigation entries; 279 unique documented HTTP method/path requests remain after duplicate navigation entries are deduplicated. |
| tiktok-marketing | 7 | unknown | 7 | 7 | skipped | unavailable-public-source: Chrome returned ERR_SSL_PROTOCOL_ERROR; no public API description could be retrieved. |
| linear | 7 | 539 | 543 | 4 | browser-rendered | complete-rendered-reference: Chrome-rendered public Apollo Linear schema reports 166 Query and 373 Mutation roots. Subscription roots are server-push schema members, not callable request operations. |
| buildkite | 99 | 129 | 132 | 3 | fetched | complete-rendered-reference: Official Buildkite rendered REST reference table has 129 unique HTTP method/path rows. |
| sonar-cloud | 157 | 156 | 157 | 1 | fetched | machine-readable-spec: Official SonarCloud web-services catalog lists 156 public actions. |
| launchdarkly | 7 | 397 | 399 | 2 | fetched | machine-readable-spec: Official LaunchDarkly OpenAPI endpoint documents 397 HTTP method/path operations. |
| fastly | 54 | 732 | 732 | 0 | fetched | machine-readable-spec: Official Fastly Postman collection contains 740 named requests; 732 unique normalized HTTP method/path operations remain after duplicate request examples are deduplicated. |
| squarespace | 47 | 53 | 53 | 0 | fetched | machine-readable-spec: Official Squarespace Commerce OpenAPI document has 53 HTTP method/path operations. |
| ebay-fulfillment | 11 | unknown | 11 | 11 | skipped | unavailable-public-source: Chrome returned the official eBay error page rather than API documentation. |
| shipstation | 9 | 47 | 47 | 0 | fetched | machine-readable-spec: Official ShipStation V1 documentation download is an OpenAPI document with 47 HTTP method/path operations. |
