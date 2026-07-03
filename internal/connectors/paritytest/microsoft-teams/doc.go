package microsoftteamsparity_test

// Engine-vs-legacy parity suite for the microsoft-teams quarantine-repair
// migration (docs/migration/quarantine.json's "microsoft-teams" entry,
// blocker_type ENGINE_GAP). Drives BOTH connectors live against the SAME
// httptest server, asserting RAW connectors.Record equality
// (reflect.DeepEqual) plus legacy-side sanity assertions, following the
// wave0/wave1-pilot goldens (parity_stripe_test.go, parity_sentry_test.go).
//
// Resolution: Microsoft Graph's @odata.nextLink pagination cursor is an
// absolute URL read from a response-body key containing a literal "."
// ("@odata.nextLink"). The engine's declarative next_url pagination type
// reads its cursor via connsdk.StringAt's dotted-path traversal
// (engine/paginate.go), which necessarily treats any "." in a path as a
// nesting separator -- there is no way to address a literal dotted key with
// that parser. This is a genuine ENGINE_GAP (identical to
// microsoft-entra-id/microsoft-lists), not an auth gap: legacy's auth is a
// plain OAuth2 client-credentials grant, fully expressible via the engine's
// declarative oauth2_client_credentials mode -- no AuthHook needed at all.
// hooks/microsoft-teams/hooks.go implements StreamHook only, porting
// legacy's harvest/nextLink loop verbatim.
//
// Because conformance/dynamic.go's dynamic checks call engine.Read with
// hooks=nil (they exercise the declarative fallback path only, never
// dispatching through a StreamHook), this bundle declares streams.json
// base.pagination as {"type":"none"} and marks every stream
// conformance.skip_dynamic -- the declarative path is never actually taken
// in production (the StreamHook handles every stream, always returning
// handled=true). This parity suite is the authoritative proof of real
// @odata.nextLink 2-page termination the skip_dynamic markers name.
//
// Legacy read in full (internal/connectors/microsoft-teams/{microsoft-teams,streams}.go,
// read-only reference):
//   - auth: OAuth2 client-credentials grant (connsdk.OAuth2ClientCredentials)
//     against a per-tenant Microsoft Entra ID token endpoint
//     (https://login.microsoftonline.com/<tenant_id>/oauth2/v2.0/token,
//     graphTokenURL) -> engine oauth2_client_credentials auth mode, a
//     two-candidate list: an explicit token_url override first (tests), then
//     the tenant_id-derived endpoint (login_base_url + tenant_id),
//     mirroring the sharepoint-lists-enterprise golden's identical
//     dual-candidate shape.
//   - base URL: https://graph.microsoft.com/v1.0 default, or an explicit
//     base_url override.
//   - 4 streams (microsoft-teams/streams.go's graphStreamEndpoints): users
//     (/users), groups (/groups), channels (/teams/getAllChannels),
//     team_device_usage_report (/reports/getTeamsDeviceUsageUserDetail,
//     scoped by a period query param: D7/D30/D90/D180, default D7). Every
//     endpoint returns {value:[...], "@odata.nextLink":"<absolute-url>"}.
//   - pagination: @odata.nextLink absolute-URL follow (microsoft-teams.go's
//     harvest/nextLink) -- ported into the StreamHook exactly, including the
//     literal-dotted-key JSON decode (never connsdk.StringAt).
//   - no incremental cursor is ever sent as a request param anywhere in
//     legacy (CursorFields is nil on every stream, microsoft-teams/streams.go's
//     graphStreams). This bundle matches that by declaring no incremental
//     block at all (full_refresh only).
//   - read-only: Write returns connectors.ErrUnsupportedOperation
//     (microsoft-teams.go:235-237) -> capabilities.write: false, no
//     writes.json.
//   - Check: GET /organization?$top=1, requires client_secret/tenant_id
//     secrets and client_id config (microsoft-teams.go:66-89).
import _ "polymetrics.ai/internal/connectors/hooks/microsoft-teams" // triggers engine.RegisterHooks("microsoft-teams", ...) init side effect for parity
