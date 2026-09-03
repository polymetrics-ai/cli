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

Connection check: `POST /boards/list`
Check JSON body: `apiKey`={{ secrets.api_key }}.

## Streams notes

- `boards`: `POST /boards/list`; records `boards`
  - JSON body: `apiKey`={{ secrets.api_key }}.
  - Incremental cursor: `created`.
- `posts`: `POST /posts/list`; records `posts`
  - JSON body: `apiKey`={{ secrets.api_key }}.
  - Incremental cursor: `created`.
- `comments`: `POST /comments/list`; records `comments`
  - JSON body: `apiKey`={{ secrets.api_key }}.
  - Incremental cursor: `created`.
- `categories`: `POST /categories/list`; records `categories`
  - JSON body: `apiKey`={{ secrets.api_key }}.
  - Incremental cursor: `created`.
- `companies`: `POST /companies/list`; records `companies`
  - JSON body: `apiKey`={{ secrets.api_key }}.
  - Incremental cursor: `created`.

## Write actions & risks

This connector's write surface is declared separately in the rendered execution bundle.
