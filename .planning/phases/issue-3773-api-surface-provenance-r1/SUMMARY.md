# Summary — Issue #3773 api_surface v2 provenance

## Delivered

- Added the closed `operation_ledger_version: 2` artifact/provenance schema and typed engine model.
- Added one engine-owned semantic validator shared by conformance, connectorgen, and certification.
- Preserved `covered_by` as the sole stream/write/implemented-direct-read binding; provenance is
  evidence only and cannot make an endpoint or capability executable.
- Kept v1/pre-ledger bundles loadable and certifiable as `legacy_unverified`; no provider bundle
  was migrated in this foundation.
- Added certification JSON/text evidence, exact help behavior, generated CLI docs, and generated
  website docs.

## Evidence

- The focused tests cover complete and malformed v2 artifacts/citations, duplicate IDs, v1
  compatibility, unchanged classifier resolution, connectorgen enforcement, and complete/invalid/
  legacy certification reports.
- The all-bundle gate checked 550 definitions with zero findings and zero surface-sync changes.
- The full scoped local verification list is recorded in `VERIFICATION.md`.

## Handoff

PR #3740 remains open and has not been used as a base. This branch was rebased only onto current
`origin/main` `7d34a0794`, producing implementation commit `6ab5f8bab`, and the focused matrix plus
individual repository gates passed afterward. The separate provider-artifact sweep owns all
`internal/connectors/defs/**` migration work.
