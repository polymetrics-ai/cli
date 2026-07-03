# Overview

Facebook Pages is a wave4 declarative-HTTP migration. It reads Facebook Page metadata and posts
through the Facebook Graph API v19.0 (`GET https://graph.facebook.com/v19.0/...`). The legacy
package (`internal/connectors/facebook-pages`) is connsdk-HTTP-based — a plain `connsdk.Requester`
issuing `GET` requests, no signature auth, no custom hooks — so this migrates to a pure Tier-1
declarative bundle per `docs/migration/conventions.md`'s tier ladder (§6 item 1). This bundle is
parity-tested against the legacy package, which stays registered and unchanged until wave6's
registry flip.

## Auth setup

Provide a Facebook Page access token via the `access_token` secret; it is sent as a Bearer token
(`Authorization: Bearer <access_token>`) and is never logged, matching legacy's
`connsdk.Bearer(token)` (`facebook_pages.go:151`). `page_id` is required and identifies the page
whose metadata/posts are read (legacy: `facebook-pages connector requires config page_id`).
`base_url` defaults to `https://graph.facebook.com/v19.0` and may be overridden for tests/proxies,
matching legacy's own `defaultBaseURL` fallback with the same absolute-http(s)-URL validation
intent (`facebook_pages.go:184`); the engine itself does not re-validate the URL shape at bundle
level, but every parity/conformance fixture only ever points at an httptest server, so this is not
exercised differently on either side.

## Streams notes

`page` is a single-object, non-paginated read (`GET /{page_id}?fields=id,name,category,fan_count,
link`); the whole response body is one record (`records.path: ""`, `single_object: true`),
matching legacy's `connsdk.RecordsAt(resp.Body, "")` call in `Read`'s `"page"` case
(`facebook_pages.go:84`). No incremental cursor — legacy's catalog declares none for this stream.

`posts` reads `GET /{page_id}/posts?fields=id,message,created_time,updated_time,permalink_url&
limit={{ config.page_size, default 100 }}`; records live at the `data` key
(`facebook_pages.go:114`). Pagination follows the Graph API's own `paging.next` **absolute** URL
convention (`pagination.type: next_url`, `next_url_path: "paging.next"`), matching legacy's own
`connsdk.StringAt(resp.Body, "paging.next")` loop (`facebook_pages.go:123-131`) exactly: an empty
`paging.next` value stops pagination. `page_size` is wired via the opt-in optional-query dialect
(`"limit": {"template": "{{ config.page_size }}", "default": "100"}`) so an unset `page_size`
config value falls back to the literal `100`, matching legacy's own `pageSize` clamp-to-100-when-
invalid-or-unset behavior (`facebook_pages.go:158-164`) for the common (unset) case. A declared
`x-cursor-field: updated_time` matches legacy's own `CursorFields: []string{"updated_time"}`
catalog declaration (`facebook_pages.go:61`); legacy never sends a server-side incremental filter
parameter for this stream (it always reads the full post list unfiltered), so this bundle declares
a bare `incremental.cursor_field` with no `request_param`, per conventions.md §8's incremental
truth table.

## Write actions & risks

None. This is a read-only source — `capabilities.write` is `false` and this bundle ships no
`writes.json`, matching legacy's `Write` always returning `connectors.ErrUnsupportedOperation`.

## Known limits

- **Post insight/metric fan-out is not modeled.** The legacy package's own doc comment states this
  explicitly: "post insight fan-out is not implemented because it requires per-post metric
  selection and permissions." This bundle carries the identical scope narrowing forward
  (`api_surface.json` excludes `/{post-id}/insights` as `out_of_scope`).
- **Legacy's fixture-mode-only `fixture: true` marker field is not modeled.** Legacy's
  `readFixture` path (only reached when `config.mode == "fixture"`, a credential-free
  conformance-harness affordance) stamps an extra `fixture: true` field onto every canned record
  (`facebook_pages.go:166-182`). This is not part of the LIVE record shape; this bundle's schemas
  target the live path only. The engine's own conformance/fixture-replay harness provides the
  credential-free test affordance this bundle needs, so no fixture-mode equivalent is needed here.
