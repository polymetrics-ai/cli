# Verification — Issue #204 Crisp all-ops executable coverage

## Completed commands

Red direct-read policy test:

```bash
go test ./internal/connectors/commandrunner -run TestRunDirectReadSupportsGenericJSONResponsePolicy -count=1
```

Result: failed as expected before implementation because `json_response` was not a supported output policy.

Targeted green:

```bash
go test ./internal/connectors/commandrunner ./internal/connectors/engine -count=1
go test ./cmd/connectorgen -count=1
go run ./cmd/connectorgen validate internal/connectors/defs
go test ./internal/connectors/conformance -run 'TestConformance/crisp' -count=1
./pm docs validate --connectors-dir docs/connectors
./pm connectors inspect crisp --json
./pm help connectors
./pm connectors --help
```

Result: pass. Connector inspection confirmed `read=true`, `write=true`, and 129 write actions without reading credentials.

All-ops inventory check:

```bash
python3 - <<'PY'
import json, collections
api=json.load(open('internal/connectors/defs/crisp/api_surface.json'))
cli=json.load(open('internal/connectors/defs/crisp/cli_surface.json'))
writes=json.load(open('internal/connectors/defs/crisp/writes.json'))
print(len(api['endpoints']), collections.Counter(tuple(e['covered_by'].keys())[0] for e in api['endpoints']))
print(len(cli['commands']), collections.Counter((c['intent'], c['availability']) for c in cli['commands']))
print(len(writes['actions']), collections.Counter(a['method'] for a in writes['actions']))
PY
```

Result: 220 endpoints covered; 91 implemented direct reads; 129 implemented reverse-ETL write actions.

Full local gate:

```bash
make verify
```

Result: pass.

Post-verify documentation whitespace check:

```bash
git diff --check
./pm docs validate --connectors-dir docs/connectors
go run ./cmd/connectorgen validate internal/connectors/defs
```

Result: pass.

## Current status

All required local verification for this slice passed. No credentialed Crisp API checks were run.
