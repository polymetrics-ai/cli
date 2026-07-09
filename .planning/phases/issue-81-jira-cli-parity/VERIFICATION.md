# Verification: Jira CLI Parity Parent

## Preflight Completed

```bash
scripts/gsd doctor
scripts/gsd verify-pi
scripts/gsd list --json
scripts/gsd prompt plan-phase issue-81-jira-cli-parity --skip-research
scripts/gsd prompt programming-loop init --phase issue-81-jira-cli-parity --dry-run
```

Results:

- `doctor`: pass.
- `verify-pi`: pass.
- `list --json`: ran; harness truncated display due output size.
- `plan-phase`: generated prompt successfully.
- `programming-loop`: blocked (`unknown GSD command: programming-loop`); manual GSD fallback recorded in `PLAN.md` and `TDD-LEDGER.md`.

## Baseline Read-Only Checks

```bash
go run ./cmd/pm help connectors
go run ./cmd/pm connectors inspect jira --json
```

Results:

- Help rendered successfully.
- Jira metadata inspection succeeded without credential access.

## Required Parent Handoff Gates

Not yet run for the full parent branch:

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
go run ./cmd/connectorgen validate internal/connectors/defs
```

## Issue #104 Targeted Gates

Planned after red test and implementation:

```bash
go test ./internal/connectors/engine -run TestBundleLoadEmbeddedJiraCLISurface -count=1
go test ./internal/connectors/engine -run CLISurface -count=1
go test ./cmd/connectorgen -run CLISurface -count=1
go run ./cmd/connectorgen validate internal/connectors/defs --json
```

## CLI Help / Docs / Website Parity

Applies to future CLI runtime/help/docs slices. For #104 metadata-only work:

- Runtime help checked: `go run ./cmd/pm help connectors` (baseline only).
- Connector inspection checked: `go run ./cmd/pm connectors inspect jira --json`.
- Bare namespace behavior: pending for help-renderer/runtime slices.
- `docs/cli/**`: pending/not changed in #104 unless generated metadata consumers require docs updates.
- `website/**`: pending/not changed in #104 unless generated metadata consumers require docs updates.
- Generated help/manual artifacts: pending/not changed in #104.
