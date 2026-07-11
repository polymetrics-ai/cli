# Twenty S5 destructive delete actions verification (#282)

Status: GREEN_LOCAL_GSD_EVIDENCE_PASSED. Manual GSD fallback active because `scripts/gsd prompt programming-loop init --phase twenty-s5-deletes --dry-run` returned `scripts/gsd: unknown GSD command: programming-loop`.

## Pre-production red evidence

```text
red writes actions 84
red kind counts {'create': 56, 'update': 28, 'delete': 0}
red name counts {'create_names': 28, 'update_names': 28, 'batch_names': 28, 'delete_names': 0}
red api_surface rows 140
red methods {'GET': 56, 'POST': 56, 'PATCH': 28, 'DELETE': 0}
```

## Local gate results

```bash
jq . internal/connectors/defs/twenty/writes.json internal/connectors/defs/twenty/api_surface.json >/tmp/twenty-s5-jq.out && echo 'jq ok'
```

```text
jq ok
```

```bash
python3 - <<'PY'
# corrected S5 assertion block from task body
PY
```

```text
s5 destructive deletes ok 28 {'create': 56, 'update': 28, 'delete': 28} {'create_names': 28, 'update_names': 28, 'batch_names': 28, 'delete_names': 28} {'GET': 56, 'POST': 56, 'PATCH': 28, 'DELETE': 28}
```

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

```bash
go test ./internal/connectors/conformance -run 'TestConformance/twenty' -count=1
```

```text
ok  	polymetrics.ai/internal/connectors/conformance	1.256s
```

```bash
go test ./internal/connectors/defs ./internal/connectors/engine ./internal/connectors/conformance ./cmd/connectorgen -count=1
```

```text
ok  	polymetrics.ai/internal/connectors/defs	0.850s
ok  	polymetrics.ai/internal/connectors/engine	1.302s
ok  	polymetrics.ai/internal/connectors/conformance	11.144s
ok  	polymetrics.ai/cmd/connectorgen	4.758s
```

```bash
go vet ./...
go build ./cmd/pm
gofmt -l cmd internal
```

```text
# no output; exit 0 for each command
```

```bash
go test ./... -count=1
```

Result: all packages passed. Slowest observed exact lines:

```text
ok  	polymetrics.ai/internal/cli	187.716s
ok  	polymetrics.ai/internal/connectors/certify	443.670s
```

```bash
scripts/verify-gsd-workflow bc014ef6
```

```text
verify-gsd-workflow: implementation changes have GSD/TDD evidence against bc014ef6
Implementation files changed:
internal/connectors/defs/twenty/api_surface.json
internal/connectors/defs/twenty/writes.json
Evidence files changed:
.planning/phases/twenty-s5-deletes/PLAN.md
.planning/phases/twenty-s5-deletes/RUN-STATE.json
.planning/phases/twenty-s5-deletes/SUMMARY.md
.planning/phases/twenty-s5-deletes/TDD-LEDGER.md
.planning/phases/twenty-s5-deletes/VERIFICATION.md
```

## make verify decision

Do not run `make verify` for S5 because prior S4 evidence shows `verify` depends on a smoke target that runs `./pm reverse run ... --approve ...`; S5 forbids reverse ETL execution and destructive external action execution.

## Scope / parity / safety

- CLI/help/docs/website parity: connector surface change only; user forbids docs/CLI/website edits in S5. Deferred to parent S6/S7 parity work.
- Live Twenty credentials: NOT USED.
- Raw HTTP / reverse ETL execution: NOT RUN.
- Dependencies: NONE.
