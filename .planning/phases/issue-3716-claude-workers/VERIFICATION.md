# VERIFICATION — issue #3716 clean project-local Claude workers

Status: review corrections applied; the outer executor owns remaining delivery gates.

## Checklist

- [x] Exactly two required `.claude/agents` projections exist.
- [x] Both are complete generated files derived from the canonical source.
- [x] YAML frontmatter has required `name` and `description`, explicit base tools, scoped required
  skills, an `Agent`/`Task` denylist, and `permissionMode: default`.
- [x] Neither worker grants `Agent`, bare `Skill`, MCP, an orchestration persona, or an unlisted
  fork-capable skill route.
- [x] Mutating a generated file to grant `Agent` fails the real drift check; sync restores it.
- [x] Installed Claude CLI smoke discovers both project workers by name in this trusted checkout.
- [x] Prior live ambient CLI/plugin fixtures demonstrate direct `Agent` delegation is blocked for
  the selected project worker.
- [ ] **NOT PERFORMED:** a real authenticated Claude session selecting each role from a clean
  trusted home containing unrelated global agent and skill definitions.
- [x] Documentation records project discovery, precedence, scoped skill behavior, and managed/CLI
  override caveats using https://code.claude.com/docs/en/sub-agents and
  https://code.claude.com/docs/en/slash-commands.
- [x] No global Claude, installed plugin, `.codex`, `.pi`, legacy-role, or connector path changed.
- [ ] Outer no-mistakes, push, PR, and CI phases are complete.

## Isolation evidence and boundary

The generated YAML `tools` field deliberately omits `Agent`, bare `Skill`, and MCP tools.
`disallowedTools` also removes `Agent` and its legacy `Task` alias. The official Claude Code
subagents documentation states that omitting `Agent` blocks spawning through that tool. The
official skills documentation states that bare `Skill` reaches project, user, plugin, and bundled
skills, while `context: fork` executes a skill in a subagent. The generator therefore emits only
scoped `Skill(name)` rules and the validator rejects any other skill surface.

The reachable Go skill names are `golang-cli`, `golang-concurrency`, `golang-context`,
`golang-database`, `golang-design-patterns`, `golang-documentation`, `golang-error-handling`,
`golang-graphql`, `golang-how-to`, `golang-lint`, `golang-safety`, `golang-security`,
`golang-spf13-cobra`, `golang-spf13-viper`, `golang-structs-interfaces`, and `golang-testing`.
Each Go name is allowed both directly and through its `cc-skills-golang:` plugin-qualified alias.
The reachable design names are `frontend-design`, `frontend-design:frontend-design`,
`vercel-composition-patterns`, `vercel-react-best-practices`, and `web-design-guidelines`. No other
skill name is reachable through the worker's Skill tool.

The allowed skill definitions available in the review environment were checked for alternate
spawning mechanisms. None uses `context: fork`, an `agent` frontmatter field, or a Bash instruction
that launches Claude. Some Go skills include `Agent` in their own `allowed-tools` metadata or
recommend subagents for broad audits; the worker-level denylist removes `Agent` and `Task`, so those
grants and instructions cannot restore a delegation tool. Unlisted agent-oriented skills are
outside the scoped surface.

Claude Code v2.1.222 was also run from this trusted repository with an ambient CLI fixture and a
temporary local plugin fixture available. A prompt requiring `pm-delivery-worker` to use `Agent`
to invoke either fixture returned `AGENT_UNAVAILABLE` without a tool invocation. This is direct
runtime evidence that the selected project worker cannot delegate to ambient agents.

## Clean-home runtime boundary

The clean-home runtime smoke is **NOT PERFORMED**. It requires a real authenticated Claude session
in a clean trusted home containing unrelated global definitions and must select both project roles
by name. The earlier isolated home had no login and stopped before model execution; no credentials
were copied or supplied. That result is not runtime precedence evidence.

Static proof is separate: the generated definitions omit `Agent` and bare `Skill`, documented
precedence puts project agents above user/plugin agents, and canonical full-file drift enforcement
matches both project files. Managed definitions and CLI `--agents` remain higher-precedence and can
replace a same-name project definition; this harness cannot prevent that override.

## Wave 1 dependency disposition

The captain accepted Wave 1 PR #3724 after 23 successful checks, 6 skipped checks, and zero
failures. There was no separate automated code-review pass. That review-coverage gap is known and
captain-approved; Wave 2 does not retrofit coverage or block on it.

## Historical local commands before review corrections

- `gofmt -w internal/agentcontract cmd/agentcontractgen`
- `go test ./internal/agentcontract ./cmd/agentcontractgen -count=1`
- `go vet ./internal/agentcontract ./cmd/agentcontractgen ./internal/cli`
- `go test ./internal/cli -count=1`
- `go build ./cmd/pm`
- `make tidy-check`, `make docs-check-no-build`, `make smoke-no-build`, and `make lint`
- `make agent-contract-check`, `make connector-boundary`, and `make release-workflow-check`
- `make connectorgen-validate` and `make connectorgen-surface-sync`
- `scripts/verify-gsd-workflow origin/main`

These recorded results predate the current review corrections and do not substitute for the outer
executor's remaining validation phases.

## CI corrective loop

PR #3728 initially failed `connector-boundary` because the neutral
`HarnessPolicyFor` identifier in `internal/agentcontract` matched the connector lexicon's
provider-policy heuristic. This was RED evidence for the boundary rule, not a connector-policy
change. The lookup was renamed to `ProjectionFor` without changing the canonical contract or
generated output.

After the rename, `go test ./internal/agentcontract ./cmd/agentcontractgen -count=1`,
`go run ./cmd/agentcontractgen check`, `go vet ./internal/agentcontract ./cmd/agentcontractgen`,
and `go run ./cmd/connectorgen boundary . --json` passed. The boundary report covered 145 Go files
and 550 connector definitions with `outcome: clean`; a fresh no-mistakes child pipeline and PR CI
remain required for the corrected commit.
