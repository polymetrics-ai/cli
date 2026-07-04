# Overview

SharePoint Lists Enterprise reads and writes SharePoint lists and list items through Microsoft
Graph (`https://graph.microsoft.com/v1.0/sites/<site id>/lists[/<list id>/items]`). Read behavior
is capability-parity migrated from `internal/connectors/sharepoint-lists-enterprise` (the
hand-written connector it migrates; the legacy package stays registered and unchanged until
wave6's registry flip and was read-only). This Pass B expansion adds list/list-item create+update
write actions the legacy connector never implemented, going beyond strict read-only parity per the
Pass B full-surface-expansion charter (see Write actions & risks).

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

Both streams declare `projection: "passthrough"` (conventions.md §8 rule 1): legacy's `Read` uses a
single shared `connsdk.Harvest` call for both streams whose per-record callback is
`return emit(connectors.Record(rec))` (`sharepoint_lists_enterprise.go:98-100`) — the raw decoded
page record is emitted verbatim with no field-built `connectors.Record{...}` mapping anywhere in
the read path. Schema-mode projection would silently drop any wire fields not enumerated in
`schemas/lists.json`/`schemas/list_items.json`; passthrough is required to preserve full-record
parity. The schemas remain a documentation surface only (conventions.md §8 rule 1's "schema stays
documentation surface").

## Write actions & risks

Pass B adds 4 write actions covering the create+update surface of Microsoft Graph's `list`/
`listItem` resources (Graph docs: `list-create.md`, `list-update` shape documented alongside
`list-get`, `listitem-create.md`, `listitem-update.md`):

- `create_list` (`kind: create`, `POST /sites/{{ config.site_id }}/lists`) — creates a new list
  (with any custom `columns`/`list.template` declared in the submitted record) on the configured
  site. Low-risk, no approval required.
- `update_list` (`kind: update`, `PATCH /sites/{{ config.site_id }}/lists/{{ record.id }}`) —
  mutates an existing list's `displayName`/`description` by id.
- `create_list_item` (`kind: create`, `POST /sites/{{ config.site_id }}/lists/{{
  config.list_id }}/items`) — creates a new item (row) in the configured list; the submitted
  record must wrap column values in a `fields` object, matching Graph's own listItem-create
  request shape exactly.
- `update_list_item` (`kind: update`, `PATCH /sites/{{ config.site_id }}/lists/{{
  config.list_id }}/items/{{ record.id }}/fields`) — mutates an existing item's column values via
  Graph's `fields` sub-resource (a `fieldValueSet`); the request body is the record minus `id`
  (the path already carries it), matching Graph's own partial-update semantics for this endpoint
  exactly — only the submitted column names change, every other column is left alone. This is the
  practically useful update shape (Graph's separate bare `PATCH .../items/{item-id}` endpoint,
  documented under the identical "Update listItem" method, is excluded in `api_surface.json` as
  `duplicate_of` this action — see there for the reasoning).

**Deliberately NOT implemented**: `DELETE /sites/{site-id}/lists/{list-id}` (whole-list delete) and
`DELETE /sites/{site-id}/lists/{list-id}/items/{item-id}` (item delete) — both excluded in
`api_surface.json` as `destructive_admin`. Legacy exposed zero mutation capability at all; this
expansion's write additions are scoped to the reversible/low-risk create+update surface, never a
destructive delete of an entire list or its rows.

`create_list`/`update_list` need only `site_id`; `create_list_item`/`update_list_item` additionally
need `list_id` configured (the list whose items are being written), mirroring the `list_items`
read stream's own `list_id` requirement.

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
  `base.pagination` block (set to `100`, matching legacy's own default,
  `sharepoint_lists_enterprise.go:21`, `defaultPageSize`), and there is no per-request `max_pages`
  override mechanism at all (conventions.md §3). Neither key is declared in `spec.json` (a
  declared-but-unwireable key is worse than an absent one — searxng precedent). The `lists` stream's
  committed 2-page conformance fixture (`fixtures/streams/lists/{page_1,page_2}.json`, 100 records
  then 1) proves pagination termination at this same page size, per conventions.md §4's
  2-page-fixture-required rule.
- Legacy's own default `max_pages` is `1` (`sharepoint_lists_enterprise.go:22`,
  `defaultMaxPages`); this bundle's unset `MaxPages` (unbounded, stopped only by the short-page
  signal) is a strict superset of legacy's default single-page behavior — no caller-visible
  behavior regresses for the common case (fewer records than one page).
- `base_url` config override exists in `spec.json` for test/proxy use (matching legacy's own
  override check, `sharepoint_lists_enterprise.go:127`) but is not exercised by any fixture
  (dynamic checks are skipped bundle-wide, above). `token_url`'s override precedence IS fully
  modeled via the dual-candidate `when`-gated auth mechanism described above, not left as a gap.
