# Zoom direct-write operation transport foundation, R1

## Purpose

Customer Managed Keys Hybrid is a customer-hosted Key Connector API, not the ordinary Zoom OAuth
API. Its declared rest_write must bind its own origin and bearer credential together so an
ordinary Zoom OAuth bearer cannot be sent to that customer-hosted origin.

## TDD evidence

- **Red:** 4927731d9 added the JSON-loaded loopback test. It failed because operations.json
  rejected rest.auth as an unknown field:
  Load declared per-operation origin/auth bundle: ... /operations/0/rest/auth: additional property not allowed.
- **Green:** 833a2d9d4 adds paired rest.base_url and rest.auth support for declared rest_write
  operations only. The test passes with zero requests to the ordinary API loopback and one to the
  declared operation origin. It also passed the engine package, connectorgen package, and focused
  go vet checks recorded in the phase TDD ledger.

## Safety contract

- rest.base_url and rest.auth must be present together; either alone is rejected at load and
  direct-write runtime preflight.
- The pair is refused for reads and all non-rest_write operation kinds.
- The preview URL and execution runtime use the same declared operation transport.
- Existing approval, typed confirmation, redaction, no-retry, redirect-refusal, and rate-limit
  paths remain in force.
- No real credential, token, certificate, or key material is recorded.

## Reuse disclosure

The capability is reusable by future declarative customer-hosted rest_write connectors that require
a separate origin plus credential. No other existing connector bundle is changed or claimed as
converted by this foundation; Zoom Customer Managed Keys Hybrid is its first adopter.
