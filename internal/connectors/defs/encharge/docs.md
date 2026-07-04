# Overview

Encharge is a wave2 fan-out migration. It reads Encharge people, segments, fields, account tags,
and schemas through the Encharge REST API (`https://api.encharge.io/v1`). This bundle migrates
`internal/connectors/encharge`; the legacy package stays registered and unchanged until wave6's
registry flip. Encharge is a marketing automation platform; the upstream source supports
full-refresh extraction only and there is no obviously-safe reverse-ETL write surface, matching
legacy's `Capabilities.Write: false`.

## Auth setup

Provide an Encharge API token via the `api_key` secret; it is sent as the `X-Encharge-Token`
header (`mode: api_key_header`) and is never logged.

## Streams notes

`peoples` (`GET /people/all`, records at `people`) is limit/offset paginated
(`pagination.type: offset_limit`, `limit_param: limit`, `offset_param: offset`, `page_size: 100` —
legacy's real default page size); a page shorter than 100 records stops the read, matching legacy's
`harvest` loop exactly. `segments` (`GET /segments`, records at `segments`), `fields`
(`GET /fields`, records at `items`), `account_tags` (`GET /tags-management`, records at `tags`),
and `schemas` (`GET /schemas`, records at `objects`) are single-page reads with no pagination
declared, matching legacy's `readSinglePage` path for every stream but `peoples`. Every stream's
primary key matches legacy's published catalog (`id` for `peoples`/`segments`, `name` for
`fields`/`schemas`, `tag` for `account_tags`); none declare an incremental cursor field, matching
legacy's full-refresh-only catalog.

## Write actions & risks

None. Encharge is read-only; `capabilities.write` is `false` and this bundle ships no
`writes.json`.

## Known limits

- Only the 5 legacy-parity read streams are implemented; the full Encharge surface (event
  ingestion, flows, tag/segment mutation) is out of scope for this wave — see
  `api_surface.json`'s `excluded: {category: out_of_scope, reason: "not implemented in this bundle"}`
  entries.
- Legacy's `decodeRecords` defensively wraps a **scalar string** array element into
  `{"tag": <value>}` for the `account_tags` stream, in case `/tags-management`'s real payload is a
  plain array of tag strings rather than tag objects (legacy's own comment: "may be a plain array
  of tag strings"). The engine's generic `connsdk.RecordsAt` extraction (used by every declarative
  stream read) only keeps array elements that are already JSON objects — a scalar string element
  is silently dropped, with no filter in this dialect able to re-shape a bare scalar into an
  object. This bundle implements the common/documented object-array shape (matching the
  `account_tags` schema's `tag`/`id`/`createdAt` fields) faithfully; the defensive scalar-string
  fallback is an `ENGINE_GAP` if the real API ever actually sends that shape — not silently
  reproduced, and not observed in this connector's own tests (`encharge_test.go` only exercises the
  object-array shape for every stream). Flagged here rather than worked around.
- Legacy's `max_pages` config override (accepting `0`/`all`/`unlimited` for unbounded, or a
  positive integer cap) has no equivalent in this dialect: `PaginationSpec.MaxPages` is a fixed
  bundle-authored literal, not runtime-config-driven; this bundle leaves it unset (unbounded),
  matching legacy's own default behavior.
- Legacy's `page_size` config override (1-100) is likewise not modeled: `pagination.page_size` is
  a fixed JSON literal in this dialect with no runtime config-driven override mechanism; the
  bundle's declared `page_size: 100` reproduces legacy's real default.
