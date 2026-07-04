# Overview

Incident.io is a read-only declarative-HTTP migration (wave2 fan-out) of
`internal/connectors/incident-io` (legacy Go package `incidentio`). It reads incident.io
incidents, severities, incident roles, users, and follow-ups through the incident.io REST API
(`https://api.incident.io`). This bundle targets capability parity with the legacy connector; the
legacy package stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide an incident.io API key via the `api_key` secret; it is used only for Bearer auth
(`Authorization: Bearer <api_key>`) and is never logged. `base_url` defaults to
`https://api.incident.io` and may be overridden for tests or proxies.

## Streams notes

- `incidents` (`GET /v2/incidents`) and `users` (`GET /v2/users`) and `follow_ups`
  (`GET /v2/follow_ups`) are paginated: `pagination.type: cursor` with `token_path:
  pagination_meta.after` and `cursor_param: after`, matching legacy's `harvest` loop exactly
  (next page requested with `after=<pagination_meta.after>`; an absent/empty `after` stops the
  read). Every request to these three streams sends `page_size` (default `100`, configurable via
  the `page_size` spec property, matching legacy's `pageSize`/`maxPageSize` bounds of 1-250) as a
  static query param — legacy only ever sends `page_size` for paginated endpoints, never for the
  two single-page v1/v2 endpoints below, so this bundle mirrors that exactly by only declaring the
  `page_size` query entry on the three paginated streams.
- `severities` (`GET /v1/severities`) and `incident_roles` (`GET /v2/incident_roles`) are
  single-page (`pagination.type: none`), matching legacy's `endpoint.paginated == false` v1/v2
  resources that never return a `pagination_meta.after`.
- `incidents` flattens the raw API's nested `severity`/`incident_status` objects into
  `severity_id`/`severity_name`/`status_id`/`status_name`/`status_category` via `computed_fields`
  dotted-path extraction (`{{ record.severity.id }}`, etc.), matching legacy's `incidentRecord`
  mapper. `users` flattens `base_role` into `base_role_id`/`base_role_name`
  (legacy's `userRecord`). `follow_ups` flattens `assignee` into `assignee_id`/`assignee_name`
  (legacy's `followUpRecord`). Every computed field is silently skipped for a record if its source
  path is absent (e.g. a follow-up with no assignee) — matching legacy's nil-safe
  `nestedObject` helper, which returns an empty map (all-nil fields) rather than panicking or
  erroring.
- Primary key is `id` on every stream. `incidents`/`severities`/`incident_roles`/`follow_ups`
  declare `x-cursor-field: updated_at` for manifest-surface parity with legacy's `CursorFields`;
  legacy never actually sends an incremental filter query param for any incident.io stream (no
  `incremental` block is declared here either), so this is descriptive only, matching legacy's own
  informational-cursor behavior. `users` has no cursor field, matching legacy's `CursorFields: nil`.

## Write actions & risks

None. incident.io is read-only in this bundle (`capabilities.write: false`); legacy also rejects
every write with `connectors.ErrUnsupportedOperation`. No `writes.json` is shipped.

## Known limits

- Full incident.io API surface (actions, incident updates, alerts, schedules, workflows, custom
  fields) is out of scope for this wave; see `api_surface.json`'s `excluded: {category:
  out_of_scope, reason: "not implemented in this bundle"}` entries. Only the 5 legacy-parity streams
  are implemented.
- `page_size`'s legacy validation range (1-250) is not enforced by this bundle's `spec.json` (the
  draft-07 subset has no numeric bounds check wired for a string-typed config field here); an
  out-of-range value is passed through to the API as-is rather than rejected client-side the way
  legacy's `pageSize()` helper does. This is a narrowing of legacy's input validation, not a change
  to any successfully-processed request's shape.
- `max_pages` is not declared in `spec.json`. `PaginationSpec.MaxPages` (the engine's only
  request-count cap) is a static JSON integer with no `{{ }}` template support, so a runtime
  `config.max_pages` value can never actually bound anything — declaring the config property
  anyway would be dead config a bundle author cannot wire to any real behavior (F6,
  `docs/migration/conventions.md`). Legacy's own `max_pages` config (0/`all`/`unlimited` = no cap)
  is reproduced here as the engine's default unbounded behavior (`MaxPages` omitted); an operator
  needing a hard cap would require a per-connector fixed value in `streams.json`, not a runtime
  override.
