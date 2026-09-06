# FastBill Connector

## Overview

Reads FastBill billing records through fixed JSON SERVICE envelopes.

Readable streams: `customers`, `invoices`, `products`, `recurring_invoices`, `revenues`.

Service API documentation: https://www.fastbill.com/api/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); FastBill API key.
- `username` (required, string); FastBill username.

Authentication uses declared mode(s): `basic`.

## Execution contract

Default stream pagination: `offset_limit`.

Connection check: `POST `
Check JSON body: `LIMIT`=1, `OFFSET`=0, `SERVICE`=customer.get.

## Streams notes

- `customers`: `POST `; records `RESPONSE.CUSTOMERS`
  - JSON body: `LIMIT`=100, `SERVICE`=customer.get.
- `invoices`: `POST `; records `RESPONSE.INVOICES`
  - JSON body: `LIMIT`=100, `SERVICE`=invoice.get.
- `products`: `POST `; records `RESPONSE.ARTICLES`
  - JSON body: `LIMIT`=100, `SERVICE`=article.get.
- `recurring_invoices`: `POST `; records `RESPONSE.INVOICES`
  - JSON body: `LIMIT`=100, `SERVICE`=recurring.get.
- `revenues`: `POST `; records `RESPONSE.REVENUES`
  - JSON body: `LIMIT`=100, `SERVICE`=revenue.get.

## Write actions & risks

This connector's write surface is declared separately in the rendered execution bundle.
