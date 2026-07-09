# Summary: Chatwoot CLI Parity Parent Orchestration

Status: in progress.

## Done

- Read required repo, GSD, parent-orchestration, review-routing, CLI parity, connector migration, and Go skill references.
- Confirmed issue #148 and sub-issues #149-#155 are open.
- Confirmed parent PR for `feat/148-chatwoot-cli-parity` -> `main` was missing, then opened draft parent PR #223 after the plan seed commit.
- Recorded manual GSD fallback because `scripts/gsd prompt programming-loop ...` is not available in the repo-local adapter registry.
- Created parent planning, TDD, verification, run-state, and orchestration-state artifacts.
- Recorded runtime fanout blocker: current Pi API tool surface lacks `subagent`, so issue #149 is local critical path.
- Opened sub-PR #227 for issue #149; CodeRabbit skipped automatic review because the base branch is non-default.
- Ran full local handoff gates on the #149 branch: `go vet ./...`, `go test ./...`, `go build ./cmd/pm`, `make verify`, and `go run ./cmd/connectorgen validate internal/connectors/defs` passed. The first uncached `go test ./...` attempt timed out at the default 10m package timeout in `internal/connectors/certify`; `make verify` passed using the project timeout and a follow-up `go test ./...` passed from cache.
- PR #227 remote checks passed and the branch was squash-merged into the parent as `573b89f5cf8952723213cd55bfa19cb5e3165618`.
- Requested manual CodeRabbit review on parent PR #223 because #227 skipped review and #223 had new integrated commits while still draft.
- Started issue #150 locally as the next dependency-unblocked lane and created its GSD/TDD/verification artifacts before production edits.
- Implemented #150 help/docs/website parity and passed targeted tests, website typecheck/build, `go vet ./...`, `go test ./...`, `go build ./cmd/pm`, and `make verify` locally.
- Opened PR #240 for #150; remote checks passed, CodeRabbit skipped the non-default-base sub-PR, and the branch was squash-merged into the parent as `80db5020b297f1323f94c0c965f4a80ab6b08eb3`.
- Requested parent PR #223 manual CodeRabbit review after #150 integration; CodeRabbit replied `Review finished` and GitHub returned no inline CodeRabbit findings.

## Next

1. Continue the next dependency-unblocked lane (#152 operation ledger, then #151/#153/#154/#155 as dependencies permit).
2. Keep parent PR #223 draft until all sub-issues are integrated and the human gate is ready.
3. Route automated review for each subsequent sub-PR and use parent PR coverage/fallback whenever non-default-base sub-PRs are skipped.

## Safety

No secrets requested or used. No credentialed connector checks. No dependency changes. No external writes. Parent PR merge to `main` remains human-gated.
