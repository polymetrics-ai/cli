# Slice 0 code review

## Scope reviewed

- Generic mutation candidate generation and fail-closed classification.
- Generated REST collection fixture provenance and URL-depth ordering.
- Connector-owned GitHub cohort/classification declarations and generator-only mutation artifact.
- Schema/loader validation and generated-artifact tests.

## Result

No unresolved findings.

- Shared Go contains no connector-specific identifier added by this slice; the whole-tree `connectorgen boundary` report is clean with no new allowlist entry.
- A missing family match resolves to the explicit `unassessed` classification; selector validation rejects out-of-cohort typos and multiple matching families error.
- A missing REST collection creator, missing API surface, or shared GraphQL transport is explicit `named_exception` provenance. None may silently borrow a static fixture.
- Generated candidates are compared byte-for-byte with the committed projection; manual override logic permits only exact named commands with `override_reason`.
- The 856-row generator product is no longer embedded in runtime `defs.FS`. The red/green guard keeps exhaustive validation in `cmd/connectorgen` and preserves the existing representative GitHub reverse-plan CLI proof without weakening or excluding it.
- Slice 0 does not call any executor, credential provider, or evidence writer.
