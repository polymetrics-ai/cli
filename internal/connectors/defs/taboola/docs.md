# Overview

Taboola reads campaigns through the Taboola Backstage API (`https://backstage.taboola.com`). This
bundle migrates `internal/connectors/taboola` (the hand-written legacy connector) at capability
parity; the legacy package stays registered and unchanged until wave6's registry flip. Taboola is
read-only here — legacy has no write surface, so `capabilities.write` is `false` and no
`writes.json` is shipped.

## Auth setup

Provide `client_id` and `client_secret` secrets (a Taboola Backstage OAuth2 client-credentials
application). The bundle exchanges them for a Bearer access token via
`mode: oauth2_client_credentials`, `token_url: {{ config.base_url }}/backstage/oauth/token` — the
identical derived endpoint legacy's `requester` constructs
(`strings.TrimRight(base, "/") + "/backstage/oauth/token"`). Both secrets are never logged.
`account_id` is a required config value sent as a path segment on the campaigns endpoint
(`urlencode`d and traversal-guarded by the engine's path interpolation, matching legacy's own
`ContainsAny(id, "/?#")`/`Contains(id, "..")` validation intent).

## Streams notes

The single `campaigns` stream (`GET /backstage/api/1.0/{account_id}/campaigns`) uses page-number
pagination (`pagination.type: page_number`, `page_param: page`, `size_param: page_size`,
`start_page: 1`, default `page_size: 100`), records at `results` — matching legacy's
`connsdk.PageNumberPaginator{PageParam: "page", SizeParam: "page_size", StartPage: 1}` exactly.
Legacy's published catalog declares `CursorFields: ["created_at"]`, mirrored here via the schema's
`x-cursor-field`, but legacy never actually derives a request filter or client-side drop from a
persisted cursor for this stream — every read re-emits the complete campaign set. This bundle
reproduces that exactly: no `incremental` block is declared, so `campaigns` is full-refresh only
(declaring `client_filtered` here would silently drop records legacy would still emit).

## Write actions & risks

None. Taboola is read-only in legacy (`Capabilities.Write` is `false`); `Write` always returns
`connectors.ErrUnsupportedOperation`. No `writes.json` is shipped for this bundle.

## Known limits

- Legacy accepts an optional `token_url` config override to replace the derived OAuth2 token
  endpoint. Since the derived default (`{{ config.base_url }}/backstage/oauth/token`) is itself a
  template (not a fixed spec-default literal) and the engine's `oauth2_client_credentials` auth spec
  has exactly one `token_url` field with no override-vs-derive branching mechanism, this bundle
  always derives `token_url` from `base_url` and drops the standalone override. Documented
  config-surface narrowing — the derived value matches legacy's own default, and no caller who
  relies on the default token endpoint is affected; only the (rarely used) explicit override is
  dropped.
- **Dynamic conformance is skipped for this bundle** (`metadata.json`'s `conformance.skip_dynamic`):
  `token_url`'s derivation from `config.base_url` resolves, under conformance's synthetic non-secret
  config value, to an unreachable non-URL (`synthetic-conformance-value/backstage/oauth/token`), so
  the OAuth token exchange fails before any declarative stream/check request is ever issued. Static
  checks (spec/schema validity, interpolation resolution, docs/fixtures presence, secret redaction)
  still run and pass. This bundle has no Tier-2 `AuthHook` (auth is fully declarative
  `oauth2_client_credentials`), so there is no `paritytest/taboola` package for this wave; the
  read/pagination/schema-projection shape is proven by structural review against legacy
  `internal/connectors/taboola` instead — the same documented precedent as dwolla, clazar, and
  sendpulse.
- Full Taboola Backstage surface (campaign items, reports, audiences, creative previews) is out of
  scope for this wave; see `api_surface.json`'s `excluded: {category: out_of_scope}` entries.
- `metadata.json` declares no `rate_limit` block: legacy enforces no client-side rate limiting for
  Taboola, so none is added here (matching legacy's real, lack-of, throttling behavior).
