# Overview

ZapSign is an electronic signature platform. This bundle reads documents, signers, and templates
from the ZapSign REST API (`GET {base_url}/docs/`, `/signers/`, `/templates/`). It migrates
`internal/connectors/zapsign` (the hand-written legacy connector) to a declarative Tier-1 bundle at
capability parity; the legacy package stays registered and unchanged until wave6's registry flip.

## Auth setup

Requires a single secret, `api_token` (ZapSign API token), sent as `Authorization: Token
<api_token>` — an `api_key_header` auth spec with `header: Authorization` and `prefix: "Token "`,
matching legacy's `connsdk.APIKeyHeader("Authorization", token, "Token ")` exactly
(`zapsign.go:118`). `base_url` defaults to `https://api.zapsign.com.br/api/v1`
(`zapsign.go:17`'s `defaultBaseURL`), materialized via `spec.json`'s `"default"` when unset.

## Streams notes

All three streams (`documents`, `signers`, `templates`) are single-page GET reads with no
pagination and no incremental support — legacy performs one unconditional request per stream and
emits every record from the response's top-level `results` array; this bundle does the same
(`records.path: "results"`, no `pagination` block declared).

Each stream's raw API record carries its identifier under the field name `token`, not `id`; a
`computed_fields` rename (`"id": "{{ record.token }}"`) maps it to the schema's `id` property,
matching legacy's `mapDocument`/`mapSigner`/`mapTemplate` output field name exactly. Legacy's
`signers`/`templates` mappers additionally fall back to a bare `id` field via a `first(token, id)`
helper when `token` is absent from a record; this bundle models only the primary `token`-present
case (ZapSign's documented wire shape always includes `token` for every one of these object types),
since the declarative dialect has no ordered-fallback-across-two-fields primitive — see Known
limits.

## Write actions & risks

None. `capabilities.write` is `false` and this bundle ships no `writes.json`, matching legacy's
`Write` returning `connectors.ErrUnsupportedOperation` unconditionally.

## Known limits

- **`signers`/`templates`' `token`-absent id fallback is not modeled.** Legacy's `mapSigner`/
  `mapTemplate` compute `id` as `first(item["token"], item["id"])` — an ordered fallback that only
  matters if a record ever omits `token` and instead carries a bare `id` field. ZapSign's documented
  API always returns `token` as the canonical identifier for signers and templates, so this fallback
  is defensive/unreachable on the real wire shape; the declarative dialect has no
  ordered-multi-field-fallback primitive (only a single bare `{{ record.<path> }}` reference or a
  filter chain), so only the `token`-present case is expressed. Deliberately out-of-scope
  edge case, not a defect.
