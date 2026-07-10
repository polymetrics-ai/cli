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

Pending.

## Refactor Notes

- Do not mark a write implemented unless it has a named write action and reverse-ETL gates.
- Do not add raw `payload` body escape hatches or generic HTTP command execution.
