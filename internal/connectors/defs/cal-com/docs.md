# Overview

Cal.com is a wave2 fan-out declarative-HTTP migration, expanded to full documented API-surface
coverage in Pass B. It reads Cal.com bookings, event types, availability schedules, webhooks, and
the authenticated user's profile, and creates/updates/cancels/deletes them, through the Cal.com v2
REST API (`https://api.cal.com/v2/...`). This bundle targets capability parity with
`internal/connectors/cal-com` (the hand-written connector it migrates, Go package `calcom`, which was
read-only); the legacy package stays registered and unchanged until wave6's registry flip. Pass B
research verified this bundle against Cal.com's real published OpenAPI 3.0 spec (121 method+path
endpoints across 82 paths, `docs/api-reference/v2/openapi.json` in `github.com/calcom/cal.com`); see
`api_surface.json` for the full endpoint-by-endpoint disposition.

## Auth setup

Provide a Cal.com API key via the `api_key` secret; it is sent as a Bearer token (`Authorization:
Bearer <api_key>`), matching legacy's `connsdk.Bearer(token)` (`cal_com.go:282`), and is never
logged. Every request also sends a `cal-api-version` header, resolved from the `api_version` config
value (default now `2024-06-14`, changed from wave2's `2024-08-13` — see Known limits below for why);
`spec.json`'s `"default"` materializes this value into `RuntimeConfig.Config` before header
resolution runs, so the header is always present. `base_url` defaults to `https://api.cal.com` and
may be overridden for tests/proxies.

## Streams notes

- `bookings` (`GET /v2/bookings`) and `schedules` (`GET /v2/schedules`) both use Cal.com's offset
  (`skip`/`take`) pagination (`pagination.type: offset_limit`, `limit_param: take`, `offset_param:
  skip`) — records live at `data`, and a page shorter than `take` stops pagination; `page_size` is
  `100` (see Known limits for why it is not runtime-configurable).
- `event_types` (`GET /v2/event-types`) is now migrated (see Known limits — the prior wave's
  ENGINE_GAP was disproven by Pass B research): records live at `data`, a flat array, identical
  envelope shape to every other stream here.
- `webhooks` (`GET /v2/webhooks`) is new in Pass B: the account's registered webhook subscriptions,
  same `data`+offset/limit pagination shape.
- `my_profile` (`GET /v2/me`) is not paginated (`pagination.type: none`); its `data` envelope is a
  single object rather than an array, which `records.path: "data"` handles identically to an array
  of one.

None of the 5 streams expose an incremental cursor field in Cal.com's v2 API — this bundle declares
no `incremental` block for any of them, so reads are full refresh, matching legacy's original 3
streams and extending the same (accurate) pattern to the 2 new ones.

## Write actions & risks

This bundle adds write support beyond legacy (which was read-only); `capabilities.write` is now
`true` and `writes.json` declares 13 actions, grouped by resource:

- **Bookings**: `create_booking` (`POST /v2/bookings`, nested `attendee` object, JSON body),
  `cancel_booking`/`confirm_booking`/`decline_booking`/`reschedule_booking` (`POST
  /v2/bookings/{{ record.uid }}/{cancel,confirm,decline,reschedule}`, `path_fields: ["uid"]`).
  `confirm_booking` sends no request body (`body_type: none`) — Cal.com's confirm endpoint takes
  none. **Risk: these mutate real scheduled meetings and trigger attendee-facing notifications
  (cancellation/confirmation/reschedule emails); approval required for all 5.**
- **Event types**: `create_event_type`/`update_event_type`/`delete_event_type`
  (`POST`/`PATCH`/`DELETE /v2/event-types[/{{ record.id }}]`). **Risk: changes what is publicly
  bookable on the account; `delete_event_type` breaks any existing public booking link for that
  event type; approval required.**
- **Schedules**: `create_schedule`/`update_schedule`/`delete_schedule`
  (`POST`/`PATCH`/`DELETE /v2/schedules[/{{ record.id }}]`). **Risk: directly changes real
  availability windows that determine when the account can be booked; approval required.**
- **Webhooks**: `create_webhook` (`POST /v2/webhooks`, requires `subscriberUrl`/`triggers`/`active`
  per `CreateWebhookInputDto`)/`delete_webhook` (`DELETE /v2/webhooks/{{ record.id }}`). An in-place
  `update_webhook` (`PATCH /v2/webhooks/{webhookId}`) is documented but not implemented in this wave
  (see `api_surface.json`) — create+delete cover the common lifecycle. **Risk: a new webhook
  subscription receives live booking-event payloads at an operator-supplied URL; approval required.**

## Known limits

- **`event_types`'s prior ENGINE_GAP was a mis-diagnosis, now RESOLVED.** wave2's `docs.md` reported
  a blocked, two-level-nested envelope (`data.eventTypeGroups[].eventTypes[]`) mirroring legacy's own
  `emitNested` flattening logic. Pass B research against Cal.com's real, current OpenAPI spec proved
  that nested shape belongs to an OLDER, non-default API version — the documented CURRENT version
  (`cal-api-version: 2024-06-14`, the version the spec's own parameter description says is
  "required") returns a flat `{status, data: EventTypeOutput_2024_06_14[]}` envelope, structurally
  identical to `bookings`/`schedules`/`webhooks`. Legacy's own hardcoded `defaultAPIVersion =
  "2024-08-13"` (applied globally to every stream, `cal_com.go:31`) was itself never sending the
  correct version header for the event-types endpoint, which is very likely why legacy's authors
  built the `emitNested` flattening in the first place — against a response shape from an
  unintentionally-old, unversioned fallback handler. No engine gap exists once the correct header
  value is sent; `event_types` is migrated as an ordinary flat-array stream in this wave.
- **A single shared `cal-api-version` header, not a per-stream one — genuine engine-dialect
  constraint, resolved by picking a universally-compatible value, not by an engine change.** The
  engine's `HTTPBase.Headers` (`streams.json`'s `base.headers`) applies identically to every stream
  in a bundle; there is no per-`StreamSpec` header-override field. Cal.com's OpenAPI spec documents a
  DIFFERENT "required" `cal-api-version` value per resource family (bookings: `2024-08-13`,
  schedules: `2024-06-11`, event-types: `2024-06-14`; webhooks/me document no version requirement at
  all) — naively honoring each resource's own documented value would need a per-stream header the
  dialect cannot express (a real `ENGINE_GAP` candidate). Live verification against
  `https://api.cal.com` on 2026-07-04 (unauthenticated probe requests, comparing 404
  "Cannot GET"/wrong-version-routing responses against 401 Unauthorized/correctly-routed responses)
  found that **`cal-api-version: 2024-06-14` alone correctly routes ALL FIVE migrated resources**
  (bookings, schedules, event-types, webhooks, my_profile) to their current handler — no other single
  value achieves this (`2024-08-13`, wave2's prior default, 404s on both `schedules` and
  `event-types`). `spec.json`'s `api_version` default was changed from `2024-08-13` to `2024-06-14`
  accordingly. This is a genuine, provable fix, not a guess: no engine change was needed once the
  right shared value was identified.
- **`page_size`/`max_pages` are not runtime-configurable.** The engine's `offset_limit` paginator's
  `PageSize` is a static bundle-authored int (not templated), and there is no `MaxPages`-equivalent
  config-driven knob either; `max_pages` is unbounded. `page_size` is fixed at `100` to match
  legacy's own default; conformance fixtures for `bookings`/`schedules`/`event_types`/`webhooks` are
  each a single short page, so `pagination_terminates` observes exactly one request per stream.
- **Legacy's fixture-mode-only fields are not modeled.** Legacy's `readFixture` path (only reached
  when `config.mode == "fixture"`) stamps extra fields (`connector`, `fixture`) onto every
  fixture-mode record (`cal_com.go:226-262`); none are part of the live record shape. This bundle's
  schemas and fixtures target the live path only.
- **Nested/complex sub-objects on `event_types`/`bookings` (locations, bookingFields, hosts,
  attendees) are typed as loosely-shaped `array`/`object` schema fields, not fully modeled with
  nested property schemas.** Draft-07 schema projection preserves them verbatim (schema mode does
  not require every nested field to be individually declared to pass through as a JSON value once
  the top-level property itself is declared with a permissive `type`); a future wave could add
  dedicated nested schemas for these if a downstream consumer needs typed access to their internals.
- The full platform/reseller surface (OAuth clients + their managed users/webhooks), connected
  calendar-provider account wiring (Google/Outlook/ICS), conferencing-app account wiring
  (Zoom/Google Meet), Stripe Connect, phone/email verification, and ephemeral slot reservations are
  out of scope — see `api_surface.json`'s per-endpoint `excluded` categories and reasons (mostly
  `requires_elevated_scope`: they need a calendar/conferencing/Stripe/platform OAuth grant this
  bundle's single `api_key` credential does not hold).
