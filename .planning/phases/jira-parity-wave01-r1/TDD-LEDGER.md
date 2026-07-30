# TDD Ledger — Jira Parity Wave 01 R1

## Red / baseline validation

Planned first red gate before production edits:

```bash
python3 - <<'PY'
import json
from pathlib import Path
surface = json.loads(Path('internal/connectors/defs/jira/api_surface.json').read_text())
writes_path = Path('internal/connectors/defs/jira/writes.json')
writes = json.loads(writes_path.read_text())['actions'] if writes_path.exists() else []
assert len(surface['endpoints']) == 616, len(surface['endpoints'])
assert len(writes) == 296, len(writes)
PY
```

Actual current failure captured before production edits: `red_baseline_exit=1` because Jira has only the old partial surface and no reverse-write file.

## Green

- Generated connector-local Jira operation ledger from official Atlassian OpenAPI v3.
- Official source hash recorded: `sha256:8439da27e1b2dd7b013a0ae721b8aeaa7746bc8e2d816fa28aa1a582e8597501`.
- Green invariant passed: 616 surface rows, 269 executable write actions, 27 reverse-ETL shared-foundation blockers, 84 DELETE actions all typed `destructive`, 101 destructive-confirmed writes total.
- Generated bounded direct-read command metadata and operation specs for JSON GET/POST read/search operations, with unsupported typed-body command shapes marked partial and blocked in the operation ledger.
- Recorded binary/raw-body, required-header, JSON Patch media-type, repeated form-field, and required raw/dynamic JSON body shared executor gaps as blocked ledger/operation rows rather than fake execution.
- Existing sanitized read fixtures still pass; representative write fixtures added for `create_issue`, `delete_issue`, and `remove_attachment`.

## Verification targets

- `python3` ledger invariant checks for source hash, operation counts, endpoint uniqueness, and action counts.
- `jq empty internal/connectors/defs/jira/*.json internal/connectors/defs/jira/schemas/*.json` where files exist.
- `go run ./cmd/connectorgen validate internal/connectors/defs`
- `go test ./internal/connectors/conformance -run 'TestConformance/jira' -count=1`
- `go test ./internal/cli -run 'Connector|Dynamic|Golden|Jira' -count=1` if package filters build locally.
- `go vet ./internal/connectors/... ./internal/cli/...`
- `go build ./cmd/pm`
- `make connector-boundary`
- `git diff --check`

## Refactor notes

- No shared runtime code changes.
- No dependency additions.
- No live provider calls or credentialed checks.
