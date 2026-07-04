# Overview

Uptick is a Tier-2 AuthHook connector for Uptick field service management tenants. It reads the
legacy five streams from `internal/connectors/uptick` (tasks, clients, properties, invoices, and
assets) plus the additional statically known Uptick REST list resources named in the existing bundle
surface: quotes, purchase orders, forms, users, teams, and stock items. The API is per tenant and the
public Uptick docs describe the v2 API as REST plus JsonAPI, with `/api/v2.x/` exposing the tenant's
endpoint list and each endpoint exposing `OPTIONS` metadata for available fields.

## Auth setup

Provide `base_url` for the tenant, `username`, and secrets `client_id`, `client_secret`, and
`password`. The custom `uptick` AuthHook exchanges those values at
`{base_url}/api/oauth2/token/` using the OAuth2 password grant and attaches the resulting bearer
token to API requests. The hook mirrors the legacy connector's password-grant path and keeps the
same validated per-tenant base URL behavior.

## Streams notes

All streams read `/api/v2.14/<resource>/`, use `links.next` next-url pagination, and extract records
from `data`. The legacy streams keep their exact sparse `fields[...]` lists, `ordering=-updated`,
and `updatedsince` incremental filter to preserve legacy record data and cursor behavior. The Pass B
streams use passthrough projection with permissive schemas because the complete field list is
tenant-discovered through live `OPTIONS` responses rather than static public documentation.

## Write actions & risks

None. `capabilities.write` remains `false` and there is no `writes.json`. Uptick's public docs make
write request bodies endpoint- and tenant-specific through `OPTIONS` metadata, and the legacy Go
connector is read-only. Declaring generic POST/PUT/PATCH/DELETE writes without those concrete
schemas would create an unsafe broad write surface.

## Known limits

- Dynamic conformance remains skipped because this bundle's only auth mode is the custom Uptick
  OAuth2 password-grant hook, which cannot be exercised with the conformance harness's synthetic
  credentials.
- No config-driven `max_pages` runtime override is declared. The engine's next-url paginator still
  terminates on absent `links.next` and protects against repeated next URLs.
- Tenant-specific endpoints, fields, and write bodies exposed only through live `OPTIONS` metadata
  are not statically materialized in this JSON bundle.
