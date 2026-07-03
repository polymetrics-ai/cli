// Engine-vs-legacy parity suite for the microsoft-entra-id quarantine-repair
// migration (docs/migration/quarantine.json's microsoft-entra-id ENGINE_GAP
// entry). Drives BOTH connectors live against the SAME httptest server,
// asserting RAW connectors.Record equality (reflect.DeepEqual), following
// the sentry golden pattern (internal/connectors/paritytest/sentry).
//
// Resolution: Microsoft Graph's list pagination is "@odata.nextLink" — a
// JSON key containing a literal dot — carrying the next page's full
// absolute URL. The engine's declarative next_url pagination type reads its
// cursor via connsdk.StringAt's dotted-path parser, which splits on "." and
// therefore cannot address a literal key containing a dot. This is the
// confirmed ENGINE_GAP behind this connector's prior quarantine; it is
// resolved by a Tier-2 StreamHook (hooks/microsoft-entra-id/hooks.go)
// porting legacy's harvest/nextLink logic exactly, rather than an engine
// change (below the ≥3-recurrence bar at the time of this migration; see
// defs/microsoft-entra-id/docs.md's Streams notes for the full reasoning).
//
// Legacy read in full (internal/connectors/microsoft-entra-id/{microsoft-entra-id,streams}.go,
// read-only reference):
//   - auth: connsdk.OAuth2ClientCredentials against
//     https://login.microsoftonline.com/<tenant_id>/oauth2/v2.0/token (or an
//     explicit token_url override), scope
//     "https://graph.microsoft.com/.default" -> engine
//     oauth2_client_credentials auth mode, dual when-gated candidates
//     (token_url override first, derived tenant endpoint fallback).
//   - base URL: https://graph.microsoft.com/v1.0 (default) or an explicit
//     base_url override.
//   - 5 streams, routed by a fixed resource table (streams.go's
//     streamEndpoints): users ("/users"), groups ("/groups"), applications
//     ("/applications"), serviceprincipals ("/servicePrincipals"),
//     directoryroles ("/directoryRoles"). Every endpoint returns
//     {"value": [...], "@odata.nextLink": "<url>"}.
//   - pagination: @odata.nextLink absolute-URL follow (harvest/nextLink) —
//     ported into the StreamHook exactly, including the $top query param
//     sent only on the first request of each stream's sub-sequence.
//   - no incremental cursor is ever sent as a request param anywhere in
//     legacy — this bundle matches that by declaring no incremental block
//     on any stream (full_refresh only).
//   - read-only: Write returns connectors.ErrUnsupportedOperation ->
//     capabilities.write: false, no writes.json.
//   - Check: GET /users?$top=1, requires client_id/client_secret/tenant_id
//     secrets and a resolvable base URL/token URL.
package microsoftentraidparity_test

import (
	_ "polymetrics.ai/internal/connectors/hooks/microsoft-entra-id" // triggers engine.RegisterHooks("microsoft-entra-id", ...) init side effect for parity, mirroring the sentry/github precedent of blank-importing the hooks package from the parity test
)
