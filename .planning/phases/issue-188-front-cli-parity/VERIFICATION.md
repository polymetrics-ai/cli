# Verification: Front CLI Parity Parent Orchestration

## Completed preflight

```bash
scripts/gsd doctor
```

Result: pass.

```bash
scripts/gsd verify-pi
```

Result: pass.

```bash
scripts/gsd list --json
```

Result: completed; output was large and truncated by the harness display.

```bash
scripts/gsd prompt plan-phase 188 --skip-research --tdd
```

Result: pass; planning prompt generated.

```bash
scripts/gsd prompt programming-loop init --phase issue-188-front-cli-parity --dry-run
```

Result: blocked; `scripts/gsd` reported `unknown GSD command: programming-loop`. Manual GSD fallback recorded in `PLAN.md` and `TDD-LEDGER.md`.

```bash
gh issue view 188 --json number,title,state,url,body,labels,assignees
gh issue view 189 --json number,title,state,url,body,labels,assignees
gh issue view 190 --json number,title,state,url,body,labels,assignees
gh issue view 191 --json number,title,state,url,body,labels,assignees
gh issue view 192 --json number,title,state,url,body,labels,assignees
gh issue view 193 --json number,title,state,url,body,labels,assignees
gh issue view 194 --json number,title,state,url,body,labels,assignees
gh issue view 195 --json number,title,state,url,body,labels,assignees
```

Result: pass; all issues are open and reference #188.

```bash
gh pr list --head feat/188-front-cli-parity --base main --state all --json number,title,state,url,isDraft,headRefName,baseRefName
```

Result: no parent PR exists yet.

```bash
python3 - <<'PY'
from urllib.request import Request, urlopen
url = 'https://dev.frontapp.com/llms.txt'
req = Request(url, headers={'User-Agent': 'polymetrics-agent/1.0'})
with urlopen(req, timeout=30) as r:
    data = r.read()
print(f'fetched {len(data)} bytes from {url}')
PY
```

Result: pass; public Front docs index fetched (69,184 bytes). No credentials used.

## Pending before parent seed commit

```bash
jq empty .planning/phases/issue-188-front-cli-parity/*.json
git diff --check
```

## Required before implementation handoff

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
go run ./cmd/connectorgen validate internal/connectors/defs
```

## CLI help/docs/website parity checks

Applicable once CLI-visible Front surfaces change:

```bash
pm help <topic>
pm <namespace>
pm <command> --help
rg -n "front|Front" docs/cli website
```

No credentialed connector checks are planned or allowed unless explicitly requested.
