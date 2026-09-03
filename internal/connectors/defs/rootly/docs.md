# Rootly Connector

## Overview

Reads Rootly incidents, services, and users through fixed JSON:API routes.

Readable streams: `incidents`, `services`, `users`.

Service API documentation: https://docs.rootly.com/api-reference.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Rootly API token.
- `start_date` (required, string); Retained initial ETL lower-bound configuration.

Authentication uses declared mode(s): `bearer`.

## Execution contract

Connection check: `GET /v1/incidents`
Check query: `page[size]`=`1`.

## Streams notes

- `incidents`: `GET /v1/incidents`; records `data`
  - Query: `page[number]`=`1`; `page[size]`=`100`.
- `services`: `GET /v1/services`; records `data`
  - Query: `page[number]`=`1`; `page[size]`=`100`.
- `users`: `GET /v1/users`; records `data`
  - Query: `page[number]`=`1`; `page[size]`=`100`.

## Write actions & risks

This connector's write surface is declared separately in the rendered execution bundle.
