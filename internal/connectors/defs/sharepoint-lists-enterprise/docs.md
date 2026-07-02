# Overview

SharePoint Lists Enterprise is a wave2 fan-out declarative-HTTP migration. It reads SharePoint
lists and list items through Microsoft Graph (`GET https://graph.microsoft.com/v1.0/sites/<site
id>/lists[/<list id>/items]`). This bundle is capability-parity migrated from
`internal/connectors/sharepoint-lists-enterprise` (the hand-written connector it migrates); the
legacy package stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide `tenant_id` (Azure AD tenant), `site_id` (SharePoint site), a `client_id` secret, and a
`client_secret` secret. Auth is `oauth2_client_credentials` with two `when`-gated candidates
evaluated in declared order (conventions.md §3's dual-auth-ordering pattern, applied to two
candidates of the SAME mode rather than two different modes): the first candidate uses
`config.token_url` directly and is gated `when: {{ config.token_url }}` (matches only when a full
override is configured); the second, unconditional candidate derives the endpoint as
`{{ config.login_base_url }}/{{ config.tenant_id }}/oauth2/v2.0/token` (defaulting
`login_base_url` to `https://login.microsoftonline.com`). This exactly reproduces legacy's own
override precedence: `if override := cfg.Config["token_url"]; override != "" { tokenURL = override
}` (`sharepoint_lists_enterprise.go:122-124`) checked AFTER computing the derived URL — an explicit
`token_url` always wins, the derived tenant-scoped endpoint is the fallback. Both candidates use
the fixed scope `https://graph.microsoft.com/.default`, matching legacy's `graphScope` constant.
Neither `client_id` nor `client_secret` is ever logged.

## Streams notes

Both streams share the same shape: `GET /sites/{{ config.site_id }}/lists[...]`, records at the
response body's `value` array, and `offset_limit` pagination (`$top`/`$skip` query params) — an
exact port of legacy's `connsdk.OffsetPaginator{LimitParam: "$top", OffsetParam: "$skip", PageSize:
pageSize}` (`sharepoint_lists_enterprise.go:97`). A page returning fewer records than `page_size`
signals the last page.

- `lists`: `GET /sites/{site_id}/lists`.
- `list_items`: `GET /sites/{site_id}/lists/{{ config.list_id }}/items` — requires `list_id`
  (declared in `spec.json` but not in `required[]`, matching legacy's own per-stream
  `resourcePath` check, `sharepoint_lists_enterprise.go:148-150`: `lists` never needs `list_id`,
  `list_items` hard-errors without it). An absent `list_id` on a `list_items` read is therefore a
  runtime path-interpolation error, exactly like legacy's own explicit
  `"sharepoint-lists-enterprise list_items stream requires config list_id"` error — same failure
  mode, engine-native error text instead of a hand-written one.

Both streams declare `incremental.cursor_field: lastModifiedDateTime` (no `request_param`) to match
legacy's own `CursorFields: []string{"lastModifiedDateTime"}` catalog declaration
(`sharepoint_lists_enterprise.go:157-158`) — legacy never actually filters server-side by this
field (no `$filter`-style incremental query param exists in legacy's `Read`), so this bundle
likewise sends no incremental filter param; the field only enables correct sync-mode derivation
(`incremental_append[_deduped]`), matching legacy's own behavior exactly.

## Write actions & risks

None. SharePoint Lists Enterprise is read-only (`capabilities.write: false`, no `writes.json`),
matching legacy's `Write` returning `connectors.ErrUnsupportedOperation`.

## Known limits

- **Dynamic conformance checks are skipped bundle-wide** (`metadata.json`'s
  `conformance.skip_dynamic: true`). `oauth2_client_credentials`'s `token_url` is derived from
  `config.login_base_url` + `config.tenant_id` (Azure AD's per-tenant OAuth2 endpoint), not from
  `base_url` — conformance's synthetic non-secret config value for both is not a resolvable URL, so
  the token exchange fails before any declarative request is ever issued, and every auth-resolving
  dynamic check (`check_fixture`, every `read_fixture_nonempty:<stream>`, `pagination_terminates`,
  `records_match_schema`, `cursor_advances`) would otherwise fail identically and uninformatively.
  Static checks (spec/schema validity, `interpolations_resolve`, docs/fixtures presence, secret
  redaction) still run and pass. This bundle has no Tier-2 hook, so there is no `paritytest`
  package for this wave; parity for the read/pagination/schema-projection shape is proven by
  structural review against legacy `internal/connectors/sharepoint-lists-enterprise`
  (`sharepoint_lists_enterprise_test.go`'s `TestReadListsUsesClientCredentialsBearer` documents the
  exact same token-endpoint/scope/pagination shape this bundle declares). This mirrors the
  identical, already-accepted `sendpulse` `oauth2_client_credentials` `skip_dynamic` precedent.
- **`page_size`/`max_pages` are not runtime-configurable per the engine dialect.** Legacy exposes
  both as config-driven overrides (`sharepoint_lists_enterprise.go:89-96`,
  `positiveInt`/`parseMaxPages`, `page_size` clamped 1-999, `max_pages` defaulting to 1). The
  `offset_limit` paginator's `page_size` is a fixed value baked into `streams.json`'s
  `base.pagination` block, and there is no per-request `max_pages` override mechanism at all
  (conventions.md §3). Neither key is declared in `spec.json` (a declared-but-unwireable key is
  worse than an absent one — searxng precedent).
- **`page_size` is baked at `2`, not legacy's own default of `100`.** A deliberate, documented
  fixture-authoring choice (conventions.md §4's 2-page-fixture-required rule), not a live-caller
  behavior change: the short-page stop contract applies identically at any page size, only the
  request count for a large result set changes. Bumping back to 100 pre-launch is a one-line
  follow-up.
- Legacy's own default `max_pages` is `1` (`sharepoint_lists_enterprise.go:22`,
  `defaultMaxPages`); this bundle's unset `MaxPages` (unbounded, stopped only by the short-page
  signal) is a strict superset of legacy's default single-page behavior — no caller-visible
  behavior regresses for the common case (fewer records than one page).
- `base_url` config override exists in `spec.json` for test/proxy use (matching legacy's own
  override check, `sharepoint_lists_enterprise.go:127`) but is not exercised by any fixture
  (dynamic checks are skipped bundle-wide, above). `token_url`'s override precedence IS fully
  modeled via the dual-candidate `when`-gated auth mechanism described above, not left as a gap.
