# Mode Connector

## Overview

Reads Mode workspace collections through fixed HAL+JSON REST routes.

Readable streams: `spaces`, `reports`, `data_sources`, `groups`, `memberships`.

Service API documentation: https://mode.com/developer/api-reference/.

## Auth setup

Connection fields:

- `api_secret` (required, secret, string);
- `api_token` (required, secret, string);
- `workspace` (required, string);

Authentication uses declared mode(s): `basic`.

## Execution contract

Connection check: `GET /{{ config.workspace }}/spaces`

## Streams notes

- `spaces`: `GET /{{ config.workspace }}/spaces`; records `_embedded.spaces`
  - Incremental cursor: `updated_at`.
- `reports`: `GET /{{ config.workspace }}/reports`; records `_embedded.reports`
  - Incremental cursor: `updated_at`.
- `data_sources`: `GET /{{ config.workspace }}/data_sources`; records `_embedded.data_sources`
  - Incremental cursor: `updated_at`.
- `groups`: `GET /{{ config.workspace }}/groups`; records `_embedded.groups`
  - Incremental cursor: `updated_at`.
- `memberships`: `GET /{{ config.workspace }}/memberships`; records `_embedded.memberships`
  - Incremental cursor: `created_at`.

## Write actions & risks

This connector's write surface is declared separately in the rendered execution bundle.
