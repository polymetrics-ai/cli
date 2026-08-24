# Discussion log — issue 4336 request-contract bounding

The launch brief and the completed scout report settle the product decision:
keep finite resource bounds, but stop treating valid OpenAPI documents as
malformed merely because they omit optional semantic maxima. The authoritative
inventory contains 10,051 deduplicated request units; 8,625 are missing common
string/numeric/array bounds, while only four declarations are malformed.

The concrete release blocker is Gong `GET /v2/all-permission-profiles`: its
required query parameter `workspaceId` has the valid source schema
`{"type":"string"}` and no `maxLength`. The importer currently aborts before
descriptor projection even though the runtime already applies a 4 KiB encoded
query cap.

The chosen implementation is additive provenance plus typed classification:
retain exact source schema, attach the finite PM envelope, preserve genuine
schema/serialization gaps, and enforce the effective cap before provider I/O.
No connector-specific branch, provider allowlist, defaulted provider numeric
range, synthetic source `maxLength`, or truncation is permitted.

The design report is treated as the phase research. No new dependency or
external framework is required.
