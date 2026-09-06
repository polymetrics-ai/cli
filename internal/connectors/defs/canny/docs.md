# Canny Connector

## Overview

Reads Canny boards, posts, comments, categories, and companies through fixed Canny REST form requests.

Readable streams: `boards`, `posts`, `comments`, `categories`, `companies`.

Service API documentation: https://developers.canny.io/api-reference.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Canny API key.

Authentication uses declared mode(s): `none`.

## Execution contract

Default stream pagination: `offset_limit`.

Connection check: `POST /boards/list`
Check Form body: `apiKey`={{ secrets.api_key }}.

## Streams notes

- `boards`: `POST /boards/list`; records `boards`
  - Form body: `apiKey`={{ secrets.api_key }}.
  - Pagination: `offset_limit`. Form `skip`/`limit` windows stop at `hasMore=false`.
- `posts`: `POST /posts/list`; records `posts`
  - Form body: `apiKey`={{ secrets.api_key }}.
  - Pagination: `offset_limit`. Form `skip`/`limit` windows stop at `hasMore=false`.
- `comments`: `POST /comments/list`; records `comments`
  - Form body: `apiKey`={{ secrets.api_key }}.
  - Pagination: `offset_limit`. Form `skip`/`limit` windows stop at `hasMore=false`.
- `categories`: `POST /categories/list`; records `categories`
  - Form body: `apiKey`={{ secrets.api_key }}.
  - Pagination: `offset_limit`. Form `skip`/`limit` windows stop at `hasMore=false`.
- `companies`: `POST /companies/list`; records `companies`
  - Form body: `apiKey`={{ secrets.api_key }}.
  - Pagination: `offset_limit`. Form `skip`/`limit` windows stop at `hasMore=false`.

## Write actions & risks

This connector's write surface is declared separately in the rendered execution bundle.
