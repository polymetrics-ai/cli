# Verification: Issue #137 HubSpot Operation Ledger

Date: 2026-07-09

## Adapter checks

Passed in this session:

```bash
scripts/gsd doctor
scripts/gsd verify-pi
scripts/gsd list --json
scripts/gsd prompt plan-phase issue-137-hubspot-operation-ledger --skip-research
```

`programming-loop` prompt is unavailable in the pinned registry; manual GSD fallback recorded in `PLAN.md` and `TDD-LEDGER.md`.

## Targeted checks

Pending:

```bash
gofmt -w cmd internal
go test ./cmd/connectorgen -run 'HubSpot|APISurfaceOperationLedger' -count=1
go run ./cmd/connectorgen validate internal/connectors/defs
python3 -m json.tool internal/connectors/defs/hubspot/api_surface.json >/dev/null
go run ./cmd/pm help connectors
go run ./cmd/pm connectors inspect hubspot --json
```

## Broad checks

Pending:

```bash
go vet ./...
go test ./...
go build ./cmd/pm
make verify
go run ./cmd/connectorgen validate internal/connectors/defs
```

## CLI help/docs/website parity

- [ ] `pm help connectors` checked.
- [ ] `pm connectors inspect hubspot --json` checked; manifest inspection is credential-free.
- [ ] `docs/connectors/hubspot/**` updated if generated manual/skill output changes.
- [ ] Connector catalog/docs validation updated or marked not applicable.
- [ ] Website docs explicitly marked not applicable for this ledger-only slice unless touched.

## Review route

Stacked PR review coverage remains governed by parent issue #132. For non-default-base sub-PRs, record CodeRabbit skip/manual/fallback status and use parent PR fallback when needed; do not treat a skipped CodeRabbit status as approval.
