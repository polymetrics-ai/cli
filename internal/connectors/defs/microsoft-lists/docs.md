# Overview

Microsoft Lists is a Tier-2 quarantine-repair migration (previously quarantined `ENGINE_GAP`,
`docs/migration/quarantine.json`). It reads SharePoint/Microsoft Lists, list items, columns, and
content types from a site through the Microsoft Graph API (`GET /sites/<site
id>/lists[/<list id>/items|columns|contentTypes]`), read-only. This bundle is capability-parity
migrated from `internal/connectors/microsoft-lists` (the hand-written connector it migrates); the
legacy package stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide `site_id` (SharePoint site) plus `client_id`, `client_secret`, and `tenant_id` secrets.
Auth is `oauth2_client_credentials` with two `when`-gated candidates evaluated in declared order
(`docs/migration/conventions.md` §3's dual-auth-ordering pattern, applied to two candidates of the
SAME mode — the identical mechanism `sharepoint-lists-enterprise`/`microsoft-teams`/
`microsoft-entra-id` use for the same Graph token-endpoint shape): the first candidate uses
`config.token_url` directly and is gated `when: {{ config.token_url }}` (matches only when a full
override is configured, e.g. for a test server); the second, unconditional candidate derives the
endpoint as `{{ config.login_base_url }}/{{ secrets.tenant_id }}/oauth2/v2.0/token` (defaulting
`login_base_url` to `https://login.microsoftonline.com`). This exactly reproduces legacy's own
override precedence (`microsoft-lists.go`'s `tokenURL`: an explicit `token_url` config override
always wins; the derived tenant-scoped endpoint is the fallback). Both candidates use the
`config.scope` value (default `https://graph.microsoft.com/.default`, matching legacy's
`graphDefaultScope` constant). None of `client_id`/`client_secret`/`tenant_id` is ever logged.

## Streams notes

Four streams, all scoped under `/sites/{{ config.site_id }}/`: `lists` (site-scoped, no `list_id`
needed), `list_items`/`columns`/`content_types` (all require `config.list_id`, declared in
`spec.json` but not in `required[]` — matching legacy's own per-stream `needsListID` check,
`microsoft-lists/streams.go`: `lists` never needs `list_id`, the other three hard-error without it).
An absent `list_id` on a `list_items`/`columns`/`content_types` read is a runtime path-interpolation
error from the StreamHook, exactly like legacy's own explicit
`"microsoft-lists stream ... requires config list_id"` error — same failure mode, hook-native error
text instead of a hand-written one. Every endpoint returns `{"value": [...], "@odata.nextLink":
"<url>"}` — records live at `records.path: "value"`, matching Graph's real wire shape.
`list_items` additionally sends `$expand=fields` on every request (legacy's
`streamEndpoints["list_items"].query`), which the hook applies identically.

Every schema property is a snake_case rename of the raw Graph camelCase field (e.g. `display_name`
from `displayName`, `last_modified_date_time` from `lastModifiedDateTime`), matching legacy's own
`mapRecord` functions (`microsoft-lists/streams.go`) field-for-field; `lists`' nested
`list.template` facet is flattened to `list_template`, and `list_items`' nested `contentType.id`
facet is flattened to `content_type_id`, exactly as legacy's `listRecord`/`listItemRecord` do.
`lists` and `list_items` declare `x-cursor-field: last_modified_date_time` (no `request_param`) to
match legacy's own `CursorFields: []string{"last_modified_date_time"}` catalog declaration — legacy
never actually filters server-side by this field, so no stream sends an incremental filter param;
the field only enables correct sync-mode derivation, matching legacy's own behavior exactly.

**Pagination — Tier-2 StreamHook, not declarative (the actual blocker behind this connector's prior
quarantine)**: identical `@odata.nextLink` shape to `microsoft-entra-id`/`microsoft-teams` — a JSON
key containing a literal dot, carrying the next page's full absolute URL. Legacy hand-rolls this
exactly (`microsoft-lists.go`'s `harvest`/`nextLink`): GET the resource (with any per-stream extra
query merged in, `$top` set from `page_size` on the FIRST request only), extract `value[]`, decode
the top-level `@odata.nextLink` string directly (not via a dotted-path helper), and if non-empty,
re-request that exact URL verbatim. The engine's declarative `next_url` pagination type reads its
cursor via `connsdk.StringAt`'s dotted-path parser, which splits on `.` and therefore cannot address
a literal key containing a dot — a genuine, confirmed `ENGINE_GAP` (see
`docs/migration/quarantine.json`'s `microsoft-lists` entry), not a config or fixture issue. This
recurs identically for `microsoft-entra-id`/`microsoft-teams` (all three share the exact same Graph
shape) — below the ≥3-recurrence bar for a general dotted-path-escape engine mechanism at the time of
this migration, and Tier-2 fully resolves it without an engine change.

`hooks/microsoft-lists/hooks.go` implements `StreamHook`, porting legacy's `harvest`/`nextLink`
logic exactly. Every stream in this bundle carries an explicit `"conformance": {"skip_dynamic":
true, "reason": "..."}` marker: `internal/connectors/conformance/dynamic.go` honors this marker by
Skipping every dynamic fixture-replay check for these streams, since the StreamHook (always
`handled=true`) is what every real `Read()` call actually dispatches through. The authoritative
substitute this marker names is `paritytest/microsoft-lists`'s dedicated 2-page `@odata.nextLink`
test (`TestParityMicrosoftLists_ListsNextLinkPagination`) and `hooks/microsoft-lists/hooks_test.go`.
`streams.json`'s own `base.pagination` stays declared `{"type": "none"}` since it is never
dynamically exercised now.

`max_pages` is a hook-consumed `spec.json` config value (permissive parse: empty/`all`/`unlimited`
means unbounded), matching legacy's own `graphMaxPages` parsing exactly — it is NOT a declarative
`streams.json` field, since pagination itself is entirely hook-driven.

## Write actions & risks

None. Microsoft Lists is read-only in legacy (`Write` returns
`connectors.ErrUnsupportedOperation`, `microsoft-lists.go`); `capabilities.write` is `false` and no
`writes.json` is declared.

## Known limits

- Full Microsoft Graph SharePoint site surface (drives, pages, permissions, site activity, etc.) is
  out of scope for this migration; see `api_surface.json`'s `excluded: {category: out_of_scope,
  reason: "Pass B capability expansion"}` entries. Only the 4 legacy-parity read streams are
  implemented.
- **Dynamic conformance checks are skipped, stream-by-stream (every stream carries the marker) and
  also at the bundle level in `metadata.json`** — pagination is hook-driven for every stream, so no
  stream's dynamic fixture-replay check can usefully exercise the real code path. Static checks
  (spec/schema validity, `interpolations_resolve`, docs/fixtures presence, secret redaction) still
  run and pass. Parity for the pagination/schema-projection shape is proven by
  `paritytest/microsoft-lists` (`TestParityMicrosoftLists_ListsNextLinkPagination`, a live
  `httptest.Server`-backed 2-page `@odata.nextLink` follow test) and
  `hooks/microsoft-lists/hooks_test.go`'s unit coverage — this mirrors the identical, already-accepted
  `sentry`/`microsoft-teams`/`microsoft-entra-id`/`sharepoint-lists-enterprise` `skip_dynamic`
  precedents.
- `page_size` is runtime-configurable (`config.page_size`, default 100, matching legacy's
  `graphDefaultPageSize`) and forwarded as the `$top` query parameter on the FIRST request of each
  stream's sub-sequence only — subsequent pages follow `@odata.nextLink` verbatim, exactly matching
  legacy's `harvest` loop.
- Candidate future engine feature: a `next_url_path` "literal key" escape (e.g. a
  `next_url_literal_key` field naming an exact top-level JSON key to read verbatim, bypassing
  dotted-path splitting) would let this connector, `microsoft-entra-id`, and `microsoft-teams` all
  drop their StreamHooks. Not implemented in this migration per the ENGINE_GAP recurrence rule
  (`conventions.md` §6) scoping decision recorded at the time each of these three connectors was
  migrated — revisit if a 4th connector hits the identical shape.
