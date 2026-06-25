# Observability

This phase does not add runtime telemetry because it does not execute new connector data-plane operations.

Verification visibility comes from:

- Unit test failures when generated catalog entries lack capabilities.
- CLI JSON for deterministic agent inspection.
- Generated manuals for human inspection.
- Phase verification artifacts under `.planning/phases/native-runtime-capability-matrix/`.
