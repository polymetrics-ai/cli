# Overview

Zendesk is an issue #160 operation-ledger bundle for the official Zendesk Support API OAS. It
tracks 617 operations across 429 paths in `operations.json`, enables 70 top-level ETL streams, and
keeps bounded JSON direct reads for remaining typed GET operations, and models 295 non-deprecated
mutations as gated reverse-ETL writes without enabling binary downloads yet.

Operation method baseline: GET=320, PUT=89, POST=110, DELETE=85, PATCH=13.

Operation ledger candidate baseline: binary_read=37, deprecated=3, destructive_action=85, direct_read=282, sensitive_reverse_etl=210.

## Auth setup

Connection fields:

- `base_url` (required): Zendesk account root URL, for example `https://example.zendesk.com`.
- `access_token` (optional, secret): OAuth2 bearer token. It takes precedence when configured.
- `api_token` (optional, secret): Zendesk API token for Basic auth.
- `email` (optional, secret): email paired with `api_token` for Basic auth.

Secrets are redacted in logs, previews, docs examples, and JSON output. Do not pass secret values in
prompt text or command arguments.

## Streams notes

This bundle enables 70 top-level collection streams from official OAS object-array responses.
Provider `next_page` pagination is modeled with same-origin `next_url` pagination where available;
child collections that require arbitrary path IDs remain direct reads.

## Write actions & risks

Zendesk write actions are endpoint-specific reverse-ETL actions generated from the official OAS.
All writes require plan → preview → approval → execute. DELETE actions include `confirm:
destructive`; deprecated mutating operations stay blocked in the operation ledger.

## Known limits

- This bundle enables top-level ETL streams, typed bounded JSON direct reads, and gated reverse-ETL writes; binary downloads remain policy-only until the binary lane is completed.
- `api_surface.json` uses blocked-by-default operation rows and `operations.json` carries typed REST
  and binary operation metadata so the official OAS inventory can be reviewed without exposing raw
  generic HTTP tools.
- `cli_surface.json` is command/help metadata only and does not add runtime `pm zendesk` dispatch.
- No credentialed Zendesk checks were run while authoring this metadata.
