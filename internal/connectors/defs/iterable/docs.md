# Overview

Iterable is migrated as a pure Tier-1 declarative bundle: legacy
`internal/connectors/iterable/iterable.go` is a thin connsdk-HTTP wrapper (a `connsdk.Requester`
with `Api-Key` header auth plus a hand-rolled `nextPageToken` pagination loop over three uniform
list endpoints) with no auth/stream Go hooks and no writes — every behavior it implements maps
directly onto `streams.json`/`spec.json`/schemas with zero Go. It reads Iterable lists, campaigns,
and templates through the Iterable REST API (`GET {base_url}/<resource>`). Read-only;
`capabilities.write` is `false` and this bundle ships no `writes.json`. The legacy package stays
registered and unchanged until the wave6 registry flip.

## Auth setup

A single required secret, `api_key`, sent as the `Api-Key` request header with no prefix
(`iterable.go`'s `connsdk.APIKeyHeader("Api-Key", key, "")`) — modeled here as
`{"mode": "api_key_header", "header": "Api-Key", "value": "{{ secrets.api_key }}"}`. `base_url`
defaults to `https://api.iterable.com/api` (legacy's `iterableDefaultBaseURL`), materialized via
`spec.json`'s `"default"` when unset.

## Streams notes

Three streams, all primary-keyed on `id`, no incremental cursor (legacy publishes no cursor field
or state-driven filter — every read is a full walk of the resource):

- `lists` — `GET /lists`, `records.path: "lists"`, fields `id`/`name`/`listType`/`createdAt`/
  `updatedAt`.
- `campaigns` — `GET /campaigns`, `records.path: "campaigns"`, fields `id`/`name`/`createdAt`/
  `updatedAt`.
- `templates` — `GET /templates`, `records.path: "templates"`, fields `id`/`name`/`createdAt`/
  `updatedAt`.

All three share the identical pagination shape: `pageSize` is sent from `config.page_size`
(default 100, matching legacy's `iterableDefaultPageSize`, max 1000 matching
`iterableMaxPageSize`), and the next page is requested via a `pageToken` query param whose value is
read from the response body's `nextPageToken` field (`cursor` pagination, `token_path:
"nextPageToken"`) — this reproduces `iterable.go`'s `harvest` loop (`iterable.go:104-135`) exactly:
an absent/empty `nextPageToken` stops the walk. No `stop_path` is declared because legacy's own stop
condition is solely "the token is empty" — there is no separate boolean has-more flag to also
honor (unlike Stripe/Zendesk), so the paginator's default stop-on-empty-token behavior is already
byte-for-byte parity.

Legacy's `max_pages` config (accepting an integer, `"all"`, or `"unlimited"`) has no engine-side
equivalent to wire it into — `PaginationSpec.MaxPages` is a fixed value declared once in
`streams.json`, not templated from `config.*` at read time (see `docs/migration/conventions.md`
§3's pagination table and searxng's docs.md for the identical rationale) — so it is not declared in
`spec.json` at all (F6: a declared-but-unwireable key is worse than an absent one). See Known
limits.

## Write actions & risks

None. Legacy's `Write` is an unconditional `connectors.ErrUnsupportedOperation` stub
(`iterable.go:100-102`); `capabilities.write` is `false` and this bundle ships no `writes.json`.

## Known limits

- **`max_pages` is not modeled.** Legacy accepts a per-read `max_pages` config value (integer,
  `"all"`, or `"unlimited"`) as a hard page-count cap independent of page fullness. The engine's
  `PaginationSpec.MaxPages` is a fixed integer declared once in `streams.json` at bundle-author
  time; there is no dialect mechanism to source it from `config.*` per read (the same gap searxng's
  docs.md documents for its own `page_size`/`max_pages`). This bundle's `streams.json` therefore
  declares no `max_pages` at all (unbounded reads, matching legacy's own default of `0`/unlimited
  when unset) rather than baking in an arbitrary fixed cap; an operator who needs a bounded read
  has no config-time lever here, same as every other bundle with this identical dialect gap.
  Deliberate, documented scope narrowing — not silently wrong for any input legacy itself accepts
  with its own default settings.
