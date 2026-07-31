# Recurly connector

## Overview

Declarative Recurly V3 connector generated from the official v2021-02-25 OpenAPI definition. The operation ledger partitions all 197 documented operations once: 93 ETL streams, 96 typed reverse-ETL write actions, 5 bounded JSON read/preview operations, and 3 official binary/export endpoints recorded as bounded metadata with shared-runtime execution blocked in this slice.

Official source: https://recurly.com/developers/api/spec/v2021-02-25.yaml

## Auth setup

Provide a Recurly API key through the connector credential flow or environment-backed secret input. Do not paste, print, summarize, or commit secret values. The declarative runtime uses HTTP Basic auth with the API key as the username and an empty password, plus the `Accept: application/vnd.recurly.v2021-02-25` header.

Path-scoped operations use fixed connector config keys such as `account_id`, `subscription_id`, `invoice_id`, and related Recurly identifiers. Inspect `spec.json` or `pm connectors inspect recurly --json` before creating credentials for scoped reads.

## Streams notes

Streams are fixed Recurly endpoints with schema-referenced JSON records and synthetic conformance fixtures. List endpoints read the Recurly `data` array; singular GET endpoints emit the response object. No arbitrary query, body, path, shell, file, or generic HTTP passthrough is exposed.

## Write actions & risks

Reverse ETL actions are generated from official POST, PUT, and DELETE operations and use closed `record_schema` definitions. Mutation execution must stay plan → preview → explicit approval → execute. Destructive lifecycle actions are marked with `confirm: destructive`. Recurly supports provider idempotency for POST, PUT, PATCH, and DELETE through `Idempotency-Key`; do not reuse idempotency keys for different mutation records.

## Known limits

The bundle records three official binary/export endpoints as bounded `binary_download` operation metadata, but shared direct-read command execution for binary payload transfers is not implemented in this connector slice. The connector has not been live-certified in this wave; fixtures are synthetic and no credentialed provider calls were run.
