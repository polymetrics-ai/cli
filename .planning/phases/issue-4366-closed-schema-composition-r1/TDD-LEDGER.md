# TDD ledger — #4366

| Slice | Red | Green | Refactor | Status |
| --- | --- | --- | --- | --- |
| Closed composition projection | Red recorded in `traces/red-composition.txt`: existing projection rejects both union keywords and cannot represent root `allOf`. | Green: closes nested local-reference `oneOf` inside `allOf`, retains nullable, discriminated and all-of contracts, and rejects duplicate/non-closed composition. | Extracted shared recursive conversion with no connector conditionals. | green |
| Engine validation | Red recorded: `anyOf`/`allOf` unknown and duplicate `oneOf` compiled. | Green: compiler and request validator enforce exact-one, at-least-one, all constraints and known contradictory intersections. | Deterministic canonical duplicate checks; shared node validation. | green |
| Pre-I/O deferred boundary | Red fixture added for an open composition arm. | Green: open/cyclic source operations retain exact identity and source-cited foundation disposition; malformed external ref is rejected before any provider request; commandrunner validates named closed composition before creating a write plan. | Reused existing typed deferred/preflight seams. | green |
| 608 reconciliation | Immutable, compressed fixture pins the historical authoritative manifest at `d842f739c`. | Green: reconciles 608 rows, exact source/target/CLI identity, all six lanes, provider totals, and one exact regression sample from each Batch 1 provider. | Sorts deterministically, no map-order assertions. | green |
| Provider regressions | Batch 1 fixture carries source-locked representative rows for Bitbucket, CircleCI, GitLab, Jira, and Vercel. | Green: the reconciliation test asserts each representative method/path/composition disposition, every source URL/SHA/location, six-lane classification, and all 608 row identities. The historical provider source directories are absent from this checkout, so the immutable manifest is the available source-cited regression boundary. | No connector-special case in common code. | green |
| Scalar-union provenance disposition | Firstmate inbox `001.msg` identified the old synthetic `TestSourceProjectionStringUnionKeepsTextCLIAndProviderArms` as a provenance-invalid expectation: it asserted Twilio/Xero-style scalar-union behavior without a retained source citation. | Replaced it with the ordinary non-union scalar projection regression. Nullable scalar unions remain covered only by the source-neutral engine schema test required by this issue, not as a provider-promotion claim. | No source-import scalar-union promotion was added for this issue. | green |

## Required red evidence

Before modifying `cmd/` or `internal/`, run the exact new focused test commands and capture their nonzero result in `traces/red-composition.txt`. A green run is not a replacement for red evidence.

## CLI/help parity

The source projection operates on existing connector commands. If it changes a command path, flag, output, or help topic, the ledger must add runtime help, bare namespace, docs/CLI, website, and generated manual evidence. If no user-visible command contract changes, verification records the inspected paths and a not-applicable result.
