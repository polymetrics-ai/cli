# Mendeley Connector

## Overview

Reads documents, folders, groups, and annotations through fixed Mendeley REST routes and OAuth refresh-token authentication.

Readable streams: `documents`, `folders`, `groups`, `annotations`.

Service API documentation: https://dev.mendeley.com/reference.

## Auth setup

Connection fields:

- `client_id` (required, secret, string); Mendeley OAuth client identifier.
- `client_refresh_token` (required, secret, string); Mendeley OAuth refresh token.
- `client_secret` (required, secret, string); Mendeley OAuth client secret.
- `name_for_institution` (required, string); Retained Mendeley institution query configuration.
- `query_for_catalog` (required, string); Retained Mendeley catalog query configuration.
- `start_date` (required, string); Initial incremental lower bound.

Authentication uses declared mode(s): `oauth2_refresh_token`.

## Execution contract

Connection check: `GET /documents`
Check query: `limit`=`1`.

## Streams notes

- `documents`: `GET /documents`; records ``
  - Query: `order`=`asc`.
  - Incremental cursor: `last_modified`.
- `folders`: `GET /folders`; records ``
  - Query: `order`=`asc`.
  - Incremental cursor: `modified`.
- `groups`: `GET /groups`; records ``
  - Query: `order`=`asc`.
- `annotations`: `GET /annotations`; records ``
  - Query: `order`=`asc`.
  - Incremental cursor: `last_modified`.

## Write actions & risks

This connector's write surface is declared separately in the rendered execution bundle.
