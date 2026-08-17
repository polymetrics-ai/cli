# Context — Issue #3978: final PostgreSQL certification and publication

## Task Delivery Header

- Issue: `Refs #3978 — Postgres Parity: certify parity and publish truthful capabilities`.
- Base branch: `integration/4015-mvp-flat-r1` at `4a0289bcc490d705b12640093f5779df48a28cfe`.
- Merges into: `integration/4015-mvp-flat-r1 → main`.
- Delivery: committed, verified direct PR opened against `integration/4015-mvp-flat-r1`; GitHub API base read-back must match exactly.
- Working branch: `fm/cli-3978-postgres-certify-publish-r1`.
- Target connector: PostgreSQL only. Shared certification code may change only where it consumes definition-owned data and is required to keep the PostgreSQL certification matrix truthful; it must contain no PostgreSQL name branch.
- Task: certify and publish the already-merged PostgreSQL transport and CDC behavior. Do not add a generic writer, direct source-to-target hop, API changefeed sink, broker, MCP, UI, or other core behavior.
- Verification: current-SHA built-binary/live PostgreSQL certification and CDC proofs; four existing warehouse-mediated binary-flow tests; deliberate post-schema scratch failure and restore; sync-mode/matrix drift checks; consumers (`cmd/connectorgen`, `internal/app`, `internal/cli`); repository generator, docs, lint, boundary, and release gates.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Four warehouse-mediated flows remain real-binary | live | Existing built-binary tests independently read destination/warehouse rows, receipts, and checkpoints; writer reports alone are not accepted. |
| All six declared target modes are executable | live | `connectors certify postgres --full --write` independently reads each managed target and requires positive read/load/checkpoint counts. |
| PostgreSQL CDC is publishable | live | Fresh `pm` creates a pgoutput slot, captures the inserted transaction's LSN, independently reads the warehouse row, requires a durable receipt with no pending stage manifest and checkpoint, then observes acknowledgement at/after that LSN. |
| `write` and `cdc` publication is truthful | live | Matrix evidence binds each promoted capability to the exact definition-owned executable route; `query` stays false and the legacy unsupported direct `Write` entry point is never rebranded as a generic writer. |
| Certification can fail after schema validation | live | Scratch-validating `sslmode=bananas` makes the current built certificate exit non-zero; restoring the declaration makes the same run green. |
| Unexecuted work is honest | live | Declared-but-unadapted certification stages retain `unexecutable` rather than `pass`; source-only and unavailable matrix cells retain concrete non-pass reasons. |

## Binding reconciliation

1. #3978's `incremental_dedupe_history` rejection clause is stale. The current PostgreSQL transport declares and PRs #4187/#4199 live-prove all six modes, including `incremental_dedupe_history`.
2. #3978's Podman-only wording is stale. `native/dbtest` supports Docker and this run uses the explicit Colima Docker socket; no runtime is restarted or reconfigured.
3. PR #4199 introduced `certification.json`, the exact `postgres_polling_watermark → postgres_managed_target` adapter, and twelve redacted sync-mode live-evidence records. The current base therefore has live write-flow proof even though the legacy metadata still says `write=false`.
4. PR #4201 introduced current live PostgreSQL 16 pgoutput → durable warehouse receipt → sealed workset → managed PostgreSQL target proof, plus pre-I/O destination `change_capture` refusals. The deferred PostgreSQL CDC→API route remains non-executable.
5. The remaining mismatch is publication bookkeeping: `metadata.json` and native tests publish `cdc=true, write=false, query=false`; the matrix's generic capability cells classify the legacy direct `Connector.Write` stub as unsupported and contain no capability-scoped CDC proof. The six-mode sync matrix has twelve passing live cells but correctly remains globally incomplete because fixture requirements remain false. This task must not call that global false value a failed live transport certificate. The final projection may publish `write=true` only for the definition-owned warehouse transport, never for the direct stub.

## GSD execution and required skills

Resolved inline/manual path: `scripts/gsd doctor`; `scripts/gsd sources` for `discuss-phase`, `plan-phase`, `execute-phase`, `verify-work`, and `code-review`; then the corresponding generated `scripts/gsd prompt` commands. The canonical single-worker contract forbids role spawning, so discussion, TDD planning, execution, verification, and review run inline in this directory.

Loaded skills: `golang-how-to`, `golang-cli`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-testing`, `golang-context`, `golang-concurrency`, `golang-database`, and `golang-lint`. The PostgreSQL/runtime and CLI help/manual/docs/website parity references were read.
