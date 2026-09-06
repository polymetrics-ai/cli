# Feishu / Lark Connector

## Overview

Reads Feishu/Lark Bitable records, tables, and field schemas through declared Bitable REST routes and a bounded tenant token exchange.

Readable streams: `records`, `tables`, `fields`.

Service API documentation: https://open.feishu.cn/document/server-docs/docs/bitable-v1/bitable-overview.

## Auth setup

Connection fields:

- `app_id` (required, secret, string); Feishu/Lark app ID used only for the tenant token exchange.
- `app_secret` (required, secret, string); Feishu/Lark app secret used only for the tenant token exchange.
- `app_token` (required, secret, string); Feishu/Lark Bitable app token.
- `region` (optional, string); Declared Feishu/Lark provider host suffix.
- `table_id` (required, string); Bitable table identifier for records and fields streams.

Authentication uses declared mode(s): `custom`.

## Execution contract

Default stream pagination: `cursor`.

Connection check: `GET /open-apis/bitable/v1/apps/{{ secrets.app_token }}/tables`
Check query: `page_size`=`1`.

## Streams notes

- `records`: `GET /open-apis/bitable/v1/apps/{{ secrets.app_token }}/tables/{{ config.table_id }}/records`; records `data.items`
- `tables`: `GET /open-apis/bitable/v1/apps/{{ secrets.app_token }}/tables`; records `data.items`
- `fields`: `GET /open-apis/bitable/v1/apps/{{ secrets.app_token }}/tables/{{ config.table_id }}/fields`; records `data.items`

## Write actions & risks

This connector's write surface is declared separately in the rendered execution bundle.
