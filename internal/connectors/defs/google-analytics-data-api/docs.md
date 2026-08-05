# Overview

Google Analytics 4 (GA4) reads are backed by the Google Analytics Data API v1beta and v1alpha. The provider-derived audit retrieved on 2026-08-05 uses Google's published discovery artifacts (both revision `20260803`): 11 v1beta methods plus 15 v1alpha methods, with equivalent `getMetadata` and `runReport` methods counted once under the programme counting policy. The resulting semantic inventory is 24 operations: 20 reads and 4 writes, not the prior invalid 10-row hook-only baseline.

Readable streams: `daily_active_users`, `website_overview`, `traffic_sources`, `devices`, `pages`.

Executable operations: `runReport` through five native presets plus 10 bounded direct GET reads: `metadata get`, `audience-exports list`, `audience-exports get`, `property-quotas get`, `audience-lists list`, `audience-lists get`, `recurring-audience-lists list`, `recurring-audience-lists get`, `report-tasks list`, and `report-tasks get`.

Official source evidence: https://analyticsdata.googleapis.com/$discovery/rest?version=v1beta, https://analyticsdata.googleapis.com/$discovery/rest?version=v1alpha, and https://developers.google.com/analytics/devguides/reporting/data/v1/rest. The connector ledger records the retrieval date, revisions, semantic de-duplication, operation mapping, and each blocker in `api_surface.json`.

## Auth setup

Use a GA4 OAuth2 bearer access token with Analytics Data API read access. Add it through credential storage from an environment variable or stdin, never in prompt text, docs, shell history, or committed fixtures. The connector accepts `access_token`, legacy `credentials`, and the older flattened `credentials.access_token` secret key forms.

Required config: `property_ids` (one or more numeric GA4 property IDs). Direct commands may also use `property_id`; if omitted, native code falls back to the first `property_ids` entry. `base_url` defaults to `https://analyticsdata.googleapis.com` and should only be overridden for local fixture tests.

## Streams notes

The five stream commands are fixed GA4 `runReport` presets implemented by the native connector. GA4 report requests use POST on the wire, but the streams emit warehouse records rather than exposing a raw POST body. Native pagination uses `offset`/`limit`, defaults to page size `10000`, caps page size at `250000`, and honors `max_pages` when set. Date-dimensioned streams carry `date` as their cursor and use client-side cursor filtering in this native slice.

- `daily_active_users`: date, active users, new users, sessions.
- `website_overview`: date-level active users, new users, sessions, page views, average session duration, and bounce rate.
- `traffic_sources`: date/source/medium sessions and users.
- `devices`: date/device category/OS/browser users and sessions.
- `pages`: date/page path/page title page views and engagement.

Bounded direct reads use fixed endpoint metadata, required path variables, byte caps, and `json_redacted` output policy. They do not accept method, URL, or arbitrary body flags.

## Write actions & risks

No reverse-ETL write actions are executable. The four official creates — `audienceExports.create`, `audienceLists.create`, `recurringAudienceLists.create`, and `reportTasks.create` — are typed as planned operations because each initiates provider-side work and needs a connector-owned closed action schema, redaction, idempotency or operation handling, sanitized fixtures, and plan -> preview -> explicit approval -> execute evidence before it can be advertised. There are no destructive delete operations in the audited v1beta/v1alpha Data API surface.

## Known limits

- Current provider-derived inventory: 24 semantic operations (20 reads, 4 writes). Truthful disposition: 11 executable read operations (`runReport` through five presets plus 10 direct GET operations), 13 blocked/planned (9 POST reads and 4 creates), 0 excluded/not-applicable, 0 certified. The old 10-row baseline was not provider-derived and is superseded.
- The 9 blocked POST reads are `batchRunReports`, `runPivotReport`, `batchRunPivotReports`, `runRealtimeReport`, `checkCompatibility`, `audienceExports.query`, `runFunnelReport`, `audienceLists.query`, and `reportTasks.query`. They require the precise shared provider-query/redaction dependency #2985 and connector-owned bounded closed schemas plus fixtures; none is exposed through generic raw API flags.
- All four creates remain blocked as future typed reverse-ETL operations, not as executable commands. They require the lifecycle and operation/idempotency evidence described above.
- Sanitized fixture coverage covers all 11 executable operations: five `runReport` stream fixtures and ten direct-read fixtures. The direct-read fixture tests assert the exact fixed GET path against a local `httptest` server.
- Conformance dynamic replay is intentionally skipped for this native-backed bundle because the real stream behavior is native POST body construction. Connector-owned native tests and sanitized fixtures are the substitute proof.
- No live provider behavior is certified by these fixtures or docs.
