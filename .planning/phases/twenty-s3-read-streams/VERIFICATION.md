# Twenty S3 read streams — verification

Issue: #280
PR: #290
Branch: `feat/280-twenty-streams`

## Pre-production red validation

```bash
python3 - <<'PY'
import json
from pathlib import Path
d=json.loads(Path('internal/connectors/defs/twenty/api_surface.json').read_text())
print('endpoints', len(d['endpoints']))
print('covered_stream', sum(1 for e in d['endpoints'] if e.get('covered_by', {}).get('stream')))
print('excluded', sum(1 for e in d['endpoints'] if 'excluded' in e))
print('direct_read', sum(1 for e in d['endpoints'] if e.get('covered_by', {}).get('direct_read')))
assert len(d['endpoints']) == 56
assert sum(1 for e in d['endpoints'] if e.get('covered_by', {}).get('stream')) == 56
assert not any('excluded' in e for e in d['endpoints'])
assert not any(e.get('covered_by', {}).get('direct_read') for e in d['endpoints'])
PY
```

Result: fail as expected before production edit.

```text
endpoints 56
covered_stream 28
excluded 28
direct_read 0
Traceback (most recent call last):
  File "<stdin>", line 9, in <module>
AssertionError

Command exited with code 1
```

```bash
test ! -f internal/connectors/engine/twenty_bundle_test.go
```

Result: fail as expected before production edit (file exists).

```text
Command exited with code 1
```

## Green gates

```bash
jq . internal/connectors/defs/twenty/api_surface.json internal/connectors/defs/twenty/streams.json >/tmp/twenty-jq.out && wc -l /tmp/twenty-jq.out
```

Result: pass.

```text
     834 /tmp/twenty-jq.out
```

```bash
python3 - <<'PY'
import json
from pathlib import Path
d=json.loads(Path('internal/connectors/defs/twenty/api_surface.json').read_text())
print('endpoints', len(d['endpoints']))
print('covered_stream', sum(1 for e in d['endpoints'] if e.get('covered_by', {}).get('stream')))
print('excluded', sum(1 for e in d['endpoints'] if 'excluded' in e))
print('direct_read', sum(1 for e in d['endpoints'] if e.get('covered_by', {}).get('direct_read')))
assert len(d['endpoints']) == 56
assert sum(1 for e in d['endpoints'] if e.get('covered_by', {}).get('stream')) == 56
assert not any('excluded' in e for e in d['endpoints'])
assert not any(e.get('covered_by', {}).get('direct_read') for e in d['endpoints'])
PY
```

Result: pass.

```text
endpoints 56
covered_stream 56
excluded 0
direct_read 0
```

```bash
go run ./cmd/connectorgen validate internal/connectors/defs --json
```

Result: pass.

```json
{
  "findings": [],
  "warnings": [],
  "connectors_checked": 548
}
```

```bash
go test ./internal/connectors/conformance -run 'TestConformance/twenty' -count=1
```

Result: pass.

```text
ok  	polymetrics.ai/internal/connectors/conformance	1.301s
```

```bash
scripts/verify-gsd-workflow b4895064
```

Result: pass.

```text
verify-gsd-workflow: implementation changes have GSD/TDD evidence against b4895064
Implementation files changed:
internal/connectors/defs/twenty/api_surface.json
internal/connectors/defs/twenty/fixtures/streams/attachments/page_1.json
internal/connectors/defs/twenty/fixtures/streams/attachments/page_2.json
internal/connectors/defs/twenty/streams.json
internal/connectors/engine/twenty_bundle_test.go
Evidence files changed:
.planning/phases/twenty-s3-read-streams/PLAN.md
.planning/phases/twenty-s3-read-streams/TDD-LEDGER.md
.planning/phases/twenty-s3-read-streams/VERIFICATION.md
```

```bash
go test ./internal/connectors/defs ./internal/connectors/engine ./cmd/connectorgen -count=1
```

Result: pass.

```text
ok  	polymetrics.ai/internal/connectors/defs	0.889s
ok  	polymetrics.ai/internal/connectors/engine	1.688s
ok  	polymetrics.ai/cmd/connectorgen	5.459s
```

```bash
gofmt -w cmd internal
go vet ./...
```

Result: pass; no output.

```bash
go test ./... -count=1
```

Result: pass. Full session output listed all packages as `ok`; slowest packages included:

```text
ok  	polymetrics.ai/internal/cli	153.647s
ok  	polymetrics.ai/internal/connectors/certify	351.605s
ok  	polymetrics.ai/internal/connectors/conformance	15.586s
```

```bash
go build ./cmd/pm
gofmt -l cmd internal
```

Result: pass; no output.

## make verify

Not run. Reason: not feasible under task safety scope. `make verify` depends on `smoke`, and `smoke` executes:

```make
./pm reverse run "$$PLAN_ID" --approve "$$APPROVAL" --root "$$SMOKE_DIR" --json >/dev/null
```

Task explicitly forbids reverse ETL execution. Required non-`make verify` gates above passed.

## Safety / parity

- No credentialed checks.
- No reverse ETL.
- No docs/website/CLI behavior changes in S3 correction; S6 #283 owns CLI surface/help/docs/website.
- Minimal fixtures kept under `internal/connectors/defs/twenty/fixtures/streams/**`; S7 #284 refines/expands later.
- Parent PR #285 merge remains human-gated.
