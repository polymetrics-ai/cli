# Overview

Generates deterministic sample users, purchases, and products without network access.

This connector discovers available streams and schemas from the configured service at runtime.

This connector is read-only; no write actions are declared.

Service API documentation: https://faker.readthedocs.io/en/master/.

## Auth setup

Connection fields:

- `count` (optional, string); default `1000`; Number of records generated per read for the
  users/purchases streams (products always generates 10). Must be a positive integer; defaults to
  1000.
- `seed` (optional, string); default `0`; Non-negative integer offset added to every generated
  record's sequence number, so different seeds produce different (but still deterministic) ids.
  Defaults to 0; negative values are clamped to 0.

Default configuration values: `count=1000`, `seed=0`.

Authentication is handled by the connector-specific implementation for this service.

## Streams notes

The connector discovers catalogs and records directly from the configured service instead of using
fixed stream declarations.

## Write actions & risks

This connector is read-only; no reverse-ETL write actions are declared.

## Known limits

- Schemas and stream availability depend on the configured service at runtime.
