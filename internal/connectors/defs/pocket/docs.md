# Pocket Connector

## Overview

Reads saved Pocket items through the fixed v3 retrieve API.

Readable streams: `items`.

Service API documentation: https://getpocket.com/developer/docs/v3/retrieve.

## Auth setup

Connection fields:

- `access_token` (required, secret, string);
- `consumer_key` (required, secret, string);
- `contentType` (optional, string);
- `detail_type` (optional, string);
- `favorite` (optional, string);
- `since` (optional, string);
- `sort` (optional, string);
- `state` (optional, string);
- `tag` (optional, string);

Authentication uses declared mode(s): `none`.

## Execution contract

Connection check: `POST /get`
Check JSON body: `access_token`={{ secrets.access_token }}, `consumer_key`={{ secrets.consumer_key }}, `count`=1, `detailType`={{ config.detail_type }}, `offset`=0.

## Streams notes

- `items`: `POST /get`; records `list`
  - JSON body: `access_token`={{ secrets.access_token }}, `consumer_key`={{ secrets.consumer_key }}, `count`=100, `detailType`={{ config.detail_type }}, `offset`=0.

## Write actions & risks

This connector's write surface is declared separately in the rendered execution bundle.
