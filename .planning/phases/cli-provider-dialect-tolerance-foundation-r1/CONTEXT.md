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
| Stripe GET/DELETE identities remain source-cited and unique | live | The restored connector-owned immutable Stripe lock and retained artifact are imported through the production retained-artifact reader. The test asserts the known GET/DELETE IDs, method/path, source URL, and source locations remain in the 589-descriptor result. |
| Stripe reference failure is retained only for its operation | live | The exact 7,967,776-byte immutable retained artifact is imported through the production source resolver. The lock-local depth cap yields one typed response-reference missing-foundation condition for every affected locked operation, with no fabricated response/output contract. |
| Unsafe references remain rejected | fake | Focused hermetic fixtures are necessary to bound malformed/external/cyclic/ambiguous/reference-budget input deterministically. Each asserts a non-nil import error rather than a retained descriptor. |
| Existing source rows and all six lane views remain represented | live | Restored Batch 1 source lock, crosswalk, and disposition artifacts pair every locked Stripe source ID with its source-cited declaration evidence. Existing ledger/evidence generators retain the rows; no Stripe command is added or claimed runnable. |
| A source descriptor gap reaches runtime preflight before I/O | live | A temporary declaration-owned bundle is projected, loaded by the bundle registry, and refused by `commandrunner.Preflight` with the exact `missing_foundation` marker before credentials or an executor are reached. |

## Current evidence and constraints

- `origin/main` is `cf29d302c` after #4358 source-reference projection integration.
- `internal/connectors/defs/stripe/api_surface.json` contains 589 endpoints. The restored source lock, crosswalk, declaration disposition, and canonical descriptor provide the authoritative per-operation source citations across lane views.
- The immutable Batch 1 Stripe evidence is retained in repository history at `c01b852f4`: lock schema v2, 589 operations, source artifact `3653ad45…b2cbdee5`, 7,967,776 bytes, OpenAPI 3.0.0. The independent audit requires it to be restored in this narrowly scoped foundation so the production retained-artifact reader is tested against exact source evidence; no source byte or hash is changed.
- `cmd/connectorgen/sourceimport.go` keeps #4358's `preflightDocument` over all paths/components. A byte-backed v1/v2 lock may declare only a lower `rest.reference_depth_limit`; this foundation preserves malformed/dynamic/resource rejection while making that one typed finite reference-depth condition operation-local for gap-enabled source locks.
- PR #4358 preserves source-reference descriptors where complete byte-backed source contracts are unavailable. It does not alter the source resolver and must not be represented as resolving Stripe reference depth.

## Scope guard

This is shared importer work plus immutable retained Stripe test evidence. It changes neither executable Stripe bundle behavior nor command materialization, credential behavior, source hashes/certificates, deletes, reverse ETL, runtime safety policy, or provider I/O. Its direct-PR review route is frozen after push for Firstmate's independent audit.
