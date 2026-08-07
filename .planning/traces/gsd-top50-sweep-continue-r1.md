# GSD lifecycle trace — cli-top50-sweep-continue-r1

Covers the connectors delivered on `fm/cli-top50-sweep-continue-r1`: **greenhouse** and
**help-scout**, and the standing method for the remaining connectors in the sweep.

## Why this file exists

The lane's dispatch brief prescribed a leaner loop — red test → author → gates → commit → push — and
omitted the GSD command lifecycle. **The project contract in `AGENTS.md` is authoritative over a
brief that omits it**, so the lifecycle is recorded here and the divergence is stated rather than
left implicit. The connector work itself already satisfied the TDD half of the contract (red observed
and captured before every production edit); what was missing was the GSD command trace, the
`scripts/gsd sources` provenance, and the `SUMMARY.md`/`VERIFICATION.md` artifacts. Those are now
present for both delivered phases.

## Adapter health

`./scripts/gsd doctor` — all green, **69 commands** registered:

```
ok	node >=18	v24.13.1
ok	repo root	/Users/karthiksivadas/.treehouse/cli-83d592/20/cli
ok	official docs	.gsd/official-docs/README.md
ok	commands registry	.gsd/commands.json
ok	upstream lock	.gsd/upstream.lock.json
ok	pi settings / extension / skill / prompt
ok	commands	69
```

Note `scripts/gsd` is a Node script with a `#!/usr/bin/env node` shebang — `bash scripts/gsd` fails
with a syntax error. Invoke it as `./scripts/gsd` or `node scripts/gsd`.

Upstream pin: `github.com/open-gsd/gsd-core@20297a8ff941378b8615a5d3e8629e52c10a0f9d`, adapter
`pi-project-local`.

## Command provenance — `scripts/gsd sources <command>`

Each of the five lifecycle commands was resolved before use. All five resolve to the same three
sources:

| Command | Resolved sources |
| --- | --- |
| `discuss-phase` | `.gsd/commands.json` · `.gsd/upstream.lock.json` · `.gsd/official-docs/COMMANDS.md` |
| `plan-phase` | same |
| `execute-phase` | same |
| `verify-work` | same |
| `code-review` | same |

`./scripts/gsd prompt verify-work help-scout-parity-sweep-r1` was generated and followed; it is what
established the `coverage:` frontmatter contract now used in both `SUMMARY.md` files.

## Execution mode — inline/manual fallback, and why

`AGENTS.md` permits inline/manual execution "when the runtime cannot provide compatible isolated
agents or the canonical contract forbids spawning them", and requires the fallback be recorded.

**Fallback taken: inline/manual, for both reasons.**

1. This lane runs under Claude Code, not Pi, so the `/gsd-*` slash commands and their runtime
   subagents are not available. `scripts/gsd prompt` is the documented non-interactive route and was
   used to obtain the authoritative instructions, which were then executed with local tools.
2. The parent-job contract this sweep runs under states the canonical worker processes ready waves
   **inline** and must not spawn an orchestrator, planner, reviewer, verifier, or GSD role.

## Lifecycle per connector

The sweep runs one lifecycle pass per connector. Mapped onto the artifacts:

| GSD stage | Artifact | Status (greenhouse / help-scout) |
| --- | --- | --- |
| `discuss-phase` | `PLAN.md` — artifact, hazards, and the four non-mechanical judgements stated *before* authoring | done / done |
| `plan-phase --tdd` | `PLAN.md` work order + `RUN-STATE.json` shipping `red_confirmed: false` | done / done |
| *(red gate)* | `TDD-LEDGER.md` §2 — **verbatim** failing output from the real bundle | done / done |
| `execute-phase` | the connector commit itself; `RUN-STATE.json` → `green` | done / done |
| `verify-work` | `VERIFICATION.md` + `SUMMARY.md` `coverage:` block | done / done |
| `code-review` | **pending** — runs once, at the end of the sweep, against the whole consolidated PR | pending |

`code-review` is deliberately deferred to the end. The captain ordered **one consolidated PR** for
the sweep, so reviewing per connector would review a diff that is not the one being merged.

## TDD evidence, both phases

**Red observed before green, never assumed.** Both red tests were run against the real, unmodified
bundle and their output captured verbatim.

- **greenhouse** — red committed as its own commit before any production edit.
- **help-scout** — the red test was authored during a window when a corrupted shared Go build cache
  made `go test` fail to *build*. Rather than claim evidence it did not have, the test was committed
  with `red_confirmed: false` and an explicit blocker; **no authoring happened until the run
  completed**, and the observed failure landed in a separate commit that flipped the flag.

`GSD_BASE_REF=origin/main ./scripts/verify-gsd-workflow` passes: implementation changes carry GSD/TDD
evidence.

## Required skills — routed per `.agents/agentic-delivery/references/required-skills-routing.md`

Loaded:

- **`golang-how-to`** — mandatory orchestrator for any Go task. Its routing table sends "write tests"
  to `golang-testing` (plus `golang-stretchr-testify` only if testify is used; this repo is
  stdlib-only, so it is not).
- **`golang-testing`** — applied in *review mode* to the two new test files.

Not loaded, with reasons — the routing doc's connector-runtime cluster
(`golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`,
`golang-safety`, `golang-context`, `golang-concurrency`) targets **connector engine and runtime Go**.
This phase changed **no production Go**: the entire Go delta is two new `_test.go` files in
`cmd/connectorgen`. Everything else is declarative JSON bundle authoring interpreted by the existing
engine. Loading the runtime cluster would have been ceremony, not routing.

### `golang-testing` review-mode findings on the two new tests

| Guideline | Verdict |
| --- | --- |
| Named subtests for table-driven tests | **N/A by shape.** These are not case matrices; they iterate the bundle's own rows. Every failure carries its `METHOD path` key, so a failure is identifiable without `t.Run`. Wrapping 144 rows in subtests would add noise without adding signal. |
| No execution-order dependence | **Pass** — pure file reads, independently runnable. |
| `t.Parallel()` where possible | **Deliberately not added.** Both tests qualify, but no sibling test in `cmd/connectorgen` uses it; matching the surrounding package was preferred over a local improvement that reads as inconsistent. Worth doing package-wide, as its own change. |
| Never test implementation details | **Pass** — the assertions are on the documented-operation contract, which is the observable contract these files exist to hold. |
| Fast unit tests | **Pass** — ~0.5 s for the whole package. |

No changes were required as a result of the review.

## Repo safety overlay — confirmed for both phases

- No secret requested, printed, stored, or summarized.
- No dependency added.
- No credentialed connector check run; every binary invocation was `--help`, which resolves no
  credentials and calls no provider.
- Reverse ETL remains plan → preview → approval → execute; destructive actions carry typed
  confirmation.
- No generic shell, generic HTTP write, or generic SQL write tool exposed.

## CLI help/docs/website parity overlay — **open**

Per `.agents/agentic-delivery/references/cli-help-docs-website-parity.md`, a CLI change is incomplete
until runtime help, `docs/cli/**`, `website/**`, generated help/manual artifacts and tests are
updated or explicitly marked N/A.

**Runtime help is done** — both connectors render full command surfaces, and every command was
invoked to prove it.

**Docs, website and golden transcripts are NOT regenerated per connector**, by design: shared
generated artifacts regenerate **once at the end of the sweep**. Doing it per connector would churn
~1,034 files of pre-existing `main` drift (sweep finding F6) on every commit.

Consequence, stated plainly so it is not discovered late: **`TestGoldenTranscripts/root_bare_manual`
currently fails on this branch.** It was verified pre-existing — it fails with this lane's changes
stashed out, because chatwoot and gmail already added command surfaces the committed transcript
predates. It is carried in both `VERIFICATION.md` files as a known-unmet item and is discharged by
the end-of-sweep regeneration, which must happen before the PR merges.
