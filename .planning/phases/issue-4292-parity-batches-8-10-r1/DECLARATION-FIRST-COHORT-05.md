# Issue #4292 declaration-first direct-write cohort 05

This mechanical cohort preserves source identities and existing action CLI paths. It does not create or infer a provider request, response, pagination, body schema, or CLI spelling. `source-lock-projection-gap` blocks source certification uniformly; it does not turn connector-local declaration work into an engine gap.

| Batch | Connector | Direct-write source operations | Existing-schema CLI-bound | Declaration-pending | Source status |
| --- | --- | ---: | ---: | ---: | --- |
| 10 | datadog | 989 | 27 | 962 | enumerable |
| 10 | pagerduty | 254 | 0 | 254 | enumerable |
| 10 | auth0 | 270 | 8 | 262 | enumerable |
| 10 | okta | 444 | 429 | 15 | enumerable |
| 10 | firehydrant | 204 | 190 | 14 | enumerable |

All enumerable rows name the same deferred source-certification component: `source-lock-projection-gap`. Existing command paths are reported verbatim from `cli_surface.json`; missing rows deliberately contain no intended path until a bounded connector-owned typed contract exists.
