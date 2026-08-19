# Provider-surface audit — issue #4291

The source-lock recovery audit covers every connector owned by this issue. A row is complete only when its `new` count comes from an authoritative machine-readable specification, a complete rendered-reference crawl, or an explicit dynamic-instance basis.

| Connector | Legacy `api_surface` count | New count | Basis | Audit state |
| --- | ---: | ---: | --- | --- |
| close-com | 297 | 300 | provider-published Close OpenAPI 3.1.0 (137 GET, 64 POST, 47 PUT, 52 DELETE) | complete — remapped |
| outreach | 259 | 259 | complete machine-readable OpenAPI plus provider-rendered fixed generic custom-object supplement; per-account object schemas are dynamic | complete — remapped, count unchanged |
| salesloft | 12 | 211 | complete rendered reference: all 315 public API pages in provider sitemap | complete — remapped |
| copper | 5 | 89 | complete rendered reference: all 637 provider-published MkDocs search-index documentation nodes | complete — remapped |
| zoho-bigin | 50 | 75 | complete rendered reference: all 98 Bigin v2 pages in the provider sitemap, including the separate `/bigin/bulk/v2` endpoint family | complete — remapped |
| klaviyo | 9 | 345 | official OpenAPI 3.0.2 full GA specification | complete — remapped |
| braze | 95 | — | pending provider-surface audit | pending |
| customer-io | 159 | 166 | official OpenAPI 3.1.0 application API specification | complete — remapped |
| intercom | 10 | 231 | official Intercom OpenAPI 3.0.1 v2.16 specification | complete — remapped |
| freshdesk | 10 | 170 | complete rendered reference: all 171 endpoint sections in the 3.2MB reference, normalizing to 170 HTTP operations | complete — remapped |
| segment | 188 | 201 | provider-published OpenAPI 3.0.3 embedded in the versioned Segment Redoc state artifact (100 GET, 43 POST, 23 PATCH, 7 PUT, 28 DELETE) | complete — remapped; legacy `/workspaces` list stream removed in favour of singleton `GET /` |
| activecampaign | 61 | — | pending provider-surface audit | pending |
| iterable | 4 | 148 | official Iterable Swagger 2.0 specification | complete — remapped |
| help-scout | 144 | — | pending provider-surface audit | pending |
| gorgias | 114 | 114 | official OpenAPI 3.1.0 specification | complete — remapped, count unchanged |
| service-now | 22 | — | pending fixed-platform and dynamic-schema basis | pending |
| chatwoot | 148 | 148 | official OpenAPI 3.1.0 specification | complete — remapped, count unchanged |
| chargebee | 428 | 527 | official Chargebee v2 SDK OpenAPI 3.1.0, the complete public surface used to generate provider client libraries (160 GET, 367 POST) | complete — remapped |
| square | 11 | 334 | official OpenAPI 3.0.0 machine-readable specification; rendered crawl retained only as partial/superseded evidence | complete — remapped |
| braintree | 73 | — | pending provider-surface audit | pending |

The legacy counts are the counts before remapping; Salesloft’s legacy count was 12 even though its source lock had already been updated to 211 in the first recovery step. They must not be treated as provider coverage.
