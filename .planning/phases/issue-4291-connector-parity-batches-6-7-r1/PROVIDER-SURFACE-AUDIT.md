# Provider-surface audit — issue #4291

The source-lock recovery audit covers every connector owned by this issue. A row is complete only when its `new` count comes from an authoritative machine-readable specification, a complete rendered-reference crawl, or an explicit dynamic-instance basis.

| Connector | Legacy `api_surface` count | New count | Basis | Audit state |
| --- | ---: | ---: | --- | --- |
| close-com | 297 | — | pending provider-surface audit | pending |
| outreach | 259 | 259 | complete machine-readable OpenAPI plus provider-rendered fixed generic custom-object supplement; per-account object schemas are dynamic | complete — remapped, count unchanged |
| salesloft | 12 | 211 | complete rendered reference: all 315 public API pages in provider sitemap | complete — remapped |
| copper | 5 | 89 | complete rendered reference: all 637 provider-published MkDocs search-index documentation nodes | complete — remapped |
| zoho-bigin | 50 | — | pending provider-surface audit | pending |
| klaviyo | 9 | 345 | official OpenAPI 3.0.2 full GA specification | complete — remapped |
| braze | 95 | — | pending provider-surface audit | pending |
| customer-io | 159 | 166 | official OpenAPI 3.1.0 application API specification | complete — remapped |
| intercom | 10 | 231 | official Intercom OpenAPI 3.0.1 v2.16 specification | complete — remapped |
| freshdesk | 10 | 170 | complete rendered reference: all 171 endpoint sections in the 3.2MB reference, normalizing to 170 HTTP operations | complete — remapped |
| segment | 188 | — | pending provider-surface audit | pending |
| activecampaign | 61 | — | pending provider-surface audit | pending |
| iterable | 4 | 148 | official Iterable Swagger 2.0 specification | complete — remapped |
| help-scout | 144 | — | pending provider-surface audit | pending |
| gorgias | 114 | — | pending provider-surface audit | pending |
| service-now | 22 | — | pending fixed-platform and dynamic-schema basis | pending |
| chatwoot | 148 | — | pending provider-surface audit | pending |
| chargebee | 428 | — | pending provider-surface audit | pending |
| square | 11 | 334 | official OpenAPI 3.0.0 machine-readable specification; rendered crawl retained only as partial/superseded evidence | complete — remapped |
| braintree | 73 | — | pending provider-surface audit | pending |

The legacy counts are the counts before remapping; Salesloft’s legacy count was 12 even though its source lock had already been updated to 211 in the first recovery step. They must not be treated as provider coverage.
