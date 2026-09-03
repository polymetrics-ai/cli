# Babelforce Connector

## Overview

Reads Babelforce call reporting, recordings, numbers, and users through fixed Babelforce v2 REST routes.

Readable streams: `calls`, `calls_extended`, `recordings`, `numbers`, `users`.

Service API documentation: https://docs.babelforce.com/.

## Auth setup

Connection fields:

- `access_key_id` (optional, secret, string); Babelforce access key ID.
- `access_token` (optional, secret, string); Babelforce access token.
- `region` (optional, string); Declared Babelforce regional provider host.

Authentication uses declared mode(s): `none`.

## Execution contract

Default stream pagination: `page_number`.

Connection check: `GET /calls/reporting/simple`
Check query: `max`=`1`.

## Streams notes

- `calls`: `GET /calls/reporting/simple`; records `items`
  - Incremental cursor: `dateCreated`.
- `calls_extended`: `GET /calls/reporting/extended`; records `items`
  - Incremental cursor: `dateCreated`.
- `recordings`: `GET /recordings`; records `items`
  - Incremental cursor: `dateCreated`.
- `numbers`: `GET /numbers`; records `items`
  - Incremental cursor: `dateCreated`.
- `users`: `GET /users`; records `items`
  - Incremental cursor: `dateCreated`.

## Write actions & risks

This connector's write surface is declared separately in the rendered execution bundle.
