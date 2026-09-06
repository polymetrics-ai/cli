# Basecamp Connector

## Overview

Reads Basecamp 3 projects, people, and account events through fixed account-bound REST routes.

Readable streams: `projects`, `people`, `events`.

Service API documentation: https://github.com/basecamp/bc3-api.

## Auth setup

Connection fields:

- `access_token` (optional, secret, string); Basecamp bearer access token.
- `account_id` (required, string); Basecamp account ID.
- `client_id` (optional, secret, string); OAuth client ID.
- `client_secret` (optional, secret, string); OAuth client secret.
- `refresh_token` (optional, secret, string); OAuth refresh token.

Authentication uses declared mode(s): `bearer`, `oauth2_refresh_token`.

## Execution contract

Default stream pagination: `link_header`.

Connection check: `GET /projects.json`

## Streams notes

- `projects`: `GET /projects.json`; records `.`
  - Incremental cursor: `updated_at`.
- `people`: `GET /people.json`; records `.`
  - Incremental cursor: `updated_at`.
- `events`: `GET /events.json`; records `.`
  - Incremental cursor: `created_at`.

## Write actions & risks

This connector's write surface is declared separately in the rendered execution bundle.
