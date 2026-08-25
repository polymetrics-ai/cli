# Issue #4292 provider-surface audit

Generated 2026-08-19T16:55:40.991Z. Each source-derived total below is the refreshed provider operation inventory; it is not derived from the previous api_surface.json. A no-change count is stated explicitly.

| Batch | Connector | Old api_surface count | New provider-operation count | Basis | Result |
| --- | --- | ---: | ---: | --- | --- |
| 8 | brex | 108 | 108 | machine-readable spec | no change |
| 8 | zoho-books | 848 | 838 | machine-readable spec | regenerated; 1 runtime binding row(s) excluded |
| 8 | testrail | 106 | — | no-public-api-description; browser | not statically countable |
| 8 | amplitude | 48 | 187 | complete rendered reference | regenerated |
| 8 | posthog | 9 | 1943 | machine-readable spec | regenerated |
| 8 | metabase | 5 | 634 | machine-readable spec | regenerated; 5 runtime binding row(s) excluded |
| 8 | dbt | 52 | 52 | machine-readable spec | no change |
| 8 | looker | 15 | 433 | machine-readable spec | regenerated |
| 8 | mode | 5 | 94 | machine-readable spec | regenerated; 5 runtime binding row(s) excluded |
| 8 | dremio | 63 | 49 | complete rendered reference | regenerated; 5 runtime binding row(s) excluded |
| 9 | coda | 45 | 124 | machine-readable spec | regenerated |
| 9 | clickup-api | 172 | 173 | complete rendered reference | regenerated |
| 9 | calendly | 35 | 61 | machine-readable spec | regenerated |
| 9 | greenhouse | 138 | — | no-public-api-description; browser | not statically countable |
| 9 | lever-hiring | 106 | 104 | complete rendered reference | regenerated; 2 runtime binding row(s) excluded |
| 9 | ashby | 212 | 193 | complete rendered reference | regenerated |
| 9 | workable | 82 | 84 | complete rendered reference | regenerated; 80 runtime binding row(s) excluded |
| 9 | recruitee | 948 | 938 | complete rendered reference | regenerated |
| 9 | hibob | 6 | 207 | complete rendered reference | regenerated; 3 runtime binding row(s) excluded |
| 9 | factorial | 6 | 155 | complete rendered reference | regenerated; 5 runtime binding row(s) excluded |
| 10 | datadog | 235 | 1739 | machine-readable spec | regenerated |
| 10 | pagerduty | 10 | 465 | machine-readable spec | regenerated |
| 10 | auth0 | 75 | 469 | machine-readable spec | regenerated |
| 10 | okta | 732 | 734 | machine-readable spec | regenerated |
| 10 | firehydrant | 479 | 373 | machine-readable spec | regenerated; 103 runtime binding row(s) excluded |
| 10 | adobe-commerce-magento | 76 | — | dynamic instance-dependent | not statically countable |
| 10 | commercetools | 3 | 821 | machine-readable spec | regenerated |
| 10 | recharge | 9 | 123 | complete rendered reference | regenerated |
| 10 | docuseal | 23 | 23 | machine-readable spec | no change |
| 10 | eventbrite | 10 | — | no-public-api-description; browser | not statically countable |

`new provider-operation count` excludes any retained runtime-only coverage rows noted in the Result column. They preserve existing conformance while a behavior change updates or retires a binding that is not in the current provider inventory; they are never counted as evidence for a provider REST operation. Adobe Commerce is instance/module-dependent; TestRail, Eventbrite, and Greenhouse have no current credential-free public API description, so their totals are deliberately not fabricated.
