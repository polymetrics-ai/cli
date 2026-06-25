# PLAN: Agentic ETL Platform

## Tasks

- [x] Record baseline verification and prompt snapshot.
- [x] Add red tests for structured errors and JSON error contracts.
- [x] Implement typed CLI errors and sanitized stderr.
- [x] Add red tests for validators and terminal sanitizer.
- [x] Implement shared validation/sanitization package.
- [x] Add red tests for connector manifests and secret redaction.
- [x] Implement connector manifests and manifest-backed inspection.
- [x] Add red tests for generated skills.
- [x] Implement `poly skills generate`.
- [x] Add red tests for streaming ETL batches.
- [x] Implement bounded ETL batch writes and checkpoint metadata.
- [x] Run full local verification and update phase artifacts.

## Out Of Scope For This Slice

- Full Temporal workflow migration.
- OS keychain integration requiring a new dependency.
- Live GitHub benchmark using real tokens.
- Real external reverse ETL writes.
