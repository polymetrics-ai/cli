# Stripe provider-dialect tolerance foundation — observed context

## Task Delivery Header

- Issue: Refs #4336 — tolerate bounded provider OpenAPI dialect gaps.
- Base branch: `main`.
- Merges into: `main`.
- Delivery: A direct PR is open against `main`, with the required local and CI checks recorded; an independent exact-head audit remains human-coordinated.
- Working branch: `fm/cli-stripe-provider-dialect-tolerance-r1`.
- Task: Make reference-depth handling operation-local for verified OpenAPI source import, retaining a source-cited descriptor and concrete missing-foundation gap when one operation cannot be represented, while continuing to reject unsafe reference inputs. Do not materialize a Stripe command or provider-I/O surface.
- Verification: Focused red/green source-import tests based on retained Stripe evidence; changed-package tests; generator/source-import, declaration-admission, operation-evidence, surface-sync, docs, lint, build, and connector-boundary gates.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| A normal Stripe GET/DELETE identity remains a complete source descriptor | fake | The public source is not fetched. A hermetic reduced fixture retains the exact lock's source URL, source IDs, operation IDs, methods, paths, and source locations; the importer must emit both complete descriptors. |
| A nested Stripe reference failure is retained only for its operation | fake | The retained artifact is not present on current `main`; a reduced fixture preserves the exact Stripe operation citation and an OpenAPI-local nested `$ref` chain. The test asserts the unaffected operation is complete and the affected operation has the precise source-descriptor foundation gap. |
| Unsafe references remain rejected | fake | Focused hermetic fixtures are necessary to bound malformed/external/cyclic/ambiguous/reference-budget input deterministically. Each asserts a non-nil import error rather than a retained descriptor. |
| Existing source rows and all six lane views remain represented | live | Existing ledger/evidence generators read the checked-in Stripe source-cited surface and operation evidence. The checks must retain each source ID/row; no Stripe command is added or claimed runnable. |

## Current evidence and constraints

- `origin/main` is `7cd0412ae388ad10342e9c1153260c6e787e5757` after integrating Firstmate's required #4360 foundation baseline.
- `internal/connectors/defs/stripe/api_surface.json` contains 589 endpoints, 581 of which have `operation.source_url` citations.
- The previously retained exact Stripe lock and artifact are present in repository history at `47c606453`: lock schema v2, 589 operations, source artifact `3653ad45…b2cbdee5`, 7,967,776 bytes, OpenAPI 3.0.0. Current `main` intentionally does not carry that source directory, so this foundation does not restore or materialize it.
- `cmd/connectorgen/sourceimport.go` currently sets a global reference depth of 64 and invokes `preflightDocument` over all paths/components before importing any operation. The preflight is what turns a one-operation reference traversal into a connector-wide abort.
- PR #4358 preserves source-reference descriptors where complete byte-backed source contracts are unavailable. It does not alter the source resolver and must not be represented as resolving Stripe reference depth.

## Scope guard

This is shared importer work only. It changes neither Stripe bundle definitions nor command materialization, credential behavior, hashes, certificates, deletes, reverse ETL, runtime safety policy, or provider I/O. Its direct-PR review route is frozen after push for Firstmate's independent audit.
