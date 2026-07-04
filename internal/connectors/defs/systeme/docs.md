# Overview

Systeme is a Tier-1 declarative-HTTP connector for the Systeme.io public API
(`https://api.systeme.io/api`). This is a Pass B full-surface expansion: 6 read streams
(`contacts`, `tags`, `contact_fields`, `funnels`, `funnel_steps`, `webhooks`) and 14 write actions.

**Research method and its limits** (recorded honestly per `docs/migration/conventions.md` §1's
`api_surface.json` depth requirement): Systeme.io's public API has **no publicly fetchable
OpenAPI/Swagger document**. Its official reference (`developer.systeme.io/reference/api`) is a
ReadMe.io-hosted site whose per-endpoint detail renders client-side from a private, authenticated
project session (`dash.readme.com/api/v1/api-specification` requires an API key this review does
not have); no public Postman collection or third-party mirror exists either. Every endpoint in this
bundle and in `api_surface.json` was therefore independently confirmed **live** against
`https://api.systeme.io/api` via unauthenticated `OPTIONS` requests: the server's `Allow` response
header names exactly which HTTP methods it accepts at that path today (a live, current, and
verifiable source of truth, cross-checked against Systeme.io's own help-center article "How to use
systeme.io's public API", `help.systeme.io/article/2329`, last updated 2026-06-05). Where the live
probe and the help article disagreed (newsletters, subscriptions), the live probe's result governs
and the discrepancy is documented in `api_surface.json` as `requires_elevated_scope` (no confirmable
live path) rather than silently guessed at.

**Tier justification**: a plain declarative-HTTP bundle. Auth is a single static header, pagination
is the engine's `page_number` type throughout, and the one sub-resource stream (`funnel_steps`) is
an ordinary single-level `fan_out` over `funnels` — nothing needs a Go hook.

## Auth setup

Provide a Systeme.io API key via the `api_key` secret; it is sent as the `X-API-Key` header
(`auth.mode: api_key_header`), no prefix. Never logged. `base_url` defaults to
`https://api.systeme.io/api` and may be overridden for tests/proxies.

## Streams notes

All streams share the base's `page_number` pagination (`page_param: page`, `size_param: limit`,
`page_size: 100`), stopping on a short page — the same shape the prior version of this bundle
already established for `contacts` and now confirmed live (via `OPTIONS`) to apply identically to
every other listed resource.

`contacts` (`GET /contacts`) is unchanged in shape from the prior version of this bundle
(`records.path: items`, `computed_fields.created_at` renamed from the raw `createdAt`); this
version additionally declares `locale`/`fields`/`tags` as optional passthrough-typed properties
(Systeme.io's contact object embeds a `fields: [{slug, value}, ...]` array of custom-field values
per the API's own documented shape) without asserting a rigid structure for `fields`' contents,
since custom fields are workspace-defined and have no fixed schema.

`tags` (`GET /tags`) and `contact_fields` (`GET /contact_fields`) are plain top-level list streams,
each with a matching `GET /{resource}/{id}` detail endpoint (live-confirmed, excluded as
`duplicate_of`) and full CRUD write actions (also live-confirmed via each detail path's `Allow`
header: `tags/{id}` allows `GET, PUT, DELETE`; `contact_fields/{id}` allows `GET, PATCH, DELETE`).

`funnels` (`GET /funnels`) is read/create-only: `funnels/{id}` live-probes to `Allow: GET` only (no
`PATCH`/`DELETE` exist for this resource today), so there is no `update_funnel`/`delete_funnel`
write action — not an oversight, a faithful reflection of what the live API actually accepts.
`funnel_steps` (`GET /funnels/{funnel_id}/steps`) `fan_out`s over every funnel id
(`ids_from.request`: `GET /funnels`, `records_path: items`, `id_field: id`), stamping `funnel_id`
onto every emitted record; `funnels/{id}/steps` live-probes to `Allow: GET, POST` (list + create,
confirmed), with no per-step detail/update/delete endpoint found (`funnels/{id}/steps/{id}`
live-probes to a bare 404).

`webhooks` (`GET /webhooks`) is list/create-only: `webhooks/{id}` live-probes to a bare 404 with **no
`Allow` header at all** — the same "no route registered" signature every genuinely nonexistent path
in this review produced (contrasted with every real id-scoped resource's 405-with-`Allow`-header
response to an unsupported method). This is a materially different, stronger signal than "detail
endpoint not yet implemented by this bundle" — it means the live API itself has no per-webhook
detail/update/delete route, so none is modeled here.

Two resources named in Systeme.io's own help article could **not** be independently confirmed and
are excluded rather than guessed at (`api_surface.json`): **newsletters** ("create/update/list/retrieve
newsletters") — `GET`/`POST /newsletters` both returned a bare 404 with no `Allow` header during
live probing, the same "no route" signature as a genuinely nonexistent path; and **subscriptions**
("retrieve subscription resources... unsubscription") — no path variant tried
(`/subscriptions`, `/subscription`, `/subscription-plans`) could be confirmed live. **communities**
is a similar partial case: `communities/{id}` IS live (`Allow: GET`), but bare `GET /communities`
404s with no `Allow` header — there is no list/discovery endpoint to enumerate community ids, so it
cannot be modeled as a syncable stream (same shape as bitly's `custom_bitlinks` gap in the goldens).
**courses** could not be confirmed live at all, consistent with Systeme.io's own public product
roadmap listing "API: add endpoints for managing courses and students" as an open, unshipped feature
request (`roadmap.systeme.io/c/228`) — the API genuinely does not have this surface yet, not merely
undocumented.

## Write actions & risks

- `create_contact`/`update_contact`/`delete_contact` — contact lifecycle. `delete_contact` is
  irreversible; approval required. Create/update are lower-risk (no approval required).
- `create_tag`/`update_tag`/`delete_tag` — tag definition lifecycle. `delete_tag` removes the tag
  from every contact it is assigned to; approval required.
- `add_contact_tag`/`remove_contact_tag` — assign/remove a tag on a contact
  (`POST`/`DELETE /contacts/{id}/tags[/{tag_id}]`). Per Systeme.io's own docs, tag assignment can
  trigger workspace automations (course enrollment, campaign entry) as a side effect — this is
  documented but not itself modeled as a distinct action (the API gives no way to suppress or
  observe automation side effects from the tag-assignment call itself).
- `create_contact_field`/`update_contact_field`/`delete_contact_field` — custom contact-field
  definition lifecycle. `delete_contact_field` removes the field's stored value from every contact;
  approval required.
- `create_funnel` — creates a new funnel. No `update_funnel`/`delete_funnel` — see Streams notes
  (the live API has no `PATCH`/`DELETE` for this resource).
- `create_funnel_step` — creates a new step within an existing funnel. No update/delete — same
  reasoning (no confirmed live route).
- `create_webhook` — creates a new outgoing webhook subscription. No update/delete — see Streams
  notes (the live API has no per-webhook detail route at all).

## Known limits

- **`funnels`/`funnel_steps`/`webhooks` are create-only beyond read** (no update/delete write
  actions) — this reflects what the live API actually accepts (confirmed via `OPTIONS` `Allow`
  headers), not an incomplete migration.
- **Newsletter and subscription-resource management are excluded** — named in Systeme.io's own
  help-center article but could not be independently confirmed live at any probed path; see
  `api_surface.json`'s `requires_elevated_scope` entries.
- **`communities` has no list stream** — `communities/{id}` is live but has no discovery/list
  endpoint to enumerate ids from.
- **Courses are not part of the live public API** — confirmed both by the absence of any live
  route and by Systeme.io's own product roadmap listing course/student API endpoints as an open
  feature request.
- **`fields` (custom contact-field values) has no fixed schema** — `contacts.fields` is typed as a
  passthrough array; individual custom-field slugs/values vary per workspace and are not enumerated
  in this schema (see `contact_fields` for the field DEFINITIONS stream, which is separate from
  each contact's field VALUES).
- All fixtures (`fixtures/streams/**`, `fixtures/check.json`) represent Systeme.io's real wire
  shape, including the camelCase `createdAt` field on contacts.
