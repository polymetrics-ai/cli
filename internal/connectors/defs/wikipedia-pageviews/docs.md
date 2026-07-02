# Overview

Wikipedia Pageviews is a wave2 fan-out declarative-HTTP migration. It reads Wikimedia pageview
metrics through the public Wikimedia REST API (`https://wikimedia.org/api/rest_v1/metrics/pageviews/...`),
requiring no authentication. This bundle migrates `internal/connectors/wikipedia-pageviews` (the
hand-written connector it replaces); the legacy package stays registered and unchanged until
wave6's registry flip.

## Auth setup

None. Wikimedia's pageviews REST API is public and unauthenticated (`base.auth: [{"mode": "none"}]`),
matching legacy's `requester` which never attaches any `Auth` to the underlying `connsdk.Requester`.
`base_url` defaults to `https://wikimedia.org` and may be overridden for tests/proxies, matching
legacy's `defaultBaseURL` constant and `baseURL` validation (scheme + host required).

## Streams notes

`pageviews` builds the per-article daily pageviews path
`api/rest_v1/metrics/pageviews/per-article/{project}/{access}/{agent}/{article}/daily/{start}/{end}`
directly from six required config values (`project`, `access`, `agent`, `article`, `start`, `end`),
matching legacy's `pageviewsPath` exactly (each segment percent-encoded, matching legacy's
`url.PathEscape` per segment — `InterpolatePath`'s default per-segment `urlencode` filter). Records
live at the `items` key. There is no pagination on this endpoint (legacy never paginates it either).
`id` is a computed field synthesizing legacy's own `recordID(item, cfg)`: `"{{ record.project }}:{{
record.article }}:{{ record.timestamp }}"`, reproducing legacy's colon-joined composite key exactly
(legacy only overrides `parts[0]` from `cfg.Config["project"]` when the raw record's own `project`
field is nil, which never happens on the real per-article pageviews wire shape — every real record
carries its own `project` field identical to the requested one).

`top_articles` builds the top-per-country path
`api/rest_v1/metrics/pageviews/top-per-country/{project}/{country}/{access}/{year}/{month}/{day}`.
Records live at the `items` key; there is no pagination (legacy never paginates it either).

## Write actions & risks

None. This connector is read-only in both legacy and this bundle (`capabilities.write: false`); no
`writes.json` is shipped.

## Known limits

- **`top_articles`'s date derivation from a single `start` value is not modeled.** Legacy's
  `topArticlesPath` derives `year`/`month`/`day` by truncating the `start` config value to its
  first 8 characters (silently accepting either `YYYYMMDD` or `YYYYMMDDHH` and discarding any hour
  suffix) and slicing that into `date[:4]`/`date[4:6]`/`date[6:8]`. The engine's template dialect
  has no string-slicing filter, so this bundle instead declares three separate required-when-reading
  `top_articles` config keys (`year`, `month`, `day`) that the caller must supply directly, rather
  than deriving them from a combined `start` value. This is a documented config-surface narrowing,
  not a silent behavior change: any caller who previously passed a single `start` value for
  `top_articles` must now split it into `year`/`month`/`day` themselves; the emitted path segments
  (and therefore the request and its data) are byte-identical once the caller does so.
- **`top_articles`'s emitted `id` field reproduces a real but degenerate legacy quirk.** Legacy's
  `recordID` helper builds `id` from `item["project"]`/`item["article"]`/`item["timestamp"]`, none
  of which exist on a real `top_articles` API item (that shape carries `project`/`access`/
  `year`/`month`/`day`/`country`/`articles[]` instead, with no per-item `article`/`timestamp`
  field and no `id`). Because `fmt.Sprint(nil)` renders `"<nil>"`, legacy's own `top_articles`
  records are therefore emitted with the literal string id `"<config.project>:<nil>:<nil>"` for
  every record — a real, if odd, legacy behavior. This bundle reproduces it exactly via the
  computed field `"id": "{{ config.project }}:<nil>:<nil>"` rather than "fixing" it into something
  more useful, per the meta-rule that a deviation from legacy's actual emitted data is never
  acceptable even when the legacy behavior looks like a bug.
- **`country` is declared as a raw schema property but is rarely present on real `items` entries**
  (it is a path parameter, not part of the response body, on the real Wikimedia API); legacy's own
  Field declaration includes it regardless, so this bundle's schema does too, for byte-for-byte
  field-set parity with legacy's catalog.
