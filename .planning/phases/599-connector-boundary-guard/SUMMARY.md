# Summary — Issue 599 Connector Boundary Guard

## Current status

Implementation is complete on branch `fm/cli-connector-boundary-guard-r1`, but branch-name verification is blocked until a supported migration preserves the active PR lineage on a conventional `<type>/<description>` branch.

## What changed

- Added stdlib-only `internal/connectors/boundary` scanner with deterministic path classification and AST-aware shared production Go scanning.
- Added `connectorgen boundary [repo-root] [--json] [--base <ref>]` with stable outcomes: clean, policy violations, invalid invocation/configuration.
- Built the connector lexicon from `internal/connectors/defs/*/metadata.json`, display names, and source/destination aliases rather than a hand-maintained provider list.
- Added exact exception ledger at `internal/connectors/boundary/exceptions.json` for current baseline GitHub/Bahmni residue; each row binds rule, connector, path, match, reason, public migration issue URL, owner, expiry, and max matches.
- Added synthetic tests for shared provider literals, provider policy helper placement, allowed definitions/native/hooks/generated/test/docs paths, exception stale/expired/broadened behavior, stable sorting, JSON shape, base-diff mode, CLI exit behavior, and current-main baseline.
- Added `make connector-boundary` and a standalone always-present GitHub Actions workflow check named `connector-boundary` with read-only permissions and no path filters.
- Added focused developer runbook at `docs/migration/connector-boundary-guard.md`.
- Confirmed the repository convention workflow has no `fm/` branch-family bypass; the active legacy branch requires supported migration before the `branch-name` check can be reported green.

## Verification

- `go test ./internal/connectors/boundary ./cmd/connectorgen` — pass.
- `go run ./cmd/connectorgen boundary . --json` — pass (`findings=0`, `exceptions=24`, `gong_exceptions=0`).
- `go run ./cmd/connectorgen validate internal/connectors/defs --json` — pass (`connectors_checked=548`, `findings=0`, `warnings=0`).
- `make connector-boundary` — pass.
- Extracted `conventions.yml` branch-name run block — blocked for `fm/cli-connector-boundary-guard-r1`; supported migration to a conventional `<type>/<description>` branch is required.
- `make verify` — pass.
- `git diff --check` — pass.

## GSD fallback note

The repo-local GSD adapter is healthy, but the advertised `programming-loop` prompt command is absent from the command registry in this checkout. The implementation followed the manual GSD universal loop with planning, TDD ledger, verification checklist, and run-state artifacts.

## Review notes

- Gong behavior from the completed Gong work remains definition-owned; no shared-code exception exists for Gong.
- This guard-only PR does not migrate GitHub/Twenty/WhatsApp behavior, does not update existing connector issue/PR matrices, and does not change branch protection.
