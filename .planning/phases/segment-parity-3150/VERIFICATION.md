# Verification — Segment parity (#3150-#3157)

| Gate | Status | Evidence |
| --- | --- | --- |
| GSD adapter | pass | `scripts/gsd doctor` passed earlier in session; prompt names loaded for plan/execute; `programming-loop` unavailable so manual fallback recorded. |
| Official source audit | pass | Re-audited Segment Redocly/OpenAPI source: OpenAPI 3.0.3, Segment Public API 73.0.8, 197 operations, SHA-256 `8ec383035b25880b2fc9e8a6cd7d1fc431f968ee308a612d866c6809e609c6a5`. |
| Segment count test | pass | `go test ./cmd/connectorgen -run Segment -count=1` passed. |
| Segment conformance | pass | `go test ./internal/connectors/conformance -run 'TestConformance/segment' -count=1` passed. |
| Connectorgen validation | pass | `go run ./cmd/connectorgen validate internal/connectors/defs` passed with 549 connectors and 0 findings. |
| Focused CLI/docs tests | pass | `go test ./internal/cli -run 'TestGoldenTranscripts|TestGoldenDocsGenerateMatchesTrackedCLIManuals|TestConnector' -count=1` passed. |
| Build | pass | `go build ./cmd/pm` passed. |
| Connector boundary | pass | `make connector-boundary` passed with clean outcome. |
| Full local verify | pass | `make verify` passed after keeping output alive while it ran. |
| Diff whitespace | pass | `git diff --check` passed. |

## CLI/help/docs/website parity checklist

- [x] `pm help docs` read before/docs parity verification.
- [x] `pm connectors inspect segment --json` returns updated streams/actions/command metadata without credentials.
- [x] `pm segment --help` renders Segment command surface and exits 0.
- [x] `pm segment <group> --help` works for representative `streams` and `writes` groups.
- [x] `pm docs generate --dir docs/cli --connectors-dir docs/connectors` run.
- [x] Skill/manual surfaces regenerated through the connector docs generator (`docs/connectors/segment/SKILL.md`, `MANUAL.md`, catalog entries).
- [x] Website docs source impact checked; connector docs are generated from the same registry/bundle metadata, no separate Segment website page exists.
