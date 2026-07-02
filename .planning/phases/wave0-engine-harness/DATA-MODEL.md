# DATA-MODEL — wave0-engine-harness

The phase introduces file-based data contracts only (no DB/state schema changes). Source of truth:
`docs/architecture/connector-architecture-v2-design.md` §A (layout + examples) and §B.2 (Go
types). Machine-checkable contracts ship as meta-schemas at
`internal/connectors/engine/schema/{metadata,spec,streams,writes,api_surface}.schema.json`
(written in the engine's minimal draft-07 dialect; enforced by `connectorgen validate` and the
bundle loader).

## 1. Bundle directory contract (`internal/connectors/defs/<name>/`)

| File | Required | Notes |
|---|---|---|
| `metadata.json` | yes | identity, capabilities, batch, rate_limit, risk (design §A example) |
| `spec.json` | yes | pure JSON Schema draft-07 object schema; `x-secret: true` marks secret properties — the ONLY config/secret split |
| `streams.json` | yes, EXCEPT when `metadata.capabilities.dynamic_schema == true` (postgres) | `{"base": HTTPBase, "streams": [StreamSpec…]}` |
| `writes.json` | optional | `{"actions": [WriteAction…]}` |
| `api_surface.json` | yes | §E.1 rules 1–5; goldens may use `out_of_scope` exclusions with `scope` prose in wave0 |
| `schemas/<stream>.json` | one per declared stream | draft-07 + `x-primary-key` (array), `x-cursor-field` (string); doubles as the record projection |
| `fixtures/streams/<stream>/page_N.json` | first stream mandatory | replay envelope, §5 below |
| `fixtures/writes/<action>.json` | per write action | `{"record": {...}, "expect": {"method","path","body"|"form"}}` |
| `docs.md` | yes | fixed headings: Overview, Auth setup, Streams notes, Write actions & risks, Known limits |

Naming: dir name == `metadata.name` == registry key, regex `^[a-z0-9][a-z0-9-]*$`. Streams and
actions `snake_case`; actions verb-first.

## 2. Key JSON shapes (abridged; full field lists in design §B.2 / API-CONTRACT.md)

`streams.json.base` (HTTPBase): `url` (template), `user_agent`, `headers` (templated values;
empty-after-interpolation ⇒ omitted), `auth` (ordered AuthSpec list with `when`), `pagination`
(default PaginationSpec), `check` ({method,path}), `error_map`
(`[{status, match_body?, class?, hint?}]`), `rate_limit?` ({requests_per_minute}).

`StreamSpec`: `name`, `method?` (default GET), `path` (template), `query?` (templated map),
`body?` (POST-body streams), `records` ({path, single_object?, filter?{field_absent|field_equals}}),
`pagination?` (overrides base), `incremental?` ({cursor_field, request_param?, param_format?
(rfc3339|unix_seconds|date|github_date_range), start_config_key?, client_filtered?}),
`computed_fields?` (name → template), `projection?` ("schema" default | "passthrough"),
`schema` (relative ref, e.g. `schemas/issues.json`).

`PaginationSpec` (wave0 dialect — extends design examples to cover stripe):
```json
{ "type": "none|link_header|page_number|offset_limit|cursor|next_url",
  "size_param": "per_page", "page_param": "pageno", "start_page": 1,
  "limit_param": "limit", "offset_param": "offset",
  "cursor_param": "starting_after",
  "token_path": "meta.next_cursor",         // cursor: token from body
  "last_record_field": "id",                 // cursor: token from last record (stripe)
  "stop_path": "has_more",                   // cursor: falsy body value stops (stripe)
  "next_url_path": "meta.next_page_link",    // next_url type
  "page_size": 100, "max_pages": 0 }
```
Constraint: `cursor` requires exactly one of `token_path` | `last_record_field`.

`WriteAction`: `name`, `kind` (create|update|upsert|delete|custom), `method`, `path` (template),
`path_fields?`, `body_type?` (json default|form|none), `body_fields?`, `record_schema`
(inline draft-07), `delete?` ({idempotent, missing_ok_status[]}), `risk`,
`confirm?` (""|"destructive"), `hook?`.

`api_surface.json`: `api`, `docs`, `reviewed_at`, `scope`, `endpoints[]` each with exactly one of
`covered_by{stream|write}` XOR `excluded{category, reason}`; categories closed vocabulary:
`destructive_admin, requires_elevated_scope, binary_payload, deprecated, non_data_endpoint,
duplicate_of, out_of_scope`.

## 3. Golden bundle specifics

- **stripe**: base url `{{ config.base_url }}` (default `https://api.stripe.com/v1`), bearer
  `{{ secrets.client_secret }}`, optional `Stripe-Account: {{ config.account_id }}` header
  (omitted when empty); 5 streams (customers, charges, invoices, subscriptions, products), each
  `records.path: "data"`, pagination `cursor` + `last_record_field: id` + `stop_path: has_more` +
  `limit_param: limit` + `page_size: 100`; incremental `{cursor_field: created,
  request_param: "created[gte]", param_format: unix_seconds, start_config_key: start_date}`;
  writes `create_customer` (POST `customers`, body_type form, record_schema properties
  email/name/description/phone, minProperties 1 — documented deviation ledger) and
  `update_customer` (POST `customers/{{ record.id }}`, path_fields [id], body_type form).
- **searxng**: base url `{{ config.base_url }}` (required), optional bearer
  `{{ secrets.token }}` with `when` truthiness; streams `search` and `reddit` both GET `/search`,
  `records.path: "results"`, query `{"q": <templated>, "format": "json"}` where reddit scopes
  `{{ config.query }} site:reddit.com` (+ optional subreddit); pagination `page_number`
  `{page_param: pageno, start_page: 1, page_size: 10, max_pages: 1}` with NO size param sent;
  schemas PK `["url"]`, `x-cursor-field: published_date`.
- **postgres**: `dynamic_schema: true`, NO streams.json; spec host/port/database/username/
  sslmode/schema + `password` x-secret; served by Tier-3 `internal/connectors/native/postgres/`
  via `engine.Base`.

## 4. Migration program artifacts

- `docs/migration/inventory.json`:
  `{"generated_at", "connectors": [{"name","path","loc","runtime_kind","bucket",
  "catalog_slugs":[],"documentation_url","stream_count"}]}`.
- `docs/migration/result.schema.json` (agent output): per orchestration-plan §Per-agent task spec —
  `{name, status: migrated|partial|blocked, files_changed[], streams_before, streams_after,
  write_actions_added, escape_hatches[{file,reason}], fixtures_added,
  conformance{passed,failing_tests[]}, blockers[{type,reason,evidence}], notes}`.
- `docs/migration/review.schema.json` (reviewer verdict): `{connector, verdict: pass|fail,
  findings[{severity, rule, file, detail}], checked: {schema_fidelity, write_actions,
  fixture_realism, escape_hatch_justification, secret_redaction, surface_completeness}}`.

## 5. Fixture replay envelope (`fixtures/streams/<stream>/page_N.json`)

```json
{
  "request":  { "method": "GET", "path": "/v1/customers",
                "query": { "limit": "100", "starting_after": "cus_2" } },
  "response": { "status": 200,
                "headers": { "Link": "<...>; rel=\"next\"" },
                "body": { "data": [ { "...": "sanitized records" } ], "has_more": true } }
}
```
The conformance replay server matches incoming requests on method + path + exact query multiset;
an unmatched request fails the run; each page must be consumed exactly once
(`pagination_terminates`). All fixture values are synthetic (THREAT-MODEL §4).

## 6. Certification report

`.polymetrics/certifications/<connector>.json` — exact shape in
`docs/architecture/connector-certification-design.md` §A ("Report artifact"); wave0 populates
`kind, schema_version, connector, pm_version, started_at, completed_at, mode, passed,
capabilities.{check,catalog,read,sync_modes,resume,query/json_contract,secret_redaction}` and
`stages[]`; `leaks`, `write_actions`, `flow`, `schedule`, `budget` remain empty/absent until the
later certify phases. History: `.polymetrics/certifications/history/<connector>/<timestamp>.json`.
