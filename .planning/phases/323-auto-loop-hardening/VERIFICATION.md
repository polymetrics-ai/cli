# Verification: Autonomous Delivery Control-Plane Hardening

## Completed for parent scaffold and implementation alignment

- Clean isolated worktree created from `origin/main`, then normally merged with the committed
  `feat/pi-shepherd-loop` implementation after PR #324 was retargeted to that base.
- Parent branch: `fix/323-auto-loop-hardening`.
- Draft parent PR #324 targets `main`.
- GitHub native sub-issue listing for #323 returns fifteen phase/slice issues #325-#339.
- `scripts/gsd doctor` passes when run with the repository-supported Node 24 runtime on `PATH`.
- Installed programming-loop dry-run detects subagent mode.
- Parent issue #323 and local forensic report headings were inspected without loading raw session
  transcripts or copying sensitive material.
- `scripts/pi-shepherd-loop.sh` is covered by the Phase 0 first-action guard and both canonical
  inventories; run, resume, help, and enable/force bypass attempts are tested in a clean sandbox.
- Validator model routing is isolated to GPT-5.6 Sol/high. Orchestrator and worker model defaults
  remain unchanged.
- Pi model compatibility was checked without a model invocation: 0.80.3 lacks Sol, while 0.80.6
  lists `openai-codex/gpt-5.6-sol`.

## Parent roster checkpoint

```bash
jq empty .planning/phases/323-auto-loop-hardening/RUN-STATE.json
jq empty .planning/phases/323-auto-loop-hardening/ORCHESTRATION-STATE.json
git diff --check
go test ./internal/agentloop/... ./cmd/loopctl/... -count=1
bash scripts/tests/auto-loop-control.sh
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
