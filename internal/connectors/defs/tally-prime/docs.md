# Overview

Reads TallyPrime accounting data (companies, ledgers, groups, stock items, vouchers) via TDL
Export/Collection envelope requests POSTed to a locally-running TallyPrime Gateway Server. Read-only
source; schema is discovered dynamically since TallyPrime has no static REST resource surface.

This connector discovers available streams and schemas from the configured service at runtime.

This connector is read-only; no write actions are declared.

Service API documentation: https://help.tallysolutions.com/tally-prime-integration-using-json-1/.

## Auth setup

Connection fields:

- `company` (required, string); Exact TallyPrime company name to scope every request to (sent as
  SVCURRENTCOMPANY in STATICVARIABLES). TallyPrime resolves masters/vouchers relative to the active
  company context, so this is required.
- `envelope_format` (optional, string); default `json`; allowed values `json`, `xml`; Envelope
  body/response encoding. json selects TallyPrime 7.0+'s native JSON export mode
  (SVEXPORTFORMAT=$$SysName:UTF8JSON); xml is the universally-supported fallback
  (SVEXPORTFORMAT=$$SysName:XML) for TallyPrime releases before native JSON export shipped.
- `from_date` (optional, string); Optional SVFROMDATE (YYYYMMDD) lower bound applied to the vouchers
  stream. Ignored by the master streams (companies/ledgers/groups/stock_items), which are always
  exported in full.
- `gateway_url` (required, string); default `http://localhost:9000`; format `uri`; Base URL of the
  locally-running TallyPrime Gateway Server that accepts TDL Export/Collection envelope requests
  over HTTP POST. TallyPrime listens on port 9000 by default; this is never a public internet
  address - the gateway is a loopback/LAN-only service, matching TallyPrime's own security model.
- `http_timeout_seconds` (optional, string); default `30`; Timeout in seconds for each envelope POST
  to the local Gateway Server.
- `mode` (optional, string); allowed values `fixture`.
- `to_date` (optional, string); Optional SVTODATE (YYYYMMDD) upper bound applied to the vouchers
  stream. Ignored by the master streams.

Default configuration values: `envelope_format=json`, `gateway_url=http://localhost:9000`,
`http_timeout_seconds=30`.

Authentication is handled by the connector-specific implementation for this service.

## Streams notes

The connector discovers catalogs and records directly from the configured service instead of using
fixed stream declarations.

## Write actions & risks

This connector is read-only. Read behavior: local TallyPrime Gateway Server read of accounting
masters (companies/ledgers/groups/stock items) and transactional vouchers via TDL Export/Collection
envelopes over HTTP POST to a locally-running TallyPrime instance (no public network egress).

## Known limits

- Schemas and stream availability depend on the configured service at runtime.
