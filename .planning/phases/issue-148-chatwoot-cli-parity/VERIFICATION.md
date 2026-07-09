# Verification: Chatwoot CLI Parity Parent Orchestration

## Completed preflight

```bash
scripts/gsd doctor
scripts/gsd verify-pi
scripts/gsd list --json
gh issue view 148 --json number,title,state,url,body,labels,assignees,milestone,projectItems
gh issue view 149 --json number,title,state,url,body,labels,assignees
gh pr list --head feat/148-chatwoot-cli-parity --base main --json number,title,state,url,isDraft,baseRefName,headRefName,mergeStateStatus,reviewDecision,statusCheckRollup
```

Results:

- GSD/Pi adapter health checks passed.
- Parent issue #148 is open.
- Sub-issues #149-#155 are open.
- Parent PR for `feat/148-chatwoot-cli-parity` -> `main` is currently missing.

## Required before parent seed commit

```bash
jq empty \
  .planning/phases/issue-148-chatwoot-cli-parity/RUN-STATE.json \
  .planning/phases/issue-148-chatwoot-cli-parity/ORCHESTRATION-STATE.json

git diff --check
```

## Required before issue #149 handoff

```bash
python3 .planning/phases/issue-149-chatwoot-cli-surface-metadata/traces/verify-official-surface-count.py
jq empty internal/connectors/defs/chatwoot/api_surface.json internal/connectors/defs/chatwoot/cli_surface.json
go test ./cmd/connectorgen -run CLISurface -count=1
go test ./internal/connectors/engine -run CLISurface -count=1
go run ./cmd/connectorgen validate internal/connectors/defs/chatwoot
go run ./cmd/connectorgen validate internal/connectors/defs
```

## Required before final parent handoff

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
go run ./cmd/connectorgen validate internal/connectors/defs
```

Runtime-backed checks are optional and are not required for this planning/metadata slice because no credentials or local services are needed.
