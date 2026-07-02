# Overview

Zenefits is an HR platform. This bundle reads people, companies, and departments from the Zenefits
Core API (`GET {base_url}/people`, `/companies`, `/departments`). It migrates
`internal/connectors/zenefits` (the hand-written legacy connector) to a declarative Tier-1 bundle at
capability parity; the legacy package stays registered and unchanged until wave6's registry flip.

## Auth setup

Requires a single secret, `token`, sent as `Authorization: Bearer <token>` — a `bearer` auth spec,
matching legacy's `connsdk.Bearer(token)` exactly (`zenefits.go:118`). `base_url` defaults to
`https://api.zenefits.com/core` (`zenefits.go:17`'s `defaultBaseURL`), materialized via
`spec.json`'s `"default"` when unset.

## Streams notes

All three streams (`people`, `companies`, `departments`) are single-page GET reads with no
pagination and no incremental support — legacy performs one unconditional request per stream and
emits every record from the response's top-level `data` array; this bundle does the same
(`records.path: "data"`, no `pagination` block declared).

Every stream's raw API record already carries an `id` field matching the schema directly, so plain
schema projection copies every field through with no `computed_fields` rename needed —
`people` maps `id`/`first_name`/`last_name`/`status`; `companies` and `departments` both map
`id`/`name` (legacy's `departments` stream reuses the identical `mapCompany` mapper function, so
this bundle gives `departments` its own schema with the same two fields for stream-name clarity,
rather than aliasing the `companies` schema file).

## Write actions & risks

None. `capabilities.write` is `false` and this bundle ships no `writes.json`, matching legacy's
`Write` returning `connectors.ErrUnsupportedOperation` unconditionally.

## Known limits

None beyond the standard wave2 scope narrowing (write endpoints are Pass B, see `api_surface.json`).
Every legacy read-path behavior (streams, auth, base URL default, record mapping) is fully modeled
declaratively with no deviations.
