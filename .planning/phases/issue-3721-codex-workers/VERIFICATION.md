# VERIFICATION — issue #3721 Codex project-local workers

Status: locally verified; awaiting the worker lifecycle's post-commit no-mistakes/PR instruction.

## Checklist

- [x] Canonical source expresses Codex standalone TOML and isolation facts.
- [x] Both target TOML files are generated, parse, and contain all official required fields.
- [x] A test first fails because the worker can reach an ambient agent, then passes with
  `agents.enabled = false`.
- [x] Drift check rejects a changed Codex worker configuration.
- [x] Trust requirement and untrusted fail-closed behavior are documented without inventing a
  trust command.
- [x] Same-filename user/project collision behavior is explicitly treated as undocumented.
- [x] No global `~/.codex` file is touched.
- [x] Focused tests, generation, static checks, and the required inline GSD review route are
  recorded.
- [ ] The child no-mistakes pipeline and sub-PR are intentionally deferred until firstmate gives
  the post-commit instruction required by this worker lifecycle.

## Evidence

- RED evidence: `TestCodexWorkersCannotDelegateToAmbientAgents` first failed because the former
  rendering did not set `agents.enabled = false`; the exact failure is recorded in the TDD ledger.
- GREEN evidence: `go test ./internal/agentcontract ./cmd/agentcontractgen -count=1`,
  `go vet ./internal/agentcontract ./cmd/agentcontractgen`, `go test ./internal/cli -count=1`,
  and `go vet ./internal/cli` passed. The CLI package completed in 398.873s.
- Generation/drift evidence: `go run ./cmd/agentcontractgen sync` generated the two local TOML
  projections, and `go run ./cmd/agentcontractgen check` passed.
- Individual `make verify` gates passed: `tidy-check`, `lint`, `docs-check-no-build`,
  `smoke-no-build`, `agent-contract-check`, `connectorgen-validate`,
  `connectorgen-surface-sync`, `connector-boundary`, and `release-workflow-check`.
- `codex doctor --json` reported a successful configuration load from this worktree. This is a
  configuration-health smoke only; it does not enumerate project agents or execute a live worker.
- Generated TOML is parser-verified and the tests prove that the generated setting is false and
  drift-protected. The official documentation says that setting disables multi-agent tools, but
  this phase does not run a live LLM delegation attempt. Project trust and standalone filename
  collision behavior are likewise documented constraints, not runtime behavior inferred here.

## GSD and review fallback

`scripts/gsd prompt verify-work issue-3721-codex-workers --auto` and
`scripts/gsd prompt code-review issue-3721-codex-workers --auto` were generated and followed
inline. The adapter's standard review role was not spawned because the canonical worker contract
and this task prohibit delegation. The resulting manual review is recorded in `REVIEW.md`.
