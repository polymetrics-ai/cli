# BambooHR connector

## Overview

BambooHR (`bamboo-hr`) is implemented from the current official BambooHR OpenAPI 3.1 public specification (`https://openapi.bamboohr.io/main/latest/docs/openapi/public-openapi.yaml`) plus the OpenAPI top-level `webhooks` object. This connector-local final-wave ledger covers every documented operation exactly once.

Post-change operation counts:

| total | streams | direct reads | writes | blocked/planned | excluded | certified |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 316 | 138 | 9 | 149 | 20 | 0 | 0 |

Blocked/planned rows by model: admin_reverse_etl=7, binary_read=6, disallowed=1, local_workflow=6.

## Auth setup

Use `pm credentials add bamboo-hr` with a stored BambooHR API key from an approved secret source. The API key is sent as the HTTP Basic username with password `x`. The connector also defines an optional `access_token` secret for official OAuth-scoped operations; provide it from environment variables or stdin-backed credential storage when those scopes are needed. Never place secrets in command text, fixtures, docs, or issue comments.

## Streams notes

JSON object and array operations are modeled as fixture-backed streams. Generated final-wave streams use `projection: passthrough` with permissive schemas so BambooHR account-specific fields are preserved instead of silently dropped. Existing hand-authored BambooHR streams and fixtures are retained when their method/path still match the current official OpenAPI. Path parameters are connector config fields and conformance fixtures use synthetic values only.

## Write actions & risks

JSON and no-body mutations are named reverse-ETL actions with closed record schemas, path fields, synthetic request fixtures, action risk text, and destructive confirmation for DELETE/remove/clear-style operations. Reverse ETL remains plan -> preview -> explicit approval -> execute; fixture replay is the only write validation performed here. File, multipart, form-login, binary download/export, and inbound webhook delivery operations are not exposed as generic escape hatches.

## Known limits

- Binary/file/export reads are blocked as `binary_read` until the shared command runner supports a bounded binary output policy for connector commands.
- Multipart/file upload and form-login operations remain blocked; this wave does not add raw local file, login, or payload passthrough surfaces.
- The OpenAPI top-level webhooks describe provider-to-consumer event deliveries. They are tracked as blocked local workflow/changefeed operations because this connector has no inbound webhook/CDC receiver and durable state foundation in the current runtime.
- Certification is fixture-only. No live BambooHR credentials were requested or used, no provider calls were made, and this change does not claim live certification.
