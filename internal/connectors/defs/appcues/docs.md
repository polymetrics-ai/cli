# Overview

Appcues is a declarative-HTTP connector for the Appcues REST API v2. It reads every listable
in-app-guidance experience type (flows, Flows 2.0, pins, mobile experiences, launchpads, banners,
checklists, embeds, NPS 2.0), audience data (segments, tags), and operational resources (offline
jobs, SDK authentication keys) for a configured account, and it manages publish state, segments,
SDK keys, and individual end-user/group profiles through Pass B's full-surface expansion. This
bundle originally migrated `internal/connectors/appcues` (the hand-written connector); the legacy
package stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide the Appcues account ID via the `account_id` config value (every resource path is scoped to
`accounts/{account_id}/...`), the Appcues API key via the `username` config value, and the Appcues
API secret via the `password` secret. Auth is HTTP Basic (`username` as the Basic-auth username,
`password` as the Basic-auth password); the secret is never logged. All streams and writes share
this one credential pair.

## Streams notes

12 streams share the identical read shape: `GET` against `accounts/{account_id}/<resource>`,
records at the response root (a top-level JSON array), 1-based `page_number` pagination
(`page_param: page`, `size_param: limit`, `start_page: 1`, `page_size: 100`), stopping on a short
page (fewer records returned than the requested page size):

- `flows`, `flows_v2`, `pins`, `mobile_experiences`, `launchpads`, `banners`, `checklists`,
  `embeds`, `nps` — the 9 in-app experience types Appcues publishes, all primary-keyed on `id` with
  `updatedAt` as the catalog cursor field (no server-side incremental filter is sent — see Known
  limits).
- `segments`, `tags` — audience-management resources, same shape.
- `jobs` — offline/async job records (id/name/status/started_at/url); no `updatedAt`-shaped field
  is published for this resource, so no `x-cursor-field` is declared.
- `sdk_keys` — SDK authentication key metadata (id/name/tag_field/created_at); no update timestamp
  is published, so no `x-cursor-field` is declared.

Matching legacy exactly, no stream declares a server-side incremental filter — every request reads
the same full list regardless of `req.State`; only the catalog cursor field is declared (where the
resource has one) so a `*_deduped`/`incremental_append_deduped` sync mode is available if a future
capability adds server-side filtering.

## Write actions & risks

`capabilities.write: true`. 34 actions, all requiring reverse-ETL plan approval before executing
(`metadata.json`'s `risk.approval`):

- **Publish/unpublish, 9 pairs** (`publish_flow`/`unpublish_flow`, and the same pair for
  `flow_v2`/`pin`/`mobile_experience`/`launchpad`/`banner`/`checklist`/`embed`/`nps`) —
  `POST .../{id}/publish` or `/unpublish`, no request body. Each immediately changes whether an
  experience is live to end users.
- **Segments** — `create_segment`/`update_segment`/`delete_segment` (`POST`/`PATCH`/`DELETE
  /segments[/{id}]`) manage segment definitions; `add_segment_user_ids`/`remove_segment_user_ids`
  (`POST /segments/{id}/add_user_ids` or `/remove_user_ids`) enqueue an async membership-update job
  for an explicit `user_ids` array. Mutating or deleting a segment changes which users any flow,
  banner, or checklist targeting rule referencing it reaches.
- **SDK keys** — `create_sdk_key`/`update_sdk_key`/`delete_sdk_key` manage authentication keys used
  by client SDKs to ingest data; `enable_sdk_key_enforcement`/`disable_sdk_key_enforcement` and
  `enable_sdk_key_secure_data_ingest`/`disable_sdk_key_secure_data_ingest` toggle per-key security
  posture. Deleting a key or enabling enforcement/secure-ingest can immediately break an
  already-deployed client SDK's ability to send data.
- **End users** — `update_user_profile` (`PATCH .../users/{user_id}/profile`, arbitrary attribute
  key/value pairs beyond the required `user_id` path field) and `delete_user_profile` (`DELETE
  .../users/{user_id}/profile`, an async job) mutate or permanently erase a specific end user's
  targeting profile and completion history; `track_user_event` (`POST .../users/{user_id}/events`)
  injects a synthetic behavioral event that can itself trigger targeting rules.
- **Groups** — `update_group_profile` (`PATCH .../groups/{group_id}/profile`) and
  `associate_group_users` (`PATCH .../groups/{group_id}/users`, an explicit `user_ids` array)
  manage group-level targeting attributes and membership.

None of these actions require a value from a read stream that isn't already available — every
write's addressable id (`id`/`user_id`/`group_id`) is either a value already emitted by a covered
read stream (`flows`, `segments`, `sdk_keys`, etc.) or an operator-supplied identifier from the
customer's own product data (end-user/group ids Appcues has no way to enumerate itself — see Known
limits).

## Known limits

- No incremental sync mode is server-side-wired for any of the 12 streams — this mirrors legacy's
  own full-refresh-only behavior for the original 5 streams, extended identically to every stream
  added in this expansion, not a capability gap introduced by migration.
- Appcues publishes no list-all-users or list-all-groups endpoint (a deliberate CDP-style privacy
  posture): reading or writing a specific end-user/group profile always requires an operator to
  already know that user/group id from their own product data. This connector therefore has no
  `users`/`groups` READ stream (there is no id source to enumerate from and no fan_out
  `ids_from.request` this API supports), but DOES expose the addressable per-id write actions
  (`update_user_profile`, `delete_user_profile`, `track_user_event`, `update_group_profile`,
  `associate_group_users`) since a write's target id comes from the caller, not from a list.
- Bulk import (`import/profiles`, `import/groups`, `import/events`) and async bulk export
  (`export/events`, `segments/{id}/segment_membership_export`) are out of scope: these are
  arbitrary-size batch operations (an array-of-objects import payload, or a job-id-then-poll export
  protocol), not single-record mutations the create/update/delete write model targets. See
  `api_surface.json` for the full per-endpoint rationale.
- `screenshots` and `ingestion_filtering_rules` are out of scope for this wave (a QA-tooling image
  resource and a narrow operational-config resource, respectively) — deliberate depth cuts, not
  technical blockers; see `api_surface.json`.
- `page_size`/`max_pages` config values declared in `spec.json` are not wired into any template
  (the engine's `page_number` paginator's `PageSize` is a static int set once in `streams.json`, not
  template-resolvable) — this predates this wave's expansion and is unchanged; every stream added
  here uses the same fixed `page_size: 100` as the original 5 streams.
