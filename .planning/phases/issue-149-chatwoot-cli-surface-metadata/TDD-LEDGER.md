# TDD Ledger: Chatwoot CLI Surface Metadata

## 2026-07-10 — planned red validation

Task type: connector metadata and CLI-surface authoring.

Required GSD mode: manual fallback from parent phase because `scripts/gsd prompt programming-loop ...` is unavailable. Parent trace: `.planning/phases/issue-148-chatwoot-cli-parity/traces/gsd-programming-loop-unavailable.txt`.

Required skills: `gsd-core`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-context`, `golang-concurrency`, `golang-documentation`, `golang-lint`.

### Red validation command

```bash
python3 .planning/phases/issue-149-chatwoot-cli-surface-metadata/traces/verify-official-surface-count.py
```

Expected before production edits: fail because official Swagger contains 144 operations but current `api_surface.json` contains 71 entries.

### Green validation commands

```bash
python3 .planning/phases/issue-149-chatwoot-cli-surface-metadata/traces/verify-official-surface-count.py
jq empty internal/connectors/defs/chatwoot/api_surface.json internal/connectors/defs/chatwoot/cli_surface.json
go test ./cmd/connectorgen -run CLISurface -count=1
go test ./internal/connectors/engine -run CLISurface -count=1
go run ./cmd/connectorgen validate internal/connectors/defs/chatwoot
```

## Evidence log

### Red evidence — 2026-07-10

```bash
python3 .planning/phases/issue-149-chatwoot-cli-surface-metadata/traces/verify-official-surface-count.py; echo exit:$?
```

Result: exit 1, as expected before production edits.

Key output:

```json
{
  "official_total": 144,
  "surface_total": 71,
  "missing_count": 73,
  "official_method_counts": {"DELETE": 18, "GET": 62, "PATCH": 21, "POST": 41, "PUT": 2},
  "surface_method_counts": {"DELETE": 6, "GET": 36, "PATCH": 6, "POST": 22, "PUT": 1}
}
```

- Green evidence: pending.
- Refactor evidence: pending.
