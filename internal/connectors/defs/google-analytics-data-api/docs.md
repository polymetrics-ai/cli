# Overview

Google Analytics 4 (GA4) reads are backed by the Google Analytics Data API v1beta. The current official discovery audit (revision `20260729`) lists 11 operations. This connector exposes five native `runReport` preset streams and three bounded GET direct reads for metadata and audience-export metadata. Other POST report/query/check operations remain blocked/planned rather than exposed as raw API calls.

Readable streams: `daily_active_users`, `website_overview`, `traffic_sources`, `devices`, `pages`.

Executable direct reads: `metadata get`, `audience-exports list`, and `audience-exports get`.

Official source evidence: https://analyticsdata.googleapis.com/$discovery/rest?version=v1beta and https://developers.google.com/analytics/devguides/reporting/data/v1/rest.

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

No reverse-ETL write actions are declared. The official `audienceExports.create` operation initiates a provider-side asynchronous export and remains blocked/planned until it has a named action, closed schema, redaction, idempotency or operation polling evidence, fixture coverage, and plan -> preview -> explicit approval -> execute support. There are no destructive delete operations in the audited v1beta Data API surface.

## Known limits

- Current official inventory: 11 operations. Truthful post-change disposition: 4 official operations executable (`runReport` via five preset streams plus GET `getMetadata`, `audienceExports.list`, and `audienceExports.get`), 4 official operations fixture-tested locally, 7 blocked/planned, 0 excluded/not-applicable, 0 certified.
- POST report/query/check operations (`runRealtimeReport`, `runPivotReport`, `batchRunReports`, `batchRunPivotReports`, `checkCompatibility`, and `audienceExports.query`) are blocked pending shared provider search/query foundation #2985 and connector-owned POST-read fixtures/redaction evidence.
- `audienceExports.create` is blocked as a future typed reverse-ETL-style create operation; it is not advertised as executable in this slice.
- Conformance dynamic replay is intentionally skipped for this native-backed bundle because the real stream behavior is native POST body construction. Connector-owned native tests and sanitized fixtures are the substitute proof.
- No live provider behavior is certified by these fixtures or docs.
