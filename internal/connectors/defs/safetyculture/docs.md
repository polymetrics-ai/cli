# SafetyCulture Connector

## Overview

Reads SafetyCulture audits, templates, and users through fixed REST routes.

Readable streams: `audits`, `templates`, `users`.

Service API documentation: https://developers.safetyculture.com/.

## Auth setup

Connection fields:

- `access_token` (required, secret, string);
- `start_date` (required, string);

Authentication uses declared mode(s): `bearer`.

## Execution contract

Connection check: `GET /audits`
Check query: `page`=`1`; `page_size`=`1`.

## Streams notes

- `audits`: `GET /audits`; records `audits`
  - Query: `page`=`1`; `page_size`=`100`.
- `templates`: `GET /templates`; records `templates`
  - Query: `page`=`1`; `page_size`=`100`.
- `users`: `GET /users`; records `users`
  - Query: `page`=`1`; `page_size`=`100`.

## Write actions & risks

This connector's write surface is declared separately in the rendered execution bundle.
