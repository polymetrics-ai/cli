# Lokalise Connector

## Overview

Reads Lokalise project keys, languages, translations, contributors, and comments through fixed API v2 routes.

Readable streams: `keys`, `languages`, `translations`, `contributors`, `comments`.

Service API documentation: https://developers.lokalise.com/reference/api-overview.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Lokalise read-access API token.
- `project_id` (required, string); Lokalise project identifier.

Authentication uses declared mode(s): `api_key_header`.

## Execution contract

Connection check: `GET /projects/{{ config.project_id }}/languages`
Check query: `limit`=`1`.

## Streams notes

- `keys`: `GET /projects/{{ config.project_id }}/keys`; records `keys`
  - Incremental cursor: `modified_at_timestamp`.
- `languages`: `GET /projects/{{ config.project_id }}/languages`; records `languages`
- `translations`: `GET /projects/{{ config.project_id }}/translations`; records `translations`
  - Incremental cursor: `modified_at_timestamp`.
- `contributors`: `GET /projects/{{ config.project_id }}/contributors`; records `contributors`
- `comments`: `GET /projects/{{ config.project_id }}/comments`; records `comments`

## Write actions & risks

This connector's write surface is declared separately in the rendered execution bundle.
