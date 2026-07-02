# Overview

NewsData.io is a read-only news feed API. This bundle migrates `internal/connectors/newsdata-io`
(the hand-written legacy connector) to a declarative Tier-1 bundle at full capability parity: it
reads the same four streams (`latest`, `crypto`, `archive`, `sources`) through the same NewsData.io
v1 REST endpoints, using the identical `apikey` query-param auth and `nextPage`/`page` body-token
pagination. The legacy package stays registered and unchanged until the wave6 registry flip.

## Auth setup

Provide a NewsData.io API key via the `api_key` secret; it is sent as the `apikey` query parameter
(`api_key_query` auth mode) on every request, matching legacy's `connsdk.APIKeyQuery("apikey",
secret)`. It is never logged.

## Streams notes

`latest`, `crypto`, and `archive` share an identical shape: `GET` against the matching NewsData.io
endpoint, records at `results`, primary key `article_id`, `pubDate` declared as `x-cursor-field` for
manifest-surface parity only — legacy never actually filters or advances reads by `pubDate` (no
server-side or client-side incremental filter exists in `newsdata_io.go`), so no `incremental` block
is declared here either; both connectors always perform a full stream read every sync.

Pagination follows NewsData.io's `nextPage`/`page` body-token convention
(`pagination.type: cursor` with `token_path: nextPage`, `cursor_param: page`): the next page is
requested with `page=<nextPage>` until the response's `nextPage` is null/empty, matching legacy's
`harvest` loop exactly. `sources` is NOT paginated (legacy's `harvest` returns after the first page
for this stream unconditionally); this bundle overrides `pagination: {"type": "none"}` at the
stream level to reproduce that exact single-page behavior.

Every article stream sends `size=<page_size>` (default 10, matching legacy's
`newsdataDefaultPageSize`) via the `stream.Query` optional-query dialect's `default` (never omitted,
mirrors legacy always setting `size` for non-sources streams). The optional filter passthroughs
(`q`/`category`/`country`/`language`/`domain`, plus `archive`'s `from_date`/`to_date`) use
`omit_when_absent: true` so an unset filter is left off the request entirely, matching legacy's
`newsdataQueryFilters`' `setIf` helper (only set a param when the corresponding config value is
non-empty).

`max_pages` defaults to `10` (legacy's own default bound for the otherwise-unbounded NewsData.io
corpus); `0`/`all`/`unlimited` config values mean unbounded, matching legacy's `newsdataMaxPages`.

## Write actions & risks

None. NewsData.io is a read-only news feed with no reverse-ETL write surface; `capabilities.write`
is `false` and this bundle ships no `writes.json`, matching legacy's `ErrUnsupportedOperation` Write
stub.

## Known limits

- `pubDate` is declared as `x-cursor-field` for catalog/manifest parity with legacy's
  `CursorFields: []string{"pubDate"}`, but neither legacy nor this bundle filters reads by it — both
  always perform a full sync. This is not a narrowing; it mirrors legacy's actual (non-)behavior
  exactly.
- Full NewsData.io API surface (news author search, generate podcast, other premium endpoints) is
  out of scope; only the 4 legacy-parity streams are implemented, matching legacy's own stream set
  one-for-one.
