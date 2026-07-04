# Overview

Amplitude is a declarative-HTTP connector for the Amplitude Analytics REST APIs (Behavioral
Cohorts, Chart Annotations, Taxonomy, and the event-list resource). It reads behavioral cohorts,
cohort-download usage, chart annotations and their categories, the event-list catalog, and the
governed taxonomy (categories, event types, event/user/group properties), and it manages chart
annotations, annotation categories, and taxonomy categories/event types through Pass B's
full-surface expansion. This bundle originally migrated `internal/connectors/amplitude` (the
hand-written connector); the legacy package stays registered and unchanged until wave6's registry
flip.

## Auth setup

Provide an Amplitude project API key via the `api_key` secret and its paired secret key via the
`secret_key` secret. Both flow only into HTTP Basic auth (`api_key` as username, `secret_key` as
password) and are never logged. All streams and writes share the same Basic-auth credential pair —
Amplitude's Behavioral Cohorts, Chart Annotations, and Taxonomy APIs are all authenticated
identically.

## Streams notes

10 streams, all full-refresh with no pagination and no incremental cursor — every Amplitude list
endpoint this connector reads returns its full collection in one response:

- `cohorts` — `GET /api/3/cohorts`, records at `cohorts`, primary key `id`.
- `cohorts_usage` — `GET /api/3/cohorts/usage`, a single-object response with no records array;
  `records.path: ""` treats the whole body as one record, and `computed_fields` flattens the
  nested `rest_download.{limit,usage,resets_at}` fields onto the emitted record's top level
  (typed bare-reference extraction keeps `limit`/`usage` as native integers, per
  `docs/migration/conventions.md` §3). Primary key `resets_at` (a fresh timestamp each reset
  window; this is a usage-quota snapshot, not a stable-id resource, so no other candidate key
  exists).
- `annotations` — `GET /api/3/annotations`, records at `data`, primary key `id`.
- `annotation_categories` — `GET /api/3/annotation-categories`, records at `data`, primary key
  `id`.
- `events_list` — `GET /api/2/events/list`, records at `data`, primary key `value` (Amplitude's own
  event-name identifier).
- `taxonomy_categories` — `GET /api/2/taxonomy/category`, records at `data`, primary key `id`.
- `taxonomy_events` — `GET /api/2/taxonomy/event`, records at `data`, primary key `event_type`. An
  optional `taxonomy_show_deleted` config value (a `'true'`/`'false'` string) is forwarded as the
  `showDeleted` query parameter via the opt-in optional-query dialect (`omit_when_absent: true`);
  left unset, Amplitude applies its own default (deleted event types excluded).
- `taxonomy_event_properties` — `GET /api/2/taxonomy/event-property`, records at `data`, composite
  primary key `[event_property, event_type]` (an event property name may repeat across different
  event types).
- `taxonomy_user_properties` — `GET /api/2/taxonomy/user-property`, records at `data`, primary key
  `user_property`. Same optional `taxonomy_show_deleted` -> `showDeleted` wiring as `taxonomy_events`.
- `taxonomy_group_properties` — `GET /api/2/taxonomy/group-property`, records at `data`, composite
  primary key `[group_type, group_property]`.

Field names are copied verbatim from the raw API response (Amplitude's own camelCase/snake_case
mix, e.g. `lastComputed`, `non_active`, `is_hidden_from_dropdowns`, preserved as-is) via schema
projection; no `computed_fields` renames are needed for any list-shaped stream.

## Write actions & risks

`capabilities.write: true`. 12 actions, all requiring reverse-ETL plan approval before executing
(`metadata.json`'s `risk.approval`):

- `create_annotation` / `update_annotation` / `delete_annotation` — `POST`/`PUT`/`DELETE
  /api/3/annotations[/{id}]`. Creates, mutates, or permanently deletes a chart annotation visible
  to every user of the Amplitude project.
- `create_annotation_category` / `update_annotation_category` / `delete_annotation_category` —
  `POST`/`PUT`/`DELETE /api/3/annotation-categories[/{id}]`. Manages the shared category taxonomy
  annotations are organized under.
- `create_taxonomy_category` / `update_taxonomy_category` / `delete_taxonomy_category` —
  `POST`/`PUT`/`DELETE /api/2/taxonomy/category[/{id}]`. Manages the Amplitude project's governed
  event-category taxonomy.
- `create_taxonomy_event` / `update_taxonomy_event` / `delete_taxonomy_event` — `POST`/`PUT`/`DELETE
  /api/2/taxonomy/event[/{event_type}]`. Registers, edits, or soft-deletes a governed event type
  in the taxonomy. `delete_taxonomy_event` is a soft delete on Amplitude's side (recoverable via
  `POST /api/2/taxonomy/event/{event_type}/restore`); the restore endpoint itself is not modeled as
  a separate write action (see `api_surface.json`) since it targets an already-deleted resource
  outside the create/update/delete model.

None of these writes ever touch raw behavioral event data or user-level analytics — the write
surface is scoped entirely to Amplitude's own metadata-management resources (annotations,
categories, taxonomy definitions), never the events/users/cohort membership an Amplitude project
actually tracks.

## Known limits

- EU-residency projects: legacy derived `https://analytics.eu.amplitude.com` automatically from a
  `data_region` config value containing "eu". The engine's `spec.json` `"default"` mechanism only
  materializes a fixed literal default, not one derived from another config key's value (see
  `docs/migration/conventions.md` §3's `default` materialization note — this is the same
  base-URL-derivation gap documented for sentry/chargebee). This bundle narrows the config surface:
  `data_region` is no longer a config option; EU-residency users must set `base_url` to
  `https://analytics.eu.amplitude.com` explicitly. Documented scope narrowing, not a silent
  behavior change — every request legacy would route to the EU host still reaches it, just via an
  explicit `base_url` instead of an inferred one.
- Cohort membership download/upload (`/api/5/cohorts/request/*`, `/api/3/cohorts/upload`,
  `/api/3/cohorts/membership`) is out of scope: the download side is a 2-step async job-poll
  protocol that can redirect to a presigned S3 URL for large cohorts, and the upload/membership
  side is a bulk id-list mutation with no single addressable record — neither shape fits this
  connector's per-record read/write model. See `api_surface.json` for the full per-endpoint
  rationale.
- The raw event Export API (`GET /api/2/export`) is out of scope: it returns a zipped archive of
  newline-delimited JSON files, not a single JSON body, and this connector deliberately ships zero
  Tier-2 hooks (decompression/archive-fan-out is a hook-only concern per
  `docs/migration/conventions.md` §1).
- Event ingestion (`POST /2/httpapi`) is out of scope: it is a write-only telemetry sink for
  creating new raw analytics events, not a reverse-ETL mutation of an existing Amplitude-managed
  resource.
- Taxonomy event-property/user-property/group-property mutations are read-only in this wave (their
  list endpoints are covered streams; create/update/delete are not yet implemented) — a deliberate
  write-depth cut, not a capability gap; see `api_surface.json` for each excluded mutation's
  rationale.
