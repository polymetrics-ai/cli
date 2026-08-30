# Discussion log

## Resolved source-mapping questions

| Question | Decision | Locked evidence |
| --- | --- | --- |
| What is the denominator? | 49 retained `rest.operations` rows. | `sources/notion-operation-source-lock.json:counts.total` |
| Are the two deprecated database entries rows? | No; they are crosswalk surface-only boundary evidence. | `sources/notion-operation-crosswalk.json:surface_only` |
| Are POST query operations mutations? | No; three query/search endpoints and token introspection remain semantic reads. | Each locked operation summary and request/response contract |
| Is `query-meeting-notes` an ETL candidate? | Yes, as a collection read; it is restricted because source has `results`/`has_more` and a bounded `limit`, but no continuation input. | `notion.rest.query-meeting-notes` |
| Do webhook schema names create sync rows? | No; no retained operation has callbacks or a webhook-registration route. | Lock operation inventory and embedded source contract |

## Non-decisions

- No runtime execution, command availability, certification, importer repair, or destination transport is decided here.
- No source fact is inferred from a current JSON definition, fixture, live request, or generator output.
