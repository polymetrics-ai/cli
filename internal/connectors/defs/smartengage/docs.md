# Overview

SmartEngage is a wave2 fan-out declarative-HTTP migration, expanded in Pass B to the full
documented SmartEngage API surface. It reads SmartEngage avatars, tags, custom fields, sequences,
and subscribers, and writes new subscribers/tags/custom-field values/sequence enrollments,
through the SmartEngage API (`https://api.smartengage.com/...`, documented at
`https://smartengage.com/docs/`). This bundle migrates `internal/connectors/smartengage` (the
hand-written connector); the legacy package stays registered and unchanged until wave6's registry
flip.

## Auth setup

Provide a SmartEngage API token via the `api_key` secret; it is sent as a Bearer token
(`Authorization: Bearer <api_key>`), matching legacy's `connsdk.Bearer(key)`
(`smartengage.go:143`). `base_url` defaults to `https://api.smartengage.com` and may be
overridden for tests/proxies, matching legacy's own `validatedBaseURL` default.

## Streams notes

All five streams (`avatars`, `tags`, `custom_fields`, `sequences`, `subscribers`) hit their own
`GET <resource>` endpoint, matching legacy's `streamEndpoints` map exactly (`avatars/list`,
`tags/list/`, `customfields/list/`, `sequences/list/`, `subscribers/list/`). Records are
extracted from the response body's top-level array (`records.path: ""`), matching legacy's
`recordsPath: ""` for every stream. None of the streams paginate in legacy (a single `r.Do` call
per read, no loop) — `pagination.type: none` is declared, one request per read. An optional
`avatar_id` config value is applied as a query filter on EVERY stream read, matching legacy's
`queryParams` (called uniformly regardless of which stream is being read), omitted entirely when
unset. Every stream declares `"projection": "passthrough"`: legacy's `readRecords` emits
`connectors.Record(rec)` verbatim off the wire with no field-built mapping
(`smartengage.go:102-120`), so schema-mode projection would silently drop any field the live API
returns beyond `id`/`avatar_id`/`name` — passthrough preserves legacy's actual verbatim-emit
behavior (conventions.md §8 rule 1). Records carry `id`, `avatar_id`, and `name`, the exact field
set legacy's `streams()` catalog declares for every stream; `x-primary-key` is `id` per legacy's
own catalog declaration (`PrimaryKey: []string{"id"}`), even though the live avatars endpoint's
natural identity field is `avatar_id` — `id` is not declared `required` in this bundle's schemas
for that reason. The schemas remain a documentation surface (the three declared fields), not an
enforced allowlist.

## Write actions & risks

Legacy's own connector is read-only, but Pass B full-surface expansion adds every
dialect-expressible mutation SmartEngage documents (`api_surface.json`), all form-encoded
(`body_type: "form"`, matching the API's documented `-d key=value` curl examples), each requiring
approval:

- **`add_subscriber`** (`POST subscribers/add`, `create`): creates a subscriber under `avatar_id`
  with optional name/email/Facebook-id/push-id fields.
- **`update_subscriber`** (`POST subscribers/update`, `update`): overwrites fields on an existing
  `subscriber_id`; SmartEngage's own docs describe this as accepting "mostly the same arguments as
  the creation call" with omitted fields left unchanged.
- **`create_tag`** (`POST tags/create`, `create`): creates a new tag under `avatar_id`.
- **`add_tag_to_subscriber`** / **`remove_tag_from_subscriber`** (`POST tags/add` / `POST
  tags/delete`, `custom`): attach/detach an existing tag (by name) to/from a subscriber.
- **`create_custom_field`** (`POST customfields/create`, `create`): creates a new custom-field
  definition under `avatar_id` (name/title/type/description).
- **`set_custom_field_value`** (`POST customfields/update`, `update`): sets a custom field's value
  on a specific subscriber.
- **`add_subscriber_to_sequence`** / **`remove_subscriber_from_sequence`** (`POST sequences/add` /
  `POST sequences/remove`, `custom`): enroll/unenroll a subscriber in an automation sequence.
  Risk: this is the highest-impact write in this bundle — enrolling a subscriber triggers
  SmartEngage's own scheduled-message delivery for that sequence, a real outbound communication
  side effect, not just a data mutation.

Every action's `record_schema` requires `avatar_id` (SmartEngage's account/list scoping key,
mirrored from every read stream's own optional `avatar_id` query filter) alongside whatever
resource-specific identifier(s) that mutation needs (`subscriber_id`, `tag`, `sequence`, `field`).

## Known limits

- The `id`/`avatar_id` primary-key note (Streams notes) still applies to every read stream: this
  is a byte-for-byte parity port of legacy's read behavior, with no read-side scope narrowing.
  Passthrough projection (Streams notes) means the schemas document the known field set but do not
  gate what flows through on a live read, matching legacy's own unfiltered emit.
- **Write request/response body shapes are inferred from SmartEngage's own public API
  documentation and Zapier/Pipedream integration references, not from a legacy Go implementation**
  (legacy has no write path to port from at all — `capabilities.write` was `false` prior to this
  Pass B expansion). Field names (`avatar_id`/`subscriber_id`/`tag`/`sequence`/`field`/`value`/
  `custom_field_name`/`custom_field_type`) are the documented parameter names; exact response
  envelopes beyond the fields SmartEngage's docs explicitly show are not guaranteed byte-for-byte,
  though the request shape (method, path, required parameters) is.
