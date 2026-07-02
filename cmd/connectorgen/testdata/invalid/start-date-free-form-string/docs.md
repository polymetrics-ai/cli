# Overview

Seeded invalid-corpus case for N2 (wave0 REVIEW.md carried flag): `events` is incremental via
`start_config_key: start_date`, `param_format: github_date_range`, but spec.json's `start_date`
property declares no date-ish `format` (free-form string) — a config value like `20260101`
(yyyymmdd, not Unix seconds) would silently pass the engine's digits-passthrough as a Unix-seconds
lower bound instead of erroring.

## Auth setup

No auth required; synthetic API.

## Streams notes

`events` is incremental on `created`.

## Write actions & risks

None; read-only bundle.

## Known limits

None; this is test fixture data.
