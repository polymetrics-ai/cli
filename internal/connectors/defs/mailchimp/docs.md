# Overview

Mailchimp is a read-only declarative-HTTP connector that reads Mailchimp Marketing API audiences
(lists), campaigns, reports, and automations through the datacenter-scoped REST API
(`https://<dc>.api.mailchimp.com/3.0`). This bundle migrates `internal/connectors/mailchimp` (the
hand-written legacy connector, which stays registered and unchanged until wave6's registry flip) to
a Tier-1 defs bundle at capability parity.

## Auth setup

Two credential shapes are accepted, matching legacy's `mailchimpAuth` precedence exactly:

1. `access_token` (x-secret) — an OAuth access token, sent as `Authorization: Bearer <access_token>`.
   Takes precedence when both secrets are set (declared FIRST in `base.auth`'s first-match-wins
   candidate list, matching legacy checking `mailchimpAccessToken` before `mailchimpAPIKey`).
2. `api_key` (x-secret) — a Mailchimp API key (e.g. `abc123-us6`), sent as HTTP Basic auth with the
   conventional username `"anystring"` (`connsdk.Basic("anystring", apiKey)` in legacy).

`data_center` (e.g. `"us6"`) is required and builds the base URL directly
(`https://{{ config.data_center }}.api.mailchimp.com/3.0`).

## Streams notes

All 4 streams (`lists`, `campaigns`, `reports`, `automations`) share Mailchimp's
`{<recordsKey>:[...], total_items:N}` envelope and `offset_limit` pagination (`limit_param: count`,
`offset_param: offset`, `page_size: 100`) — advancing by 100 until a short page is returned, matching
legacy's short-page stop condition exactly. Each stream declares an `incremental` block with its own
`since_*` request param, matching legacy's per-resource `sinceParam` table:

| stream | cursor field | since param |
|---|---|---|
| lists | `date_created` | `since_date_created` |
| campaigns | `create_time` | `since_create_time` |
| reports | `send_time` | `since_send_time` |
| automations | `create_time` | `since_create_time` |

The since param is sent only when the incremental lower bound resolves (the sync's persisted cursor,
or else the `start_date` config on a fresh sync) — matching legacy's `if since != ""` guard — and is
sent verbatim (RFC3339, `param_format: rfc3339`), matching legacy (no unix-seconds conversion for
this API).

## Write actions & risks

None. The legacy connector advertises no reverse-ETL actions (`Capabilities.Write: false`);
`capabilities.write` is `false` and no `writes.json` is shipped, matching legacy exactly.

## Known limits

- Legacy also derived the datacenter from the API key's `-us6`-style suffix when `data_center` was
  unset. This bundle requires `data_center` explicitly and drops the API-key-suffix fallback: the
  engine's `spec.json` `"default"` materialization mechanism only fills in a FIXED literal default,
  not one derived from another config/secret value (`docs/migration/conventions.md` §3's
  DERIVED-default note — mailchimp's own base URL is exactly the cited example of a base URL that is
  "a function of another config value, not a fixed literal"). Documented config-surface narrowing,
  not a silent behavior change: any caller that previously relied on API-key-suffix derivation must
  now supply `data_center` directly.
- Legacy's config-driven `page_size`/`max_pages` overrides have no declarative equivalent: the
  engine's `PaginationSpec.PageSize`/`MaxPages` are fixed values baked into `streams.json`'s
  `base.pagination` block, not runtime-config-driven (same class of dead config as searxng's
  wave0 finding, `docs/migration/conventions.md` §3). Neither is declared in `spec.json`; pagination
  is fixed at `page_size: 100` (legacy's own default) with unbounded pages (legacy's own default
  `max_pages` behavior).
- Legacy secret keys were named `credentials.access_token`/`credentials.apikey` (dotted namespacing
  internal to the legacy config store). The engine's `{{ secrets.<key> }}` reference only resolves a
  single segment after the `secrets.` namespace (`interpolate.go`'s `resolveRefValue` splits on the
  first two dot-segments only), so a dotted secret property name cannot be referenced correctly; this
  bundle renames them to `access_token`/`api_key` — a spec-property-naming translation only, with no
  change to wire behavior or credential semantics.
- Only the 4 legacy-parity read streams are implemented; the broader Mailchimp Marketing API surface
  (list members, templates, e-commerce stores, webhooks) is out of scope until Pass B — see
  `api_surface.json`.
