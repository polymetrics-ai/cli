# Overview

Smaily is a wave2 fan-out declarative-HTTP migration, expanded in Pass B to the full documented
Smaily PHP API surface. It reads Smaily campaigns, segments, contacts, templates, automations,
and organization users, and writes subscriber/segment upserts, campaign-recipient unsubscribes,
individual message sends, and automation-workflow triggers, through the Smaily PHP API
(`https://<subdomain>.sendsmaily.net/api/*.php`, documented at `https://smaily.com/help/api/`).
This bundle migrates `internal/connectors/smaily` (the hand-written connector); the legacy
package stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide `api_username` (config) and `api_password` (secret); they are sent as HTTP Basic auth
credentials (`basic` auth mode), matching legacy's `connsdk.Basic(user, pass)` (`smaily.go:144`).
`base_url` is required directly by this bundle (see Known limits for why legacy's
`api_subdomain`-derived default is not modeled).

## Streams notes

The 5 legacy-parity streams (`campaigns`, `segments`, `subscribers`, `templates`,
`automations`) hit their own `GET api/<resource>.php` endpoint, matching legacy's
`streamEndpoints` map exactly (`campaign.php`, `segment.php`, `contact.php`, `template.php`,
`autoresponder.php`). Records are extracted from the response body's top-level array
(`records.path: ""`, root-array shape), matching legacy's `recordsPath: ""` for every stream and
`connsdk.RecordsAt`'s empty-path == body-root semantics. None of these 5 streams paginate in
legacy (a single `r.Do` call per read, no loop, and no query parameters at all) —
`pagination.type: none` is declared, one request per read, no query. Every stream declares
`projection: "passthrough"`: legacy's `readRecords` emits each decoded record verbatim
(`emit(connectors.Record(rec))`, `smaily.go:112`, no field-building or `mapRecord` step), so this
bundle emits every raw field the API returns rather than narrowing to the `id`/`name`/
`created_at` triple `streams()`'s catalog happens to declare — schema-mode projection on a
verbatim-emitting legacy would silently drop real API fields (`conventions.md` §8 rule 1). The
`id`/`name`/`created_at` properties in `schemas/*.json` remain the documented, guaranteed-present
fields; they are a floor, not a ceiling, on what a record contains.

The Pass-B-added `organization_users` stream (`GET api/organizations/users.php`) has no legacy
equivalent — its shape is authored directly from Smaily's own API documentation. Unlike the 5
legacy streams, it genuinely paginates server-side (`page`/`limit` query params, confirmed by
Smaily's own docs example `?page=0&limit=250`): `pagination.type: page_number` with
`start_page: 0` (Smaily's pagination is 0-indexed, confirmed by the documented example) and a
small `page_size: 2` (an authoring choice for readable fixtures, not a documented API default —
any positive `limit` value is accepted by the live API).

Three more Pass-B streams round out the full documented surface, none with a legacy equivalent:
`segment_rules` (`GET api/list.php`) is Smaily's own documented "list segments"/"list segment
rules" read, returning `id`/`name`/`filter_type`/`subscribers_count`/`filter_data` — declared as a
separate stream from the legacy-parity `segments` stream (see Known limits for why `segments`
itself keeps calling `api/segment.php`, a path this review could not confirm still exists in
Smaily's current docs). `segment_subscribers` (`GET api/contact.php`, optionally scoped by the new
`config.segment_id` via the `list` query parameter, `omit_when_absent`) is Smaily's documented
"list subscribers of a segment" read — same underlying path as `subscribers`, but declared as its
own stream since its real (email-keyed) record shape and the `list` scoping parameter are
independently documented and meaningfully different in intent from the unscoped contact list.
`ab_tests` (`GET api/split.php`) lists A/B tests; the same path is used by the new
`launch_ab_test` write below (GET list vs POST launch, mirroring the `campaigns` GET/POST split
this API already establishes).

## Write actions & risks

Legacy's own connector is read-only, but Pass B full-surface expansion adds 5
dialect-expressible mutations Smaily documents (`api_surface.json`), each requiring approval:

- **`create_or_update_subscriber`** (`POST api/contact.php`, `upsert`): creates or updates a
  subscriber matched by `email`; Smaily's own docs note this endpoint does NOT trigger automation
  workflows (use `trigger_automation_workflow` for that).
- **`create_or_update_segment`** (`POST api/list.php`, `upsert`): creates a new segment, or
  overwrites an existing one's filter definition when `id` is set. `filter_data` is an array of
  `[field, [operator, value]]` tuples per Smaily's own filter-condition grammar (e.g.
  `[["gender", ["Equal", "women"]]]`) — the dialect declares it as a bare `"type": "array"` with
  no further structural validation (draft-07 has no tuple-shaped array schema in this dialect's
  subset), so malformed filter tuples are rejected by the live API, not client-side.
- **`unsubscribe_recipient`** (`POST api/unsubscribe.php`, `update`): unsubscribes a recipient
  from a specific campaign; Smaily's docs note the benefit over a generic global unsubscribe is
  that it is reflected in that campaign's own statistics.
- **`send_message`** (`POST api/message/send.php`, `create`): sends a single, individually
  templated outbound email using an automation workflow's template, WITHOUT triggering the
  workflow itself (filters/delays do not apply — contrast with `trigger_automation_workflow`
  below). Risk: a genuine outbound-email side effect to real recipients.
- **`trigger_automation_workflow`** (`POST api/autoresponder.php`, `custom`): opts in
  subscribers and triggers a "form submitted"-style automation workflow for them. Risk: the
  highest-impact write in this bundle — subscriber data is updated BEFORE the workflow's messages
  send, and the update itself may cascade into other workflow-driven scheduled sends; Smaily's own
  docs explicitly warn this is a reactive operation, not suited for single individually-crafted
  messages (use `send_message` for that case instead).

## Known limits

- **`api_subdomain`-derived `base_url` is not modeled; `base_url` is required directly instead.**
  Legacy derives `https://<api_subdomain>.sendsmaily.net` from a separate `api_subdomain` config
  value when `base_url` is unset (`smaily.go:151-158`), including a subdomain-label safety check
  (`strings.ContainsAny(subdomain, "/:@")`). The engine's `spec.json` `"default"` materialization
  mechanism (`docs/migration/conventions.md` §3) only fills a FIXED literal default, not a
  value derived from another config key at read time — there is no declarative base-URL-
  construction template in this dialect. Per convention, this bundle narrows the config surface:
  `base_url` is a required spec property with no default, and `api_subdomain` is not declared at
  all (a declared-but-unwireable config key is worse than an absent one, per the searxng/bitly
  precedent). An operator who previously configured only `api_subdomain` must now supply the full
  `https://<subdomain>.sendsmaily.net` URL as `base_url`; this is a documented config-surface
  narrowing, not a data-shape change — every record emitted for a given account is identical
  either way once `base_url` resolves to the same origin.
- **Write request/response body shapes are sourced from Smaily's own public API documentation**
  (`https://smaily.com/help/api/{subscribers-2,segments,campaigns-3,messages,automations-2}/...`),
  not from legacy Go (which has no write path at all — `capabilities.write` was `false` prior to
  this Pass B expansion). Field names (`email`/`name`/`filter_type`/`filter_data`/`campaign_id`/
  `autoresponder_id`/`to`/`context`/`attachments`/`autoresponder`/`addresses`) are the documented
  parameter names taken directly from Smaily's published curl/request examples.
