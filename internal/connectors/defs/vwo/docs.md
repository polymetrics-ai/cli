# Overview

VWO (Visual Website Optimizer) is an A/B testing and conversion optimization platform. This bundle
reads and writes campaigns via VWO's API v2 (`{base_url}/accounts/{account_id}/campaigns`). It
originally migrated `internal/connectors/vwo` (the hand-written legacy connector, read-only, a
single unpaginated `GET /campaigns` request), which stays registered and unchanged until wave6's
registry flip. **Pass B full-surface expansion (this revision) found and corrected two real-API
mismatches in legacy's own request shape** — see "Auth setup" and "Streams notes" below — while
researching the live, currently-published OpenAPI specs at `developers.wingify.com` (VWO's parent
company; the documented API host, `app.wingify.com/api/v2`, is the identical host family as
legacy's `app.vwo.com/api/v2`).

## Auth setup

**Pass B correction**: legacy authenticated with `Authorization: Token <api_key>` via
`connsdk.APIKeyHeader("Authorization", key, "Token ")`. The real API's OWN published OpenAPI
security scheme (confirmed identically across every fetched endpoint page's embedded spec) is:

```json
"securitySchemes": {"sec0": {"type": "apiKey", "in": "header", "name": "token"}}
```

— a bare header literally named `token` (lowercase, no `Authorization` wrapper, no `Token ` value
prefix), carrying the raw API key. This bundle now sends `token: <api_key>` via `streams.json`
`base.auth`'s `api_key_header` mode (`header: "token"`, `prefix: ""`). This is a correction, not an
accepted-input-behavior change under the parity meta-rule (conventions.md §5): legacy's
`Authorization: Token` header shape targets an auth scheme the real, documented API does not
recognize, so no config that worked against the real service is broken by this fix — the reverse:
this fix is what makes the connector actually able to authenticate against the real API for the
first time. `api_key` itself (the generated personal API token, from
`https://app.wingify.com/#/developers/tokens`) is unchanged.

`base_url` defaults to `https://app.vwo.com/api/v2`, matching legacy's `defaultBaseURL` constant —
confirmed still correct and current against the real API's own documented host.

## Streams notes

**Pass B correction**: legacy's bare `GET /campaigns` (no account scoping at all) does not match
any endpoint the real, currently-documented API exposes — every real campaign endpoint is
account-scoped: `GET /accounts/{account_id}/campaigns`. This bundle now sends the real path,
wiring `{account_id}` from `spec.json`'s new `account_id` config property, which **defaults to the
literal string `"current"`** — VWO's own documented keyword meaning "the Main Workspace of the
authenticated token" (from the real endpoint's own OpenAPI parameter doc: "Use 'current' keyword to
refer to the Main Workspace or the Integer Workspace Id"). Because of this default, **no new
required config is introduced** relative to legacy's zero-account-scoping behavior — an operator
who configures nothing beyond `api_key` gets the exact same "my own workspace's campaigns" scope
legacy intended, now via a path the real API actually serves. Setting `account_id` to a specific
numeric workspace id targets a VWO sub-account instead.

`campaigns` reads `GET /accounts/{account_id}/campaigns` and extracts records from the response's
top-level `_data` array (the real API's own documented envelope — NOT the `campaigns` key legacy's
`connsdk.RecordsAt(resp.Body, "campaigns")` targeted, another real-API-shape correction). `id` is
still forced to string output via `computed_fields`' `"id": "{{ record.id | last_path_segment }}"`
(VWO's wire `id` is a bare JSON integer; the `last_path_segment` filter forces `Interpolate`'s
string output while passing a delimiter-free numeric value through unchanged — same pattern as
before this pass). The real record shape is considerably richer than legacy's narrow
`{id, name, status, created_at}` projection, but the schema intentionally keeps that legacy
projection. `created_at` is filled from legacy-shaped `created_at` first and the current API's
`createdOn` fallback second.

Two new optional query filters are wired via the optional-query dialect (`omit_when_absent: true`):
`campaign_type` (the real API's `type` filter, one of 17 documented campaign-type values) and
`platform` (`website`/`full-stack`/`mobile-app`). Pagination is now modeled: the real endpoint
documents `limit` (default 25, this bundle sends 100 via `spec.json`'s `page_size`) and `offset`
(a numeric byte-offset, standard `offset_limit` pagination) — legacy issued exactly one unpaginated
request and would have silently returned only the API's own default first page (25 records) for any
workspace with more campaigns than that; this bundle now paginates to completion.

`start_date` (legacy's own optional passthrough config, preserved for config-shape backward
compatibility) is **not a real VWO query parameter at all** — the real "Get all campaigns" endpoint
documents no date-range filter of any kind. It was already a silent no-op against the real API
before this pass (legacy's own bug, inherited rather than introduced by migration); this bundle
keeps sending it exactly as legacy did (still harmless — VWO simply ignores an unrecognized query
param) rather than removing a config property a real deployment might already have set, but Known
limits documents this honestly rather than silently.

**`campaign_variations`** (Pass B new stream, `fan_out`): reads
`GET /accounts/{account_id}/campaigns/{campaign_id}/variations`, the per-campaign variations list
whose real response shape (`{"_data": [{"id","name","isControl","isDisabled","percentSplit",
"platform",...}]}`) is confirmed against `developers.wingify.com/reference/
get-all-the-variations-of-a-campaign`. Implemented via the engine's `fan_out` dialect
(conventions.md §3): `ids_from.request` issues one preliminary, fully-paginated
`GET /accounts/{account_id}/campaigns` call (reusing this stream's own effective pagination spec —
the base `offset_limit` block, same as `campaigns` itself), extracting each campaign's bare-integer
`id`; `into.path_var: "campaign_id"` threads each resolved id into `campaign_variations`' own `path`
as `{{ fanout.id }}`; `stamp_field: "campaign_id"` writes the originating campaign's id (as a
string, matching every other fan-out-derived id field in this repo, e.g. cisco-meraki's
`organizationId`) onto every emitted variation record after projection/`computed_fields`. Each
sub-sequence runs its own independent `offset_limit` pagination (short-page stop — a campaign's
variation count is always small, so this terminates after one page in practice), matching the
"fresh paginator per id" contract fan_out streams get uniformly (conventions.md §3). `id` here
(the variation's OWN id, distinct from the stamped `campaign_id`) is a bare integer via typed
`computed_fields` extraction, not string-forced, since VWO's own docs never surface it as a
URI-shaped value the way `campaigns.id` conventionally gets treated.

## Write actions & risks

Two write actions added in this Pass B expansion (legacy shipped none — `Write` always returned
`ErrUnsupportedOperation`). Both are external mutations of live A/B-testing configuration;
**approval required**:

- `create_campaign` — `POST /accounts/{account_id}/campaigns`. Requires `type`, `urls`,
  `primaryUrl`, and at least one `goals` entry (the real API's own required-field set).
- `update_campaign` — `PATCH /accounts/{account_id}/campaigns/{campaign_id}`. Can rename a
  campaign, change its `status` (which can **start, pause, or stop a live experiment** — the
  highest-risk mutation this connector exposes), or adjust `percentTraffic`.

## Write actions NOT modeled (Pass B breadth-vs-cost triage)

Per-campaign Variations create/update, and the entire Goals and Sections sub-resource families
(list/create/update for each), are not modeled as write actions this pass:

- **Variations create/update**: `campaign_variations` (above) now covers the READ side of this
  fan-out family. The corresponding create/update writes would need a campaign_id supplied
  per-record by the caller (the write dialect has no fan-out-derived-id mechanism analogous to
  streams' `fan_out.stamp_field` — a write action's `path_fields`/path templating draws only from
  the record itself) — a real but narrower follow-up than the read side, deferred this pass.
- **Goals**: excluded as `SCHEMA_AMBIGUOUS` — the dedicated "Get the goals of a campaign" reference
  page (`developers.wingify.com/reference/get-the-goals-of-a-campaign`) returned 404 and is not
  indexed in the site's own `llms.txt` page list, leaving no reachable source to confirm the goals
  list's real response shape independent of the `campaigns` stream's own goals COUNT summary field.
- **Sections**: the "Get all sections of a campaign" reference page's response schema/example was
  not populated in the published docs (multivariate-campaigns-only sub-resource) — an unconfirmed
  shape not worth guessing at.

See `api_surface.json` for the full excluded-endpoint list and per-endpoint reasoning. The
`campaigns` stream's own `goals`/`variations` fields still surface each campaign's goal/variation
COUNT summary (`partialCollection`/`totalCount`) regardless.

## Known limits

- **Goals and Sections sub-resources are not modeled as streams or writes; Variations create/update
  writes are not modeled** — see above.
- **`start_date` sends a real-API-ignored query parameter.** See "Streams notes" above: this is an
  inherited legacy no-op, not a new deviation introduced by this pass, and is kept for config-shape
  backward compatibility rather than silently dropped.
- **The Update-a-campaign-goal endpoint's own published path uses singular `campaign`**
  (`/accounts/{account_id}/campaign/{campaign_id}/goals/{goal_id}`), inconsistent with every
  sibling goals endpoint's plural `campaigns`. Recorded verbatim in `api_surface.json` from the
  live OpenAPI spec — not a transcription error introduced by this bundle, and moot regardless
  since this endpoint is not covered (see above).
- **No live-API verification was possible.** All endpoint shapes in this revision come from VWO/
  Wingify's own published OpenAPI specs and documentation examples (`developers.wingify.com`), not
  a live test account — standard for this migration's research methodology (conventions.md's
  Pass B mandate), but noted here since the auth-header and base-path corrections in particular are
  significant enough that a real-account smoke test is recommended before first production use.
