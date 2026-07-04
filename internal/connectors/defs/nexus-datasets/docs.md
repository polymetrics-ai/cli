# Overview

Infor Nexus Datasets is a Tier-2 (AuthHook) migration of `internal/connectors/nexus-datasets`
(quarantine.json: `AUTH_COMPLEX`, "complex auth (hook needed)"). It reads records from a single
configured Infor Nexus export dataset through the Infor Nexus Data API (v3.1), authenticating with
an HMAC-SHA256 request signature the declarative dialect cannot express (no `auth.mode` computes a
signature over method/path/timestamp). This bundle is parity-tested against
`internal/connectors/nexus-datasets` (the hand-written connector it migrates); the legacy package
stays registered and unchanged until wave6's registry flip. Read-only: legacy's `Write` always
returns `ErrUnsupportedOperation`, and this bundle declares `capabilities.write: false` with no
`writes.json` to match.

## Auth setup

Provide four secrets: `access_key_id`, `user_id`, `secret_key`, and `api_key` (all `x-secret` in
`spec.json`, never logged). `hooks/nexus-datasets/hooks.go` implements `AuthHook`, porting legacy
`nexus_datasets.go`'s `hmacAuth` field-for-field: on every outgoing request it computes
`HMAC-SHA256(secret_key, method + "\n" + path + "\n" + unix_timestamp)`, base64-encodes the
signature, and sets:

- `X-Infor-AccessKeyId: <access_key_id>`
- `X-Infor-UserId: <user_id>`
- `X-Infor-ApiKey: <api_key>`
- `X-Infor-Timestamp: <unix_timestamp>`
- `Authorization: InforNexus <access_key_id>:<signature>`

`secret_key` never appears on the wire itself (only its HMAC digest does) and is never logged; no
error path in the hook interpolates a secret value into an error string.

The bundle's `base.auth` declares exactly one candidate: `{"mode": "custom", "hook":
"nexus-datasets"}` — legacy has no alternate auth path, so there is no `when`-gated fallback to
declare (same shape as gmail's sole custom-mode candidate).

## Streams notes

One stream, `datasets`, primary-keyed on `id`, incrementally cursored on `updated_at`
(`incremental.request_param: modifiedSince`, `param_format: rfc3339`, `start_config_key:
start_date`) — matches legacy's `incrementalLowerBound` (state cursor, falling back to
`start_date`). Pagination is `offset_limit` (`limit`/`offset` query params, fixed `page_size: 100`)
with the engine's standard short-page stop, matching legacy's default page size and unbounded
`max_pages` behavior. `computed_fields` stamps the `dataset_name` marker onto every record
(legacy's `rec["dataset_name"] = dataset`, streams.go's routing table has only one dataset per
connector instance, so this is a static-literal-shaped config reference, not a per-record
derivation) and uses `coalesce` to preserve legacy's alternate `id` (`record_id`/`key`/`uid`) and
`updated_at` (`modified_at`/`last_modified`/`timestamp`) fallbacks when those alternate fields are
present.

Records are extracted from the page body's `records` array (`records.path: "records"`), the API's
documented envelope shape: each item carries `id`, `raw_data` (the dataset row payload, an opaque
JSON object), `raw_data_string` (its stringified form), and `updated_at`. Schema projection keeps
these four fields verbatim (`raw_data`/`raw_data_string` typed `["object","null"]`/`["string",
"null"]`, matching the API's real wire shape and legacy's pass-through).

## Write actions & risks

None — Infor Nexus dataset export is read-only. `capabilities.write: false`, no `writes.json` file,
matching legacy's `ErrUnsupportedOperation` (`nexus_datasets.go` `Write`).

## Known limits

- **`page_size`/`max_pages` runtime overrides are not modeled.** Legacy accepts `config.page_size`
  (1-1000, default 100) and `config.max_pages` (0/`all`/`unlimited` = unbounded). The engine's
  `offset_limit` pagination fields are fixed bundle literals, so `spec.json` intentionally does not
  declare dead config properties. The bundle matches legacy defaults: `limit=100` and unbounded
  pages.
- **The raw-record `raw_data` fallback is not modeled.** Legacy's `nexusDatasetRecord` sets
  `raw_data` to the entire raw row when the documented `raw_data` envelope is absent. The current
  `computed_fields` dialect can coalesce `record.*` paths but cannot copy the root `record` object
  into a field, so this defensive malformed-row fallback remains an engine gap. Legacy also treats
  empty strings as missing in the `id`/`updated_at` fallback chain; `coalesce` is first-non-null, so
  it models absent/null alternate keys but not empty-string fallback semantics.
- **The HMAC canonicalization scheme (`method + "\n" + path + "\n" + timestamp`) is legacy's own
  best-effort implementation**, not a scheme independently verified against Infor's own signing
  spec (legacy's own doc comment: "the exact upstream canonicalization may differ across Infor
  Nexus deployments"). This bundle reproduces legacy's exact scheme byte-for-byte (same doc caveat
  carried forward) — this is a parity port, not a new gap introduced by migration.
- **`TestConformance/nexus-datasets`'s dynamic (fixture-replay) checks are `skip_dynamic`'d** for
  the same reason as gmail: the sole auth candidate is `mode: custom`, and conformance's synthetic
  non-secret config can never carry a real `secret_key`/`api_key` that would make a signed
  request meaningful against the replay server. See `metadata.json`'s `conformance.reason`. The
  hook's own unit tests (`internal/connectors/hooks/nexus-datasets/hooks_test.go`) and the
  pre-existing legacy test suite (`internal/connectors/nexus-datasets/nexus_datasets_test.go`,
  unchanged, still passing against the read-only legacy package) remain the authoritative
  correctness bar for the HMAC auth path.
