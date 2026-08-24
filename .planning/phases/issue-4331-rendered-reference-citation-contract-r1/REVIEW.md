# Code review — Issue 4331

The generated `code-review` GSD prompt was executed inline because this direct-PR environment has no compatible isolated Pi reviewer. Manual review covered the changed contract and projection paths.

## Findings

No blocking findings.

- Strict JSON decoding remains enabled. The new wire fields are explicit and limited to the authoritative mainline v3 contract; colliding batch field names are not accepted.
- An absent document `kind` still resolves to `openapi`. Its output fields remain omitted, and the focused snapshot test verifies the OpenAPI-only projection digest while the mixed test compares bytes directly.
- Every non-OpenAPI kind validates captured artifact and published-capture metadata. Rendered references are distinguished from OpenAPI by the captured bytes: source import refuses an actual standalone OpenAPI/Swagger description under `rendered_reference`, while structured JSON/YAML and OpenAPI-path-fragment captures remain valid. Rendered citations are HTTPS, absolute, same-origin provenance only; the code comment and fetch assertions prove they are never fetched.
- Rendered and bundle operations flow through the existing inventory, route, identity, count, deduplication, source-projection provenance, and final import-identity checks. The distinct importer branch only avoids treating declared non-OpenAPI bytes as an OpenAPI parser input. Bundle media types include ZIP plus registered and legacy gzip so captured gzip archives are not rejected as an artificial edge case.
- `unavailable` is source-traced and yields a blocking `source_projection` finding rather than a successful empty inventory.

## Scope note

`bundle` verifies archive integrity and projects its already listed operation inventory. Archive extraction is deliberately deferred: extraction would be a separate parser foundation, while the lock retains the archive hash and operation locations as evidence.
