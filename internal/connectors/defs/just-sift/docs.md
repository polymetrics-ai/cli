# Overview

JustSift reads Sift people directory profiles and person field definitions through the JustSift
REST API (`https://api.justsift.com/v1`). This bundle migrates the legacy
`internal/connectors/just-sift` package to the declarative engine at capability parity; the legacy
package stays registered and unchanged until wave6's registry flip. The connector is read-only.

## Auth setup

Provide a JustSift API token via the `api_token` secret; it is used only for Bearer auth
(`Authorization: Bearer <api_token>`) and is never logged.

## Streams notes

Two streams, each with a different pagination shape (matching legacy's `streamEndpoint.pagination`
routing):

- `peoples` (`GET /search/people`): 1-based `page_number` pagination (`page_param: page`,
  `start_page: 1`), request size driven by `query.limit` (`config.page_size`, default `"100"`).
  Legacy's own page-size config also drives the short-page stop check; see the pagination
  stop-threshold note below (same class as judge-me-reviews' documented narrowing). Every record
  gets a static `connector: "just-sift"` marker stamped via `computed_fields`, matching legacy
  `peopleRecord`/`fieldRecord`'s `rec["connector"] = registryName` line.
- `fields` (`GET /fields/person`): `cursor` pagination with `token_path: links.next` and
  `cursor_param: cursor` — the next-page token is read from the response body's `links.next` and
  sent back as the `cursor` query parameter.

Neither stream declares an `incremental` block: legacy's `streams()` sets no `CursorFields` for
either stream (`peoples`/`fields` are full-refresh only in both legacy and this bundle).

## Write actions & risks

None. JustSift is read-only in both legacy and this bundle (`capabilities.write: false`, no
`writes.json`).

## Known limits

- **Pagination stop-threshold parity narrowing (ACCEPTABLE, documented, same class as
  judge-me-reviews)**: legacy's `page_size` config (1-500, default 100) drives both the `limit`
  request query param and the `peoples` stream's short-page stop check. The engine's `page_number`
  paginator's stop-threshold (`pagination.page_size`) is a fixed literal (`100`, legacy's default)
  and cannot be wired to the same runtime `config.page_size` value the `limit` query param uses.
  Never wrong for the default-`page_size` case; only imprecise for a non-default override.
- **`links.next` query-fragment parsing (ACCEPTABLE, documented)**: legacy's `applyNextCursor`
  defensively handles a `links.next` token shaped as a full `"key=value"` query fragment (e.g.
  `"cursor=abc"`) by parsing it and merging the parsed key/value pairs into the request query,
  falling back to `query.Set("cursor", next)` only when the token is NOT a parseable query string.
  The engine's `cursor`+`token_path` paginator (`tokenPathCursor`) always sends the token verbatim
  as the `cursor_param` value (`cursor_param.Set(cursor, token)`) with no fragment-parsing — for
  JustSift's real-world opaque-cursor-token shape (the common case) this is identical to legacy's
  behavior; the divergence only manifests for the defensive `"key=value"`-shaped token branch,
  which JustSift's documented API does not produce in the normal case. This never changes emitted
  record DATA (only the shape of the next-page request in an edge case legacy's own tests exercise
  via mock server but the real API is not documented to trigger) — see
  `docs/migration/conventions.md`'s parity-deviation ledger.
- Full JustSift API surface (any write/mutation endpoints, other read resources) is out of scope;
  see `api_surface.json` — only the 2 legacy-parity read streams are implemented.
- `peoples` fixture ships a 100-record page 1 (matching the fixed `pagination.page_size: 100`
  short-page threshold) plus a 1-record page 2; `fields` ships a 2-page fixture where page 1's
  `links.next` is a non-empty opaque token and page 2's `links` is empty, proving the cursor
  paginator advances then terminates.
