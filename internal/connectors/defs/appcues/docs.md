# Overview

Appcues is a read-only declarative-HTTP connector for the Appcues REST API v2. It reads flows,
segments, tags, checklists, and banners for a configured account. This bundle migrates
`internal/connectors/appcues` (the hand-written connector); the legacy package stays registered and
unchanged until wave6's registry flip.

## Auth setup

Provide the Appcues account ID via the `account_id` config value (every resource path is scoped to
`accounts/{account_id}/...`), the Appcues API key via the `username` config value, and the Appcues
API secret via the `password` secret. Auth is HTTP Basic (`username` as the Basic-auth username,
`password` as the Basic-auth password); the secret is never logged.

## Streams notes

All 5 streams (`flows`, `segments`, `tags`, `checklists`, `banners`) share the identical shape: `GET`
against `accounts/{account_id}/<resource>`, records at the response root (a top-level JSON array),
primary key `["id"]`. Pagination is 1-based `page_number` (`page_param: page`, `size_param: limit`,
`start_page: 1`, `page_size: 100` matching legacy's default), stopping on a short page (fewer
records returned than the requested page size) exactly like legacy's `harvest` loop. Every stream
carries `updatedAt` as its catalog cursor field (matching legacy's `CursorFields`), but — matching
legacy exactly — no stream declares a server-side incremental filter: legacy's `Read` always issues
the same full list request regardless of `req.State`, so no `incremental` block is declared in
`streams.json` either (declaring one would add filtering behavior legacy never had).

## Write actions & risks

Not applicable — this connector is read-only (`capabilities.write: false`), matching legacy exactly.

## Known limits

- Full Appcues API surface (users, events, step groups) is out of scope for this wave; see
  `api_surface.json`'s `excluded: {category: out_of_scope, reason: "Pass B capability expansion"}`
  entries.
- No incremental sync mode is wired for any stream (see Streams notes) — this mirrors legacy's own
  full-refresh-only behavior, not a capability gap introduced by migration.
