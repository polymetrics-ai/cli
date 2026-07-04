# Overview

Mailgun is a wave2 fan-out declarative-HTTP migration. It reads Mailgun sending domains, email
events, mailing lists, and analytics tags through the Mailgun v3 REST API
(`GET https://api.mailgun.net/v3/...`). This bundle migrates `internal/connectors/mailgun` (the
hand-written connector it replaces at capability parity); the legacy package stays registered and
unchanged until wave6's registry flip.

## Auth setup

Provide a Mailgun account private API key via the `private_key` secret; it is sent as the password
of HTTP Basic auth with the literal username `api` (`Authorization: Basic base64("api:<private_key>")`),
matching legacy's `connsdk.Basic(mailgunBasicUser, secret)`. It is never logged.

## Streams notes

`domains` (`GET /v3/domains`) pages with Mailgun's skip/limit offset convention over an
`{"items":[...],"total_count":N}` envelope (`pagination.type: offset_limit`, `limit_param: limit`,
`offset_param: skip`), matching legacy's `harvestOffset`. `events` (`GET /v3/{domain}/events`),
`mailing_lists` (`GET /v3/lists/pages`), and `tags` (`GET /v3/{domain}/tags`) instead follow
Mailgun's absolute `paging.next` URL convention over the same `{"items":[...],"paging":{"next":...}}`
envelope (`pagination.type: next_url`, `next_url_path: paging.next`), matching legacy's
`harvestPagingNext` — the engine's `next_url` paginator applies the same same-host SSRF guard and
same-URL loop guard legacy's own `seen` map implements.

`events` and `tags` are domain-scoped: their `path` templates substitute the required
`domain_name` config value (`/v3/{{ config.domain_name }}/events`, `/v3/{{ config.domain_name }}/tags`),
matching legacy's `resolveResource`'s `{domain}` token substitution. An absent `domain_name` hard-errors
on both sides (legacy: `"mailgun config domain_name is required for this stream"`; engine: an
unresolved `config.domain_name` path-template key) — same failure classification, different literal
text, per `conventions.md` §5's precedent for config-validation parity.

Legacy's `mailgunEventRecord`/`mailgunTagRecord` rename the raw API's hyphenated field names
(`message-id`, `log-level`, `first-seen`, `last-seen`) to snake_case; this bundle reproduces the
identical renames via `computed_fields` (`"message_id": "{{ record.message-id }}"`, etc. — the
hyphen is preserved as a literal map-key segment since the dotted-path walker splits only on `.`,
never `-`).

Legacy's `mailgunDomainRecord`/`mailgunTagRecord` compute `id`/`tag` via
`firstNonEmpty(item, "id", "name")` / `firstNonEmpty(item, "tag", "name")` — a fallback across two
candidate keys. Mailgun's real `/v3/domains` and `/v3/{domain}/tags` wire responses only ever carry
`name`/`tag` respectively (no `id` key on a domain object, no `name` key on a tag object per
Mailgun's own published schema), so the fallback's first candidate never actually resolves against
real traffic — this bundle's `computed_fields` (`"id": "{{ record.name }}"`,
`"tag": "{{ record.tag }}"`) reproduces the fallback's REAL effective behavior against genuine
Mailgun responses. See Known limits for the documented deviation.

`page_size` (spec default `100`, matching legacy's `mailgunDefaultPageSize`) is sent as the `limit`
query param on the first request of every stream; `next_url`-paginated streams re-send it on every
subsequent page too (the engine merges `stream.Query` onto every page, including an absolute
next-page URL — see bitly's identical documented divergence in its own `docs.md`), which is benign
here since Mailgun's own `paging.next` URL already carries an identical `limit` value.

## Write actions & risks

None. Mailgun is a read-only source connector (legacy's `Write` returns
`connectors.ErrUnsupportedOperation`); `capabilities.write` is `false` and this bundle ships no
`writes.json`.

## Known limits

- **EU region base URL is not derived from a `domain_region` config flag.** Legacy resolves the
  default base URL to `https://api.mailgun.net` (US) or `https://api.eu.mailgun.net` (EU) based on
  a `domain_region` config value when `base_url` itself is unset. The engine's `spec.json`
  `"default"` materialization mechanism only supports a single FIXED literal default, not a
  config-value-conditional derivation (conventions.md's `default` materialization note — the same
  class of gap as sentry's hostname-derived URL or chargebee's site-derived URL). This bundle
  defaults `base_url` to the US host and requires EU-region operators to override `base_url`
  explicitly to `https://api.eu.mailgun.net`; `domain_region` itself is not declared in `spec.json`
  (a declared-but-unwireable config key is worse than an absent one, per F6/REVIEW.md).
- **`max_pages` is not modeled.** Legacy exposes a config-driven `max_pages` hard request-count cap
  (`mailgunMaxPages`) applied uniformly across both its offset and paging-next read loops. Neither
  the `offset_limit` nor `next_url` paginator constructors read `PaginationSpec.MaxPages` from a
  runtime config value (only `page_number`/`offset_limit`'s own page_size are engine-configurable,
  and `next_url` has no page-count knob at all) — pagination is bounded only by the short/empty-page
  stop signal (a short offset page, or an empty/repeated `paging.next` value), matching Mailgun's
  own real termination behavior. `max_pages` is not declared in `spec.json`.
- **`id` (domains) / `tag` (tags) are single-source, not a genuine two-key fallback.** Legacy computes
  both via `firstNonEmpty(item, "id"/"tag", "name")` — a fallback across two candidate keys. The
  engine's `computed_fields` dialect has no coalesce-across-multiple-references primitive (only a
  bare single reference, a filter chain, or a static literal), so this bundle wires the field
  directly to the candidate Mailgun's real wire shape actually populates (`record.name` for
  domains' `id`, `record.tag` for tags' `tag`) rather than reproducing the fallback mechanically.
  This is DATA-equivalent for every real Mailgun response (the first candidate is never present on
  the wire in practice) but would diverge from legacy only for a hypothetical response that populated
  the normally-absent first candidate — not a shape Mailgun's documented API ever produces.
- Full Mailgun API surface (sending mail, suppressions, webhooks, templates, routes, IP pools,
  per-tag stats) is out of scope for wave2; see `api_surface.json`'s `excluded: {category:
  out_of_scope, reason: "not implemented in this bundle"}` entries. Only the 4 legacy-parity read
  streams are implemented.
