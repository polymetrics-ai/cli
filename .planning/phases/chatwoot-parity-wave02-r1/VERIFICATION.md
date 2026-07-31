# VERIFICATION — Chatwoot parity wave02 r1

## Required local gates

```bash
python3 .planning/phases/chatwoot-parity-wave02-r1/generate_chatwoot_defs.py
# Direct validate of internal/connectors/defs/chatwoot is intentionally not used because connectorgen expects a parent connectors directory.
tmp=$(mktemp -d); mkdir -p "$tmp/chatwoot"; cp -R internal/connectors/defs/chatwoot/. "$tmp/chatwoot/"; go run ./cmd/connectorgen validate "$tmp" --json
go run ./cmd/connectorgen validate internal/connectors/defs --json
go test ./internal/connectors/conformance -run 'TestConformance/chatwoot' -count=1
go test ./internal/connectors/conformance -count=1
go test ./internal/connectors/engine -count=1
go run ./cmd/pm docs validate --connectors-dir docs/connectors
go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1
```

## CLI/help/docs parity

- `go run ./cmd/pm help docs` inspected docs command help before docs validation.
- `go run ./cmd/pm help chatwoot` rendered the generated Chatwoot connector manual.
- `go run ./cmd/pm chatwoot --help` rendered provider-style command groups and global safety flags.
- `go run ./cmd/pm chatwoot contacts delete --help` showed `AVAILABILITY implemented`, `--id`, approval-token guidance, and typed `--confirm destructive` requirement.
- `go run ./cmd/pm docs validate --connectors-dir docs/connectors` passed after `MANUAL.md`/`SKILL.md` required sections were fixed.
- `POLYMETRICS_UPDATE_GOLDEN_TRANSCRIPTS=1 go test ./internal/cli -run TestGoldenTranscripts -count=1` refreshed root golden transcripts so Chatwoot appears in root connector commands.

## Results

Passed:

```text
connectorgen validate temp Chatwoot parent: findings=0 warnings=0 connectors_checked=1
connectorgen validate internal/connectors/defs: findings=0 warnings=0 connectors_checked=549
go test ./internal/connectors/conformance -run 'TestConformance/chatwoot' -count=1: ok
go test ./internal/connectors/conformance -count=1: ok
go test ./internal/connectors/engine -count=1: ok
go run ./cmd/pm docs validate --connectors-dir docs/connectors: Validated connector docs in docs/connectors
go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1: ok
```

Broader local gates passed before commit:

```text
gofmt -w cmd internal: passed
go vet ./...: passed
go test ./...: timed out at default package timeout while internal/cli golden/Bahmni matrix was still running; rerun with explicit package timeout below passed
go test -timeout 30m ./...: passed
go build ./cmd/pm: passed
make verify: passed
git diff --check: passed
/Users/karthiksivadas/karthik-agent-workspace/bin/fm-ensure-agents-md.sh .: reported pre-existing AGENTS.md/CLAUDE.md real-file conflict requiring manual reconciliation
```
