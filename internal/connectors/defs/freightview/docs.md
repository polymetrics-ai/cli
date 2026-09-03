# Freightview Connector

## Overview

Reads Freightview shipments, quotes, and tracking events through fixed Freightview v2.0 REST routes using client-credentials authentication.

Readable streams: `shipments`, `quotes`, `tracking`.

Service API documentation: https://docs.freightview.com/.

## Auth setup

Connection fields:

- `client_id` (required, secret, string); Freightview client ID.
- `client_secret` (required, secret, string); Freightview client secret.

Authentication uses declared mode(s): `oauth2_client_credentials`.

## Execution contract

Connection check: `GET /shipments`
Check query: `limit`=`1`.

## Streams notes

- `shipments`: `GET /shipments`; records `shipments`
- `quotes`: `GET /shipments/{{ fanout.id }}/quotes`; records `quotes`
- `tracking`: `GET /shipments/{{ fanout.id }}/tracking`; records `.`

## Write actions & risks

This connector's write surface is declared separately in the rendered execution bundle.
