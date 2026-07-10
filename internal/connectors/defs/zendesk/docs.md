# Overview

Zendesk is an issue #160 operation-ledger bundle for the official Zendesk Support API OAS. It
tracks 617 operations across 429 paths in `operations.json` without enabling streams,
direct reads, binary downloads, or writes yet.

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

No Zendesk streams are enabled in this ledger slice. Later lanes map durable collection reads from
the operation ledger to `streams.json`, schemas, fixtures, pagination, cursors, and primary keys.

## Write actions & risks

No Zendesk write actions are enabled in this ledger slice. Mutation candidates in `api_surface.json`
and `operations.json` remain blocked by default until later lanes add typed reverse-ETL schemas,
risk text, approval text, redaction policy, and `confirm: destructive` where required. Reverse ETL
remains plan → preview → approval → execute.

## Known limits

- This bundle is a metadata scaffold only; it does not claim executable read/write parity yet.
- `api_surface.json` uses blocked-by-default operation rows and `operations.json` carries typed REST
  and binary operation metadata so the official OAS inventory can be reviewed without exposing raw
  generic HTTP tools.
- `cli_surface.json` is command/help metadata only and does not add runtime `pm zendesk` dispatch.
- No credentialed Zendesk checks were run while authoring this metadata.
