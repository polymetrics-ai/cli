# Overview

Lever Hiring reads Lever's Data API (`https://api.lever.co/v1`): opportunities (candidate
profiles), postings, users, requisitions, and pipeline stages. Read-only, full-refresh only. This
bundle migrates `internal/connectors/lever-hiring` (the hand-written connector); the legacy
package stays registered and unchanged until wave6's registry flip.

## Auth setup

Lever accepts either of two credential shapes, with the same precedence legacy enforces
(OAuth access token wins when both are present):

- `access_token` (OAuth2 access token) — sent as `Authorization: Bearer <access_token>`.
- `api_key` — sent as the HTTP Basic auth username with a blank password (Lever's documented
  API-key auth convention).

This is the zendesk-support golden dual-auth shape: `base.auth` is declared as an ordered
first-match-wins candidate list with `access_token`'s bearer candidate FIRST and `api_key`'s basic
candidate as the fallback, each gated by a `when` truthiness check on its own secret so the
both-present case resolves to bearer, exactly matching legacy's `leverAuth`
(`access_token` checked before `api_key`).

`base_url` defaults to `https://api.lever.co/v1`; pass an explicit override to reach
`https://api.sandbox.lever.co/v1` or a test/proxy endpoint.

## Streams notes

All 5 streams share the same shape: `GET` against the Lever list endpoint with `limit=100`,
records at `data`, primary key `["id"]`. Pagination follows Lever's `hasNext`/`next` envelope
(`pagination.type: cursor` with `token_path: next`, `cursor_param: offset`, `stop_path: hasNext`):
the next page is requested with `offset=<next>`, and pagination stops when `hasNext` is falsy
(any value other than the literal `true`) OR the `next` token itself is empty — exactly legacy's
`hasNext != "true" || next == ""` stop condition, and the engine's `tokenPathCursor` paginator's
own stop_path semantics were purpose-built for this exact "boolean stop flag may be populated even
when a token function is present" shape (see Zendesk's `meta.has_more` precedent,
`docs/migration/conventions.md` §3).

`opportunities`, `postings`, `users`, and `requisitions` expose a `createdAt` unix-millis field
that legacy declares as a `CursorFields` catalog hint, but legacy's own harvest loop never uses it
as a server-side incremental filter parameter anywhere — every legacy read is unconditional
full-refresh pagination with no `updated_after`/`since`-style request parameter at all. This
bundle preserves that exact behavior: `x-cursor-field` is declared on each schema (so a
`*_deduped`/`incremental_append` sync mode can still dedupe/order by it client-side per design
§B.6) but no `streams.json` `incremental` block is declared anywhere, matching legacy's genuine
absence of a request-level incremental mechanism. `stages` has neither a cursor field nor a
`createdAt` field upstream, matching legacy exactly.

## Write actions & risks

None. This bundle is read-only (`capabilities.write: false`); Lever Hiring exposes no reverse-ETL
write surface here (matches legacy's `Write` returning `connectors.ErrUnsupportedOperation`
unconditionally).

## Known limits

- Legacy's `environment: sandbox` config shorthand (deriving `https://api.sandbox.lever.co/v1`
  when no explicit `base_url` is set) is not modeled as a separate config key: `base_url`'s
  `spec.json` default is a fixed literal (production), and the engine has no conditional
  base-URL-derivation mechanism (`docs/migration/conventions.md` §3's "derived default" note —
  this is exactly that shape, same as chargebee/sentry's own documented scope narrowing).
  Reaching the sandbox host is still fully supported: pass `base_url` explicitly as
  `https://api.sandbox.lever.co/v1`. Documented scope narrowing (a config-surface shorthand
  removed), not a reachability loss.
- Legacy exposed runtime-configurable `page_size` (1-100) and `max_pages` (0/all/unlimited) config
  knobs. Neither is expressible in this dialect: `PaginationSpec`'s `page_size`/`max_pages` fields
  are fixed JSON literals in `streams.json`, never resolved from `RuntimeConfig.Config` at read
  time (mirrors the stripe golden's own documented `page_size`/`max_pages`-is-dead-config
  precedent, ledger item 3). This bundle fixes the page size at 100 (legacy's own default) and
  leaves pagination unbounded, matching legacy's own default behavior for a caller that never
  overrides either knob.
- Full Lever API surface (opportunity/note/feedback mutation, panels, archive reasons, webhooks)
  is out of scope for this wave; see `api_surface.json`'s `excluded: {category: out_of_scope,
  reason: "not implemented in this bundle"}` entries.
