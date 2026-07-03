# Overview

Marketo is a marketing automation platform whose REST API is served from a tenant-specific
identity host. This bundle is a pure Tier-1 declarative migration of
`internal/connectors/marketo` (the hand-written legacy connector): legacy is a thin
`connsdk.Requester`-based GET+paginate+map connector — plain Bearer auth, no signature scheme, no
async jobs, no compound writes — so every behavior it implements (base-URL validation, cursor
pagination via `nextPageToken`/`moreResult`, page-size/max-pages config, the `activities` stream's
optional `activityTypeIds` filter, and per-stream field projection) maps directly onto
`streams.json`'s declarative dialect. No `hooks/marketo/` package or native component split is
warranted. The legacy package stays registered and unchanged until wave6's registry flip; the
catalog inventory's `runtime_kind: "native_go"` label reflects only that legacy happens to be a
hand-written Go package, not that it needs one — reading `internal/connectors/marketo/marketo.go`
shows every code path is declarative-HTTP-shaped (the catalog's `type: "source"` label is
accurate; only the `runtime_kind` classification undersells how declarative the legacy
implementation already is).

## Auth setup

Provide `base_url` (required; your tenant's REST identity host, ending in `/rest/v1` — e.g.
`https://123-ABC-456.mktorest.com/rest/v1`; there is no default, matching legacy's own
"caller must supply an identity host" requirement) and an `access_token` secret, sent as a Bearer
token (`Authorization: Bearer <access_token>`). This connector, like legacy, does **not** refresh
OAuth tokens internally — the caller must supply a valid, unexpired access token obtained
out-of-band (Marketo's client-credentials token endpoint).

## Streams notes

All 3 streams share Marketo's `nextPageToken`/`moreResult` cursor pagination convention
(`pagination.type: cursor` with `token_path: "nextPageToken"`, `cursor_param: "nextPageToken"`,
`stop_path: "moreResult"`): the next page's `nextPageToken` request parameter is read from the
previous response's `nextPageToken` body field, and pagination stops when EITHER `nextPageToken` is
empty OR `moreResult` is not the literal `true` — matching legacy `harvest`'s exact
`next == "" || !strings.EqualFold(more, "true")` stop condition (the engine's `stop_path` check is
the declarative equivalent of that second condition). Every request also sends `batchSize` from the
`page_size` config value (default 300, legacy's `marketoDefaultPageSize`, also legacy's
`marketoMaxPageSize` — Marketo's REST API hard-caps batch size at 300 for every endpoint this
bundle reads).

- `leads` (`GET /leads.json`, records at `result`): emits `id`/`email`/`updatedAt`/`createdAt`
  directly, matching legacy `marketoLeadRecord` exactly.
- `programs` (`GET /programs.json`, records at `result`): emits `id`/`name`/`updatedAt`/`createdAt`
  directly, matching legacy `marketoProgramRecord` exactly.
- `activities` (`GET /activities.json`, records at `result`): emits
  `id`/`activityTypeId`/`activityDate`/`leadId` directly, matching legacy `marketoActivityRecord`
  exactly. The optional `activity_type_ids` config value is sent as the `activityTypeIds` query
  parameter on this stream ONLY, via the opt-in `omit_when_absent` query-param object form (declared
  but not in `spec.json`'s `required[]`) — matching legacy's exact per-endpoint conditional
  (`endpoint.resource == "activities.json" && activity_type_ids != ""`). When unset in live use, the
  parameter is omitted entirely, exactly like legacy.

None of the 3 streams declares an `incremental` block or `x-cursor-field`: legacy's own
`marketoStreams()` catalog declares only `PrimaryKey`, never `CursorFields`, for any stream (no
stream is read incrementally, and none is ever request-param-filtered by an updatedAt/createdAt
lower bound) — per `docs/migration/conventions.md` §8's incremental truth table, a bare
`cursor_field` is declared only when legacy actually publishes `CursorFields`, so this bundle
correctly declares none for any stream, matching legacy's real (full-refresh-only) behavior.

## Write actions & risks

None. This connector is read-only by legacy's own design (`Write` unconditionally returns
`connectors.ErrUnsupportedOperation`, and the package doc comment says so explicitly);
`capabilities.write` is `false` and this bundle ships no `writes.json`.

## Known limits

- Full Marketo API surface (campaigns, lists, lead-database bulk import/export jobs, custom
  objects) is out of scope; see `api_surface.json`'s `excluded` entries. Only the 3 legacy-parity
  streams are implemented.
- `page_size`'s legacy-enforced numeric bound (1-300) and `max_pages`'s `all`/`unlimited`
  string-literal acceptance are not separately validated by the engine dialect (no numeric-range or
  enum validator on a `spec.json` string property) — an out-of-range value is sent to the API
  verbatim rather than rejected client-side as legacy's `intConfig`/`maxPagesConfig` would. This
  changes only the client-side error surface for an already-invalid config value, never
  accepted-input behavior for any value legacy itself would accept.
- This bundle does not refresh or manage Marketo OAuth tokens, matching legacy exactly — supplying
  a fresh `access_token` before each sync (or via an external token-refresh mechanism) remains the
  operator's responsibility.
