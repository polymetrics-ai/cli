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

## Next

1. Add #150 red tests for Chatwoot runtime manual and website connector data.
2. Regenerate checked-in connector docs and website data for Chatwoot command surface parity.
3. Run targeted and docs/website parity verification.
4. Open a stacked sub-PR against `feat/148-chatwoot-cli-parity` with `Refs #150` and `Refs #148`.

## Safety

No secrets requested or used. No credentialed connector checks. No dependency changes. No external writes. Parent PR merge to `main` remains human-gated.
