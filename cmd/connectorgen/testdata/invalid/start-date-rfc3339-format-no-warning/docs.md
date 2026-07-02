# Overview

N2 no-false-positive companion to `start-date-free-form-string`: identical `incremental` shape
(`param_format: github_date_range`, `start_config_key: start_date`), but `spec.json`'s
`start_date` property declares `format: date-time`, so `start_date_free_form_string` must NOT
fire.

## Auth setup

No auth required; synthetic API.

## Streams notes

`events` is incremental on `created`.

## Write actions & risks

None; read-only bundle.

## Known limits

None; this is test fixture data.
