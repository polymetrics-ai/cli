# Required Skill Routing for Agents and Subagents

Use this reference before assigning or executing repo-local agent work. It maps common Polymetrics tasks to the required available skills.

For runtime, RLM, Pi agent, PostgreSQL, DragonflyDB/Redis, Temporal, Podman, or website documentation work, also read `.agents/agentic-delivery/references/runtime-rlm-website-integration.md`.

## Runtime-portable repo skills (.pi/skills — available to BOTH Claude Code and pi workers)

The `golang-*`, `vercel-*`, and design skills referenced below are Claude Code plugins and are
NOT visible to pi/Codex workers. The pi loop's workers load the repo-local equivalents instead,
by reading the SKILL.md paths directly:

- Any Go implementation/review task → `.pi/skills/go-implementation/SKILL.md`
  (+ `references/go-rules.md`: 46 cited rules — Uber/Google style, Go team, Cheney, Ardan Labs, clig.dev)
- Any `website/**` TS/Next.js task → `.pi/skills/ts-website/SKILL.md`
  (+ `references/ts-rules.md`: 40 cited rules — Next.js/Vercel, TkDodo, Epic React, Total TypeScript, Tailwind v4, Radix, Fumadocs, Playwright)
- Any website UI/UX/styling task → `.pi/skills/design-ui/SKILL.md`
  (+ `references/design-rules.md`: 35 cited rules — Web Interface Guidelines, Refactoring UI, WCAG 2.2, thoughtbot, Stripe-docs teardowns)

Workers record which skills they loaded in the TDD ledger / handoff, and cite rule numbers in
review findings and dispositions. Claude Code sessions may load these in addition to the plugin
skills below.

## Always-on Go skill routing

For any Go implementation, review, debugging, CLI, connector runtime, validation, or test task, load:

- `golang-how-to` — orchestrates the correct Go skill set for the task.

Then load the task-specific Go skills below.

### CLI and command behavior

For `pm` CLI commands, flags, help behavior, command groups, output formatting, shell completion, or command tests, load:

- `golang-cli`
- `golang-testing`
- `golang-error-handling`
- `golang-security` when user input, credentials, filesystem paths, command args, or external IO are involved
- `golang-documentation` when help text, manual pages, examples, or generated docs change

If a future slice introduces or edits Cobra/Viper code, also load:

- `golang-spf13-cobra`
- `golang-spf13-viper` when config/env/file layering is involved

### Connector runtime and architecture

For connector engine, hooks, native protocols, direct-read, binary, ETL, reverse ETL, operation ledgers, or declarative bundle architecture, load:

- `golang-design-patterns`
- `golang-structs-interfaces`
- `golang-error-handling`
- `golang-security`
- `golang-safety`
- `golang-testing`
- `golang-context` and `golang-concurrency` when cancellation, goroutines, workers, streams, or channels are involved
- `golang-database` for SQL/CDC/database connector work
- `golang-graphql` for GraphQL connector/runtime work

### Reviews and hardening

For code review, security review, and automated review disposition, load:

- `golang-how-to`
- `golang-security`
- `golang-safety`
- `golang-error-handling`
- `golang-lint`
- `golang-testing`

### Documentation for Go behavior

For doc comments, README/manual docs, CLI docs, generated docs, examples, or docs-only changes that explain Go behavior, load:

- `golang-documentation`
- `golang-cli` when docs describe CLI behavior
- `golang-security` when docs mention credentials, auth scopes, writes, filesystem paths, command args, or data movement

## Design and website skill routing

For website docs, React/Next pages, UI components, documentation UX, accessibility, visual presentation, or docs preview work, load:

- `frontend-design` for production-grade UI/design implementation or visual polish
- `web-design-guidelines` for accessibility/UX/design review
- `vercel-react-best-practices` for React/Next implementation, review, performance, and data fetching
- `vercel-composition-patterns` when changing reusable React component APIs or composition patterns

For CLI docs/website parity work, combine the design skills above with:

- `golang-cli`
- `golang-documentation`
- `.agents/agentic-delivery/references/cli-help-docs-website-parity.md`

## Runtime/RLM/Pi agent dependency routing

For optional runtime services, RLM agent mode, Podman/Docker Compose local runtime, PostgreSQL, DragonflyDB/Redis-compatible coordination, Temporal, `pm runtime`, `pm rlm`, `pm agent image`, `pm worker`, or runtime website docs, read `.agents/agentic-delivery/references/runtime-rlm-website-integration.md` and load:

- `golang-how-to`
- `golang-cli` for CLI surfaces
- `golang-context` and `golang-concurrency` for workers, cancellation, leases, and Temporal-style orchestration
- `golang-database` for PostgreSQL work
- `golang-security` and `golang-safety` for secrets, credentials, command args, paths, and runtime boundaries
- `golang-testing` for runtime-gated tests
- `golang-documentation` for docs/website updates
- website design skills when `website/**` changes

## Required GSD path

Skill loading does not replace GSD. For implementation or behavior-changing work:

1. Read `.agents/agentic-delivery/references/gsd-pi-adapter.md`.
2. Run `/gsd-programming-loop ...` in Pi or `scripts/gsd prompt programming-loop ...` from shell.
3. Record the GSD command path and the skills loaded in the GSD plan, TDD ledger, worker handoff, or PR body.

## PR evidence requirement

Every implementation PR should list:

- GSD command used;
- required Go/design skills loaded;
- CLI help/docs/website parity evidence when applicable;
- verification commands and results.
