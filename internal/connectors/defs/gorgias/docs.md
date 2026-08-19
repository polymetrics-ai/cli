# Gorgias

## Overview

Reads and writes Gorgias helpdesk tickets, customers, messages, satisfaction surveys, users, teams,
tags, views, macros, rules, widgets, integrations, jobs, account settings, custom fields, statistics,
voice calls, and files through the Gorgias REST API.

The documented surface is **114 operations**, re-derived on 2026-08-07 from the provider's stable
OpenAPI 3.1.0 document (`gorgias-rest-api.json`, bound to ReadMe docs version 1.5.1), fetched via
<https://dash.readme.com/api/v1/api-registry/1qfhqbgmshn434r>: 61 paths, 114 operations (46 GET, 23
POST, 27 PUT, 18 DELETE), 1 deprecated (still counted). The provider-artifact sweep had classified
this artifact as `html_reference` and carried forward a count of 114 from an older audit; both the
classification and the count are corrected here — `developers.gorgias.com/reference` is a
ReadMe.com-hosted docs site whose stable version publishes exactly one registered OpenAPI document,
and that document's own operation count reconciles exactly with the ledger's carried-forward figure.

There is no top-level `webhooks` block in the document and Gorgias publishes no webhook management
endpoints, so there is nothing in that area to implement.

Every documented operation is partitioned exactly once in `api_surface.json` and carries exactly one
disposition — executable, blocked with a named dependency, or not executable with a source citation.
None is blank.

## Auth setup

Connection fields:

- `base_url` (required, string, format `uri`) — the account's full API root, e.g.
  `https://<domain>.gorgias.com/api`. The trailing `/api` segment matters: every stream, write, and
  operation path in this bundle is declared relative to it.
- `username` (required, string) — the Gorgias account email used for HTTP Basic auth.
- `password` (required, secret, string) — the Gorgias API key used for HTTP Basic auth; never logged.
- `page_size` (optional, string, default `100`) — records per page (1-100).
- `mode` (optional, string) — `live` (default) or `fixture` for credential-free conformance.

Add the credential from an environment variable or stdin, never as prompt text:

```
pm credentials add gorgias --from-env GORGIAS_PASSWORD
```

Connection checks call `GET /tickets` with query `limit=1`.

## Streams notes

Four ETL streams, unchanged from the connector's prior read-only scope: `tickets`, `customers`,
`messages`, `satisfaction_surveys`. All four use Gorgias's standard cursor pagination (`cursor` query
parameter, next token from `meta.next_cursor`), declared once in `streams.json`'s base and shared by
every stream.

The remaining 40 read-only GET endpoints and 2 read-only POST query endpoints (`search`,
`stats get`) are implemented as **bounded direct reads** rather than additional streams — each is a
single-shot, flag-driven command (`pm gorgias <resource> get`/`list`), not a full incremental sync
target. This mirrors the sweep's own precedent (`notion`'s `comment list`, `data-source templates
list`, etc.): a "list" endpoint does not have to become a stream to be fully documented and
executable, and treating this connector's config/admin-shaped collections (account settings,
integrations, jobs, macros, rules, tags, teams, users, views, widgets, voice call data, and so on) as
bounded reads keeps 42 additional endpoints reachable without 40+ new schema/fixture files.

`stats get` (`POST /api/stats/{name}`) satisfies its required `filters` body key with a static empty
object default (`rest.body: {"filters": {}}`) — the same technique gong's `users_extensive` operation
uses — so the command is fully implemented with just the required `--name` flag.

## Write actions & risks

61 write actions (60 mutations plus one multipart file upload). Every mutation goes through plan →
preview → approval → execute; 19 are destructive (every `DELETE`, plus the file upload, which
requires `--confirm destructive` because it sends local file bytes to Gorgias).

15 actions are `partial`, not `implemented`: each has a required field with no typed scalar leaf (an
array, a free-form object, or a provider-defined discriminated union), so no honest CLI flag contract
exists for that field. Each keeps its full typed `record_schema` in `writes.json` and stays usable
through reverse ETL from a source record; only the direct flag-driven command form is partial.

| Action | Non-flaggable required field | Why |
| --- | --- | --- |
| `create_custom_field` | `definition` | discriminated union over the field's data type (text/dropdown/checkbox/etc.) |
| `update_custom_fields` | `custom_fields` | bare top-level array of custom field objects |
| `create_customer` | `channels` | array of channel objects |
| `delete_customers` | `ids` | array of integer customer IDs |
| `update_customer_custom_field_values` | `values` | bare top-level array of `{id, value}` objects |
| `update_customer_data` | `data` | arbitrary free-form document |
| `create_job` | `params` | job-type-specific object with no fixed shape |
| `bulk_archive_macros` / `bulk_unarchive_macros` | `ids` | array of integer macro IDs |
| `update_rules_priorities` | `priorities` | array of `{id, priority}` objects |
| `delete_tags` | `ids` | array of integer tag IDs |
| `merge_tags` | `source_tags_ids` | array of integer tag IDs |
| `create_ticket` | `messages` | array of message objects, each with its own required unioned `channel` |
| `update_ticket_custom_fields` | `values` | bare top-level array of `{id, value}` objects |
| `create_widget` | `template` | widget-type-specific rendering object with no fixed shape |

### Bulk bodies

`update_custom_fields`, `update_customer_custom_field_values`, and `update_ticket_custom_fields` each
send a **bare top-level JSON array** as the entire request body (`body_type: "json_array"`), not an
object — the same mechanism gong's `upload_crm_entity_schema` uses.

### The multipart upload

`POST /api/upload` is implemented as `upload_file`, a `writes.json` action with
`body_type: "multipart"` and a single `file` part — the same proven shape as gong's
`upload_call_media`/`upload_crm_entities`. It streams a project-relative local file (`--file-path`)
into the request; file path and content are redacted in plans.

### The binary download

`GET /api/{file_type}/download/{domain_hash}/{resource_name}` is implemented as `files download`, a
`binary_download` operation. The provider responds with an HTTP 307 redirect to signed, cross-host
file storage (not a direct 200 with bytes), so the operation declares `allow_cross_host: true`; the
Gorgias Basic-auth credential is never forwarded across that redirect. The response is written only
beneath an explicit `--dest-root`.

## Known limits

- **`from-agent` on message writes is a plain string flag.** The message `channel` field is
  documented as a discriminated union of channel-specific shapes (email/chat/phone/etc.); this bundle
  exposes it as a plain string flag naming the channel, the common and honest case, rather than
  splitting every channel shape into its own command.
- **`PUT /api/views/{view_id}/items` ("Search for view's items") is blocked with a named
  dependency.** It takes its filter/order criteria in a request body but requires PUT, and this
  repository's operation direct-read executor accepts only GET or POST. The common case — an
  unfiltered, cursor-paginated listing of a view's items — is already served by the GET sibling,
  implemented here as `views items list`.
- **`POST /api/stats/{name}/download` ("Download a statistic") is blocked with a named
  dependency.** It returns `text/csv` from a POST request, and this repository's binary download
  executor requires GET. There is no POST-capable binary/CSV download executor today.
- **`POST /api/reporting/stats` ("Retrieve a statistic") is blocked with a named dependency.** Its
  required `query` field is a 50-arm discriminated union implementing the Statistics API's recursive
  filter-expression grammar, with no typed scalar leaf and no neutral empty default — unlike
  `stats get`'s flat `filters` object, no static literal can honestly stand in for an arbitrary
  filter tree.
- **`PUT /api/customers/{customer_id}/custom-fields/{id}` and
  `PUT /api/tickets/{ticket_id}/custom-fields/{id}` ("Update customer/ticket field value") are
  blocked with a named dependency.** Each sends a **bare top-level JSON scalar** (string, integer, or
  boolean) as the entire request body, not an object of record fields. This repository's write
  `body_type` dialect (`json`/`form`/`json_array`/`multipart`/`base64_upload`) always produces an
  object or array wrapper and has no mode for a bare scalar body. The bulk siblings
  (`update_customer_custom_field_values`, `update_ticket_custom_fields`) cover the same data via a
  `{id, value}` array and are implemented.
- **`GET /api/tickets/{ticket_id}/messages` ("List messages of a ticket") is provider-deprecated,
  but executable.** The OpenAPI document marks it `deprecated: true` and recommends
  `GET /api/messages`; both remain reachable: the legacy route is `pm gorgias tickets messages list
  --ticket-id <id> --json`, while `messages` remains the ETL stream.
- **Declarative typed destination proof.** `sync_transport.json` declares every Gorgias ETL stream
  as a `declarative_stream_source` and binds `update_ticket` to `tickets` with the exact `id` and
  `status` source fields. The route is keyed, warehouse-acknowledged, and has explicit strategies
  for overwrite, append, and upsert modes. This is fixture/dry declaration evidence only: provider
  live certification and persisted App/CLI destination dispatch remain pending the #4304 foundation.
  The declaration disposition records all 61 typed-action eligibility decisions; the remaining
  ordinary actions need closed, definition-owned exact-action selection, while multipart `upload_file`
  needs a separate bounded binary/multipart destination contract and remains reachable as `files upload`.
- **Direct reads use `json_redacted`.** That is the only general-purpose output policy the runtime
  supports for direct reads (`commandrunner.supportedDirectReadOutputPolicies`).
- **Flags exist for scalar fields only.** An object, array, or union-typed field is supplied by the
  reverse-ETL source record, not by a shell flag — the same rule the merged `notion`/`gong` bundles
  follow.
- **No `fixtures/writes/*.json` are shipped.** Per `internal/connectors/conformance/dynamic.go`, a
  missing write fixture is a skipped check, not a failure; `gong` (27 write actions, several
  multipart) ships the same way.
