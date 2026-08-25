# Issue #4292 declaration-first direct-write cohort 02

This mechanical cohort preserves source identities and existing action CLI paths. It does not create or infer a provider request, response, pagination, body schema, or CLI spelling. `source-lock-projection-gap` blocks source certification uniformly; it does not turn connector-local declaration work into an engine gap.

| Batch | Connector | Direct-write source operations | Existing-schema CLI-bound | Declaration-pending | Source status |
| --- | --- | ---: | ---: | ---: | --- |
| 8 | metabase | 329 | 0 | 329 | enumerable |
| 8 | dbt | 26 | 13 | 13 | enumerable |
| 8 | looker | 222 | 0 | 222 | enumerable |
| 8 | mode | 45 | 0 | 45 | enumerable |
| 8 | dremio | 31 | 10 | 21 | enumerable |

All enumerable rows name the same deferred source-certification component: `source-lock-projection-gap`. Existing command paths are reported verbatim from `cli_surface.json`; missing rows deliberately contain no intended path until a bounded connector-owned typed contract exists.
