# Overview

SearXNG is the reference read-only, credential-free declarative-HTTP golden migration (wave0
`wave:F`). It reads web and Reddit search results from a self-hosted SearXNG metasearch instance's
JSON API (`GET {base_url}/search?q=<query>&format=json&pageno=<n>`). This bundle is
engine-vs-legacy parity-tested against `internal/connectors/searxng` (the hand-written connector it
migrates); the legacy package stays registered and unchanged until wave6's registry flip. It is a
pm-native connector (no `catalog_data.json` entry) and is live-registered via
`connectors.RegisterNativeLive("searxng")` in the legacy package's `init()`.

## Auth setup

No credentials are required by default: public SearXNG instances are open. Only `base_url` (your
instance's URL) must be provided; there is no default, so an instance must be named explicitly.

Legacy also supports an optional `api_key` secret sent as a Bearer token for instances behind an
auth proxy (`searxng.go:184-189`). This bundle wires the identical behavior via `streams.json`
`base.auth`'s `when`-gated bearer spec: `{"mode":"bearer","token":"{{ secrets.api_key }}","when":"{{
secrets.api_key }}"}` falling back to `{"mode":"none"}`. This is safe in the common no-credential
case because the engine's `when` truthiness check treats an entirely-absent `secrets.*` key as
falsy (not an error) — the first spec's `when` evaluates false, so `selectAuth` falls through to the
unconditional `none` spec. Parity-tested both ways by `TestParitySearxng_ApiKeySecretSendsBearerAuth`
/ `TestParitySearxng_ApiKeyAbsentSendsNoAuth`.

## Streams notes

Both streams (`search`, `reddit`) hit the same `GET /search` endpoint with `format=json`; they
differ only in how the `q` query parameter is scoped. `search` sends `config.query` verbatim.
`reddit` prefixes it with `site:reddit.com ` so any general SearXNG instance (without a dedicated
Reddit engine installed) returns Reddit results, matching legacy's `searxngQuery`
(`searxng.go:225-242`) base case (no `subreddit` configured). Records are extracted from the
top-level `results` array; primary key is `url`; `published_date` is declared as the cursor field
for manifest-surface parity, but neither legacy nor this bundle actually filters or advances reads
by it — SearXNG results aren't reliably orderable by a server-side cursor, so both connectors always
perform a full stream read.

`engines` (the raw API's `engines[]` array) is comma-joined into a string via a `join:,`
`computed_fields` filter, matching legacy's `joinAny` output byte-for-byte. `stream` is emitted as a
static-literal `computed_fields` value (`"search"`/`"reddit"`), matching legacy's derived marker
field. Both were previously documented deviations (parity-deviation ledger entries 4 and 6,
`docs/migration/conventions.md`) that changed the emitted record shape versus legacy; both are now
RESOLVED via the engine's `join:<sep>` filter and static-literal `computed_fields` support, and
`parity_searxng_test.go` asserts RAW record equality (no normalization/stripping) for both fields.

Pagination is `pageno` (1-based `page_number` pagination), matching SearXNG's own pagination
scheme. No page-size query parameter is ever sent (SearXNG's per-page result count is
engine/instance-defined; the bundle's `page_size: 10` is used purely as the client-side short-page
stop threshold, exactly like legacy's `searxngDefaultPageSize`/`PageNumberPaginator.SizeParam: ""`).
`max_pages: 1` (legacy's own default, `searxngDefaultMaxPages`) IS enforced by the engine read path
as a hard request-count cap, independent of page fullness — see
`TestParitySearxng_MaxPagesStop`.

## Write actions & risks

None. SearXNG is a read-only metasearch API with no mutation endpoints; `capabilities.write` is
`false` and this bundle ships no `writes.json`.

## Known limits

- **Subreddit-narrowing is not modeled.** Legacy's `reddit` stream additionally supports a
  `subreddit` config value that scopes the query to `site:reddit.com/r/<sub>` instead of the bare
  `site:reddit.com`. The engine's declarative query templating (`stream.Query`) has no
  conditional/default-value filter (an unconditional `{{ config.subreddit }}` reference hard-errors
  when the key is absent, since only `auth`/`when` templating tolerates absent-key-falsy — query
  resolution does not), so a subreddit-present-vs-absent branch cannot be expressed without risking
  an unresolved-key error when `subreddit` is unset (the common case). This bundle models only the
  base case (no subreddit configured, matching legacy's own default query-scoping behavior when
  `subreddit` is unset); the parity suite documents and tests against this base case only
  (`TestParitySearxng_RedditStreamScopesQuery`). Deliberately out-of-scope config, not a defect.
  `categories`/`engines`/`language`/`time_range`/`safesearch` (legacy's other optional SearXNG filter
  passthroughs) have the identical limitation and are, for the same reason, not declared in
  `spec.json` at all (F6, REVIEW.md: a declared-but-unwireable config key is worse than an absent
  one — see `docs/migration/conventions.md`'s "declared config must be consumed" rule).
