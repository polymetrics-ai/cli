# Overview

Genesys is a declarative-HTTP bundle migrated from `internal/connectors/genesys` (the hand-written
legacy connector, which stays registered and unchanged until wave6's registry flip). It reads
Genesys Cloud users, routing queues, groups, and authorization divisions through the Genesys Cloud
Platform API v2. Legacy is a plain connsdk-based HTTP connector (single OAuth2 client-credentials
auth mode, page-number pagination, no writes, no auth/stream hooks needed) — a pure Tier-1
declarative bundle fully expresses it; no Tier-2 hook or Tier-3 native package is warranted. It is
read-only in both legacy and this bundle (`capabilities.write: false`, no `writes.json`).

**Catalog-label note**: this connector was catalog-labeled native/destination in the source catalog
metadata prior to migration; that label is wrong for this connector. Legacy
(`internal/connectors/genesys/genesys.go`) is a read-only (`Write: false`) connsdk-HTTP source with
no protocol-native (SQL/queue/SDK) dependency of any kind — it is a textbook Tier-1 candidate, not a
Tier-3 native connector and not a write-capable destination. This bundle correctly ships no
`writes.json` and no `internal/connectors/native/genesys/` package.

## Auth setup

Legacy authenticates with an OAuth2 client-credentials grant
(`connsdk.OAuth2ClientCredentials{TokenURL, ClientID, ClientSecret}`, `genesys.go:148`), with an
optional single `scope` config value applied only when non-empty
(`if scope := strings.TrimSpace(cfg.Config["scope"]); scope != ""`, `genesys.go:149-150`). This
bundle's `base.auth` reproduces the identical shape via the engine's native
`oauth2_client_credentials` mode:

```json
{
  "mode": "oauth2_client_credentials",
  "token_url": "{{ config.token_url }}",
  "client_id": "{{ secrets.client_id }}",
  "client_secret": "{{ secrets.client_secret }}",
  "scopes": "{{ config.scope }}"
}
```

`scope` defaults to `""` in `spec.json` (materialized via the engine's `default`-materialization
mechanism, conventions.md §3) so the template always resolves; `buildOAuth2ClientCredentials`
splits the resolved string on whitespace into zero-or-more scopes, so an empty resolved value
produces exactly zero scopes — matching legacy's own "only set scopes when non-empty" branch, not
an approximation of it.

**Narrowed from legacy: `base_url`/`token_url` have no derived default.** Legacy derives both from
a `region` config value (default `mypurecloud.com`): `base_url` = `https://api.<region>/api/v2`
(`genesys.go:162-170`), `token_url` = `https://login.<region>/oauth/token` (`genesys.go:184-193`).
The engine's `spec.json` `"default"` mechanism only materializes a FIXED literal for an absent key —
there is no mechanism to derive one config value's default from another config value (the identical
narrowing auth0's `docs.md` documents for its own region/tenant-derived `audience` default;
conventions.md §3's `spec.json` `"default"` section explicitly calls out this "DERIVED default"
case as requiring either an explicit required field or a not-yet-existing computed-field-style
mechanism). This bundle takes the documented "require it explicitly" path: both `base_url` and
`token_url` are `required` in `spec.json` with no default; operators set them directly (typically
`https://api.mypurecloud.com/api/v2` and `https://login.mypurecloud.com/oauth/token` for the
default US region, matching legacy's own default region's derived values) instead of relying on a
`region` shorthand. This never changes emitted record data for any configuration legacy itself
would accept; it only requires two explicit config values instead of one optional `region`
shorthand. Documented parity deviation, ACCEPTABLE.

## Streams notes

All 4 streams (`users`, `queues`, `groups`, `divisions`) share the identical shape: a flat
`GET <resource>` request, `records.path: "entities"` (Genesys Cloud's uniform collection envelope,
matching legacy's hardcoded `connsdk.RecordsAt(resp.Body, "entities")` call at `genesys.go:105`),
and `pageNumber`/`pageSize` page-number pagination with a short-page stop — matching legacy's
`harvest`, which advances `pageNumber` until a page returns fewer than `pageSize` records.
`page_size` defaults to 100 and is bounded 1-500 in `spec.json`, matching legacy's
`defaultPageSize`/`maxPageSize` constants (`genesys.go:20-21`); `max_pages` defaults to unbounded
(`0`), matching legacy's `maxPages` config parsing (`0`/`all`/`unlimited` all mean unbounded).

Field mapping is a direct 1:1 projection of legacy's `mapRecord` functions:

- `users`: `id`, `name`, `display_name` (= `name`, legacy's `userRecord` deliberately duplicates the
  `name` field onto `display_name`, `streams.go:27-29`), `email`, `state`. This bundle's
  `computed_fields.display_name: "{{ record.name }}"` reproduces the identical duplication (a bare
  single `{{ record.<path> }}` reference copies the raw typed value verbatim per conventions.md §3's
  typed-extraction rule — here a string, so no type-widening concern).
- `queues`: `id`, `name`, `description` (`queueRecord`, `streams.go:31-33`).
- `groups`: identical to `queues` — legacy's `groupRecord` is a bare alias for `queueRecord`
  (`streams.go:35`).
- `divisions`: identical to `queues` — legacy's `divisionRecord` is also a bare alias for
  `queueRecord` (`streams.go:37`).

No stream declares an `incremental` block: legacy's `streams()` catalog declares no
`CursorFields` for any of the 4 streams (unlike freshcaller's `calls.call_time`), so no schema here
declares `x-cursor-field` either — full parity, not a narrowing.

## Write actions & risks

None. Genesys is a read-only source in this bundle, matching legacy
(`Capabilities: connectors.Capabilities{..., Write: false}`, `genesys.go:38`).

## Known limits

- **Conformance dynamic checks are skipped** (`metadata.json`'s `conformance.skip_dynamic`):
  `oauth2_client_credentials` auth's `token_url` is a separate declared `config.token_url`
  property; conformance's replay-server rewiring (`withReplayURL`) only overrides the bundle's
  base request URL used for stream/check paths, never `RuntimeConfig.Config["token_url"]` itself,
  so the token exchange always targets the synthetic non-secret placeholder value
  (`"synthetic-conformance-value"`, not a real URL) and fails before any declarative stream/check
  request is issued — every auth-resolving dynamic check would otherwise fail identically and
  uninformatively. Static checks (spec/schema validity, `interpolations_resolve`, docs/fixtures
  presence, secret redaction) still run and pass. Genesys has no Tier-2 `AuthHook` (auth is fully
  declarative `oauth2_client_credentials`), so there is no `paritytest/genesys` package for this
  wave; the read/pagination/schema-projection shape is proven by structural review against legacy
  `internal/connectors/genesys` instead. Matches `box`/`clazar`/`sendpulse`/`kyriba`'s identical
  documented precedent.
- **`base_url`/`token_url` have no derived default** — see Auth setup above. Documented parity
  deviation, ACCEPTABLE: operators must set both explicitly (typically the default US-region
  derived values) rather than relying on legacy's `region`-shorthand derivation; never changes
  emitted data for any configuration legacy itself would accept.
- Only the 4 legacy-parity streams are implemented; the broader Genesys Cloud Platform API surface
  (conversations, analytics, routing configuration, presence, recordings) is out of scope for this
  wave — see `api_surface.json`'s `excluded` entries.
