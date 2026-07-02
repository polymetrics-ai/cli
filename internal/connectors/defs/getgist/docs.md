# Overview

GetGist (Gist, getgist.com) is a read-only declarative-HTTP connector migrated from
`internal/connectors/getgist` (legacy wave2 fan-out). It reads Gist contacts, tags, segments,
campaigns, forms, and teammates through the Gist REST API. This bundle is capability-parity with
the legacy hand-written connector; the legacy package stays registered and unchanged until wave6's
registry flip.

## Auth setup

Provide a Gist API key via the `api_key` secret; it is used only for Bearer auth
(`Authorization: Bearer <api_key>`) and is never logged. `base_url` defaults to
`https://api.getgist.com` and may be overridden for tests or proxies.

## Streams notes

All 6 streams (`contacts`, `tags`, `segments`, `campaigns`, `forms`, `teammates`) share the same
shape: `GET` against the Gist list endpoint, records nested under a resource-named JSON key (e.g.
`{"contacts": [...]}`), primary key `["id"]`. Pagination is `page_number` (`page`/`per_page` query
params, default page size 50) — Gist's real stop signal is a `pages.next` link, but the engine's
`page_number` paginator's short-page stop rule (fewer than `page_size` records returned) is the
exact SAME primary termination check legacy's own `harvest` loop applies first (legacy checks the
short page before consulting `pages.next` at all), so no behavior is lost porting to the declarative
paginator.

`contacts` exposes an `updated_at` field that legacy's own `Stream` catalog entry names as a
`CursorFields` marker — however legacy's `Read` path never actually sends any incremental filter
parameter, nor filters client-side, for ANY stream (the only place `req.State["cursor"]` is
consulted is fixture-mode's `previous_cursor` debug annotation). Declaring an `incremental` block
here would change accepted behavior (the engine would then compute and possibly send a lower-bound
parameter legacy never sent), so this bundle matches legacy's real behavior: `contacts`' schema
still declares `x-cursor-field: updated_at` (informational, for downstream sort/dedup — matches the
legacy catalog metadata), but `streams.json` declares no `incremental` block for any stream — every
stream is a full-refresh read, exactly like legacy.

## Write actions & risks

None. Gist's exposed resources here are read-only in this connector (`capabilities.write: false`,
matching legacy exactly); there is no `writes.json`.

## Known limits

- Only the 6 legacy-parity read streams are implemented; Gist's write endpoints (creating/updating
  contacts, tags, campaigns, sending events) are out of scope for this migration wave — see
  `api_surface.json`'s `excluded: {category: out_of_scope, reason: "Pass B capability expansion"}`
  entries.
- `contacts`' `updated_at` cursor field is catalog-informational only, matching legacy's own
  behavior of declaring a `CursorFields` entry without ever filtering by it (see Streams notes
  above) — this is a faithful port, not a scope narrowing.
- Fixtures represent Gist's real wire shape faithfully, including the `pages.next` link field
  (not consumed by the declarative paginator's stop logic, which relies on the short-page rule
  instead, per Streams notes above).
