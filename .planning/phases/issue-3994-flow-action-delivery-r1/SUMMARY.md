---
coverage:
  - id: D1
    description: Durable scope drift stops an action before connector dispatch.
    verification:
      - kind: integration
        ref: internal/cli/flow_action_runner_test.go:TestConnectorFlowActionRunnerScopeDriftStopsBeforeTargetRequest
        status: pass
    human_judgment: false
  - id: D2
    description: Typed write acknowledgement and independent read-back precede receipt and checkpoint.
    verification:
      - kind: integration
        ref: internal/app/flow_action_test.go:TestExecuteAuthorizedFlowActionAcknowledgesReadsBackThenRecordsReceipt
        status: pass
      - kind: integration
        ref: internal/cli/flow_action_runner_test.go:TestConnectorFlowActionRunnerWritesReadsBackAndThenCheckpoints
        status: pass
    human_judgment: false
  - id: D3
    description: Real provider proof using the captain runbook.
    verification:
      - kind: manual_procedural
        ref: Issue #3994 live-proof procedure
        status: unknown
    human_judgment: true
    rationale: This worktree has no authorized real provider credential and policy forbids requesting or exposing one.
---

# Summary — Issue #3994

## Delivered

- Flow actions now use a connector-backed `StepActionRunner`, assembled by
  `flowRun`, that relies on the #4132 durable authorization reference instead
  of a per-run approval string.
- The action re-derives its content-free authorization scope before connector
  validation, maps selected warehouse records, supplies typed durable approval
  evidence, acknowledges the write, reads the configured target stream back,
  persists an opaque receipt, and only then lets Engine checkpoint success.
- `HTTPActionRunner` remains a legacy flow-package test fixture and is absent
  from the app/CLI production call path.
- Runtime help, generated CLI manual/transcript, website source, and generated
  website docs now describe `--authorization`, `destination_table`, and
  `read_back_stream`.

## Automated verification

- `go test -timeout 20m ./internal/flow/...` — pass
- `go test -timeout 20m ./internal/app/...` — pass
- `go test -timeout 20m ./internal/cli` — pass
- `go vet ./...` and `go build ./cmd/pm` — pass
- `make tidy-check docs-check agent-contract-check connectorgen-validate connectorgen-surface-sync connector-boundary release-workflow-check` — pass
- `make smoke-no-build` and `make lint` — pass
- `npm --prefix website run typecheck` — pass; `npm --prefix website run lint` — pass with pre-existing unrelated warnings

## Live-proof handoff

No credentialed provider was available in this isolated worktree. The deferred
proof is the exact captain-runbook procedure recorded on #3994; it must run
against the approved live target before merge. The hermetic tests prove the
observable connector writes, zero-send scope refusal, read-back, receipt, and
checkpoint ordering but do not substitute for that live provider proof.
