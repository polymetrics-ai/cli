# Linkrunner

## Overview

Reads Linkrunner mobile attribution campaigns and attributed users from the Linkrunner Data API
(`https://api.linkrunner.io/api/v1`). Read-only: Linkrunner has no approved reverse-ETL write
surface, matching the legacy `internal/connectors/linkrunner` package.

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

## Known limits

- **Fixture pagination page size**: `streams.json`'s `base.pagination.page_size` is `2` (rather
  than legacy's real default of `100`) so a committed 2-page fixture can prove pagination
  termination without embedding 100 synthetic records — an implementation detail of the fixture
  chunking, never a change to emitted record data. `config.page_size`/`config.max_pages` are
  declared in `spec.json` for parity with legacy's config surface but, like stripe's documented
  `page_size`/`limit_param` deviation (conventions.md §5 item 3), are not wired into the engine's
  fixed pagination block.
- Linkrunner's public docs (`docs.linkrunner.io/sdk-less/api-reference`) describe the client SDK
  surface (`init`/`attribution-data`/`trigger`), not this connector's Data API endpoints
  (`/campaigns`, `/attributed-users`); the legacy Go connector and its test suite are the
  authoritative source for this bundle's request/response shapes, per migration convention (legacy
  is ground truth over docs).
- Pass B (any additional Linkrunner Data API surface beyond the 2 legacy streams) is out of scope;
  see `api_surface.json`.
