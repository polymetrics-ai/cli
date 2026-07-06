# Overview

Acme is a synthetic connector used as the conformance v2 self-test control case.

## Auth setup

Provide a bearer token via the `token` secret; falls back to no auth when unset.

## Streams notes

`widgets` is incremental on `updated_at` and paginates by page number.
`notes` is a single-page, full-refresh-only stream.

## Write actions & risks

`update_widget` is a low-risk PATCH. `delete_widget` is a high-risk, destructive-confirm DELETE
with idempotent 404 handling.

## Known limits

None; this is test fixture data.
