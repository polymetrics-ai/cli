# TDD ledger — Google Ads connector parity wave03-r1 (#3021-#3028)

## Red / pre-edit evidence

### 2026-07-31 — official v22 inventory outstrips local ledger and write/direct surfaces

Credential-free source audit command:

```bash
python3 - <<'PY'
import json, urllib.request, collections
spec=json.load(urllib.request.urlopen('https://googleads.googleapis.com/$discovery/rest?version=v22', timeout=60))
methods=[]
def walk(res,prefix=''):
    for name,m in res.get('methods',{}).items():
        methods.append((prefix+name,m))
    for rn,sub in res.get('resources',{}).items():
        walk(sub,prefix+rn+'.')
walk(spec)
print('version', spec.get('version'), 'revision', spec.get('revision'))
print('official_methods', len(methods), dict(collections.Counter(m.get('httpMethod') for _,m in methods)))
print('schemas', len(spec.get('schemas',{})))
local=json.load(open('internal/connectors/defs/google-ads/api_surface.json'))
print('local_surface_rows', len(local.get('endpoints',[])))
PY
```

Observed red evidence:

```text
version v22 revision 20260721
official_methods 163 {'POST': 151, 'GET': 11, 'DELETE': 1}
schemas 1363
local_surface_rows 12
```

Expected green evidence:

- Source audit artifact records exact v22 method count and classification.
- `api_surface.json` is v22, operation-ledger mode, duplicate-free, and no longer a coarse v24 grouping.
- Every declared stream and write action has a `covered_by` row.
- Every executable direct-read command referenced by `api_surface.json` exists in `cli_surface.json` and `operations.json`.

### 2026-07-31 — required connector-local validate command currently fails on baseline

Command:

```bash
go run ./cmd/connectorgen validate internal/connectors/defs/google-ads
```

Observed red evidence:

```text
fixtures: metadata.json: [missing_file] load bundle fixtures: missing required file metadata.json
schemas: metadata.json: [missing_file] load bundle schemas: missing required file metadata.json
connectorgen validate: 2 connector(s) checked, 2 finding(s)
exit status 1
```

Expected green evidence:

- The exact user-mandated command validates the Google Ads connector successfully, or the blocker is recorded if shared tooling changes are disallowed.

## Green / verification log

### 2026-07-31 — source inventory and fixture-backed parity generated

Generator:

```bash
python3 .planning/phases/google-ads-parity-wave03-r1/generate_google_ads_parity.py
```

Observed:

```text
generated {'classification': {'blocked_duplicate': 1, 'blocked_raw_query': 1, 'blocked_raw_write_schema': 97, 'blocked_required_body_direct_read': 12, 'blocked_reserved_path': 22, 'direct_read': 21, 'stream': 2, 'write': 7}, 'api_surface_rows': 164, 'writes': 7, 'write_fixtures': 7, 'direct_reads': 21, 'blocked': 133}
```

Green evidence:

- `internal/connectors/defs/google-ads/api_surface.json` uses v22 operation ledger mode and records 164 rows: 3 stream-covered rows, 21 direct reads, 7 write-covered rows, and 133 blocked/planned operation rows.
- `SOURCE-AUDIT.json` records discovery revision `20260721`, 163 raw methods, HTTP split `GET=11`, `POST=151`, `DELETE=1`, and the reserved path-variable gap.
- The extra local row is documented: `customers.googleAds.search` is split into fixed `campaigns` and `ad_groups` stream rows.

### 2026-07-31 — exact validate gate fixed generically and passed

A small generic `connectorgen validate` compatibility change now accepts either a defs root or a single connector bundle directory. This was required to make the user-mandated gate validate Google Ads without treating `schemas/` and `fixtures/` as sibling connectors.

Command:

```bash
go run ./cmd/connectorgen validate internal/connectors/defs/google-ads
```

Observed:

```text
connectorgen validate: 1 connector(s) checked, 0 findings
```

### 2026-07-31 — fixture-only conformance and CLI/help parity gates passed

Commands:

```bash
go test ./internal/connectors/conformance -run 'TestConformance/google-ads' -count=1
go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1
go test ./internal/connectors/hooks/google-ads -count=1
go build ./cmd/pm
make connector-boundary
make verify
git diff --check
```

Observed: all passed. `make verify` initially timed out twice in the long-running, unrelated `internal/connectors/certify` package, then `go test ./internal/connectors/certify -timeout 20m` passed in 1084.020s and the final full `make verify` run passed.
