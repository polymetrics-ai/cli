# Plan — #3792 operation direct-read preflight and surface reconciliation

## GSD setup

- `scripts/gsd doctor` passed.
- Resolved `discuss-phase`, `plan-phase --tdd`, `execute-phase`,
  `verify-work`, and `code-review` with `scripts/gsd sources`.
- `go run ./cmd/agentcontractgen check` passed.
- The manual-GSD fallback and required skills are recorded in `CONTEXT.md`.

## Goal

Turn `availability: "implemented"` into a runtime-enforced claim for
operation-backed direct reads, then give maintainers a deterministic way to
reconcile stale `api_surface` operation reasons from that runtime result.

## Ordered TDD slices

1. **Engine-owned direct-read metadata contract.** Add a closed connector
   interface and engine implementation that validates the named operation's
   supported kind, fixed method/path, positive/effective response cap, endpoint
   ledger presence, and command output policy without a network call. Refactor
   the executor to use that static admission helper.
2. **Commandrunner admission.** Require the metadata provider in
   `validateOperationDirectReadCommand`, reject metadata lookup failures and
   command/metadata mismatches, and retain the existing full-bundle runtime
   preflight sweep unchanged. Add focused red/green cases for missing metadata,
   unsupported kind, endpoint mismatch, cap, and policy.
3. **`api_surface` reconciliation generator.** Add `connectorgen
   surface-reconcile [dir] [--check] [--json] [--reason-contains text]`. Load
   each disk bundle through the declarative engine and call the real runner
   preflight before deriving coverage. For a direct-read operation row, write
   `covered_by.direct_read`/`direct_reads` only from successful commands;
   otherwise derive a factual blocked reason from no candidate, non-implemented
   candidate, or preflight failure. Refuse unknown/ambiguous row shapes.
4. **Proof and reporting.** Add a tiny fixture that starts with stale
   foundation prose and becomes covered only through a real successful
   preflight, plus refusal and blocked-reason fixtures. Run check-only,
   reason-filtered reporting for the six #2985 stale-reason connectors; do not
   apply its changes.
5. **Shipped endpoint ledger repair.** Generate and embed a compact
   operation direct-read endpoint ledger from disk `api_surface.json` rows,
   carrying only method, path, operation kind, and response cap. Attach it at
   bundle load, fail preflight closed for missing or incomplete entries, and
   record the missing-ledger rejection count and binary-size comparison.

## Owned paths

- `internal/connectors/connectors.go`
- `internal/connectors/engine/{bundle.go,connector.go,direct_read.go,*_test.go}`
- `internal/connectors/defs/{defs.go,operation_endpoint_ledger.json}`
- `internal/connectors/commandrunner/{runner.go,runner_test.go}`
- `internal/connectors/native/{amazon-sqs,ashby}/{direct_read.go,engine_delegate.go,*_test.go}`
- `cmd/connectorgen/{main.go,surfacereconcile.go,surfacereconcile_test.go}`
- `docs/migration/conventions.md` and this phase record.

No connector definition under `internal/connectors/defs/{zendesk-support,hubspot,asana,bitbucket,freshchat,youtube-analytics}` is an owned write target.

## Verification plan

- Focused red then green `commandrunner`, `engine`, and `connectorgen` tests.
- The unchanged `TestEveryImplementedCommandPassesRuntimePreflight` must pass
  and cover the strengthened path.
- Run the maintainer help, docs check, focused packages, `go vet ./...`, build,
  and individual non-suite repository gates from `AGENTS.md`; leave aggregate
  `go test ./...` and `make verify` to CI.
