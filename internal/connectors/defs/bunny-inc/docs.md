# Bunny, Inc. Connector

## Overview

Reads Bunny subscription-billing data through declared per-tenant GraphQL connection routes.

Readable streams: `accounts`, `contacts`, `invoices`, `payments`, `subscriptions`.

Service API documentation: https://docs.bunny.com/.

## Auth setup

Connection fields:

- `apikey` (required, secret, string); Bunny GraphQL API key.
- `subdomain` (required, string); Bunny tenant subdomain.

Authentication uses declared mode(s): `bearer`.

## Execution contract

Default stream pagination: `cursor`.

Connection check: `POST /graphql`
Check JSON body: `query`=query { accounts(first: 1) { nodes { id } } }.

## Streams notes

- `accounts`: `POST /graphql`; records `data.accounts.nodes`
  - Incremental cursor: `updatedAt`.
- `contacts`: `POST /graphql`; records `data.contacts.nodes`
  - Incremental cursor: `updatedAt`.
- `invoices`: `POST /graphql`; records `data.invoices.nodes`
  - Incremental cursor: `updatedAt`.
- `payments`: `POST /graphql`; records `data.payments.nodes`
  - Incremental cursor: `updatedAt`.
- `subscriptions`: `POST /graphql`; records `data.subscriptions.nodes`
  - Incremental cursor: `updatedAt`.

## Write actions & risks

This connector's write surface is declared separately in the rendered execution bundle.
