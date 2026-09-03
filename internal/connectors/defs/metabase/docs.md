# Metabase Connector

## Overview

Reads Metabase cards, dashboards, collections, databases, and users through the Metabase REST API using a declared session token.

Readable streams: `cards`, `dashboards`, `collections`, `databases`, `users`.

Service API documentation: https://www.metabase.com/docs/latest/api-documentation.

## Auth setup

Connection fields:

- `instance_api_url` (required, string); Required Metabase instance HTTPS origin, or declared loopback HTTP for local testing.
- `password` (optional, secret, string); Metabase password used only for the fixed declared session exchange when session_token is absent.
- `session_token` (optional, secret, string); Existing Metabase session token sent only as X-Metabase-Session.
- `username` (required, string);

Authentication uses declared mode(s): `api_key_header`, `declared_session`.

## Execution contract

Connection check: `GET /card`

## Streams notes

- `cards`: `GET /card`; records `.`
- `dashboards`: `GET /dashboard`; records `.`
- `collections`: `GET /collection`; records `.`
- `databases`: `GET /database`; records `.`
- `users`: `GET /user`; records `.`

## Write actions & risks

This connector's write surface is declared separately in the rendered execution bundle.
