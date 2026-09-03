# Apify Dataset Connector

## Overview

Reads Apify dataset items and dataset metadata through fixed Apify API v2 routes.

Readable streams: `item_collection`, `dataset_collection`, `dataset`.

Service API documentation: https://docs.apify.com/api/v2.

## Auth setup

Connection fields:

- `dataset_id` (required, string); Apify dataset identifier.
- `token` (required, secret, string); Apify API token.

Authentication uses declared mode(s): `bearer`.

## Execution contract

Connection check: `GET /datasets`
Check query: `limit`=`1`.

## Streams notes

- `item_collection`: `GET /datasets/{{ config.dataset_id }}/items`; records `.`
- `dataset_collection`: `GET /datasets`; records `data.items`
  - Incremental cursor: `createdAt`.
- `dataset`: `GET /datasets/{{ config.dataset_id }}`; records `data`
  - Incremental cursor: `modifiedAt`.

## Write actions & risks

This connector's write surface is declared separately in the rendered execution bundle.
