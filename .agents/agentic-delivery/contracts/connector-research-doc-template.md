# Connector API-surface research doc template

`pm-web-researcher` produces this during the RESEARCH stage for a connector task. It is the durable
input `pm-planner` turns into `api_surface.json` / `cli_surface.json` and the 7 parity sub-issues.
Write both a human `RESEARCH.md` (this shape) and a machine `RESEARCH.json`
(`.agents/agentic-delivery/schemas/connector-research.schema.yaml`) under
`.planning/auto-loop/RESEARCH/<connector>/`.

The rule the whole loop depends on: **every provider API operation must appear here, classified,
with an official `source_url`.** `unclassified_endpoints` must be `0` and `complete` must be `true`
before PARENT_PLAN starts.

```markdown
# <connector> — API surface research

## Provider
- name: <e.g. Twenty CRM>
- docs_root: <official docs URL>
- base_urls: [<REST base>, <GraphQL endpoint>]
- api_styles: [rest, graphql]         # whichever the provider actually exposes
- auth: { scheme: <bearer|oauth2|api-key|app>, header: <e.g. Authorization: Bearer>, scopes: [...] }
- rate_limits: <known limits / pagination caps, with source_url>

## Read endpoints (each → an ETL stream or direct_read)
| object | method | path / graphql_op | pagination | cursor_field | execution_model | source_url |
|--------|--------|-------------------|------------|--------------|-----------------|------------|
| people | GET    | /rest/people      | cursor(pageInfo) | updatedAt | stream_read | https://... |
| person | GET    | /rest/people/{id} | none       | —            | direct_read     | https://... |

## Write verbs (each → a reverse-ETL action)
| action | method | path | risk_tier | execution_model | source_url |
|--------|--------|------|-----------|-----------------|------------|
| create_person | POST | /rest/people | normal | reverse_etl | https://... |
| delete_person | DELETE | /rest/people/{id} | destructive | destructive_admin | https://... |

## execution_model vocabulary (from docs/migration/conventions.md — do not invent)
stream_read · direct_read · reverse_etl · sensitive_reverse_etl · admin_reverse_etl ·
destructive_admin · graphql_query · binary_transfer · local_workflow · excluded:<reason>

## Coverage self-check (must pass before planning)
- read_endpoints_found: <N>
- write_verbs_found: <M>
- unclassified_endpoints: 0        # MUST be 0
- all_source_urls_present: true    # every row above has an official source_url
- complete: true                   # set false (and explain) if the full surface could not be confirmed

## Notes / risks
- <REST vs GraphQL choice per object and why>
- <endpoints intentionally excluded and the reason>
- <auth/pagination quirks the implementer must handle>
```
