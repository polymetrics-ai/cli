// Engine-vs-legacy parity suite for the microsoft-lists quarantine-repair
// migration (docs/migration/quarantine.json's microsoft-lists ENGINE_GAP
// entry). Drives BOTH connectors live against the SAME httptest server,
// asserting RAW connectors.Record equality (reflect.DeepEqual), following
// the sentry/microsoft-entra-id golden pattern.
//
// Resolution: identical @odata.nextLink dotted-literal-key pagination gap
// as microsoft-entra-id/microsoft-teams — the engine's declarative
// next_url pagination type reads its cursor via connsdk.StringAt's
// dotted-path parser, which cannot address the literal dotted key
// "@odata.nextLink" at all. Resolved by a Tier-2 StreamHook
// (hooks/microsoft-lists/hooks.go) porting legacy's harvest/nextLink logic
// exactly.
//
// Legacy read in full (internal/connectors/microsoft-lists/{microsoft-lists,streams}.go,
// read-only reference):
//   - auth: connsdk.OAuth2ClientCredentials against
//     https://login.microsoftonline.com/<tenant_id>/oauth2/v2.0/token (or an
//     explicit token_url override), scope
//     "https://graph.microsoft.com/.default" -> engine
//     oauth2_client_credentials auth mode, dual when-gated candidates.
//   - base URL: https://graph.microsoft.com/v1.0 (default) or an explicit
//     base_url override.
//   - 4 streams, all scoped under sites/{site_id}/: lists (site-scoped),
//     list_items/columns/content_types (require config.list_id; list_items
//     additionally sends $expand=fields).
//   - pagination: @odata.nextLink absolute-URL follow (harvest/nextLink) —
//     ported into the StreamHook exactly.
//   - no incremental cursor is ever sent as a request param anywhere in
//     legacy — this bundle matches that by declaring no request_param on
//     any stream's incremental block (CursorFields are catalog-only).
//   - read-only: Write returns connectors.ErrUnsupportedOperation ->
//     capabilities.write: false, no writes.json.
//   - Check: GET /sites/{site_id}/lists, requires
//     client_id/client_secret/tenant_id secrets, site_id config, and a
//     resolvable base URL/token URL.
package microsoftlistsparity_test

import (
	_ "polymetrics.ai/internal/connectors/hooks/microsoft-lists" // triggers engine.RegisterHooks("microsoft-lists", ...) init side effect for parity
)
