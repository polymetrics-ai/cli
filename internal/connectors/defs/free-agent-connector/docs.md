# FreeAgent Connector

## Overview

Reads FreeAgent contacts, invoices, bills, projects, and tasks through fixed FreeAgent v2 REST routes and OAuth2 refresh-token authentication.

Readable streams: `contacts`, `invoices`, `bills`, `projects`, `tasks`.

Service API documentation: https://dev.freeagent.com/.

## Auth setup

Connection fields:

- `client_id` (required, secret, string); FreeAgent OAuth client ID.
- `client_refresh_token_2` (required, secret, string); FreeAgent OAuth refresh token.
- `client_secret` (required, secret, string); FreeAgent OAuth client secret.
- `updated_since` (optional, string); Optional FreeAgent update lower bound.

Authentication uses declared mode(s): `oauth2_refresh_token`.

## Execution contract

Default stream pagination: `page_number`.

Connection check: `GET /contacts`
Check query: `per_page`=`1`.

## Streams notes

- `contacts`: `GET /contacts`; records `contacts`
  - Incremental cursor: `updated_at`.
- `invoices`: `GET /invoices`; records `invoices`
  - Incremental cursor: `updated_at`.
- `bills`: `GET /bills`; records `bills`
  - Incremental cursor: `updated_at`.
- `projects`: `GET /projects`; records `projects`
  - Incremental cursor: `updated_at`.
- `tasks`: `GET /tasks`; records `tasks`
  - Incremental cursor: `updated_at`.

## Write actions & risks

This connector's write surface is declared separately in the rendered execution bundle.
