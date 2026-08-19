# TDD ledger — Stripe connector parity (#3005)

## Red / pre-edit validation

### 2026-07-30 — official operation inventory parity is incomplete

Command (credential-free, no provider call beyond fetching public OpenAPI source):

```bash
python3 - <<'PY'
import json, urllib.request
from collections import Counter
with urllib.request.urlopen('https://raw.githubusercontent.com/stripe/openapi/master/openapi/spec3.json', timeout=60) as r:
    spec=json.load(r)
official=[]
for p,item in spec.get('paths',{}).items():
    for m,obj in item.items():
        if m.lower() in {'get','post','put','patch','delete'}:
            official.append((m.upper(),p))
local=json.load(open('internal/connectors/defs/stripe/api_surface.json'))
local_ops={(e.get('method','').upper(),e.get('path')) for e in local.get('endpoints',[])}
missing=[op for op in official if op not in local_ops]
print('RED stripe official operation inventory parity')
print('official_count',len(official),'local_count',len(local_ops),'missing_count',len(missing))
print('official_methods',dict(Counter(m for m,_ in official)))
print('sample_missing',missing[:10])
PY
```

Observed red evidence:

```text
RED stripe official operation inventory parity
official_count 589 local_count 15 missing_count 576
official_methods {'GET': 263, 'POST': 294, 'DELETE': 32}
sample_missing [('GET', '/v1/account'), ('POST', '/v1/account_links'), ('POST', '/v1/account_sessions'), ('GET', '/v1/accounts'), ('POST', '/v1/accounts'), ('DELETE', '/v1/accounts/{account}'), ('GET', '/v1/accounts/{account}'), ('POST', '/v1/accounts/{account}'), ('POST', '/v1/accounts/{account}/bank_accounts'), ('DELETE', '/v1/accounts/{account}/bank_accounts/{id}')]
```

Expected green evidence after ledger update:

- `api_surface.json` endpoint count equals 589.
- Each official `(method,path)` appears exactly once.
- Existing implemented Stripe streams/writes remain covered.
- Unimplemented destructive/DELETE operations are represented as in-scope blocked/planned operation rows, not blanket unsafe exclusions.

## Green / verification log

### 2026-07-30 — official operation inventory parity green

Observed after regenerating `internal/connectors/defs/stripe/api_surface.json` from the official Stripe OpenAPI source:

```text
GREEN stripe official operation inventory parity
official_count 589 local_count 589 missing_count 0 extra_count 0 duplicate_count 0
covered 8 operation_blocked 581
methods {'GET': 263, 'POST': 294, 'DELETE': 32}
official_lanes {'direct_read_query_search': 9, 'reverse_etl_write': 316, 'etl_read': 242, 'excluded_not_applicable': 8, 'cdc_changefeed': 7, 'binary_file': 7}
```

### 2026-07-30 — destructive write slice green

- Added `delete_customer` write action with `kind: delete`, `body_type: none`, `delete.missing_ok_status: [404]`, and `confirm: "destructive"`.
- Added `fixtures/writes/delete_customer.json`.
- Hardened customer create/update/delete write schemas with `additionalProperties: false`, non-empty string patterns for mutable fields, and `^cus_[A-Za-z0-9_]+$` customer ID patterns so empty effective writes/deletes do not validate.
- Restored optional defaulted `base_url` config and updated shared write-default materialization so write preview/execution preserve configured test/proxy overrides.
- `go test ./internal/connectors/conformance -run 'TestConformance/stripe' -count=1` passed, including delete semantics.

### 2026-07-30 — connector metadata and CLI/help surface green

- Added Stripe-owned `cli_surface.json` for existing stream commands and customer create/update/delete plan commands.
- Regenerated `docs/connectors/stripe/MANUAL.md`, `docs/connectors/stripe/SKILL.md`, and CLI golden transcripts affected by the new provider command listing.
- Verified `pm help stripe`, `pm stripe`, `pm stripe customers --help`, and `pm stripe customers delete --help` without credentials or provider calls.
