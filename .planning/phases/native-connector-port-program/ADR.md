# ADR: Native Ports Are Enabled By Conformance, Not By Catalog Presence

## Decision

Expose native port plans for every catalog connector, but keep data-plane operations disabled until a native Go implementation passes conformance tests.

## Rationale

The connector catalog is discovery metadata. A production native connector must prove check, catalog, read/write, state, secret redaction, pagination, rate limiting, and retry behavior before agents can use it.

## Consequences

- Agents can plan connector work across the full catalog.
- Humans can inspect CDC and reverse ETL requirements before implementation.
- The runtime remains safe because planned connectors are not registered as executable connectors.
