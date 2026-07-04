# Linkrunner

## Overview

Reads Linkrunner mobile attribution campaigns and attributed users from the Linkrunner Data API
(`https://api.linkrunner.io/api/v1`). Read-only: Linkrunner has no approved reverse-ETL write
surface, matching the legacy `internal/connectors/linkrunner` package. Pass B also reviewed
Linkrunner's SDK-less API reference; those documented endpoints are client-device telemetry calls,
not server-side sync resources.

## Auth setup

An API key (`secrets.linkrunner-key`) is sent on every request as the `linkrunner-key` header (a
custom API-key-header scheme, not Bearer/Basic), matching legacy's `connsdk.APIKeyHeader` usage
exactly.

## Streams notes

`campaigns` reads `/campaigns` and emits records from the `data.campaigns` array; the optional
`config.filter`/`config.channel` values are forwarded as query params only when set
(`omit_when_absent`), matching legacy's conditional filter passthrough. `attributed_users` reads
`/attributed-users` (`data.users` array) and **requires** `config.display_id` to scope the read —
declared as a plain (always-resolved) query template, so a missing `display_id` hard-errors exactly
as legacy's own `errors.New("linkrunner attributed_users stream requires config display_id")` does;
`config.start_timestamp`/`config.end_timestamp`/`config.timezone` are optional passthrough filters
(`omit_when_absent`), matching legacy.

Both streams use page/limit pagination (`page` starting at 1, `limit` page size); a page shorter
than the declared size stops the read, matching legacy's `harvest` loop exactly. Legacy performs no
automated incremental filtering for either stream (a full sync always re-reads every campaign/user
from page 1); each schema still declares its natural cursor field (`update_at`/`attributed_at`) for
downstream state tracking, but no `request_param` is sent — this mirrors legacy's own behavior, not
a narrowing.

## Write actions & risks

None. Linkrunner is read-only in pm (`capabilities.write: false`), matching legacy.
The documented SDK-less POST endpoints (`/api/client/init`, `/api/client/attribution-data`,
`/api/client/trigger`, `/api/client/set-user-data`, `/api/client/capture-event`,
`/api/client/capture-payment`, `/api/client/remove-captured-payment`,
`/api/client/update-push-token`, `/api/client/integrations`, and
`/api/client/handle-deeplink`) all require a project token in the JSON body plus device/install
context. The current declarative write dialect cannot safely inject secrets into JSON bodies, and
these calls represent client telemetry, attribution, or payment/revenue side effects rather than
ordinary reverse-ETL object mutations.

## Known limits

- **`page_size`/`max_pages` config overrides are not modeled.** `streams.json`'s
  `base.pagination.page_size` is set to legacy's real production default, `100` (legacy:
  `linkrunnerDefaultPageSize = 100`) — this is the actual value a live deployment's paginator
  sends; it is not a fixture convenience. `config.page_size`/`config.max_pages` are declared in
  `spec.json` for parity with legacy's config surface but, like stripe's documented
  `page_size`/`limit_param` deviation (conventions.md §5 item 3), are not wired into the engine's
  fixed pagination block. The mandatory 2-page conformance fixture
  (`fixtures/streams/campaigns/{page_1,page_2}.json`) is sized to match: page 1 returns a full
  100-record page (so the paginator continues to page 2), page 2 returns the 1-record remainder —
  matching aviationstack's and awin-advertiser's identical repaired precedent
  (`docs/migration/conventions.md`, wave2 sweep class C3). The `attributed_users` single-page
  fixture requests `limit=100` to match.
- Linkrunner's public docs (`docs.linkrunner.io/sdk-less/api-reference`) describe the client SDK
  surface, not this connector's Data API endpoints (`/campaigns`, `/attributed-users`); the legacy
  Go connector and its test suite are the authoritative source for this bundle's Data API
  request/response shapes, per migration convention (legacy is ground truth over docs when the
  public docs omit the legacy API).
- Pass B did not add SDK-less writes because they require body-carried secrets and device context
  that would either leak credentials through write records or create client-telemetry side effects
  from a server-side ETL connector. Each documented SDK-less endpoint is accounted for in
  `api_surface.json`.
