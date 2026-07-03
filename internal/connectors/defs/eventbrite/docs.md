# Overview

Eventbrite is a full legacy-parity declarative-HTTP migration. It reads Eventbrite organizations,
events, attendees, orders, and ticket classes through the Eventbrite v3 REST API (`GET
https://www.eventbriteapi.com/v3/...`). This bundle is a capability-parity port of the
hand-written connector at `internal/connectors/eventbrite` (`eventbrite.go`/`streams.go`), which
stays registered and unchanged until wave6's registry flip. Eventbrite is read-only in legacy (no
reverse-ETL writes), so `capabilities.write` is `false` and this bundle ships no `writes.json`.

## Auth setup

Provide an Eventbrite private OAuth token via the `private_token` secret; it is used only for
Bearer auth (`Authorization: Bearer <private_token>`, `streams.json`'s `base.auth`) and is never
logged, matching legacy's `connsdk.Bearer(secret)` (`eventbrite.go:271`). `base_url` defaults to
`https://www.eventbriteapi.com/v3` and may be overridden for tests/proxies (legacy's own
`eventbriteBaseURL` validates scheme+host the same way).

## Streams notes

`organizations` reads from `/users/me/organizations/` and needs no scoping id (legacy's
`scopeUser`). `events` is scoped to the configured `organization_id`
(`/organizations/{{ config.organization_id }}/events/`, legacy's `scopeOrg`); `attendees`,
`orders`, and `ticket_classes` are scoped to the configured `event_id`
(`/events/{{ config.event_id }}/...`, legacy's `scopeEvent`) — an absent required id hard-errors on
both sides (legacy: `"eventbrite stream requires config organization_id/event_id"`; engine: an
unresolved `config.*` path-template key), matching conventions.md §5's precedent for
config-validation parity.

Every request sends `expand=venue,ticket_classes` (a static per-stream `query` entry), matching
legacy's `harvest`'s unconditional `base.Set("expand", "venue,ticket_classes")`
(`eventbrite.go:151-152`) — harmless on streams that ignore the expansion.

Pagination follows Eventbrite's `pagination.has_more_items`/`pagination.continuation`
continuation-token convention (`pagination.type: cursor`, `cursor_param: continuation`,
`token_path: pagination.continuation`, `stop_path: pagination.has_more_items`), matching legacy's
`harvest` loop exactly: the next page is requested with `continuation=<token>`, and pagination
stops when `has_more_items` is falsy (any value other than the literal `true`) — the engine's
`stop_path` semantics read the SAME `pagination.has_more_items` boolean legacy's own
`hasMore != "true"` check reads. The engine's `tokenPathCursor` paginator also loop-guards against
the same continuation token repeating twice in a row, matching legacy's own `next == continuation`
guard (`eventbrite.go:193`).

`events`, `attendees`, and `orders` carry legacy's `changed_since` incremental filter
(`incremental.cursor_field: changed`, `request_param: changed_since`, `start_config_key:
start_date`), wired through the opt-in optional-query dialect (`{{ incremental.lower_bound }}`
with `omit_when_absent: true`) so `changed_since` is sent only once a lower bound resolves (a
state cursor from a prior sync, or the `start_date` config on a first run) — matching legacy's
`incrementalLowerBound` helper (`eventbrite.go:299-307`), which returns the persisted cursor or
else `start_date`, empty meaning a full sync with no filter.

`ticket_classes` does **not** declare an `incremental` block, even though legacy's own catalog
declares `CursorFields: []string{"changed"}` for it (`streams.go:81`) and legacy's `harvest`
unconditionally applies `changed_since` to every stream's request regardless of endpoint,
including ticket_classes. This is a genuine parity gap in the ORIGINAL legacy catalog, not a
migration shortcut: `ticketClassFields()`/`ticketClassRecord()` (`streams.go:149-163,234-248`)
never include a `changed` field at all — no ticket-class response payload carries a "changed"
timestamp, so legacy's own declared cursor field is dead on arrival (state advances against a
field that is never populated). The engine's `x-cursor-field`/`incremental.cursor_field`
mechanism, unlike legacy's Go struct literals, is validated (`cursor_field_missing`) against the
stream's own schema properties, so declaring `changed` here would be a hard validate failure, not
a silent no-op like it was in legacy. To preserve the WIRE-REQUEST behavior (Eventbrite still
receives a `changed_since` query param when `start_date` is configured) without inventing a
schema field that doesn't exist, `ticket_classes`'s `changed_since` is wired directly off
`{{ config.start_date }}` via the optional-query dialect (omitted when unset) rather than through
the `incremental` cursor-state machinery — this reproduces legacy's actual request shape exactly
while being honest about the stream having no real incremental cursor. See Known limits.

`events.name`/`description` are Eventbrite's `{text, html}` multipart-text objects, flattened to
their `.text` value via `computed_fields` (matching legacy's `flattenText`); `events.start`/`end`
are `{timezone, local, utc}` datetime-tz objects, flattened to `.utc` (matching legacy's
`flattenUTC`). `attendees.name`/`email` come from the nested `profile.{name,email}` object
(matching legacy's `flatten(profile["name"])`/`flatten(profile["email"])`).
`ticket_classes.cost`/`fee` are `{display, value, currency}` currency objects, flattened to
`.display` (matching legacy's `flattenCurrency`). `ticket_classes.name`/`description` are the same
`{text, html}` shape as events. `organizations.name` and `orders.name`/`email` are plain wire
strings on the real Eventbrite API (legacy's generic `flatten()` helper tolerates either shape but
these three fields are never actually nested), so no `computed_fields` rename is needed for them —
plain schema projection copies them by exact key match.

## Write actions & risks

None. Eventbrite is a read-only source in legacy (`eventbrite.go`'s package doc: "Eventbrite is a
read-only data source (no reverse-ETL writes)"); `capabilities.write` is `false` and no
`writes.json` is shipped, matching legacy's `Write` returning
`connectors.ErrUnsupportedOperation`.

## Known limits

- **`ticket_classes` has no `incremental` block / no `x-cursor-field`**, even though legacy's
  catalog declares `CursorFields: []string{"changed"}` for it. As detailed in Streams notes above,
  this is because `ticketClassFields()`/`ticketClassRecord()` never emit a `changed` field at
  all — legacy's own cursor-field declaration was already dead code (no field to advance state
  against). The wire-level `changed_since` request behavior is still reproduced (via
  `{{ config.start_date }}`, omitted when unset), so no legacy-accepted request shape is lost; only
  the (already-nonfunctional) state-tracking declaration is not carried forward. ACCEPTABLE parity
  deviation (conventions.md §5): documented here rather than declaring a schema field that would
  fail `cursor_field_missing` validation.
- **`max_pages` is not modeled as a bundle-level config knob.** Legacy exposes `max_pages` as a
  config override (`eventbriteMaxPages`, `eventbrite.go:337-350`, accepting an integer, `all`, or
  `unlimited`). The engine's `cursor`/`token_path` paginator has no config-driven `max_pages`
  override wired to a spec property — pagination is bounded only by the `has_more_items` stop
  signal and the token-repeat loop guard, matching Eventbrite's own real termination behavior (the
  same "unbounded by default" outcome as legacy's `max_pages` unset/`0`/`all`/`unlimited` case).
- **Legacy's fixture-mode-only synthetic fields are not modeled.** Legacy's `readFixture` path
  (only reached when `config.mode == "fixture"`, a credential-free conformance-harness affordance)
  stamps a broad synthetic superset of fields shared across all 5 streams' mappers, which is not
  the live wire shape any single migrated stream actually returns. This bundle's schemas and
  fixtures target the LIVE record shape only; the engine's own fixture-replay conformance harness
  (`internal/connectors/conformance`) supersedes the need for an in-connector fixture-mode branch.
