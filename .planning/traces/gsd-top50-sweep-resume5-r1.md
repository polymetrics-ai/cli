# GSD lifecycle trace — cli-top50-sweep-resume5-r1 (jira)

Lane: `cli-top50-sweep-resume5-r1`. Branch: `fm/cli-top50-sweep-resume2-r1` (continued, not re-cut).
Program: `cli-top50-fixed-schema-sweep-r1`. Connector: **jira**, landing order 4 of 21 under the
captain's largest-first reversal, behind github (1220), workday-rest (907) and zendesk-support (625).

## Lifecycle

`discuss-phase` → `plan-phase --tdd` → `execute-phase` → `verify-work` → `code-review`.

Each command resolved with `./scripts/gsd sources <command>` before use. All five resolve to the
same three provenance files, which is the adapter reporting that the official GSD Core command
definitions back every one of them:

```
$ ./scripts/gsd sources discuss-phase          # identical for plan-phase, execute-phase,
/…/cli/.gsd/commands.json                      # verify-work and code-review
/…/cli/.gsd/upstream.lock.json
/…/cli/.gsd/official-docs/COMMANDS.md
```

`scripts/gsd` is a **Node** script; `bash scripts/gsd` dies with a syntax error (handoff finding 28).

## Execution mode — inline, and why

Inline/manual execution, recorded as the canonical contract requires. The runtime here provides no
isolated agent that can hold a 617-operation derivation plus its artifact, and
`.agents/agentic-delivery/contracts/parent-orchestrator-contract.md` forbids spawning an
orchestrator, shepherd, planner, reviewer, verifier or extra worker for this job: the one canonical
worker processes ready waves inline and persists state in issues, branches, PRs and GSD artifacts.

## TDD

Strict red-first, enforced rather than asserted. `tools/check_red_observed.py` refuses to clear a
connector until `RUN-STATE.json.red_failure` holds real captured `go test` output. jira's red was
written against the bundle as shipped, run, and recorded verbatim in
`.planning/phases/jira-parity-sweep-r1/TDD-LEDGER.md` before any production edit.

## Required skills

`golang-how-to`, `golang-testing`, `golang-cli`, `golang-security`, `golang-safety`,
`golang-error-handling` — per `.agents/agentic-delivery/references/required-skills-routing.md`.

## Deliberate deviation from the dispatch brief

The dispatch brief for this lane said the artifact should be re-fetched and its **byte count**
checked against `MASTER-PLAN.json`. That check is not satisfiable for jira and the reason is a
finding, not a failure — Atlassian serves a rolling snapshot from a single unpinned URL. Recorded in
`PLAN.md` hazard 1 and in the surface test's own doc comment; the artifact sha256 replaces it.
