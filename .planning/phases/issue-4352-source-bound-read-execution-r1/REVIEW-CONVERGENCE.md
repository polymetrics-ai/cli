# Firstmate exhaustive review convergence — source-bound read repair r4

## Immutable review custody

- Review source: `/Users/karthiksivadas/karthik-agent-workspace/data/cli-source-bound-read-execution-reaudit-codex-r3/report.md`
- PR/issue: [#4356](https://github.com/polymetrics-ai/cli/pull/4356) / Refs #4352.
- Base: `main` at `b33983927d863032dac8220949990506e812937d`; immutable review SHA and PR head: `19b2bd8dc470d6fa92da1a500173c8c8c30ba59c`.
- Branch: `fm/cli-source-bound-read-execution-r1-continuation`, verified locally and remotely at the immutable SHA before this record.
- Delivery: direct-PR; only normal fast-forward commits to the existing PR branch. Do not merge, force-push, alter `main`, source locks, retained bytes, other worktrees, credentials, or provider state.
- GSD: inline/manual `discuss-phase` → `plan-phase --tdd` → `execute-phase` → `verify-work` → `code-review`. The repository contract forbids generic role spawning and no compatible isolated runtime is available; this records the fallback without weakening any gate.

## Architecture and complete lane map

`retained Asana source → sourceprojection → operations.source_operation → cli_surface source_operation/flags → hand parser → commandrunner → App credential boundary → engine OperationDirectRead or Connector.Read → authentication admission → requester`.

The review maps provider operations across ETL, reverse ETL, direct read, direct write, binary download, and binary upload. A source-complete operation must take its exact existing executor lane; a true gap remains declared with the source citation and exact `missing_foundation`. No generic route, method, body, opaque cursor, or write executor is admissible.

## Frozen initial finding set

| ID | Invariant and repair | Required red/green regression | State |
| --- | --- | --- | --- |
| AUDIT-001 | Projection must not emit raw paging `offset` or dead/colliding command `limit`; use only the declared closed page contract. | Red: current generated source-bound Asana flags contain raw controls. Green: projection/help/direct APIs omit them while `--page`/`--page-cursor` retain derived navigation. | open |
| AUDIT-002 | Caller-selected source-bound origin must fail before credential lookup, auth cohort/protected state, or provider I/O. Coordinate rather than duplicate #4351's generic input-ordering repair. | Red: a CLI/App observer sees credential/auth work before rejection. Green: origin error with zero credential reads, auth construction, and requester calls. | open |
| AUDIT-003 | Source-bound ETL source identity, method, path, records, pagination, and origin must be enforced at direct App `Read`/`ReadWithOutcome`, not only CLI preflight. | Red: direct stream literal drift reaches the later boundary. Green: each mismatch rejects before auth/requester use. | open |
| AUDIT-004 | Nineteen DELETEs and two no-body POSTs have the existing reverse-ETL/delete action foundation and must be promoted; absence of local action is not a foundation. | Red: 21 source-complete mutations remain declared blocked. Green: all reach the credential boundary through plan → preview → approval → execute; real named gaps remain. | open |
| AUDIT-005 | Fixed-100 isolated inputs must include every connector its cohort references. | Red: fixed-100 isolated tests name missing Asana input. Green: both tests and full generator package pass deterministically. | open |
| AUDIT-006 | Generated Asana docs/manual/help/website must state actual source-backed counts, implemented lanes, closed pagination, and only real named gaps. | Red: source-derived semantic assertions find stale counts/blockers. Green: regenerated artifacts assert current truth. | open |

## Lens coverage

Architecture/data flow, declaration reachability, happy/bad/edge behavior, CLI/App parity, secret-taint ordering, provider semantics, output integrity, and tests/evidence are complete in the independent audit. State/concurrency and retry/rate-limit/resume are not applicable: this repair introduces no state machine, persistence, goroutine, timer, retry, or rate-limit change. Each frozen finding has a behavioral red/green row in `TDD-LEDGER.md`.

## Fix and re-review protocol

Production work may now address only this frozen set and required generated artifacts. Each dependency group records exact files/symbols, intended regression, observed red failure, green result, and staged file list before its coherent fast-forward checkpoint. A new finding is classified `initial_snapshot_miss`, `fix_created:<sha>`, `requirement_changed:<decision>`, or `new_external_evidence:<source>` and requires another cycle. Final acceptance is a fresh independent Codex audit of the whole original PR change plus repair delta at the pushed code SHA; a later artifact-only audit commit may name that code SHA only if it changes no behavior.
