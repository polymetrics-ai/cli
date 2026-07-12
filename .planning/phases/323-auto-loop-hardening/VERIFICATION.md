# Verification: Autonomous Delivery Control-Plane Hardening

## Completed for parent scaffold

- Clean isolated worktree created from `origin/main`.
- Parent branch: `fix/323-auto-loop-hardening`.
- `scripts/gsd doctor` passes when run with the repository-supported Node 24 runtime on `PATH`.
- Installed programming-loop dry-run detects subagent mode.
- Parent issue #323 and local forensic report headings were inspected without loading raw session
  transcripts or copying sensitive material.

## Before first parent push

```bash
jq empty .planning/phases/323-auto-loop-hardening/RUN-STATE.json
jq empty .planning/phases/323-auto-loop-hardening/ORCHESTRATION-STATE.json
git diff --check
```

## Child gates

Each child issue records its exact targeted tests. The accumulated minimum is:

```bash
go test ./internal/agentloop/...
go test -race ./internal/agentloop/...
go test ./cmd/loopctl/...
scripts/tests/auto-loop-control.sh
```

## Final parent gate

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
```

Runtime-backed services and live provider APIs are not required. No credentialed check is in scope.
The merge-disabled canary uses local/fake adapters unless a later explicit human gate authorizes
anything broader.
