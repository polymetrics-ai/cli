# Code review — Zoom ETL certification parity

## Scope

- `internal/connectors/defs/zoom/command_surface_test.go`
- `internal/connectors/defs/zoom/certification-sweep.json`
- Wave #4266 planning and verification artifacts

## Manual review result

No critical, warning, or informational finding.

The new test parses only the committed generated artifact, asserts its connector identity and declared cardinality, verifies every existing ETL command remains `implemented` with `fixture_required` evidence, and ensures the read-capability projection remains a pending fixture row. It neither resolves credentials nor invokes the Zoom provider. The generated JSON passes its derivation check. All production-path changes are inside `internal/connectors/defs/zoom/`; no auth, engine, certification allowlist, or certification-status file changed.

## Review fallback

`scripts/gsd prompt code-review zoom-etl-certification-parity-r1` was resolved. The canonical contract forbids the reviewer-role spawn expected by the official workflow, so this file records the required inline/manual review and its disposition.
