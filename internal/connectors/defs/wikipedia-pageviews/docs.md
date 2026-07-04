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

Both streams declare `projection: "passthrough"` (§8 rule 1): legacy's `Read` decodes each stream's
response body with `connsdk.RecordsAt(resp.Body, endpoint.recordsPath)` and, for every decoded item,
only conditionally backfills the `id` key in place (`if item["id"] == nil { item["id"] =
recordID(item, req.Config) }`) before emitting the *same* raw map verbatim —
`emit(connectors.Record(item))` (`wikipedia_pageviews.go`) never field-builds a new
`connectors.Record{...}` from named keys. This bundle's `computed_fields.id` reproduces that
conditional backfill declaratively (matching the `defs/picqer` precedent of pairing
`projection: "passthrough"` with a `computed_fields`-derived `id`); `passthrough` on top of it
reproduces legacy's verbatim pass-through of every other raw field. The real wire shape carries more
fields than either schema previously declared — confirmed by this bundle's own fixtures:
`pageviews` items also carry `granularity` and `agent`; `top_articles` items also carry `project`,
`access`, `year`, `month`, `day` alongside the per-item `articles[]` array. Both schemas now list
these as documented (nullable) properties for accuracy; schema-mode projection would have silently
dropped them without `passthrough`.

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
  `recordID` helper builds `id` from `item["project"]`/`item["article"]`/`item["timestamp"]`. A real
  `top_articles` API item carries `project`/`access`/`year`/`month`/`day`/`country`/`articles[]`,
  with no per-item `article`/`timestamp` field and no `id`. Because the record's own `project` field
  *is* present, `recordID` uses it directly (`parts[0] = fmt.Sprint(item["project"])`); the
  `cfg.Config["project"]` fallback only fires when `parts[0] == "<nil>"`, which never happens on this
  wire shape. The `article`/`timestamp` fields are absent, so `fmt.Sprint(nil)` renders them as
  `"<nil>"`. Legacy therefore emits the id `"<record.project>:<nil>:<nil>"` for every record — a
  real, if odd, legacy behavior. This bundle reproduces it exactly via the computed field
  `"id": "{{ record.project }}:<nil>:<nil>"` rather than "fixing" it into something more useful, per
  the meta-rule that a deviation from legacy's actual emitted data is never acceptable even when the
  legacy behavior looks like a bug.
- **`country` is declared as a raw schema property but is rarely present on real `items` entries**
  (it is a path parameter, not part of the response body, on the real Wikimedia API); legacy's own
  Field declaration includes it regardless, so this bundle's schema does too, for byte-for-byte
  field-set parity with legacy's catalog.
