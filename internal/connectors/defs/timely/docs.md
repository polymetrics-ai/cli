# Overview

Timely is a wave2 fan-out declarative-HTTP migration. It reads Timely users, projects, clients,
and calendar/time events through the Timely REST API v1.1
(`GET https://api.timelyapp.com/1.1/<account_id>/<resource>`). This bundle is capability-parity
migrated from `internal/connectors/timely` (the hand-written connector it migrates); the legacy
package stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Timely OAuth access token via the `bearer_token` secret; it is sent as a Bearer token
(`Authorization: Bearer <bearer_token>`), matching legacy's `connsdk.Bearer(token)`
(`timely.go:123`) exactly, and is never logged. `account_id` is required and prefixes every
stream's path (`<account_id>/<resource>`), matching legacy's `accountPath` helper
(`timely.go:166-172`). `base_url` defaults to `https://api.timelyapp.com/1.1`, matching legacy's
`defaultBaseURL` fallback.

## Streams notes

All 4 streams (`users`, `projects`, `clients`, `events`) read the full JSON array response
(`records.path: "."`) with no pagination, matching legacy's single unpaginated `Do` request per
stream (`timely.go:91`, `RecordsAt(resp.Body, ".")`). Primary key `["id"]` on every stream.

`events` additionally sends a `since` query param sourced from the `start_date` config value via
the opt-in optional-query dialect (`{"template": "{{ config.start_date }}", "omit_when_absent":
true}`) — present only when `start_date` is configured, omitted entirely otherwise, matching
legacy's own stream-specific gating exactly (`timely.go:86-90`: `since` is only ever set for the
`events` stream, and only when `start_date` is non-empty). This is a plain optional config
passthrough, not a true `incremental` block: legacy tracks no persisted cursor and applies no
client-side re-filtering — it is a one-shot "start from this timestamp" hint the Timely API
itself interprets, so this bundle deliberately does not declare an `incremental` block (which
would imply cursor-based state tracking legacy never had).

## Write actions & risks

None. Timely is read-only (`capabilities.write: false`, no `writes.json`), matching legacy's
`Write` returning `connectors.ErrUnsupportedOperation`.

## Known limits

- **`events`'s `since` is a one-shot config hint, not a stateful incremental cursor.** Legacy
  never persists or advances a cursor value for this stream — every sync either passes the
  configured `start_date` verbatim or omits `since` entirely. This bundle reproduces that exact
  behavior; it is not eligible for `incremental_append[_deduped]` sync modes since no
  `incremental` block is declared (conventions.md §2, "Sync-mode derivation — never declared").
