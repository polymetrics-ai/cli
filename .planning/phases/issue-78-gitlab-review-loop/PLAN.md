# Plan: GitLab CLI Parity Review Loop (#78 / PR #127)

Branch: `feat/78-gitlab-cli-parity`
PR: https://github.com/polymetrics-ai/cli/pull/127
Issues: #78, #83-#89

## GSD evidence

- `scripts/gsd doctor` ✅
- `scripts/gsd list --json` ✅
- `scripts/gsd prompt code-review issue-78-gitlab-cli-parity-review --tdd` ✅
- Requested workflow `.agents/agentic-delivery/workflows/claude-review-loop.md` is absent in this checkout; fallback policy sources loaded:
  - `.agents/agentic-delivery/contracts/code-review-disposition-template.md`
  - `.agents/agentic-delivery/workflows/automated-review-routing-loop.md`
  - `.agents/agentic-delivery/workflows/coderabbit-review-loop.md`
  - `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md`

## Required skills loaded

- `gsd-core`
- `golang-how-to`
- `golang-security`
- `golang-safety`
- `golang-error-handling`
- `golang-lint`
- `golang-testing`
- `golang-cli`
- `golang-documentation`
- CLI help/docs/website parity reference

## Ordered review-loop plan

1. Rebase `feat/78-gitlab-cli-parity` onto `origin/main` after fetching latest refs.
2. Resolve conflicts while preserving GitLab connector parity work:
   - `internal/connectors/defs/gitlab/**`
   - `docs/connectors/gitlab/**`
   - CLI surface metadata/tests/docs/website updates.
3. Run local gates before any push:
   - `gofmt -w cmd internal`
   - `go vet ./...`
   - `go build ./cmd/pm`
   - `go test -timeout 20m ./...`
   - `make connectorgen-validate`
   - `golangci-lint run`
   - `make verify` if time allows
   - CLI parity checks for root/help/connector leaf help/docs.
4. Push rebased branch with `--force-with-lease` only after local gates pass.
5. Request one Claude review pass with `gh pr comment 127 --body "@claude review this PR"`.
6. Collect inline and summary comments, triage every finding, and reply in-thread using the required disposition template.
7. Implement accepted fixes only after disposition analysis, with tests and docs/website parity updates when behavior/docs change.
8. Re-run targeted + broad gates, push reviewed fix commits, request one more Claude pass only if needed, and report final status.

## Safety boundaries

- Never push to `main`.
- No secrets, credentialed GitLab checks, new dependencies, raw generic write/API tools, shell/SQL write tools, or arbitrary GraphQL mutation executors.
- GitLab writes remain reverse ETL plan → preview → approval → execute.
- Destructive/admin/elevated-scope findings are `Needs human` unless already represented as metadata-only safety gates.
