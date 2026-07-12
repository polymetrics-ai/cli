# Verification: Freshdesk CLI Parity Parent

Date: 2026-07-09

## Preflight Commands

```bash
scripts/gsd doctor
scripts/gsd verify-pi
scripts/gsd list --json
scripts/gsd prompt plan-phase 1 --skip-research
scripts/gsd prompt programming-loop init --phase issue-172-freshdesk-cli-parity --dry-run
python3 - <<'PY'
import json, pathlib, sys
p=pathlib.Path('internal/connectors/defs/freshdesk/api_surface.json')
data=json.loads(p.read_text())
count=len(data.get('endpoints', []))
if count != 170:
    print(f'freshdesk api_surface endpoints={count}, want 170')
    sys.exit(1)
print('freshdesk api_surface endpoint count matches 170')
PY
test -f internal/connectors/defs/freshdesk/cli_surface.json
```

## Results

- `scripts/gsd doctor`: pass.
- `scripts/gsd verify-pi`: pass.
- `scripts/gsd list --json`: pass; 69 commands found.
- `scripts/gsd prompt plan-phase 1 --skip-research`: pass; generated prompt successfully.
- `scripts/gsd prompt programming-loop init --phase issue-172-freshdesk-cli-parity --dry-run`: blocked; adapter returned `unknown GSD command: programming-loop`. Manual universal-loop fallback is active through `.pi/prompts/pm-gsd-loop.md`.
- Freshdesk endpoint-count red check: expected fail; current count is 10, official baseline target is 170.
- Freshdesk `cli_surface.json` red check: expected fail; file is absent.

## Focused #173 Green Slice

```bash
python3 <json/count validation>
go run ./cmd/connectorgen validate internal/connectors/defs --json
go test ./internal/connectors/engine -run CLISurface
go test ./cmd/connectorgen -run CLISurface
go test ./cmd/connectorgen ./internal/connectors/engine
go test ./internal/connectors/conformance -run 'TestConformance/freshdesk'
```

Results:

- Freshdesk JSON/count validation passed: 170 endpoints (`GET:117`, `POST:10`, `PUT:10`, `DELETE:33`), 5 covered streams, 165 blocked operation rows.
- Full defs validation passed: 547 connectors checked, 0 findings, 0 warnings.
- Focused CLISurface, engine/connectorgen, and Freshdesk conformance tests passed.

## Parent Verification After #173

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go test ./internal/connectors/certify -run TestWritePlanPreviewJSONHasNoApprovalToken -count=1 -timeout 20m
go test ./... -timeout 20m
go build ./cmd/pm
make verify
go run ./cmd/connectorgen validate internal/connectors/defs
```

Results:

- `gofmt -w cmd internal`: passed, no diff after the committed metadata slice.
- `go vet ./...`: passed.
- `go test ./...`: failed on the known/default timeout path: `internal/connectors/certify` exceeded Go's default 10-minute package timeout in `TestWritePlanPreviewJSONHasNoApprovalToken` while full-package execution was still running.
- `go test ./internal/connectors/certify -run TestWritePlanPreviewJSONHasNoApprovalToken -count=1 -timeout 20m`: passed.
- `go test ./... -timeout 20m`: passed.
- `go build ./cmd/pm`: passed.
- `make verify`: passed, including `go test -timeout 20m ./...`, docs validation, smoke, lint, and connectorgen validation.
- `go run ./cmd/connectorgen validate internal/connectors/defs`: passed, 547 connectors checked, 0 findings.

Runtime-backed checks remain optional and require explicit credential/service approval.
