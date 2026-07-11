# Twenty S4 writes verification (#281)

Status: GREEN for PR #304 F1 review fix. Prior S4 gates were green. `make verify` remains intentionally not run because `Makefile` target `smoke-no-build` executes `./pm reverse run ... --approve ...`; S4 safety forbids reverse-ETL execution.

## GSD adapter / manual fallback

Command attempted:

```bash
scripts/gsd doctor && scripts/gsd list | head -80 && scripts/gsd prompt programming-loop init --phase twenty-s4-writes --dry-run
```

Result: doctor/list OK; final command failed with `scripts/gsd: unknown GSD command: programming-loop` (exit 1). Manual GSD fallback active and recorded.

## Pre-production red / preflight evidence

```text
research preflight:
true
0
true
field manifest preflight:
28
546
api surface GET preflight:
56
api surface covered streams preflight:
56
initial red counts:
writes.json missing (expected before S4)
56
56
0
```

## Review fix F1 required gates

Planned exact commands:

```bash
jq . internal/connectors/defs/twenty/writes.json internal/connectors/engine/schema/writes.schema.json
python3 - <<'PY'
import json
from pathlib import Path
w=json.loads(Path('internal/connectors/defs/twenty/writes.json').read_text())
b=[a for a in w['actions'] if a['name'].startswith('batch_')]
assert len(b)==28
assert all(a.get('body_field')=='records' for a in b)
assert not any('body_fields' in a for a in b)
assert all(a['record_schema']['required']==['records'] for a in b)
print('batch body_field ok', len(b))
PY
go test ./internal/connectors/engine -run 'TestWrite.*BodyField|TestWriteRawBodyField' -count=1
go run ./cmd/connectorgen validate internal/connectors/defs --json
go test ./internal/connectors/conformance -run 'TestConformance/twenty' -count=1
go test ./internal/connectors/defs ./internal/connectors/engine ./internal/connectors/conformance ./cmd/connectorgen -count=1
go vet ./...
go build ./cmd/pm
gofmt -l cmd internal
go test ./... -count=1
scripts/verify-gsd-workflow 1a86cc1a
```

Red evidence captured before production implementation:

```bash
go test ./internal/connectors/engine -run 'TestWrite.*BodyField|TestWriteRawBodyField' -count=1
```

```text
# polymetrics.ai/internal/connectors/engine [polymetrics.ai/internal/connectors/engine.test]
internal/connectors/engine/write_test.go:209:3: unknown field BodyField in struct literal of type WriteAction
internal/connectors/engine/write_test.go:255:3: unknown field BodyField in struct literal of type WriteAction
FAIL	polymetrics.ai/internal/connectors/engine [build failed]
FAIL
```

Green review-fix gate results:

```text
jq . internal/connectors/defs/twenty/writes.json internal/connectors/engine/schema/writes.schema.json >/dev/null
# exit 0

python batch shape check:
batch body_field ok 28

go test ./internal/connectors/engine -run 'TestWrite.*BodyField|TestWriteRawBodyField' -count=1
ok  	polymetrics.ai/internal/connectors/engine	0.390s

go run ./cmd/connectorgen validate internal/connectors/defs --json
{
  "findings": [],
  "warnings": [],
  "connectors_checked": 548
}

go test ./internal/connectors/conformance -run 'TestConformance/twenty' -count=1
ok  	polymetrics.ai/internal/connectors/conformance	1.265s

go test ./internal/connectors/defs ./internal/connectors/engine ./internal/connectors/conformance ./cmd/connectorgen -count=1
ok  	polymetrics.ai/internal/connectors/defs	1.749s
ok  	polymetrics.ai/internal/connectors/engine	1.152s
ok  	polymetrics.ai/internal/connectors/conformance	10.059s
ok  	polymetrics.ai/cmd/connectorgen	5.147s

go vet ./...
# no output; exit 0

go build ./cmd/pm
# no output; exit 0

gofmt -l cmd internal
# no output; exit 0

go test ./... -count=1
# all packages passed; slowest observed: internal/connectors/certify 371.909s, internal/cli 162.848s

scripts/verify-gsd-workflow 1a86cc1a
verify-gsd-workflow: implementation changes have GSD/TDD evidence against 1a86cc1a
Implementation files changed:
internal/connectors/defs/twenty/api_surface.json
internal/connectors/defs/twenty/writes.json
Evidence files changed:
.planning/phases/twenty-s4-writes/PLAN.md
.planning/phases/twenty-s4-writes/RUN-STATE.json
.planning/phases/twenty-s4-writes/SUMMARY.md
.planning/phases/twenty-s4-writes/TDD-LEDGER.md
.planning/phases/twenty-s4-writes/VERIFICATION.md
```

## Original local artifact gates

```bash
jq . internal/connectors/defs/twenty/writes.json internal/connectors/defs/twenty/api_surface.json >/dev/null
find .planning/auto-loop/tasks/S4 .planning/phases/twenty-s4-writes -type f -name '*.json' -exec jq . {} + >/dev/null
```

Output: no output; exit 0.

Count/shape command output:

```text
actions 84
create 28
update 28
batch 28
delete 0
dupes api total 140
api GET 56
api POST 56
api PATCH 28
api DELETE 0
api write 84
```

Shape checks:

```text
true
true
true
true
true
```

Coverage/immutability checks:

```text
true
true
immutable violations: []
actions checked: 84
```

## Connector validation gates

### connectorgen full defs

```bash
go run ./cmd/connectorgen validate internal/connectors/defs --json
```

```json
{
  "findings": [],
  "warnings": [],
  "connectors_checked": 548
}
```

### twenty conformance

```bash
go test ./internal/connectors/conformance -run 'TestConformance/twenty' -count=1
```

```text
ok  	polymetrics.ai/internal/connectors/conformance	1.640s
```

### focused packages

```bash
go test ./internal/connectors/defs ./internal/connectors/engine ./internal/connectors/conformance ./cmd/connectorgen -count=1
```

```text
ok  	polymetrics.ai/internal/connectors/defs	0.874s
ok  	polymetrics.ai/internal/connectors/engine	1.160s
ok  	polymetrics.ai/internal/connectors/conformance	9.651s
ok  	polymetrics.ai/cmd/connectorgen	5.890s
```

### format / vet / build

```bash
gofmt -l cmd internal
go vet ./...
go build ./cmd/pm
```

Output: no output; exit 0 for each command.

### full tests

```bash
go test ./... -count=1
```

```text
ok  	polymetrics.ai/cmd/connectorgen	7.490s
ok  	polymetrics.ai/cmd/iconregistrygen	0.547s
ok  	polymetrics.ai/cmd/pm	0.578s
ok  	polymetrics.ai/cmd/prissueguard	0.879s
ok  	polymetrics.ai/internal/agentmode	1.183s
ok  	polymetrics.ai/internal/app	17.324s
ok  	polymetrics.ai/internal/cli	163.980s
ok  	polymetrics.ai/internal/connectors	2.354s
ok  	polymetrics.ai/internal/connectors/bundleregistry	5.203s
ok  	polymetrics.ai/internal/connectors/certify	364.419s
ok  	polymetrics.ai/internal/connectors/commandrunner	3.468s
ok  	polymetrics.ai/internal/connectors/conformance	16.583s
ok  	polymetrics.ai/internal/connectors/connsdk	4.011s
ok  	polymetrics.ai/internal/connectors/defs	4.716s
ok  	polymetrics.ai/internal/connectors/engine	4.812s
ok  	polymetrics.ai/internal/connectors/hooks/akeneo	3.895s
ok  	polymetrics.ai/internal/connectors/hooks/alpha-vantage	3.025s
ok  	polymetrics.ai/internal/connectors/hooks/amazon-ads	2.172s
ok  	polymetrics.ai/internal/connectors/hooks/amazon-seller-partner	1.931s
ok  	polymetrics.ai/internal/connectors/hooks/apify-dataset	1.305s
ok  	polymetrics.ai/internal/connectors/hooks/apple-search-ads	1.480s
ok  	polymetrics.ai/internal/connectors/hooks/appsflyer	1.789s
ok  	polymetrics.ai/internal/connectors/hooks/ashby	2.093s
ok  	polymetrics.ai/internal/connectors/hooks/aws-cloudtrail	2.287s
ok  	polymetrics.ai/internal/connectors/hooks/babelforce	2.351s
ok  	polymetrics.ai/internal/connectors/hooks/basecamp	2.614s
ok  	polymetrics.ai/internal/connectors/hooks/blogger	2.628s
ok  	polymetrics.ai/internal/connectors/hooks/bunny-inc	2.640s
ok  	polymetrics.ai/internal/connectors/hooks/canny	2.618s
ok  	polymetrics.ai/internal/connectors/hooks/chift	2.616s
ok  	polymetrics.ai/internal/connectors/hooks/copper	2.626s
ok  	polymetrics.ai/internal/connectors/hooks/dixa	2.618s
ok  	polymetrics.ai/internal/connectors/hooks/ebay-fulfillment	2.605s
ok  	polymetrics.ai/internal/connectors/hooks/elasticsearch	2.604s
ok  	polymetrics.ai/internal/connectors/hooks/fastbill	2.590s
ok  	polymetrics.ai/internal/connectors/hooks/feishu	2.599s
ok  	polymetrics.ai/internal/connectors/hooks/free-agent-connector	2.606s
ok  	polymetrics.ai/internal/connectors/hooks/freightview	2.617s
ok  	polymetrics.ai/internal/connectors/hooks/github	2.994s
ok  	polymetrics.ai/internal/connectors/hooks/gmail	2.660s
ok  	polymetrics.ai/internal/connectors/hooks/google-ads	2.633s
ok  	polymetrics.ai/internal/connectors/hooks/google-analytics-data-api	2.624s
ok  	polymetrics.ai/internal/connectors/hooks/google-calendar	2.613s
ok  	polymetrics.ai/internal/connectors/hooks/google-classroom	2.605s
ok  	polymetrics.ai/internal/connectors/hooks/google-forms	2.627s
ok  	polymetrics.ai/internal/connectors/hooks/google-pagespeed-insights	2.604s
ok  	polymetrics.ai/internal/connectors/hooks/google-search-console	2.252s
ok  	polymetrics.ai/internal/connectors/hooks/hookset	2.422s
ok  	polymetrics.ai/internal/connectors/hooks/hoorayhr	2.408s
ok  	polymetrics.ai/internal/connectors/hooks/jamf-pro	2.409s
ok  	polymetrics.ai/internal/connectors/hooks/keka	2.410s
ok  	polymetrics.ai/internal/connectors/hooks/less-annoying-crm	2.437s
ok  	polymetrics.ai/internal/connectors/hooks/lokalise	2.422s
ok  	polymetrics.ai/internal/connectors/hooks/mendeley	2.408s
ok  	polymetrics.ai/internal/connectors/hooks/mercado-ads	2.435s
ok  	polymetrics.ai/internal/connectors/hooks/metabase	2.417s
ok  	polymetrics.ai/internal/connectors/hooks/microsoft-dataverse	2.403s
ok  	polymetrics.ai/internal/connectors/hooks/microsoft-entra-id	2.412s
ok  	polymetrics.ai/internal/connectors/hooks/microsoft-lists	2.421s
ok  	polymetrics.ai/internal/connectors/hooks/microsoft-teams	2.528s
ok  	polymetrics.ai/internal/connectors/hooks/mixpanel	2.719s
ok  	polymetrics.ai/internal/connectors/hooks/mode	2.764s
ok  	polymetrics.ai/internal/connectors/hooks/monday	2.920s
ok  	polymetrics.ai/internal/connectors/hooks/my-hours	3.070s
ok  	polymetrics.ai/internal/connectors/hooks/netsuite	3.063s
ok  	polymetrics.ai/internal/connectors/hooks/nexus-datasets	3.076s
ok  	polymetrics.ai/internal/connectors/hooks/notion	3.036s
ok  	polymetrics.ai/internal/connectors/hooks/outlook	2.989s
ok  	polymetrics.ai/internal/connectors/hooks/paypal-transaction	3.016s
ok  	polymetrics.ai/internal/connectors/hooks/pinterest	3.034s
ok  	polymetrics.ai/internal/connectors/hooks/plaid	3.045s
ok  	polymetrics.ai/internal/connectors/hooks/pocket	3.090s
ok  	polymetrics.ai/internal/connectors/hooks/prestashop	3.082s
ok  	polymetrics.ai/internal/connectors/hooks/quickbooks	3.045s
ok  	polymetrics.ai/internal/connectors/hooks/rootly	3.097s
ok  	polymetrics.ai/internal/connectors/hooks/rss	18.088s
ok  	polymetrics.ai/internal/connectors/hooks/safetyculture	3.090s
ok  	polymetrics.ai/internal/connectors/hooks/salesloft	3.071s
ok  	polymetrics.ai/internal/connectors/hooks/sentry	3.094s
ok  	polymetrics.ai/internal/connectors/hooks/serpstat	3.083s
ok  	polymetrics.ai/internal/connectors/hooks/slack	3.089s
ok  	polymetrics.ai/internal/connectors/hooks/smartsheets	3.087s
ok  	polymetrics.ai/internal/connectors/hooks/snapchat-marketing	3.113s
ok  	polymetrics.ai/internal/connectors/hooks/stigg	10.571s
ok  	polymetrics.ai/internal/connectors/hooks/strava	3.102s
ok  	polymetrics.ai/internal/connectors/hooks/twilio	2.744s
ok  	polymetrics.ai/internal/connectors/hooks/uptick	2.775s
ok  	polymetrics.ai/internal/connectors/hooks/us-census	2.725s
ok  	polymetrics.ai/internal/connectors/hooks/wasabi-stats-api	2.811s
ok  	polymetrics.ai/internal/connectors/hooks/yahoo-finance-price	2.814s
ok  	polymetrics.ai/internal/connectors/hooks/youtube-analytics	2.789s
ok  	polymetrics.ai/internal/connectors/hooks/zoho-analytics-metadata-api	2.814s
ok  	polymetrics.ai/internal/connectors/hooks/zoho-bigin	2.476s
ok  	polymetrics.ai/internal/connectors/native/alpha-vantage	2.476s
ok  	polymetrics.ai/internal/connectors/native/amazon-sqs	2.413s
ok  	polymetrics.ai/internal/connectors/native/apify-dataset	2.510s
ok  	polymetrics.ai/internal/connectors/native/ashby	2.457s
ok  	polymetrics.ai/internal/connectors/native/aws-cloudtrail	2.453s
ok  	polymetrics.ai/internal/connectors/native/babelforce	2.418s
ok  	polymetrics.ai/internal/connectors/native/basecamp	2.419s
ok  	polymetrics.ai/internal/connectors/native/bing-ads	2.404s
ok  	polymetrics.ai/internal/connectors/native/bunny-inc	2.468s
ok  	polymetrics.ai/internal/connectors/native/canny	2.428s
ok  	polymetrics.ai/internal/connectors/native/copper	2.458s
ok  	polymetrics.ai/internal/connectors/native/dixa	2.453s
ok  	polymetrics.ai/internal/connectors/native/dynamodb	2.400s
ok  	polymetrics.ai/internal/connectors/native/faker	2.482s
ok  	polymetrics.ai/internal/connectors/native/fastbill	2.495s
ok  	polymetrics.ai/internal/connectors/native/feishu	2.459s
ok  	polymetrics.ai/internal/connectors/native/free-agent-connector	2.457s
ok  	polymetrics.ai/internal/connectors/native/freightview	2.471s
ok  	polymetrics.ai/internal/connectors/native/google-analytics-data-api	2.482s
ok  	polymetrics.ai/internal/connectors/native/google-calendar	2.473s
ok  	polymetrics.ai/internal/connectors/native/google-classroom	2.452s
ok  	polymetrics.ai/internal/connectors/native/google-pagespeed-insights	2.613s
ok  	polymetrics.ai/internal/connectors/native/less-annoying-crm	2.757s
ok  	polymetrics.ai/internal/connectors/native/lokalise	2.762s
ok  	polymetrics.ai/internal/connectors/native/mendeley	2.742s
ok  	polymetrics.ai/internal/connectors/native/mercado-ads	2.745s
ok  	polymetrics.ai/internal/connectors/native/metabase	2.752s
ok  	polymetrics.ai/internal/connectors/native/mode	2.742s
ok  	polymetrics.ai/internal/connectors/native/my-hours	2.771s
ok  	polymetrics.ai/internal/connectors/native/nativeset	2.695s
ok  	polymetrics.ai/internal/connectors/native/pocket	2.790s
ok  	polymetrics.ai/internal/connectors/native/postgres	2.653s
ok  	polymetrics.ai/internal/connectors/native/prestashop	2.817s
ok  	polymetrics.ai/internal/connectors/native/rootly	2.831s
ok  	polymetrics.ai/internal/connectors/native/safetyculture	2.831s
ok  	polymetrics.ai/internal/connectors/native/tally-prime	2.805s
ok  	polymetrics.ai/internal/connectors/native/yahoo-finance-price	2.884s
ok  	polymetrics.ai/internal/coordination	3.027s
ok  	polymetrics.ai/internal/coordination/issueguard	3.244s
ok  	polymetrics.ai/internal/flow	3.067s
ok  	polymetrics.ai/internal/ledger	3.156s
ok  	polymetrics.ai/internal/perf	3.314s
ok  	polymetrics.ai/internal/rlm	3.030s
ok  	polymetrics.ai/internal/rlm/router	3.323s
ok  	polymetrics.ai/internal/runtime	3.278s
ok  	polymetrics.ai/internal/runtimecheck	2.919s
ok  	polymetrics.ai/internal/safety	3.195s
ok  	polymetrics.ai/internal/schedule	3.317s
ok  	polymetrics.ai/internal/state	3.643s
ok  	polymetrics.ai/internal/temporalprobe	3.004s
ok  	polymetrics.ai/internal/vault	3.298s
ok  	polymetrics.ai/internal/worker	2.663s
```

## Scope / docs parity / safety

- `make verify`: NOT RUN. Reason: `verify` depends on `smoke`, and `smoke-no-build` runs `./pm reverse run "$PLAN_ID" --approve "$APPROVAL"`; S4 forbids reverse-ETL execution.
- CLI/help/docs/website parity: not applicable to S4; no command/flag/help/docs/website files edited. S6 #283 and S7 #284 own parity.
- Live Twenty credentials: NOT USED.
- Raw HTTP / reverse ETL execution: NOT RUN.
- DELETE/destructive actions: NOT ADDED.
- Dependencies / Go / engine schema changes: NONE.

## GSD evidence gate

```bash
scripts/verify-gsd-workflow 1a86cc1a
```

```text
verify-gsd-workflow: implementation changes have GSD/TDD evidence against 1a86cc1a
Implementation files changed:
internal/connectors/defs/twenty/api_surface.json
internal/connectors/defs/twenty/writes.json
Evidence files changed:
.planning/phases/twenty-s4-writes/PLAN.md
.planning/phases/twenty-s4-writes/RUN-STATE.json
.planning/phases/twenty-s4-writes/SUMMARY.md
.planning/phases/twenty-s4-writes/TDD-LEDGER.md
.planning/phases/twenty-s4-writes/VERIFICATION.md
```
