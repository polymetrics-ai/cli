# TDD ledger — #4366

| Slice | Red | Green | Refactor | Status |
| --- | --- | --- | --- | --- |
| Closed composition projection | Add named tests for nested/referenced `oneOf`/`anyOf`/`allOf`, nullable scalar union, discriminator, duplicate arms, contradictory `allOf`, and malformed/external/cyclic refs. Current `sourceProjectionSchema` rejects both union keywords and has no `allOf` conversion. | Convert only closed, source-backed schemas and preserve source path/provenance. | Extract common helpers without provider-specific conditionals. | planned |
| Engine validation | Add validator cases for exact-one, at-least-one, and all constraints with a multi-match/no-match/contradiction matrix. Current compiler has only `oneOf`. | Compile and validate all three without open-object fallback. | Keep error messages path-aware and deterministic. | planned |
| Pre-I/O deferred boundary | Add a source-cited deferred operation with a request spy and missing non-composition foundation. | It returns exact `missing_foundation`; spy count remains zero. | Reuse existing deferred admission seam. | planned |
| 608 reconciliation | Add a test against `docs/architecture/batch1-source-operation-mapping-manifest.json` for exact 608 composition records and per-provider totals. | Projected/disposition data covers every row or fails with its stable record key. | Sort deterministically, no map-order assertions. | planned |
| Provider regressions | Start red with retained source fixtures from Bitbucket, CircleCI, GitLab, Jira, Vercel. | Each closed contract is typed or remains deferred with an exact alternative foundation. | No connector-special case in common code. | planned |

## Required red evidence

Before modifying `cmd/` or `internal/`, run the exact new focused test commands and capture their nonzero result in `traces/red-composition.txt`. A green run is not a replacement for red evidence.

## CLI/help parity

The source projection operates on existing connector commands. If it changes a command path, flag, output, or help topic, the ledger must add runtime help, bare namespace, docs/CLI, website, and generated manual evidence. If no user-visible command contract changes, verification records the inspected paths and a not-applicable result.
