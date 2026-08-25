# Issue #4292 declaration-first direct-write cohort 01

This mechanical cohort preserves source identities and existing action CLI paths. It does not create or infer a provider request, response, pagination, body schema, or CLI spelling. `source-lock-projection-gap` blocks source certification uniformly; it does not turn connector-local declaration work into an engine gap.

| Batch | Connector | Direct-write source operations | Existing-schema CLI-bound | Declaration-pending | Source status |
| --- | --- | ---: | ---: | ---: | --- |
| 8 | brex | 49 | 14 | 35 | enumerable |
| 8 | zoho-books | 579 | 562 | 17 | enumerable |
| 8 | testrail | — | — | — | skipped: no-public-api-description |
| 8 | amplitude | 99 | 12 | 87 | enumerable |
| 8 | posthog | 1134 | 0 | 1134 | enumerable |

All enumerable rows name the same deferred source-certification component: `source-lock-projection-gap`. Existing command paths are reported verbatim from `cli_surface.json`; missing rows deliberately contain no intended path until a bounded connector-owned typed contract exists.
