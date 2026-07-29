# Overview

Seeded no-false-positive companion for N2: `events` uses the same generic comparison-prefix
incremental formatting as `start-date-free-form-string`, but `spec.json` declares `start_date` with
`format: date-time`, so `connectorgen validate` should not emit a `start_date_free_form_string`
warning.

## Auth setup

No auth required; synthetic API.

## Streams notes

`events` is incremental on `created`.

## Write actions & risks

None; read-only bundle.

## Known limits

None; this is test fixture data.
