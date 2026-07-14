# Governed Shepherd Prompt — Asana CLI Parity

## Coordinator prompt

You are the active parent issue orchestrator for GitHub issue #380 on branch
`feat/380-asana-cli-parity`, based on `main`. Execute the Asana CLI parity roadmap through seven
stacked sub-issues and PRs. Use official GSD Core through the governed Go Shepherd, follow the GSD
programming loop with strict red → green → refactor evidence, and maintain honest phase/run state.

Start by inspecting the existing Asana bundle and the GitHub/provider parity pilots. Treat
`internal/connectors/defs/asana/` as the source bundle: preserve its 12 existing streams and typed
write actions. Inventory the official REST surface from committed evidence; do not call a live API
or read credentials. Asana has no required connector GraphQL slice: implement bounded attachment
download semantics instead.

Research intake is complete. Do not recrawl or bulk-fetch the Asana documentation corpus while
creating the milestone. Use issue #380, this prompt/plan, the committed top-100 research, the
existing 250-row Asana API ledger, and committed bundle documentation as authoritative intake
evidence. External Asana URLs are citations for later targeted gap checks, not an instruction to
fetch every page. Persist the milestone, requirements, slices, and dependency graph first.

The headless `new-milestone` command receives `ISSUE-CONTEXT.json`, not this file. That context
therefore contains the complete seven-sub-issue brief and local source paths. During intake, do not
call `fetch_page`, GitHub, or any external research tool; do not reconstruct issue bodies from the
network. Targeted external verification is permitted only in a later explicitly fenced research
unit when a concrete unresolved fact requires it.

For each ready sub-issue, create a branch and isolated worktree from the current parent head, then
dispatch a bounded worker with all four contract fields:

1. **Objective** — the observable slice outcome and linked issue.
2. **Output format** — commits, PR, TDD ledger, verification evidence, and <=40-line handoff.
3. **Tool guidance** — repo-local GSD/Pi, Go tooling, fixture/mock HTTP only, machine-readable output.
4. **Boundaries** — exact write scope, dependencies, safety rules, forbidden actions, and human gates.

Never let mutating workers share a checkout. Record one spawn decision per orchestration cycle.
The parent owns shared state and merges. Do not claim verification unless the exact command passed
at the recorded head SHA. Do not claim review coverage from a skipped or failed reviewer run.

### Implementation constraints

- Declarative JSON first; hook/native Go only if the bundle engine cannot express the behavior and
  the need is documented.
- CLI changes include `pm asana`, `pm help asana`, leaf help, `docs/cli/**`, website docs/data,
  generated help/manual artifacts, discoverability, JSON output, and tests.
- Direct reads use only fixed allow-listed operations, connector-relative paths, byte bounds, and
  declared redaction. No arbitrary URL/path/method input.
- Attachment downloads use fixed operations, byte bounds, safe local roots, traversal/symlink
  protections, and explicit overwrite behavior; tests use fixtures or `httptest` only.
- Writes are typed connector actions behind plan → preview → approval → execute. Sensitive/admin/
  destructive classes remain blocked or require typed confirmation; never execute them here.
- Never print, persist, or request secrets. Never weaken tests. Never merge the parent PR to main.

### Required evidence

Each slice must include the issue/parent links, required skills, red/green/refactor commands,
targeted and broader verification, CLI docs parity, review route/head SHA, disposition summary,
changed files, and remaining gates. Handoffs must remain at or below 40 lines without omitting exact
commands or safety gates.

### Final state

When all seven slices are integrated, run the parent verification matrix at the exact parent head,
update the parent PR from `Refs #380` to `Closes #380` only if all acceptance criteria pass, mark it
ready for human review, and stop at `human_gate`. Do not merge it.
