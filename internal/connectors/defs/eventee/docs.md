# Overview

Eventee reads and writes the documented Public Eventee API
(`https://api.eventee.co/public/v1`). This Pass B bundle covers the official Apiary documentation
surface: event content, reviews, groups, participants, partners, pauses, registrations, speakers,
tracks, and the documented mutation endpoints. The legacy package stays registered and unchanged
until wave6's registry flip.

## Auth setup

Provide an Eventee public API token via the `api_token` secret; it is used only for Bearer auth
(`Authorization: Bearer <api_token>`) and is never logged.

## Streams notes

Seven content streams (`lectures`, `speakers`, `days`, `halls`, `tracks`, `workshops`, `pauses`)
share `GET /content` and select the documented nested array by name. `partners`, `reviews`,
`groups`, `participants`, and `registrations` read their dedicated endpoints. No stream declares
pagination or an incremental cursor; the public docs show single-response full-refresh endpoints.

Existing legacy-parity streams keep schema projection for the fields legacy emitted. Recorded
fixtures use the documented raw response shapes, so some raw fields are intentionally dropped by the
schemas.

## Write actions & risks

Eventee write actions cover the documented mutation endpoints: create/update/delete halls,
lectures, partners, pauses, speakers, and tracks; invite attendees and registrations; update
attendee check-in; remove attendees and registrants; and clear test content. Deletes and removals
are marked `confirm: destructive`; reverse ETL still follows plan, preview, approval, execute.

## Known limits

- No pagination or incremental sync is documented; every read is a full snapshot.
- The official `GET /registrations` example is a single JSON object despite the endpoint name. The
  stream reads the response root, which supports either a single object or a top-level array.
- Eventee documents attendee and registration removal emails as URL-encoded values in the JSON
  body; callers should provide the encoded email string.
