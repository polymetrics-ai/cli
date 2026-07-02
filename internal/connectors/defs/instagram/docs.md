# Overview

Instagram is a wave2 fan-out declarative-HTTP migration. It reads Instagram Business/Creator
account profile, media, and stories through the Facebook Graph API
(`https://graph.facebook.com/v23.0`). This bundle is engine-vs-legacy parity-tested against
`internal/connectors/instagram` (the hand-written connector it migrates); the legacy package stays
registered and unchanged until wave6's registry flip. **Status: partial** — 3 of legacy's 4
streams (`users`, `media`, `stories`) are migrated at capability parity; `user_insights` is blocked
(see Known limits) and remains on the legacy implementation.

## Auth setup

Provide a long-lived Instagram access token via the `access_token` secret; it is sent as a Bearer
token (`Authorization: Bearer <access_token>`) and is never logged, matching legacy's
`connsdk.Bearer(secret)`. `base_url` defaults to `https://graph.facebook.com/v23.0` and may be
overridden for tests/proxies. `ig_user_id` (the Instagram Business/Creator Account node id) is
required and is substituted into every stream's path (urlencoded by `InterpolatePath`'s
per-segment default, matching legacy's `instagramUserID` + implicit path composition), matching
legacy's own required-config check.

## Streams notes

`users` (`GET /{ig_user_id}`) returns a single Graph API node rather than a paginated `data[]`
edge; `records: {"path": "", "single_object": true}` maps that one object to one emitted record,
matching legacy's `readSingle` exactly. `media` (`GET /{ig_user_id}/media`) and `stories` (`GET
/{ig_user_id}/stories`) are `data[]` edges; pagination follows the Graph API's absolute
`paging.next` URL convention (`pagination.type: next_url`, `next_url_path: "paging.next"`),
matching legacy's `harvest` loop, which follows `paging.next` as-is and stops when it's empty or a
page yields zero records (the engine's `next_url` paginator's own empty-value stop signal plus its
loop guard against re-requesting the same URL cover this identically). Every list request sends
`limit=100` (matches legacy's default `page_size`) via each stream's static `query: {"limit":
"100"}`, mirroring stripe's `limit=100` static-query precedent — the `next_url` paginator has no
config-driven page-size knob, so a runtime `config.page_size` override (which legacy supports) is
not modeled; see Known limits.

## Write actions & risks

None. Instagram is exposed as a read-only source here; `capabilities.write` is `false` and this
bundle ships no `writes.json`, matching legacy's `Write` returning
`connectors.ErrUnsupportedOperation`.

## Known limits

- **`user_insights` is BLOCKED (ENGINE_GAP) and not migrated.** Legacy's
  `instagramUserInsightRecord` mapper does two things the Tier-1 dialect cannot express: (1) it
  hoists the LAST element of the raw `values[]` array's `value`/`end_time` fields onto the flat
  top-level record (`values[len(values)-1]`) — the engine's `computed_fields`/record-path
  resolution only walks dotted OBJECT-map keys (`resolveRecordPathValue`), with no array-index or
  "last element" accessor at all; and (2) it conditionally synthesizes a derived `id` (`name +
  "_" + end_time`) ONLY when the raw record has no native `id` — the dialect has no
  conditional/fallback-if-absent expression for a computed field (a computed field either resolves
  from its own single template or is a no-op skip; it cannot branch on whether ANOTHER field is
  present). Neither gap can be worked around with `join:<sep>`, `last_path_segment`, or any other
  existing filter without changing the emitted record's actual id/value/end_time DATA for at least
  some real Instagram insight rows — which the migration's meta-rule (conventions.md §5) forbids
  as a silent deviation. This is a genuine Tier-2 `RecordHook` trigger (per-record post-processing
  beyond schema projection), but Tier-2 hooks are out of scope for this wave2 fan-out per the wave
  hard rules (JSON+docs.md only); `user_insights` therefore stays on the legacy Go implementation
  until a follow-up hook-capable wave revisits it. `api_surface.json` records this as an `excluded:
  {category: out_of_scope}` entry with the full technical reason inline.
- **`instagram_business_account_id` config-key fallback is not modeled.** Legacy accepts EITHER
  `ig_user_id` OR, when that's unset, `instagram_business_account_id` as the account node id
  (`instagramUserID`'s two-key fallback). The dialect's path/query templating resolves exactly one
  named `config.*` key per reference with no "try key A, else key B" fallback primitive, so this
  bundle declares `ig_user_id` only (the canonical, first-checked legacy key) as the required
  config property. A caller who previously configured Instagram via
  `instagram_business_account_id` alone (never setting `ig_user_id`) would need to migrate that
  config key name; this is a documented config-surface narrowing, not a data-shape change for any
  caller using the canonical key.
- **`page_size`/`max_pages` are not runtime-configurable.** Legacy exposes `page_size` (default
  100, max 100) and `max_pages` (default unlimited) as config-driven overrides
  (`instagramPageSize`/`instagramMaxPages`). The engine's `next_url` paginator has no
  config-driven page-size or page-count-cap knob (unlike `page_number`/`offset_limit`'s
  `PageSize`/read.go's `MaxPages` field, which `next_url` never reads), so this bundle sends
  Instagram's own default (`limit=100`) as a fixed per-stream query literal and relies solely on
  the empty-`paging.next` stop signal, matching Instagram's own real termination behavior;
  `page_size` is still declared in `spec.json` for documentation parity with legacy's config
  surface even though no template consumes it as a runtime override.
- **`media`/`stories` ship single-page conformance fixtures**, per conventions.md §4's sanctioned
  `next_url` exception: the next-page URL is the fixture replay server's own address, unknown
  until the harness picks a port at runtime, so a static fixture file cannot embed a correct
  second-page absolute URL. Unlike the bitly/calendly precedent this exception cites, this wave's
  hard rules forbid creating any Go files (including `paritytest/instagram`), so 2-page `next_url`
  termination correctness is proven only by this bundle's own conformance dynamic checks against
  the single-page fixture (`read_fixture_nonempty`, `records_match_schema`) and by legacy's own
  existing `instagram_test.go` `TestReadPaginatesAndAuthenticates` (still exercised against the
  legacy package, unaffected by this migration) — not by a dedicated live 2-page parity test in
  this bundle. A follow-up wave with Go-authoring scope should add
  `paritytest/instagram/TestParityInstagram_MediaStreamPaginatesViaAbsoluteNextURL` to close this
  gap the same way bitly/calendly did.
