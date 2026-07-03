# Overview

SmartEngage is a wave2 fan-out declarative-HTTP migration. It reads SmartEngage avatars, tags,
custom fields, sequences, and subscribers through the SmartEngage API (`GET
https://api.smartengage.com/...`). This bundle migrates `internal/connectors/smartengage` (the
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

None. SmartEngage's legacy connector is read-only (`Write` returns
`connectors.ErrUnsupportedOperation`); `capabilities.write` is `false` and this bundle ships no
`writes.json`.

## Known limits

None beyond the `id`/`avatar_id` primary-key note above (Streams notes): this bundle is a
byte-for-byte parity port of legacy's read behavior, with no scope narrowing. Passthrough
projection (Streams notes) means the schemas document the known field set but do not gate what
flows through on a live read, matching legacy's own unfiltered emit.
