# TDD Ledger: Freshdesk Full Operation Implementation

Sub-issue: #176
Parent issue: #172

## Red Evidence

```bash
python3 - <<'PY'
import json, sys
api=json.load(open('internal/connectors/defs/freshdesk/api_surface.json'))
blocked=sum(1 for ep in api['endpoints'] if 'operation' in ep)
covered=sum(1 for ep in api['endpoints'] if 'covered_by' in ep)
print(f'freshdesk implemented coverage={covered}, blocked={blocked}, total={len(api["endpoints"])}')
if blocked:
    sys.exit(1)
PY
```

Result: `freshdesk implemented coverage=5, blocked=165, total=170`; exit 1.

Planned red tests before shared-code edits:

- `commandrunner` rejects `output_policy: "json"` today for implemented direct reads.
- `engine.DirectRead` rejects `output_policy: "json"` today.
- `connectorgen validate` rejects Freshdesk command entries using `output_policy: "json"` until the schema/validator support lands.

## Green Evidence

Generic bounded JSON direct-read policy red→green:

```bash
go test ./internal/connectors/commandrunner -run 'TestRunImplementedDirectReadCommandAllowsGenericJSONPolicy' -count=1
go test ./internal/connectors/engine -run 'TestDirectReadJSONPolicyPreservesJSONBody' -count=1
go test ./cmd/connectorgen -run 'TestValidate_CLISurfaceImplementedDirectReadGenericJSONOutputPolicyPasses' -count=1
```

Result: all passed after adding `output_policy: "json"` support to command runner, engine direct reads, connectorgen validation, and CLI surface schema.

Freshdesk operation conversion green checks:

```bash
go test ./cmd/connectorgen -run Freshdesk -count=1
go run ./cmd/connectorgen validate internal/connectors/defs --json
go test ./internal/connectors/conformance -run 'TestConformance/freshdesk' -count=1
go test ./internal/connectors/commandrunner -run DirectRead -count=1
go test ./internal/connectors/engine -run DirectRead -count=1
go test ./cmd/connectorgen -run 'CLISurface|Freshdesk' -count=1
go test ./cmd/connectorgen ./internal/connectors/engine ./internal/connectors/commandrunner
/tmp/pm docs validate --connectors-dir docs/connectors
```

Result: all passed. Freshdesk coverage is now 168 executable endpoint rows and 2 blocked-safe operation rows (`POST /contacts/imports` CSV multipart upload, custom-object dynamic query filter). Command surface has 5 ETL streams, 109 bounded JSON direct reads, and 50 reverse-ETL write commands/actions.

Broader gates also passed: `go vet ./...`, `go test ./... -timeout 20m`, `go build ./cmd/pm`, and `make verify`.

## Refactor Notes

- Do not mark a write implemented unless it has a named write action and reverse-ETL gates.
- Do not add raw `payload` body escape hatches or generic HTTP command execution.
- Two endpoints remain blocked because implementing them safely requires new typed file-upload / dynamic-query policies rather than raw payload/query escape hatches.
