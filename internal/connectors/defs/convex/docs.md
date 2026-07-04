# Overview

Convex is a backend-as-a-service platform (reactive database + serverless functions). This bundle
reads table metadata and documents from a Convex deployment through its HTTP API (`GET
{deployment_url}/api/tables`, `GET {deployment_url}/api/tables/<table>/documents`). It migrates
`internal/connectors/convex` (the hand-written connector) at capability parity; the legacy package
stays registered and unchanged until wave6's registry flip.

**Catalog label correction**: `catalog_data.json` carries both a `source-convex` entry (correct)
and a `destination-convex` entry (type `"destination"`, `supported_sync_modes:
["append","append_dedup","overwrite"]`, `config_fields: [access_key, deployment_url]`).
`docs/migration/inventory.json` likewise mis-tags this connector's `runtime_kind` as
`destination_go`. Both labels are **wrong** for the actual legacy implementation this bundle
migrates: `internal/connectors/convex/convex.go` declares `Capabilities{..., Write: false}` and its
`Write` method unconditionally returns `connectors.ErrUnsupportedOperation` — there is no write path
at all, compound or otherwise, in the Go source. This bundle therefore ships as a **read-only
source** (`capabilities.write: false`, no `writes.json`), matching the legacy Go connector — the
sole ground truth per migration convention — not the stale catalog metadata. `source-convex`'s own
catalog entry (`config_fields: [access_key (secret), deployment_url]`,
`supported_sync_modes: ["full_refresh"]`) already agrees with this bundle's shape; only
`destination-convex` is the residual mislabeled entry, presumably describing a different (Python
CDK, per `language: "python"`/`tags: ["cdk:python"]`) implementation never ported to this Go legacy
package.

## Auth setup

Provide a Convex deployment access key via the `access_key` secret; it is used only for Bearer auth
(`Authorization: Bearer <access_key>`, matching legacy's `connsdk.Bearer(key)`) and is never logged.
Also provide `deployment_url`, the Convex deployment's base URL (e.g.
`https://my-deployment.convex.cloud`).

## Streams notes

`tables` (`GET api/tables`) lists Convex table metadata; records are extracted from the top-level
`tables` array (`records.path: "tables"`), matching legacy's `connsdk.RecordsAt(resp.Body,
"tables")`. `documents` (`GET api/tables/{{ config.table }}/documents`) reads documents from one
table (`table` config, default `"data"`, matching legacy's `tableName` default); records are
extracted from the top-level `documents` array.

Legacy's `Read` emits every record's raw fields largely verbatim (`out := connectors.Record(rec)`,
no field-built mapping) — both streams declare `"projection": "passthrough"` per
`docs/migration/conventions.md` §8 rule 1, so every raw field Convex returns survives unfiltered,
not just the primary-key fields the schema documents for typing purposes. The one exception is
`documents`' `id` field: Convex's real system field is `_id`, and legacy additionally stamps a
compatibility `id` alias whenever the raw record has no native `id` of its own
(`if out["id"] == nil && out["_id"] != nil { out["id"] = out["_id"] }`, `convex.go:118-120`). This
bundle reproduces the guarded alias via a `coalesce` computed field
(`"id": "{{ coalesce record.id record._id }}"`), preserving a user-supplied raw `id` when present
and filling from `_id` only when absent/null. `x-primary-key: ["id"]` names the resulting field
directly.

Pagination for `documents` is `pagination.type: cursor` (`cursor_param: cursor`, `token_path:
cursor`) — the next-page token is read from the response body's top-level `cursor` field and sent
back as the `cursor` query parameter on the next request, stopping when the token is absent/empty;
this matches legacy's `strings.TrimSpace(next) == ""` stop condition on the identical body field
name exactly (`convex.go:126-130`). `tables` is not paginated in legacy (a single `GET api/tables`
call) and declares no `pagination` override.

## Write actions & risks

None. Convex is read-only here (`capabilities.write: false`), matching the legacy Go connector
exactly (`Capabilities.Write: false`, `Write` returns `ErrUnsupportedOperation`) — see the catalog
label correction above. This bundle ships no `writes.json` at all.

## Known limits

- **`base_url` (a `deployment_url`-equivalent test/proxy override) is not modeled.** Legacy accepts
  either `base_url` or `deployment_url`, preferring `base_url` when both are set
  (`convex.go:150-153`) — the same test/proxy-override shape commcare's `base_url` has. The engine's
  `streams.json` `base.url` is a single plain template with no `config-A-or-config-B` fallback
  primitive (unlike `auth`'s `when`-gated candidate list), so only one of the two keys can be wired.
  `deployment_url` is kept (it is the field `source-convex`'s own catalog entry documents as
  required) and `base_url` is dropped from `spec.json` entirely rather than left as an unwireable,
  dead config key (F6, `docs/migration/conventions.md`). This is a documented, accepted
  config-surface narrowing: every real deployment already sets `deployment_url` (`base_url` was a
  legacy test-only escape hatch), so this never changes emitted record DATA for any production
  config.
- Full Convex HTTP API surface (mutation/query/action endpoints) is out of scope for this pass; see
  `api_surface.json`'s `excluded: {category: out_of_scope, reason: "not implemented in this bundle"}`
  entries.
