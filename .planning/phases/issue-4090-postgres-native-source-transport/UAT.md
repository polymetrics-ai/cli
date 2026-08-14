---
status: complete
phase: issue-4090-postgres-native-source-transport
source: SUMMARY.md
started: 2026-08-14T00:00:00Z
updated: 2026-08-14T00:00:00Z
mode: autonomous-automated-acceptance
---

# Acceptance UAT — Issue #4090

The Firstmate launch brief explicitly requires autonomous execution without
waiting for a human. These are therefore automated acceptance checks rather
than conversational prompts; each is backed by the required executable proof.

## Current Test

[testing complete]

## Tests

### 1. Definition-selected preflight

expected: PostgreSQL declares the exact `native_database` source and registry
preflight resolves it; missing descriptor, wrong family, and missing
registration all refuse before any source I/O.

result: pass

evidence: `TestPostgresDefinitionDeclaresBoundedSnapshotTransportSource`,
`TestRegisterPostgresSnapshotTransportSourceMakesDefinitionSelectedSourceReachable`,
and `TestPostgresTransportRegistryPreflightRefusesBeforeSourceIO`.

### 2. Bounded full snapshot against PostgreSQL 16.10

expected: `full_append` and `full_overwrite` each read the five seeded rows in
three pages of at most two records, with a valid identity/schema/barrier/dedupe
checkpoint from the registered source.

result: pass

evidence: `TestPostgresDynamicTypedCatalogUsesLiveMetadata`; real output in
`traces/live-source-green.txt`.

### 3. Safety and quality gates

expected: catalog-derived SQL has no caller-authored query surface, all shared
boundaries remain untouched, and focused/race/lint/repository gates succeed.

result: pass

evidence: `REVIEW.md` and `VERIFICATION.md`.

## Summary

total: 3
passed: 3
issues: 0
pending: 0
skipped: 0
blocked: 0

## Gaps

[none]
