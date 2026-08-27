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
- Audit M1 repair: omission is not an opt-out. Automatic artifacts require the
  actual JSON member `metadata.capabilities.write: false`; omitted metadata is
  rejected before source-import writes a descriptor. The internal presence bit
  is JSON-hidden and is excluded from the generated certification inventory.
- The PR adds **0 user commands**. Its usable-surface delta is that a real
  write-disabled source bundle may retain its supported read/ETL commands
  instead of failing all source coverage because provider mutations coexist.
- Sentry and Vercel source-lock acceptance vectors preserve their real URLs,
  SHAs, byte counts, locations, source IDs, methods, and paths. Byte-identical
  fixtures load all 223/400 source-lock operations and retain all 103/237
  classified mutations. Writes/CLI projection bytes remain unchanged.
- Current main `1324c52bab0b224ed8958858af7676b8b8e191b4` (#4351) merged
  cleanly with no conflict. The integration regression now reads the real
  Sentry `deleteOrganizationDashboard` and Vercel
  `deleteStorageStoresBlobById` lock entries and proves each declared
  destructive delete/reverse-ETL action wins over artifact retention.

## TDD and local verification

- **Red:** focused generator tests failed before production code because
  `sourceProjectionApplyWriteDisabledMutationArtifacts` did not exist.
- **Green:** focused source-import/projection coverage passed, including
  write-capable, executable-delete, missing-citation, GraphQL, and full
  Sentry/Vercel source-lock controls.
- Audit fix: automatic artifacts now require explicit
  `metadata.capabilities.write=false` at both projection and validation; a
  red-to-green test prevents a hand-edited `write:true` descriptor bypass.
- Audit M1: a second red-to-green regression removes `write` from actual
  Sentry metadata and proves engine loading, source projection, and
  source-import reject it. Sentry/Vercel full-lock and action/delete precedence
  coverage remains green; the certification-matrix guard proves the presence
  bit cannot become a false public capability.
- Audit M1 commands: **RED** `GOFLAGS=-p=3 go test -count=1 -timeout 20m
  ./cmd/connectorgen ./internal/connectors/engine -run
  '^(TestBundleLoadRejectsMetadataCapabilitiesWithoutWrite|TestSourceProjectionAutomaticMutationArtifactRejectsOmittedWriteDeclaration|TestSourceImportCommandRejectsOmittedWriteCapabilityBeforeArtifactAdmission)$'`
  failed before implementation; **GREEN** the identical command passed after
  implementation. **PASS:** the focused Sentry/Vercel/citation/action suite
  (1.142s), full `GOFLAGS=-p=3 go test -count=1 -timeout 20m
  ./cmd/connectorgen` (229.037s), engine (9.800s), commandrunner (22.177s),
  serial CLI (488.445s), `make connectorgen-certification-matrix`, lint,
  runtime-preflight, docs, canon, smoke, `go vet`, `go build`, and
  `git diff --check`.
- Current-main evidence: `git fetch origin main` found
  `1324c52bab0b224ed8958858af7676b8b8e191b4` already equal to this branch's
  merge base, so integration was a no-op with no conflict or discarded work.
  Rebuilt binary witness: in a fresh no-credential project, `./pm github label
  delete --name bug --root <project> --json` exited 1 at `missing
  --credential`. This is the existing source-cited GitHub delete control; it
  proves the shared command foundation reaches its credential boundary without
  provider I/O and does not claim a new Sentry/Vercel user command.
- PASS after current-main integration: focused full-lock acceptance (1.118s),
  full `cmd/connectorgen` (154.269s), engine (14.192s), commandrunner
  (24.871s), runtime-preflight (8.737s), and CLI (438.166s) suites;
  `GOFLAGS=-p=3 go vet ./...`; `GOFLAGS=-p=3 go build ./cmd/pm`.
- PASS: `tidy-check`, lint, agent contract, generator validate/surface-sync/
  declaration-admission/operation-evidence, GitHub parity, certification,
  connector boundary/canon, release workflow, docs and smoke gates;
  `git diff --check`.
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
bypass. The next independent audit found and this PR fixed M1's omitted-write
admission. Pending: the Captain-requested fresh separate Codex audit of the
exact pushed repair SHA, Claude-auto review result, and required CI
status. The remaining built-binary credential-boundary proof belongs to the
named downstream source-bound-read foundation because this PR deliberately adds
zero connector commands. Do not merge from this PR.
