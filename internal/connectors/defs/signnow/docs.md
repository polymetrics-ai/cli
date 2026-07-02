# Overview

signNow is a wave2 fan-out declarative-HTTP migration. It reads signNow documents, templates, and
users through the signNow REST API (`GET https://api.signnow.com/...`). This bundle targets
capability parity with `internal/connectors/signnow` (the hand-written connector it migrates); the
legacy package stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide a signNow OAuth access token via the `access_token` secret; it is sent as a Bearer token
(`Authorization: Bearer <access_token>`) and is never logged, matching legacy's
`connsdk.Bearer(token)` (`signnow.go:131`). `base_url` defaults to `https://api.signnow.com` and may
be overridden for tests/proxies.

## Streams notes

All 3 streams (`documents`, `templates`, `users`) are `GET` list endpoints (`document`, `template`,
`user`) with records at the top-level `data` key, primary key `["id"]`. Pagination follows
signNow's own cursor convention (`pagination.type: cursor`, `cursor_param: page_token`,
`token_path: next`) — the next page's `page_token` is read from the response body's top-level
`next` field, and pagination stops on an empty/absent `next` (legacy's own
`strings.TrimSpace(token) == ""` stop condition; no `stop_path` is declared since legacy never
checks a separate boolean stop flag — an empty token is the only stop signal on both sides).
`limit` defaults to `50` (`spec.json`'s `page_size` default, matching legacy's
`signnowDefaultPageSize`) and is sent on every request via `{{ config.page_size }}`.

Legacy's shared `signnowRecord` mapper derives `name` via a fallback chain
(`document_name -> template_name -> name -> email`) and `updated_at` via another
(`updated -> updated_at -> created`) across all three endpoints' differing raw field names. Since
the engine dialect has no cross-endpoint "first non-null of N paths" filter, and each endpoint's
real wire shape only ever populates ONE of the fallback candidates for a given output field
(documents emit `document_name`/`updated`; templates emit `template_name`/`updated`; users emit
`name`/`created`, with `email` passed straight through by schema projection), this bundle expresses
the identical effective mapping per-stream directly: `documents`/`templates` use a
`computed_fields` rename (`name` from `document_name`/`template_name` respectively, `updated_at`
from `updated`); `users` relies on direct schema-projection for `name`/`email` (raw field names
already match) plus a `computed_fields` rename of `updated_at` from `created` (users have no
`updated`/`updated_at` field in their real wire shape). This produces byte-identical output to
legacy's fallback chain for every record shape either endpoint actually emits.

## Write actions & risks

None. signNow's read endpoints have no obviously-safe reverse-ETL writes modeled; `capabilities.write`
is `false` and this bundle ships no `writes.json`, matching legacy's `Write` returning
`connectors.ErrUnsupportedOperation`.

## Known limits

- **`max_pages` is not runtime-configurable.** Legacy exposes a config-driven `max_pages` override
  (`signnow.go:226-236`) that hard-caps the request count. The engine's cursor paginator reads
  `PaginationSpec.MaxPages` only from the bundle's own static `streams.json` (never templated from
  `config.*`), so a `spec.json`-declared `max_pages` property would be genuinely dead config (F6,
  REVIEW.md). This bundle does not declare `max_pages` in `spec.json` at all; pagination is
  unbounded, matching legacy's own default (`max_pages` unset -> unlimited) and stops only on
  signNow's empty-`next`-token signal, identical to legacy's own only stop condition.
- **Cross-endpoint fallback chain is expressed per-stream, not as a shared coalesce.** See
  "Streams notes" above — this is a documented, verified-equivalent expression of legacy's shared
  `signnowRecord` mapper, not a scope narrowing. If a real signNow document record were ever missing
  `document_name` but had a raw `name` field (legacy's third fallback tier), this bundle would emit
  `null` instead of falling through to `name` — this is an accepted, extremely unlikely-in-practice
  divergence given signNow's documented response shape uses `document_name`/`template_name`
  consistently; no fixture or live evidence of a document/template record omitting its own primary
  name field has been observed.
- Full signNow API surface (folders, invites, embedded signing, teams) is out of scope for wave2;
  see `api_surface.json`'s `excluded: {category: out_of_scope}` entries.
