# Overview

Eventee is a wave2 fan-out migration. It reads Eventee event agenda data (lectures, speakers, days,
halls, tracks, workshops, partners) through the Eventee public REST API
(`https://api.eventee.co/public/v1`). This bundle migrates `internal/connectors/eventee`; the
legacy package stays registered and unchanged until wave6's registry flip. Eventee's public API is
read-only for our purposes with no pagination and no incremental cursor, matching legacy's
straightforward full-refresh source.

## Auth setup

Provide an Eventee public API token via the `api_token` secret; it is used only for Bearer auth
(`Authorization: Bearer <api_token>`) and is never logged.

## Streams notes

Six of the seven streams (`lectures`, `speakers`, `days`, `halls`, `tracks`, `workshops`) share the
single `GET /content` endpoint and select their own nested array by name from the response body
(records at `lectures`/`speakers`/`days`/`halls`/`tracks`/`workshops` respectively), matching
legacy's `streamEndpoint` routing table exactly — `workshops` reuses the same field shape as
`lectures` (legacy's `eventeeLectureRecord` mapper is shared by both), so this bundle's
`schemas/workshops.json` mirrors `schemas/lectures.json` field-for-field. `partners` is a
dedicated endpoint (`GET /partners`) whose body is a top-level JSON array rather than a nested
object field, expressed here with `records.path: ""` (root-array selection). No stream declares
pagination (`base` has no `pagination` block) or an incremental cursor field, matching legacy's "a
single GET returns the full collection" full-refresh-only behavior.

## Write actions & risks

None. Eventee is read-only; `capabilities.write` is `false` and this bundle ships no
`writes.json`.

## Known limits

- Only the 7 legacy-parity read streams are implemented; legacy's own routing table documents a
  `participants` stream by convention (dedicated endpoint, top-level array, keyed by email in the
  upstream schema) that is not present in `eventeeStreamEndpoints`/`eventeeStreams` either — it was
  never implemented in legacy, so there is nothing to migrate for it here. The full Eventee surface
  (participants, feedback, polls, etc.) is out of scope for this wave — see `api_surface.json`'s
  `excluded: {category: out_of_scope, reason: "Pass B capability expansion"}` entries.
- No pagination or incremental sync exists on either side (legacy or this bundle) — every read is
  a full snapshot of the current event agenda.
