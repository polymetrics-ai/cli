# Sub-issue #3689 — refuse non-batchable actions in bulk reverse ETL

> **Filed as [#3689](https://github.com/polymetrics-ai/cli/issues/3689), sub-issue of
> [#3682](https://github.com/polymetrics-ai/cli/issues/3682).** Created under the captain's bounded
> identity exception for issue creation only.

**Title:** `feat(app): refuse non-batchable write actions in bulk reverse ETL plans`

---

## Objective

Enforce the `batchable` declaration (#3684): a write action declared `batchable: false`
must be refused by the bulk reverse-ETL path, and must remain individually executable.

## Operations unblocked

| Operation | Connector | Notes |
| --- | --- | --- |
| `vote` | `reddit` | 1 operation. Unblocked once Reddit declares `batchable: false` on the action in its own lane; this slice makes that declaration mean something. |

## Parent

- Parent issue: #3682
- Branch: `fm/cli-engine-nonbatchable-write-r1`

## Scope

Allowed write scope:

- `internal/app/app.go`, `internal/app/util.go` — the guard
- tests under `internal/app/**` and `internal/connectors/engine/**`

Do **not** edit any connector bundle.

## Design

Two enforcement points, both reading the **live connector manifest** rather than stored plan state:

1. **`PlanReverseETL`** — refuse before a plan, an approval token, or any state mutation exists.
   This is the primary gate: no plan is created, so there is nothing to approve or replay.
2. **`RunReverseETL`**, on the SourceTable branch only — re-check before executing a stored plan.
   Defence in depth for a plan created by an older binary, or a plan whose connector adopted
   `batchable: false` after the plan was made.

Reading the live manifest rather than the stored plan mirrors the existing precedent at
`internal/app/app.go:834`, where `confirmationChallengeForPlan` deliberately prefers the current
manifest "so a local state edit cannot remove a destructive-action confirmation gate from an
already-created plan". The same reasoning applies verbatim here: a hand-edited `state.json` must not
be able to launder a non-batchable action into a bulk run.

### What is deliberately NOT guarded

`PlanConnectorCommand` / `runConnectorCommandPlan` — the `pm <connector> <command>` path. That path
builds a plan for exactly one record from typed CLI flags. It is the human-proxy path the whole
primitive exists to preserve, and it keeps its full plan → preview → approval → execute pipeline.
Guarding it would defeat the purpose.

`commandrunner.BuildWriteCommand` is likewise untouched.

### Error message

The refusal must name the action, the connector, why it was refused, and what to do instead:

```
write action "vote" on connector "reddit" is declared non-batchable and cannot run from a bulk
reverse ETL plan over source table "saved_posts": the connector declares this action must be
invoked one record at a time; run it individually with `pm reddit vote` instead
```

A bare `errors.New("not batchable")` would fail the "clear and actionable" requirement.

## Verification — by executing, not by reading

Both halves must be proven against the real runtime (registry → engine → HTTP), not asserted from
the schema:

- A `batchable: false` action **runs**: plan it as a single connector command, approve it, execute
  it, and observe the real HTTP request arrive at the destination.
- A `batchable: false` action **is refused**: call `PlanReverseETL` against a warehouse table and
  observe the refusal, with no plan persisted and no approval token minted.
- A `batchable: true` (and an absent-field) action still plans and executes in bulk unchanged.

## Required skills

- `golang-how-to`, `golang-error-handling`, `golang-testing`, `golang-security`, `golang-safety`,
  `golang-design-patterns`
