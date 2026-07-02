# Overview

Nebius AI Studio exposes an OpenAI-compatible REST API at `https://api.studio.nebius.com`. This
bundle reads models, files, and batch jobs, migrated from `internal/connectors/nebius-ai` (the
hand-written connector this bundle replaces at capability parity); the legacy package stays
registered and unchanged until wave6's registry flip. The connector is read-only.

## Auth setup

Provide the `api_key` secret; the engine's declarative `bearer` auth mode sends
`Authorization: Bearer {{ secrets.api_key }}` on every request, matching legacy's
`connsdk.Bearer(secret)`. The secret only ever flows into the auth header and is never logged.

Set `base_url` to override the API host (e.g. a test proxy); it defaults to
`https://api.studio.nebius.com` (legacy's `nebiusDefaultBaseURL`).

## Streams notes

Three list streams, all sharing the OpenAI-compatible list envelope `{object:"list", data:[...],
has_more:bool}`:

- `models` (`GET /v1/models`) — no cursor field is populated by legacy's own catalog beyond
  `created`, and legacy never applies any incremental filter to this stream; this bundle declares
  no `incremental` block, matching legacy's real (non-filtering) behavior, while still recording
  `x-cursor-field: created` on the schema for catalog-metadata parity (see dwolla's identical
  precedent for a stream whose cursor field is descriptive-only).
- `files` (`GET /v1/files`) — same shape; `x-cursor-field: created_at`.
- `batches` (`GET /v1/batches`) — same shape; `x-cursor-field: created_at`.

Pagination is `cursor` with `last_record_field: id` and `stop_path: has_more` (Stripe-style
`after`/`has_more`, matching legacy's `harvest` loop exactly: the next page's `after` query param
is the last emitted record's `id`, and a `has_more: false` — or a missing/falsy value — stops the
loop regardless of page fullness). Every request sends `limit={{ config.limit }}` (default 20,
legacy's `nebiusDefaultPageSize`; legacy also accepted a `page_size` alias for the same value,
narrowed here to `limit` only — see Known limits).

None of legacy's three streams actually filters by any date range server-side or client-side:
legacy's own `harvest` and `InitialState` never reference the read request's cursor state at all
(the state cursor exists only as connector bookkeeping / a fixture-mode debug marker). This bundle
matches that real behavior exactly by declaring no `incremental` block on any stream — adding one
(with or without `client_filtered`) would be a behavior change, not a port.

## Write actions & risks

None. Nebius AI Studio is read-only in both legacy and this bundle (`capabilities.write: false`);
no `writes.json` file is shipped. File upload, batch creation/cancellation, and the
chat/embeddings inference endpoints are out of scope for this wave (Pass B) — see
`api_surface.json`.

## Known limits

- Legacy accepted a `page_size` config key as an alias for `limit` (same effective value, two
  accepted key names). This bundle narrows the config surface to `limit` only — the engine's
  `stream.Query`/`spec.json` dialect has no alias mechanism (one canonical config key maps to one
  query param). This narrows accepted CONFIGURATION surface only, never emitted record data.
- `max_pages` is not exposed as a `spec.json` property: the engine's `cursor` paginator has no
  request-count cap field beyond the bundle-wide `PaginationSpec` (no per-connector override
  parameter in this dialect for the `cursor` type), matching legacy's own default of unbounded
  (`max_pages: 0`/`all`/`unlimited`) — this bundle is unbounded by default with no configurable cap,
  a narrowing of legacy's optional cap, not a behavior change to the common (unbounded) case.
- `metadata.json` declares no `rate_limit` block: legacy enforces no client-side rate limiting for
  Nebius AI Studio, so none is added here either.
