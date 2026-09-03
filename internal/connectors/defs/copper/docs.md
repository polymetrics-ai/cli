# Copper Connector

## Overview

Reads Copper CRM records through fixed typed search routes.

Readable streams: `people`, `companies`, `opportunities`, `leads`, `tasks`.

Service API documentation: https://developer.copper.com/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Copper API key.
- `user_email` (required, string); Copper user email.

Authentication uses declared mode(s): `none`.

## Execution contract

Connection check: `POST /people/search`
Check JSON body: `page_number`=1, `page_size`=1.

## Streams notes

- `people`: `POST /people/search`; records `.`
  - JSON body: `page_size`=100.
  - Incremental cursor: `date_modified`.
- `companies`: `POST /companies/search`; records `.`
  - JSON body: `page_size`=100.
  - Incremental cursor: `date_modified`.
- `opportunities`: `POST /opportunities/search`; records `.`
  - JSON body: `page_size`=100.
  - Incremental cursor: `date_modified`.
- `leads`: `POST /leads/search`; records `.`
  - JSON body: `page_size`=100.
  - Incremental cursor: `date_modified`.
- `tasks`: `POST /tasks/search`; records `.`
  - JSON body: `page_size`=100.
  - Incremental cursor: `date_modified`.

## Write actions & risks

This connector's write surface is declared separately in the rendered execution bundle.
