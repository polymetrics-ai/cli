# Overview

Interzoid is a wave2 fan-out declarative-HTTP migration. It is an AI-powered data-quality /
data-matching API: each stream is a single lookup endpoint that, given an input value (a company
name, person name, street address, or organization name), returns one JSON object containing a
similarity key (`SimKey`) or a standardized value (`Standard`) plus a remaining credit count. This
bundle is engine-vs-legacy parity-tested against `internal/connectors/interzoid` (the hand-written
connector it migrates); the legacy package stays registered and unchanged until wave6's registry
flip.

## Auth setup

Provide an Interzoid license key via the `api_key` secret; it is injected as the `license` query
parameter on every request (`api_key_query` auth mode, `param: license`) and is never logged,
matching legacy's `connsdk.APIKeyQuery("license", secret)`. `base_url` defaults to
`https://api.interzoid.com` and may be overridden for tests/proxies.

## Streams notes

Each of the 4 streams hits one fixed Interzoid lookup endpoint and returns a single JSON object at
the response root (no array, no pagination) — `records: {"path": "", "single_object": true}` maps
that single object to one emitted record per read, matching legacy's `RecordsAt(resp.Body, "")`
one-element-result behavior exactly.

- `company_name_matching` (`GET /getcompanymatchadvanced`): requires `config.company`; optional
  `config.company_match_algorithm` is sent as `algorithm` only when set
  (`omit_when_absent: true`, matching legacy's `buildInputs`' `required: false` skip-when-empty
  behavior for that input).
- `individual_name_matching` (`GET /getfullnamematch`): requires `config.fullname`.
- `street_address_matching` (`GET /getaddressmatchadvanced`): requires `config.address`; optional
  `config.address_match_algorithm` is sent as `algorithm` only when set, same as above.
- `standardize_company_names` (`GET /getorgstandard`): requires `config.org`.

Each stream's `computed_fields` echoes its own input value back onto the record under
`query_<name>` (e.g. `query_company`) via `config.*` in `computed_fields` — the sanctioned
Tier-1 mechanism for stamping a config-scoped value onto every emitted record — matching legacy's
`echo` map / `simKeyRecord`/`standardRecord` mappers exactly. A required input that is absent hard
errors when the query template's `{{ config.<key> }}` reference fails to resolve, matching
legacy's `"interzoid stream requires config %q"` per-endpoint required-input check — same failure
classification (an absent required lookup input), different literal error text.

There is no incremental cursor and no pagination for any stream, matching legacy: every Interzoid
lookup is a one-shot, single-record GET.

## Write actions & risks

None. Interzoid is a read-only data-matching API with no reverse-ETL surface; `capabilities.write`
is `false` and this bundle ships no `writes.json`, matching legacy's `Write` returning
`connectors.ErrUnsupportedOperation`.

## Known limits

- **No declarative `check` request.** Legacy's `Check()` never issues a live lookup (each lookup
  spends an Interzoid API credit); it only validates that `api_key` and a well-formed `base_url`
  are present. This bundle declares no `streams.json` `base.check` block at all, so the engine's
  `Check()` performs the identical no-network-call validation (auth/URL resolution only, via
  `newRuntime`) without spending a credit — the closest-fidelity port available, rather than
  inventing a live check request legacy deliberately avoids.
  `conformance`'s `check_fixture` check gracefully Skips for a bundle with no declared `check`
  (there is nothing to exercise), which is the expected, honest outcome here.
- Only the 4 core lookup endpoints are migrated; Interzoid's broader API catalog (email
  validation, geocoding, currency conversion) is out of scope for wave2 — see `api_surface.json`'s
  `excluded: {category: out_of_scope, reason: "Pass B capability expansion"}` entries.
