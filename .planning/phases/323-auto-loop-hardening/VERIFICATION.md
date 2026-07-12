# Verification: Autonomous Delivery Control-Plane Hardening

## Completed for parent scaffold

- Clean isolated worktree created from `origin/main`.
- Parent branch: `fix/323-auto-loop-hardening`.
- Draft parent PR #324 targets `main`.
- GitHub native sub-issue listing for #323 returns fifteen phase/slice issues #325-#339.
- `scripts/gsd doctor` passes when run with the repository-supported Node 24 runtime on `PATH`.
- Installed programming-loop dry-run detects subagent mode.
- Parent issue #323 and local forensic report headings were inspected without loading raw session
  transcripts or copying sensitive material.

## Parent roster checkpoint

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
