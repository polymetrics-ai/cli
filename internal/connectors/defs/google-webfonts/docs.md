# Overview

Reads Google Web Fonts families through the Google Fonts Developer API's single list resource,
`GET {base_url}/webfonts` (default `https://www.googleapis.com/webfonts/v1`). Migrated from
`internal/connectors/google-webfonts` (legacy hand-written connector, read-only). The published
streams are five different sorted views of the same underlying list, distinguished only by the
`sort` query parameter — `webfonts` (API default order, no `sort` sent), `popular_fonts`
(`sort=popularity`), `trending_fonts` (`sort=trending`), `newest_fonts` (`sort=date`), and
`alpha_fonts` (`sort=alpha`), matching legacy's `streamEndpoints` routing table exactly.

## Auth setup

`api_key` (a Google API key) is a required secret, sent as the `key` query parameter via
`base.auth`'s `api_key_query` mode. Legacy hard-errors (`requester`) when the secret is unset;
this bundle's `spec.json` marks `api_key` required, matching that behavior — there is no
credential-free fallback for this connector (unlike searxng's optional Bearer proxy).

## Streams notes

All five streams hit the identical `GET /webfonts` endpoint; only the `sort` query parameter (or
its absence, for the default `webfonts` stream) differs, mirroring legacy's `streamEndpoints` table
in `google-webfonts/streams.go`. Records are extracted from the top-level `items` array. Primary
key is `family` (the API has no numeric id — Google Fonts families are uniquely named); the cursor
field is `lastModified` (kept as the API's own camelCase key, unrenamed, matching legacy's
`fontRecord` which passes it through verbatim rather than renaming it).

Legacy's optional passthrough filters (`family`, `subset`, `category`, `capability`, plus the
API-standard `alt`/`prettyPrint` params) are wired via the `stream.Query` optional-query dialect
(`omit_when_absent: true` on each): each is sent only when configured, omitted entirely otherwise —
matching legacy's `optionalQuery` helper which only sets a query key when the corresponding config
value is non-empty.

Pagination is modeled as `cursor` (`cursor_param: pageToken`, `token_path: nextPageToken`,
`max_pages: 100`), matching legacy's `harvest` loop: the Google Fonts Developer API does not
document real pagination (the full font list is typically returned in one response), but legacy's
loop nonetheless honors an optional `nextPageToken` in the response body (echoed back as
`pageToken`) so the connector keeps working if Google ever adds paging, bounded by
`maxPagesCap = 100`. This bundle reproduces the identical shape and bound. No `stop_path` is
declared — an absent/empty `nextPageToken` is legacy's own (and the engine's default) stop signal.

## Write actions & risks

None. The Google Fonts Developer API is read-only; `capabilities.write` is `false` and this bundle
ships no `writes.json`.

## Known limits

- Legacy's derived `variant_count` and `subset_count` integer fields are modeled with
  `computed_fields` using the engine's typed `length` filter, matching the in-code `lenOf` helper
  for the documented Google Fonts wire shape where `variants` and `subsets` are arrays.
- **Real-world pagination is unobserved.** The Google Fonts Developer API does not document a
  `nextPageToken` response field in its public reference; legacy's tolerance for one appears to be
  defensive/future-proofing rather than an exercised behavior. This bundle's 2-page fixture proves
  the engine's `cursor`/`token_path` paginator mechanically terminates correctly on a synthetic
  `nextPageToken`, matching legacy's identical (also unobserved-in-practice) tolerance.
