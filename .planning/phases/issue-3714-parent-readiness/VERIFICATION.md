# VERIFICATION — issue #3714 parent readiness

Status: refreshed local validation passed. A brand-new no-mistakes run, remote parent CI, and the
readiness transition remain pending.

## Checklist

- [x] The prior parent refresh contained then-current `origin/main` (`fc99e1836`).
- [x] Existing parent branch contains newly current `origin/main` (`4871df2f8`) through clean
  merge `08377a5ae`.
- [x] The merge had no conflicts, so #3730's destructive-write confirmation gate landed unchanged.
- [x] Canonical source retains complete Codex, Claude, and Pi policy/projection metadata.
- [x] Generated projections were regenerated through `agentcontractgen sync`, not hand-edited (0 changes).
- [x] `agentcontractgen check` confirms no projection drift.
- [x] Focused generator and integrated-main package tests pass.
- [ ] Refresh evidence is committed and pushed on the existing parent branch.
- [ ] A brand-new no-mistakes parent validation run completes from the refreshed committed head.
- [ ] PR #3723 checks are green and no actionable review finding remains.
- [ ] PR #3723 is ready for captain review; it is not merged.

## Executed local validation

- `git merge-base --is-ancestor origin/main HEAD`: passed after the merge; it failed as the planned
  RED check before the merge.
- `go run ./cmd/agentcontractgen sync`: synchronized 0 registered projections.
- `go run ./cmd/agentcontractgen check` and `make agent-contract-check`: passed.
- `bash scripts/tests/pi-clean-project-agents.sh`: passed.
- `go test -p 1 ./cmd/connectorgen ./internal/connectors/commandrunner
  ./internal/connectors/conformance ./internal/connectors/connsdk ./internal/connectors/engine`,
  `go test -p 1 ./internal/connectors/defs/recurly`, and `go test -p 1 ./internal/cli`: passed.
- `go test ./internal/agentcontract ./cmd/agentcontractgen`, `go test -p 1 ./internal/app`,
  `go test -p 1 ./internal/cli -run 'Test(ReverseETL|GitHub|YouTubeAnalytics)'`,
  `go test -p 1 ./internal/connectors ./internal/connectors/engine`, and
  `go test -p 1 ./internal/connectors/hooks/google-search-console`: passed.
- `go vet -p 1` on the generator and affected app, CLI, connector, engine, and hook packages:
  passed.
- `make tidy-check`, `go build ./cmd/pm`, `make docs-check-no-build`, `make smoke-no-build`,
  `make lint`, `make connectorgen-validate`, `make connectorgen-surface-sync`,
  `make connector-boundary`, `make release-workflow-check`, and `git diff --check`: passed.

The `origin/main...HEAD` diff check also passes after removing pre-existing trailing whitespace
from the phase artifacts; it does not alter any canonical or generated projection.

The initial parallel test invocation exhausted local temporary disk space. The fresh serial retries
above completed after capacity was restored. Full `make verify` is intentionally left to the PR CI
suite under the repository's per-command timeout policy.

## CLI parity

Not applicable: this phase changes parent integration evidence and generated worker projections,
not a `pm` command, flag, help topic, connector surface, manual, or website documentation.
