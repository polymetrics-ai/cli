# Overview

Test fixture bundle; not a real connector. It exists so `internal/app` tests can exercise the
non-batchable write primitive end to end without editing a shipped connector bundle.

## Auth setup

None; the fixture points `config.base_url` at an `httptest` server.

## Streams notes

Single `widgets` stream, no pagination.

## Write actions & risks

- `cast_vote` — declared `batchable: false`. Individually executable via `pm humanproxy-demo vote`;
  refused by any bulk reverse ETL plan.
- `sync_profile` — batchable (field absent, so it defaults to true). Runs in bulk unchanged.

## Known limits

Fixture only; used by `internal/app/reverse_batchable_test.go`.
