# PLAN — engine primitive: non-batchable write actions

## Scope

Foundation (shared runtime) lane. Branch: `fm/cli-engine-nonbatchable-write-r1`.
Worktree: `/Users/karthiksivadas/.treehouse/cli-83d592/26/cli`.

Issue tree (bodies mirrored in `.planning/phases/engine-nonbatchable-write-r1/subissues/`):

- **#3682** — parent: engine primitive for non-batchable write actions (`PARENT.md`)
- **#3684** — sub-issue: `batchable` declaration capability (`declaration.md`)
- **#3689** — sub-issue: bulk reverse-ETL refusal capability (`enforcement.md`)

#3684 and #3689 are linked as GitHub sub-issues of #3682.

**Identity note.** Issue creation for this program is supposed to use `alfred-polymetrics-ai`, and
no Alfred credential exists in this environment — `gh api user` resolves to `karthik-sivadas` and
`~/.config/gh/hosts.yml` holds no other identity. The bodies were first drafted in-tree and the
issues were then filed under the available account after the captain authorised a bounded exception
covering **issue creation only**. That exception does not extend to approvals, merges, or
branch-protection changes; firstmate still owns every merge.

Allowed write scope:

- `internal/connectors/engine/**` — schema, bundle load, manifest/definition synthesis, test fixture
- `internal/connectors/manifest.go`, `internal/connectors/definition.go`, `internal/connectors/guide.go`
- `internal/app/**` — the guard and its error type
- `docs/**`, `website/**` regeneration
- this phase artifact directory

Explicitly out of scope: **any connector bundle** (`internal/connectors/defs/**`) and **any native
connector** (`internal/connectors/native/**`). The path-ownership guardrail is live on `main` and
rejects connector-owned paths from a shared-runtime lane. Reddit adopts `batchable: false` in its
own lane after this lands.

## GSD mode

- `scripts/gsd doctor`: passed in this worktree (69 commands, adapter healthy).
- `scripts/gsd list`: passed.
- `scripts/gsd prompt plan-phase engine-nonbatchable-write-r1`: available; used for this plan.
- `scripts/gsd prompt programming-loop engine-nonbatchable-write-r1`: **unavailable** —
  `scripts/gsd: unknown GSD command: programming-loop`.
- **Fallback recorded:** manual GSD universal runtime loop, matching the parent-phase precedent in
  `.planning/phases/connector-guardrail-remediation-r1/`. TDD (red-before-green), local
  verification, review, and human-gate requirements are not weakened by the fallback.

## Required skills loaded

- `golang-how-to` (orchestrator) — routed to the set below
- `golang-safety` — zero-value design for the new boolean; defensive copying into manifest specs
- `golang-error-handling` — sentinel vs typed error for the refusal
- Applied from `.agents/agentic-delivery/references/required-skills-routing.md` §"Connector runtime
  and architecture": `golang-design-patterns`, `golang-structs-interfaces`, `golang-testing`,
  `golang-security`, `golang-documentation`

## Problem

`PlanReverseETL` (`internal/app/app.go:634`) accepts any write action name and fans it out over up
to 100,000 warehouse records under one approval token. `writes.json` has no field that can gate an
action out of that path. Some operations must never be bulk-automated — Reddit's API rule permits
"API clients proxying a human's action one-for-one" and forbids "bots deciding how to vote on
content or amplifying a human's vote" — and the engine currently cannot express the distinction.

## Design

### 1. Declaration

`"batchable"` boolean on each `writes.json` action, `"default": true` in the JSON schema.

Go representation is `Batchable *bool` with an `IsBatchable()` accessor, on three types:

| Type | Package | Purpose |
| --- | --- | --- |
| `engine.WriteAction` | `internal/connectors/engine` | parsed bundle |
| `connectors.WriteActionSpec` | `internal/connectors` | manifest (what the app reads) |
| `connectors.WriteActionInfo` | `internal/connectors` | definition (`pm connectors describe`) |

**Why `*bool` and not `bool`:** Go's zero value for `bool` is `false`, and `false` means
*non-batchable* — the restrictive setting. A plain `bool` would silently mark every
hand-constructed `WriteAction{...}` literal, every native connector's `WriteActionSpec{...}`
literal (e.g. `internal/connectors/native/amazon-sqs/writer.go:72`), and every test fixture as
non-batchable, inverting the default at exactly the call sites that never opt in. `nil` → batchable
gives a safe zero value for JSON-absent, Go-zero, and test-constructed values alike, which is the
zero-value-design rule from `golang-safety`.

The pointer is **copied by value**, never shared, when crossing from bundle to manifest — same
defensive-copy discipline as the adjacent `append([]string(nil), a.RedactFields...)`.

### 2. Enforcement

Two gates, both reading the **live connector manifest** rather than stored plan state:

1. `PlanReverseETL` — refuse before a plan row, an approval token, or any state mutation exists.
2. `RunReverseETL`, SourceTable branch only — re-check before executing a stored plan.

Reading the live manifest mirrors the existing precedent at `internal/app/app.go:834`, where
`confirmationChallengeForPlan` deliberately prefers the current manifest "so a local state edit
cannot remove a destructive-action confirmation gate from an already-created plan". A hand-edited
`state.json` must likewise not be able to launder a non-batchable action into a bulk run.

**Not guarded, deliberately:** `PlanConnectorCommand`, `runConnectorCommandPlan`, and
`commandrunner.BuildWriteCommand` — the `pm <connector> <command>` path. That path builds a plan for
exactly one record from typed CLI flags and keeps its full plan → preview → approval → execute
pipeline. It is the human-proxy path the primitive exists to preserve.

### 3. Error

A typed `app.NonBatchableActionError` carrying connector, action, source table, and (when the
connector's command surface declares one) the individual command to use instead. Typed rather than
sentinel because it carries data — the `golang-error-handling` decision rule — and it mirrors the
existing `commandrunner.BlockedCommandError` precedent in this repo.

### 4. Decision — batchability and confirmation stay separate concerns

The brief asked whether `batchable: false` should also force per-invocation confirmation. **It
should not.** `confirm: "destructive"` already exists and answers a different question: *how severe
is the consequence if this single call is wrong?* `batchable: false` answers: *may this action be
fanned out over N records under one approval?*

Coupling them breaks in both directions:

- A vote is not destructive. Forcing `--confirm destructive` on every `pm reddit vote` would make
  the human-proxy command hostile to the human it exists to serve.
- Bulk deletes are destructive but legitimately batchable, and several bundles already ship
  `confirm: "destructive"` bulk delete actions. Deriving batchability from `confirm` would break
  them.

They are orthogonal: an action may declare either, both, or neither.

### 5. No raw request escape hatch

The declaration is one typed boolean read from a bundle and surfaced through the existing manifest.
It adds no field through which a caller could supply a method, path, or body. The refusal path only
*removes* reachability; it never widens it.

## Task sequence

1. Red: engine loader tests for `batchable` true/false/absent/malformed. — `internal/connectors/engine/bundle_test.go`
2. Green: schema field + `WriteAction.Batchable` + `IsBatchable()`.
3. Red: manifest/definition propagation tests.
4. Green: `WriteActionSpec` / `WriteActionInfo` fields + `synthesizeManifest` / `synthesizeDefinition` copy.
5. Red: app guard tests — plan refused, no plan persisted; execute-time re-check; batchable unaffected.
6. Green: `NonBatchableActionError` + guards in `PlanReverseETL` and `RunReverseETL`.
7. Red: end-to-end test that a non-batchable action still executes individually against a real HTTP server.
8. Green: confirm no change needed on the command path (guard placement proven correct).
9. Help/manual parity line in `guide.go`; docs; regenerate website catalog.
10. Local gates + binary execution evidence.

## Verification strategy — execute, do not read

The defect class this must not join: 174 commands declared `availability: implemented` that fail at
runtime, because a validator exempted a check the runtime enforces. A schema field with no
enforcement is the same failure. So both halves are proven by running the real runtime — registry →
engine → HTTP — not by asserting on the schema.

Constraint: connector bundles are embedded at compile time (`internal/connectors/defs/defs.go`
`//go:embed`), and this lane may not edit one. So the `batchable: false` subject is an **engine test
fixture bundle** under `internal/connectors/engine/testdata/bundles/`, loaded through the real
`engine.Load` + `engine.New` and registered into the real `app` registry via the exported
`App.Registry().Register`. Same loader, same manifest synthesis, same `PlanReverseETL`, same
`connectors.Write` path, real `httptest` destination. The only thing not exercised through the
`pm` process boundary is bundle embedding, which is not what the guard depends on.

The `pm` binary is separately built and run to prove: existing connectors are unchanged, the bulk
plan path still works for a batchable action, and the new field surfaces in
`pm connectors inspect --json` / `pm help <connector>`.

## Human gates

Do not merge. Do not clear draft status. Firstmate merges.
