## Refs

Closes #4329

## Intent

Make a source-cited connector that explicitly declares `write: false` retain
its provider mutations as non-executable source artifacts. Its supported fixed
read/ETL contracts can therefore validate and materialize without inventing a
write action, request schema, transport, partial command, or source-lock edit.

`Closes #4329`

## Scope and usable-surface delta

- Shared `connectorgen` only: source-import derives the existing named
  `source-cited-non-executable-mutation-foundation-r1` artifact when the locked
  provider operation is mutating, the bundle explicitly has no write
  capability, no actual complete action exists, and the exact provider citation
  is present.
- Complete/implemented actions win; deletes and reverse-ETL are not suppressed.
- The PR adds **0 user commands**. Its usable-surface delta is that a real
  write-disabled source bundle may retain its supported read/ETL commands
  instead of failing all source coverage because provider mutations coexist.
- Sentry and Vercel source-lock acceptance vectors preserve their real URLs,
  SHAs, byte counts, locations, source IDs, methods, and paths. Byte-identical
  fixtures load all 223/400 source-lock operations and retain all 103/237
  classified mutations. Writes/CLI projection bytes remain unchanged.

## TDD and local verification

- **Red:** focused generator tests failed before production code because
  `sourceProjectionApplyWriteDisabledMutationArtifacts` did not exist.
- **Green:** focused source-import/projection coverage passed, including
  write-capable, executable-delete, missing-citation, GraphQL, and full
  Sentry/Vercel source-lock controls.
- Audit fix: automatic artifacts now require explicit
  `metadata.capabilities.write=false` at both projection and validation; a
  red-to-green test prevents a hand-edited `write:true` descriptor bypass.
- PASS: full `cmd/connectorgen` (153.286s), engine (13.126s), commandrunner
  (21.799s), and CLI (441.520s) suites; `go vet ./...`; `go build ./cmd/pm`.
- PASS: `tidy-check`, lint, agent contract, generator validate/surface-sync/
  operation-evidence, connector boundary/canon, release workflow, docs and
  smoke gates; `git diff --check`.
- CLI parity: no `pm` command/flag/help/manual/website surface changed. The
  rebuilt binary has no Sentry/Vercel topic on this foundation-only branch, so
  there are zero new implemented commands to credential-probe and no usability
  claim is made. Source-bound Sentry/Vercel materialization remains downstream
  work.

## Delivery lifecycle and skills

Inline GSD fallback was used because the canonical contract forbids role
spawning: `discuss-phase` → `plan-phase --tdd` → `execute-phase` →
`verify-work` → `code-review`. Evidence is in this phase directory.

Skills used: `golang-how-to`, `golang-cli`, `golang-testing`,
`golang-error-handling`, `golang-security`, `golang-safety`,
`golang-design-patterns`, `golang-structs-interfaces`, and
`golang-documentation`.

## Delivery record

The first independent audit found and this PR fixed the write-capable artifact
bypass. Pending: the fresh audit of the audit-fix SHA, Claude-auto review
result, and required CI status. The remaining built-binary credential-boundary
proof belongs to the named downstream source-bound-read foundation because this
PR deliberately adds zero connector commands. Do not merge from this PR.
