# VERIFICATION — issue #3716 clean project-local Claude workers

Status: review corrections applied; the outer executor owns remaining delivery gates.

## Checklist

- [x] Exactly two required `.claude/agents` projections exist.
- [x] Both are complete generated files derived from the canonical source.
- [x] YAML frontmatter has required `name` and `description`, explicit base tools, trusted
  plugin-qualified skill preloads, an `Agent`/`Task`/`Skill` denylist, and
  `permissionMode: default`.
- [x] Neither worker grants `Agent`, runtime `Skill`, MCP, an orchestration persona, or an unlisted
  `context: fork` route.
- [x] The canonical check recursively rejects extra definitions, duplicate names, symlinks, and
  canonical name/path mismatches across every repository-local `.claude/agents` scope.
- [x] Inventory prunes only exact root `.git` metadata; `.GIT`, `.Git`, and nested paths remain in
  the discoverable scope check.
- [x] The Claude renderer emits LF-canonical expected bytes, collapses repeated carriage returns
  before LF to the same idempotent fixed point, preserves bare carriage returns, and sync treats
  canonical-source or working-tree EOL-only differences as current.
- [x] Mutating a generated file to grant `Agent` fails the real drift check; sync restores it.
- [x] Installed Claude CLI smoke discovers both project workers by name in this trusted checkout.
- [x] Prior live ambient CLI/plugin fixtures demonstrate direct `Agent` delegation is blocked for
  the selected project worker.
- [ ] **NOT PERFORMED:** a real authenticated Claude session selecting each role from a clean
  trusted home containing unrelated global agent and skill definitions.
- [x] Documentation records project discovery, precedence, qualified preload behavior, unavailable
  skill cost, and managed/CLI override caveats using https://code.claude.com/docs/en/sub-agents and
  https://code.claude.com/docs/en/slash-commands.
- [x] No global Claude, installed plugin, `.codex`, `.pi`, legacy-role, or connector path changed.
- [ ] Outer no-mistakes, push, PR, and CI phases are complete.

## Isolation evidence and boundary

The generated YAML `tools` field deliberately omits `Agent`, `Skill`, and MCP tools.
`disallowedTools` also removes `Agent`, its legacy `Task` alias, and `Skill`. The official Claude
Code subagent documentation states that omitting `Agent` blocks spawning through that tool, that
the `skills` field preloads full skill content, and that omitting `Skill` prevents runtime skill
invocation. The official skills documentation states that plugin-qualified names cannot collide
with personal or project skills and distinguishes preloading from invoking a `context: fork`
skill.

The trusted preloads are the sixteen required Go identifiers under `cc-skills-golang:*` and
`frontend-design:frontend-design`. The validator rejects every unqualified or additional preload.
`vercel-composition-patterns`, `vercel-react-best-practices`, and `web-design-guidelines` have no
trusted plugin-qualified source in the reviewed Claude environment and are intentionally
unavailable. The cost is explicit: a website/docs UI job requiring them cannot satisfy repository
skill routing in this worker and must preserve state for a captain-approved harness with trusted
plugin packaging.

Official documentation demonstrates namespace collision resistance. Static tests demonstrate that
the canonical source and both generated files contain only the qualified identifiers and that
`Agent`, `Task`, and `Skill` are denied. Read-only environment inspection found enabled
`cc-skills-golang@skills-dir` v1.6.0 and `frontend-design@claude-code-plugins` v1.1.0 sources, but the
repository does not pin those installed plugin versions. No authenticated runtime preload
resolution was performed in this review correction, so source/version selection at launch remains
configured rather than demonstrated.

Claude Code v2.1.222 was also run from this trusted repository with an ambient CLI fixture and a
temporary local plugin fixture available. A prompt requiring `pm-delivery-worker` to use `Agent`
to invoke either fixture returned `AGENT_UNAVAILABLE` without a tool invocation. This is direct
runtime evidence that the selected project worker cannot delegate to ambient agents.

## Clean-home runtime boundary

The clean-home runtime smoke is **NOT PERFORMED**. It requires a real authenticated Claude session
in a clean trusted home containing unrelated global definitions and must select both project roles
by name. The earlier isolated home had no login and stopped before model execution; no credentials
were copied or supplied. That result is not runtime precedence evidence.

Static proof is separate: the generated definitions omit `Agent` and runtime `Skill`, documented
precedence puts project agents above user/plugin agents, repository-wide recursive inventory
requires exactly the two canonical definitions across all nested project scopes, and full-file
drift enforcement matches LF-canonical expected and actual files. Inventory excludes only exact
root `.git` metadata; ordinary case variants remain checked. Managed definitions and CLI
`--agents` remain higher-precedence and can replace a same-name project definition; launch
configuration can also select a different plugin installation. This harness cannot prevent those
overrides.

## Review-correction verification

- Focused command: `go test ./internal/agentcontract ./cmd/agentcontractgen -count=1`.
- Result: **PASS** — `internal/agentcontract` and `cmd/agentcontractgen` both passed.
- The first focused run exposed an incorrectly scoped `WalkDir` error variable and failed to build;
  that local defect was corrected before the final passing fix-round run.
- The nested-scope/CRLF review round adds repository-depth inventory cases, a nested scope symlink
  case, direct CRLF parsing, accepted CRLF drift checks, and no-op CRLF sync to the same focused
  package command; the result is **PASS**.
- The metadata/EOL review round adds `.GIT` and `.Git` inventory cases, exact root `.git` pruning,
  canonical CRLF rendering, successful checking, and zero-update repeated sync; the same focused
  package command passes.
- The fixed-point EOL review round adds direct repeated-run idempotence cases and canonical
  `\r\r\r\n` render/check/sync coverage; the same focused package command passes.
- Windows runtime execution was not performed; portability is source- and fixture-level evidence
  from slash-native contract validation, LF-canonical rendering/CRLF-normalized comparison, and
  Node-mediated adapter invocation.

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
