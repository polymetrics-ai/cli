# Twenty S3 read streams — verification

Issue: #280
PR: #290
Branch: `feat/280-twenty-streams`

## Review F1 correction checklist (2026-07-11)

Disposition: Accepted. Action: remove no-op cursor `pagination.limit_param: "limit"`, add static stream `query: {"limit":"60"}` to all 28 streams, and update attachments fixture request queries.

Pre-production red validation to capture before bundle edits:

```bash
python3 - <<'PY'
import json
from pathlib import Path
s=json.loads(Path('internal/connectors/defs/twenty/streams.json').read_text())
assert len(s['streams']) == 28
for stream in s['streams']:
    pag=stream['pagination']
    assert pag['type']=='cursor'
    assert pag['cursor_param']=='starting_after'
    assert 'limit_param' not in pag
    assert stream.get('query',{}).get('limit') == '60'
print('streams limit query ok', len(s['streams']))
p1=json.loads(Path('internal/connectors/defs/twenty/fixtures/streams/attachments/page_1.json').read_text())
p2=json.loads(Path('internal/connectors/defs/twenty/fixtures/streams/attachments/page_2.json').read_text())
assert p1['request']['query'] == {'limit':'60'}
assert p2['request']['query'] == {'limit':'60','starting_after':'attachment_fixture_cursor_1'}
print('fixtures limit query ok')
PY
```

Result: fail as expected before production edit.

```text
Traceback (most recent call last):
  File "<stdin>", line 9, in <module>
AssertionError

Command exited with code 1
```

Required green gates after the fix:

```bash
jq . internal/connectors/defs/twenty/streams.json internal/connectors/defs/twenty/fixtures/streams/attachments/page_1.json internal/connectors/defs/twenty/fixtures/streams/attachments/page_2.json
python3 - <<'PY'
import json
from pathlib import Path
s=json.loads(Path('internal/connectors/defs/twenty/streams.json').read_text())
assert len(s['streams']) == 28
for stream in s['streams']:
    pag=stream['pagination']
    assert pag['type']=='cursor'
    assert pag['cursor_param']=='starting_after'
    assert 'limit_param' not in pag
    assert stream.get('query',{}).get('limit') == '60'
print('streams limit query ok', len(s['streams']))
p1=json.loads(Path('internal/connectors/defs/twenty/fixtures/streams/attachments/page_1.json').read_text())
p2=json.loads(Path('internal/connectors/defs/twenty/fixtures/streams/attachments/page_2.json').read_text())
assert p1['request']['query'] == {'limit':'60'}
assert p2['request']['query'] == {'limit':'60','starting_after':'attachment_fixture_cursor_1'}
print('fixtures limit query ok')
PY
go run ./cmd/connectorgen validate internal/connectors/defs --json
go test ./internal/connectors/conformance -run 'TestConformance/twenty' -count=1
go test ./internal/connectors/defs ./internal/connectors/engine ./cmd/connectorgen -count=1
go build ./cmd/pm
gofmt -l cmd internal
```

Result: pass.

```text
jq: pass (all three JSON files parsed and pretty-printed; no stderr)
```

```text
streams limit query ok 28
fixtures limit query ok
```

```json
{
  "findings": [],
  "warnings": [],
  "connectors_checked": 548
}
```

```text
ok  	polymetrics.ai/internal/connectors/conformance	2.059s
```

```text
ok  	polymetrics.ai/internal/connectors/defs	0.979s
ok  	polymetrics.ai/internal/connectors/engine	1.310s
ok  	polymetrics.ai/cmd/connectorgen	5.666s
```

```text
go build ./cmd/pm: pass (no output)
gofmt -l cmd internal: pass (no output)
```

Additional local gates:

```bash
gofmt -w cmd internal
go vet ./...
go test ./... -count=1
scripts/verify-gsd-workflow b4895064
```

Result: pass.

```text
gofmt -w cmd internal: pass (no output)
go vet ./...: pass (no output)
go test ./... -count=1: pass; full output listed all packages as ok; slowest packages included:
ok  	polymetrics.ai/internal/cli	161.641s
ok  	polymetrics.ai/internal/connectors/certify	360.914s
ok  	polymetrics.ai/internal/connectors/conformance	19.345s
scripts/verify-gsd-workflow b4895064: pass
verify-gsd-workflow: implementation changes have GSD/TDD evidence against b4895064
Implementation files changed:
internal/connectors/defs/twenty/api_surface.json
internal/connectors/defs/twenty/fixtures/streams/attachments/page_1.json
internal/connectors/defs/twenty/fixtures/streams/attachments/page_2.json
internal/connectors/defs/twenty/streams.json
Evidence files changed:
.planning/phases/twenty-s3-read-streams/PLAN.md
.planning/phases/twenty-s3-read-streams/TDD-LEDGER.md
.planning/phases/twenty-s3-read-streams/VERIFICATION.md
```

Optional broad gate: `go test ./... -count=1` was run and passed locally after the review fix. Prior PR CI verify was also green on `0c3b2c97` before this F1 correction; CI verify remains expected after push.

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
