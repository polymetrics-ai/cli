# API → API GitHub transport live proof — context

## Task Delivery Header

| Field | Value |
| --- | --- |
| Primary issue | [#4015](https://github.com/polymetrics-ai/cli/issues/4015) |
| Delivery mode | Direct PR to `integration/4015-mvp-flat-r1` |
| Branch | `fm/cli-quadrant-api-to-api-github-github-r1` |
| Base admission | Branch created directly at `2c48e4deb34128339fccbe5d4b7daad4e13a23e7`, the stated head of PR #4184 |
| Target connector | GitHub only; changed production connector scope is none unless a GitHub-local gap is proven |
| Ownership guard | No PostgreSQL transport or #4184 atomicity edits. A discovered shared foundation gap stops this lane for a separate issue. |
| Header fallback | `.agents/agentic-delivery/contracts/task-delivery-header-template.md` is absent from this integration base. This header records its documented manual fallback fields. |

## Objective

Prove the existing closed GitHub route through a freshly built `pm` binary:

`GitHub issues` → connection-owned WAL/Parquet warehouse → reopen → typed GitHub issue-label action → independent GitHub read-back → durable acknowledgement → checkpoint.

The proof is deliberately narrow. It proves this route, not GitHub write parity:
GitHub's `issue_label_destination` exposes two actions out of 607 declared GitHub
write actions.

## Definition-derived mapping

| Definition source | Selected value | Why it is the valid closed mapping |
| --- | --- | --- |
| `sync_transport.json` source | `issues` | It is an eligible declarative GitHub source stream and its records expose positive integer `number` values. The production source selector admits exactly the configured issue record. |
| `sync_transport.json` mode | `full_append` | The declaration maps it to `append` and the allowed action `add_issue_labels`. |
| `writes.json` apply binding | `target_issue` → `issue_number`; `label` → singleton `labels` | The binding fixes both the target field and payload shape; the carrier accepts no provider-selected write shape. |
| `writes.json` cleanup binding | `target_issue` → `issue_number`; `label` → `name` | The typed inverse is `remove_issue_label`; its declared missing-status handling is part of the edge proof. |

The source issue supplies the bounded, independently read GitHub `issues` record
that admits one staged workset. The destination issue and label are saved
connection configuration consumed solely through the definition-owned bindings;
they are not command arguments or a hand-authored mapper.

## Decisions

- Use `full_append` / `append` / `add_issue_labels`: it is additive, has a typed
  cleanup inverse, and is the precise declaration mapping for the one-record
  GitHub route.
- Use only a repository controlled by the credential owner and only run-owned
  label state. Credentials enter through the normal `pm credentials add ...
  --from-env` boundary and are never emitted or persisted in this evidence.
- The live destination must be read back independently via GitHub's declared
  `issues` reader. The writer's receipt alone is insufficient.
- This is a proof/evidence lane. Existing production code and its real-binary
  test already provide the route. Do not alter production code unless live
  execution demonstrates a GitHub-local defect.

## Required skills and lifecycle

Loaded: `golang-how-to`, `golang-cli`, `golang-testing`,
`golang-error-handling`, `golang-security`, `golang-safety`,
`golang-design-patterns`, `golang-structs-interfaces`, `golang-context`, and
`golang-concurrency`.

`scripts/gsd doctor`, all five `scripts/gsd sources` commands, and
`go run ./cmd/agentcontractgen check` passed. The generated GSD prompts are
executed inline: compatible GSD role spawning is forbidden by the canonical
single-worker contract and this direct-PR task. See `RUN-STATE.json`.
