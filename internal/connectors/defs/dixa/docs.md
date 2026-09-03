# Dixa Connector

## Overview

Reads Dixa conversation export records through fixed bearer-authenticated export routes.

Readable streams: `conversations`, `conversation_queue`, `conversation_rating`, `conversation_assignment`.

Service API documentation: https://docs.dixa.io/.

## Auth setup

Connection fields:

- `api_token` (optional, secret, string); Dixa API token.
- `updated_after` (optional, string); Inclusive export lower bound in milliseconds.
- `updated_before` (optional, string); Exclusive export upper bound in milliseconds.

Authentication uses declared mode(s): `bearer`.

## Execution contract

Connection check: `GET /conversation_export`
Check query: `updated_after`=`{{ config.updated_after }}`; `updated_before`=`{{ config.updated_before }}`.

## Streams notes

- `conversations`: `GET /conversation_export`; records `.`
  - Query: `updated_after`=`{{ config.updated_after }}`; `updated_before`=`{{ config.updated_before }}`.
  - Incremental cursor: `updated_at`.
- `conversation_queue`: `GET /conversation_export`; records `.`
  - Query: `updated_after`=`{{ config.updated_after }}`; `updated_before`=`{{ config.updated_before }}`.
  - Incremental cursor: `updated_at`.
- `conversation_rating`: `GET /conversation_export`; records `.`
  - Query: `updated_after`=`{{ config.updated_after }}`; `updated_before`=`{{ config.updated_before }}`.
  - Incremental cursor: `updated_at`.
- `conversation_assignment`: `GET /conversation_export`; records `.`
  - Query: `updated_after`=`{{ config.updated_after }}`; `updated_before`=`{{ config.updated_before }}`.
  - Incremental cursor: `updated_at`.

## Write actions & risks

This connector's write surface is declared separately in the rendered execution bundle.
