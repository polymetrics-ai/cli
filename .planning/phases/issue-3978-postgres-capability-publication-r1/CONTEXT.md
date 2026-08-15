# Context — Issues 3978 and 3977: truthful PostgreSQL capability publication

## Task Delivery Header

- Issues: `Closes #3978` and `Refs #3977`.
- Base branch: `integration/4015-mvp-flat-r1`.
- Merges into: `integration/4015-mvp-flat-r1 -> main`.
- Delivery: direct PR against `integration/4015-mvp-flat-r1`; do not use no-mistakes.
- Working branch: `fm/cli-pg-capability-publication-club-r1`.
- Target connector: PostgreSQL only.
- Task: publish `write: true`, `cdc: true`, and `query: false` only after the production application can dispatch PostgreSQL change capture into its connection-owned warehouse and the existing managed-target write driver remains behaviorally proven.
- Required verification: focused app/connector/CLI tests; explicit Docker/Colima `databaseintegration` proof; generated connector/docs parity; scoped repository gates; CI.

The task-delivery-header template named by `AGENTS.md` is absent from this integration base. This context records the required fields as the repository's documented manual fallback.

## Locked decisions

1. A public capability flag is executable truth, not metadata intent. PostgreSQL CDC must reach the production `App.RunETL` path, durably publish a whole source transaction to the connection-owned warehouse, persist its full checkpoint, and only then let the native reader acknowledge the LSN.
2. The #3977 amendment is authoritative: PostgreSQL CDC -> warehouse is the only admitted source route. There is no direct CDC-to-target dispatch and no polling/cursor fallback.
3. PostgreSQL write publication relies on the already-merged managed-target driver, sealed workset delivery, approval, receipt, and live row-state proofs. This slice does not invent a generic connector `Write`, arbitrary SQL, or a direct source-to-target route.
4. `query` stays false. `incremental_dedupe_history`, arbitrary existing targets, physical-absence deletes, generic SQL, and API changefeed sinks remain explicit non-support.
5. Generated catalog, connector docs, website data, and golden transcripts are regenerated through repository generators when their source surfaces move. No derived artifact is hand-authored.
6. #4125 and #4158 are excluded even if their base failures appear during broad verification.

## GSD execution and required skills

The installed `discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and `code-review` prompts were resolved with `scripts/gsd`; `scripts/gsd doctor` and `go run ./cmd/agentcontractgen check` pass. This issue slug is not a numbered roadmap phase and the canonical contract requires one worker, so the lifecycle runs inline and is recorded in this directory.

Loaded skills: `golang-how-to`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-testing`, `golang-database`, `golang-context`, `golang-concurrency`, `golang-cli`, `golang-documentation`, and `golang-lint`. The PostgreSQL/runtime and CLI help/docs/website parity references were read.

The branch was paused while #4156 landed the shared definition-owned transport declaration and
`change_capture` route. After rebasing onto base commit `11fd27b4b`, its generic declared transport
path was preserved as the first dispatch choice. PostgreSQL's separate residual remained: a native
`ChangefeedExecutor` with no closed source/destination transport still had no production route into
the connection-owned warehouse. This change closes only that residual and does not modify shared
transport declaration or validation.
