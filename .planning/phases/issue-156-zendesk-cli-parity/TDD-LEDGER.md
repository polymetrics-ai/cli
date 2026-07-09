# TDD Ledger: Zendesk CLI Parity Parent Orchestration

## 2026-07-09 — Parent planning seed

Task type: orchestration/planning.

### Red evidence

Not applicable for the parent seed: no production behavior is changed. The missing parent PR and absent Zendesk bundle are recorded as blockers/next red checks.

### Validation evidence

- `scripts/gsd doctor` — passed.
- `scripts/gsd verify-pi` — passed.
- `scripts/gsd list --json` — passed; output was long and the harness saved the full log to a temp file.
- `scripts/gsd prompt plan-phase issue-156-zendesk-cli-parity --skip-research` — generated the planning prompt.
- `scripts/gsd prompt programming-loop init --phase issue-156-zendesk-cli-parity --dry-run` — failed with `unknown GSD command: programming-loop`; manual GSD fallback is active.
- `scripts/gsd prompt execute-phase issue-157-zendesk-cli-surface-metadata --plan 1` — generated the execution prompt to pair with the manual programming loop.

### Planned red checks for #157 before production edits

- `test -d internal/connectors/defs/zendesk` should fail before the new bundle scaffold exists.
- Add or run a focused Go validation that `engine.Load(defs.FS, "zendesk")` cannot load before metadata exists, then passes after the bundle is embedded.
- Run `go run ./cmd/connectorgen validate internal/connectors/defs` before/after the metadata slice to show the new bundle is structurally valid and does not introduce findings.

### Green target for #157

- Zendesk connector bundle scaffold loads from disk and embedded `defs.FS`.
- `api_surface.json` accounts for the official Zendesk OAS operation baseline without enabling unsafe raw write surfaces.
- `cli_surface.json` command inventory maps provider/API-inspired commands to `etl`, `direct_read`, `reverse_etl`, `docs_only`, or safe unsupported statuses.
- No secrets or secret-shaped fixture/docs literals are introduced.
