# Overview

Picqer is a wave2 fan-out declarative-HTTP migration. It reads Picqer products, customers,
orders, picklists, warehouses, and suppliers through the Picqer REST API
(`GET https://<organization>.picqer.com/api/v1/<resource>`). This bundle is a capability-parity
port of `internal/connectors/picqer` (the hand-written connector it migrates); the legacy package
stays registered and unchanged until wave6's registry flip.

## Auth setup

Picqer uses HTTP Basic auth with the API key as the username and an empty password
(`connsdk.Basic(key, password)`, `picqer.go:134`). Provide the key via the `api_key` secret
(preferred) or the plain-config `username` fallback; the dual-candidate `auth` list in
`streams.json` reproduces legacy's exact precedence
(`firstNonEmpty(secret(cfg,"api_key"), cfg.Config["username"])`, `picqer.go:129`): the `api_key`
secret candidate is declared first and gated on its own presence (`when`), falling through to the
`username` config candidate when `api_key` is unset. An optional `password` secret is honored on
either branch (Picqer's own convention leaves it blank).

`organization_name` derives the base URL as `https://<organization_name>.picqer.com/api/v1`,
matching legacy's own derivation (`picqer.go:144`). See Known limits for the one narrowing this
bundle makes versus legacy's config surface.

## Streams notes

All six streams (`products`, `customers`, `orders`, `picklists`, `warehouses`, `suppliers`) are
simple list endpoints (`GET /<resource>`) with `"projection": "passthrough"` — legacy's
`mapRecord` returns every raw field unchanged (`out := connectors.Record(rec)`) and only adds an
`id` alias, it never drops fields via schema-shaped filtering (`picqer.go:101-112`). This bundle
reproduces that exactly: every raw field survives, and a `computed_fields.id` entry copies the
resource-specific numeric key (`idproduct`/`idcustomer`/`idorder`/`idpicklist`/`idwarehouse`/
`idsupplier`) into a uniform `id` field via typed bare-reference extraction (preserves Picqer's
real integer wire type, no stringification).

Pagination is `offset_limit` with `offset_param: offset` and no `limit_param` — matching legacy's
`connsdk.OffsetPaginator{OffsetParam: "offset", PageSize: size}` (`picqer.go:82`), which never
sends a page-size query parameter at all (`LimitParam` is left empty in legacy); the paginator
stops when a page returns fewer than `page_size` (100) records, a purely client-side threshold,
never a server-enforced page size.

## Write actions & risks

None. Picqer's legacy connector is read-only (`Capabilities.Write: false`); this bundle ships no
`writes.json`.

## Known limits

- **Explicit `base_url` override is not modeled; `organization_name` is now required.** Legacy
  accepts either an explicit `base_url` (checked first) or derives
  `https://<organization_name>.picqer.com/api/v1` when `base_url` is unset (`picqer.go:136-145`).
  The engine's `streams.json` `base.url` is a single non-conditional template (unlike `auth`, which
  supports a `when`-gated candidate list) — there is no mechanism to express "prefer this literal
  override, else derive from this other config key" in one field. Per
  `docs/migration/conventions.md`'s guidance for a derived base URL, this bundle requires
  `organization_name` and drops the `base_url`-override path; a caller who previously pointed the
  legacy connector at a fixed `base_url` (e.g. a proxy or non-standard Picqer deployment) cannot do
  so through this bundle. This is a documented, deliberate config-surface narrowing, not a silent
  behavior change for the common case (an organization-name-driven Picqer tenant).
- **Legacy's defensive `out["id"] == nil` guard is not modeled.** Legacy only back-fills `id` from
  the resource-specific key when the raw record has no pre-existing `id` field at all; since
  Picqer's real API never emits a bare `id` field on these resources (only the prefixed
  `id<resource>` keys), this guard is dead code in practice and the engine's unconditional
  `computed_fields.id` copy is capability parity for every real Picqer response.
