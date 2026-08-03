# Parent issue #3682 — engine primitive: non-batchable write actions

> **Filed as [#3682](https://github.com/polymetrics-ai/cli/issues/3682).** The captain authorised a
> bounded identity exception for issue creation only, since no `alfred-polymetrics-ai` credential
> exists in this environment. That exception does not extend to approvals, merges, or
> branch-protection changes.

**Title:** `feat(engine): let a write action declare itself non-batchable`

**Labels:** `connector-architecture-v2`, `engine`, `foundation`

---

## Objective

Add a small, generic engine primitive that lets a write action declare itself **non-batchable**, so
it can never be executed from a bulk reverse-ETL plan, while remaining fully executable as its own
`pm <connector> <command>`.

## Operations unblocked

| Operation | Connector | Notes |
| --- | --- | --- |
| `vote` | `reddit` | 1 operation, unblocked indirectly. Reddit adopts the primitive in its own lane after this lands. |

This foundation PR unblocks exactly **one** operation today. It is filed anyway because it is the
only compliant path to a command the captain explicitly asked for, and because the declaration is
generic — any operation that must never be bulk-automated can adopt it.

## Background — verified against the code, not assumed

The captain asked for `pm reddit vote`. Reddit's API rule permits *"API clients proxying a human's
action one-for-one"* and forbids *"bots deciding how to vote on content or amplifying a human's
vote"*. A human typing `pm reddit vote` is the first; a scheduled reverse-ETL sync over a warehouse
table is the second.

The engine cannot currently express that distinction. The Reddit lane verified both candidate paths
and correctly refused each:

1. **`writes.json` has no field gating an action out of bulk reverse ETL.** `PlanReverseETL`
   (`internal/app/app.go:634`) accepts any action name for up to `100000` records under a single
   approval token. There is no allowlist, denylist, or per-action batch policy anywhere in the
   plan path.
2. **`operations.json`'s `rest_write` kind is schema-declared with zero execution path.** Only
   `rest_read` is wired to a runtime. Shipping `vote` as a `rest_write` operation would manufacture
   exactly the declared-implemented-but-fails-at-runtime defect class that tonight's audit found
   **174 instances of** across this repo.

So the correct move is neither path: it is to give the engine the missing vocabulary first.

## What this parent delivers

A `"batchable"` boolean on write actions in `internal/connectors/engine/schema/writes.schema.json`,
defaulting to **true** so every one of the existing write actions is unchanged, plus enforcement in
the bulk reverse-ETL path that refuses any action declared `batchable: false` with a clear,
actionable error.

## Sub-issues

| Sub-issue | Capability | Operations unblocked |
| --- | --- | --- |
| [#3684](https://github.com/polymetrics-ai/cli/issues/3684) — declaration | `batchable` declaration on write actions: schema field, bundle load, manifest/definition surface, help + docs parity | none directly — it is the vocabulary the enforcement slice and `reddit vote` both depend on |
| [#3689](https://github.com/polymetrics-ai/cli/issues/3689) — enforcement | Bulk reverse-ETL refusal guard at plan time and execute time, with an actionable error | `reddit vote` (1 operation, `reddit`) once Reddit adopts `batchable: false` in its own lane |

## Non-goals

- **No connector bundle edits.** This is a foundation PR. The path-ownership guardrail is live on
  `main` and rejects connector-bundle edits from a shared-runtime branch. Reddit adopts the
  declaration in its own lane afterwards.
- **No raw request escape hatch.** The declaration is a single typed boolean read from the bundle
  and surfaced through the existing manifest. Nothing here lets a caller supply a method, path, or
  body.
- **No new confirmation semantics.** See the design decision below.

## Design decision — batchability and confirmation are separate concerns

The brief asked whether `batchable: false` should also force per-invocation confirmation. It should
not, and they are kept orthogonal.

`confirm: "destructive"` already exists on write actions and answers a different question: *how
severe is the consequence if this single call is wrong?* It gates one invocation behind a typed
challenge string. `batchable: false` answers: *may this action be fanned out over N records under
one approval?* It gates the shape of the plan, not the severity of the call.

Coupling them would be wrong in both directions:

- A `vote` is not destructive. Forcing `--confirm destructive` on every `pm reddit vote` would make
  the human-proxy command hostile to the exact human it exists to serve — and the captain's stated
  ground for the command is that PM serves humans, not only agents.
- A bulk `delete` is destructive but legitimately batchable; several connectors already ship
  `confirm: "destructive"` bulk deletes today. Deriving batchability from `confirm` would silently
  break them.

Because they are independent, an action may declare either, both, or neither, and the existing
approval/preview/confirmation pipeline is untouched. The reverse-ETL invariant that every write goes
through plan → preview → approval → execute still holds for non-batchable actions; the guard only
refuses the *SourceTable-driven bulk* plan shape, never the single-record connector-command plan.

## Acceptance

- `batchable` is declarable in `writes.json`, defaults to `true`, and is rejected by the loader when
  malformed.
- Every existing write action across all bundles is unchanged — no bundle edits, no behavior change.
- A `batchable: false` action **is refused** when a bulk reverse-ETL plan targets it, with an error
  naming the action, the connector, and the individual command to use instead.
- A `batchable: false` action **still executes** end-to-end when invoked individually.
- Both halves proven by executing the runtime, not by reading the schema.
- Help, docs, and website catalog parity verified.

## Human gates

Do not merge and do not clear draft status. Firstmate merges.
