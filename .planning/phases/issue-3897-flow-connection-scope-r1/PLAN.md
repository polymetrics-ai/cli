# #3897 — Connection-scoped flow warehouse reads

**Issue:** #3897  
**Parent:** #3988  
**Branch:** `feat/3897-flow-connection-scope` from `feat/3988-github-certification`  
**PR base:** exactly `feat/3988-github-certification`  
**Status:** Planned — production edits have not started

## Delivery mode

The GSD adapter was verified with `scripts/gsd doctor`; `discuss-phase`,
`plan-phase --tdd`, `execute-phase`, `verify-work`, and `code-review` were
resolved using `scripts/gsd sources` and prompt traces. The normal GSD
phase registry is intentionally unavailable because `ROADMAP.md` is an archive
entry with no numbered active phases. This issue therefore uses the approved
inline/manual-GSD fallback documented in `CONTEXT.md`; the delivery contract
forbids spawning GSD roles.

Required skills loaded: `golang-how-to`, `golang-cli`, `golang-testing`,
`golang-stretchr-testify`, `golang-error-handling`, `golang-security`,
`golang-safety`, `golang-context`, `golang-documentation`, `golang-lint`, and
`no-mistakes`.

## Scope

1. Add optional connection selectors to flow query and action source reads.
2. Thread selectors into the real Parquet/DuckDB and `QueryTableRequest`
   paths; do not select by SQL string interpolation.
3. Preserve a typed ambiguity refusal for omitted selectors, with advice that
   names only a manifest field available to the caller.
4. Preserve selected action source identity through JSON serialization and the
   current action-runner/approval boundary.
5. Update runtime help/manual/website/golden surfaces only if public flow
   manifest syntax is described there.

## Explicit exclusions

- #3994 connector action dispatch, approval-token lifecycle, preview, provider
  mutation, and generic HTTP write.
- #3992 scheduling, #3990 rate policies, #3864/#3862 transport.
- Reverse ETL execution or any live provider call.

## TDD slices and checkpoints

### Slice 1 — RED: real warehouse isolation cannot be selected by a flow

**Red:** Add a focused flow/app integration test that materializes `records`
under `acme` and `globex` through normal ETL, then runs query and action source
reads with explicit selectors. Before production code, it must fail because
the selector is not propagated and/or the DuckDB bare view is unavailable.
Assert source row IDs, not only an exit code.

**Green:** Add selector-carrying request types/adapters and scoped DuckDB view
registration. Query `connection: "acme"` and action
`action_cfg.source_connection: "globex"` each return only their own Parquet
rows. The action test uses a local stub runner; it performs no provider write.

**Refactor:** Centralize the source-read request conversion and avoid raw SQL
concatenation for action `source_table`.

**Checkpoint:** commit/push the passing source-read slice.

### Slice 2 — RED: honesty and root-owned table semantics

**Red:** Add assertions that unselected duplicated source tables yield
`*warehouse.AmbiguousTableError` and that their remediation contains no CLI
flag. Add root-owned `_unattributed` fixtures that must not expose a
connection-owned table.

**Green:** Preflight the actual flow source read through warehouse ownership
resolution, decorate only the flow path with `connection` or
`action_cfg.source_connection`, and pass `_unattributed` unchanged to the
warehouse layer.

**Refactor:** Keep `warehouse.FindTable` and `WithAmbiguityRemedy` the one
source of selection and wording mechanics.

**Checkpoint:** commit/push the passing ambiguity/unattributed slice.

### Slice 3 — RED: manifest and action boundary identity

**Red:** Extend manifest parse/serialize and action-runner tests to expect both
selectors. They fail before fields/request propagation exist.

**Green:** Verify query `connection` and action `source_connection` survive
JSON round-trip; verify the action runner receives the unchanged action step
after its source records were read with that selector. If a current preview or
digest accepts source identity, include it in that request; otherwise record
the absent boundary and leave #3994 to own it.

**Refactor:** preserve backwards-compatible omitted fields and no new public
CLI flag.

**Checkpoint:** commit/push the manifest/boundary slice.

### Slice 4 — docs, binary proof, and gate evidence

Update `docs/cli/flow.md`, relevant website flow docs, runtime help, and golden
output only if they describe the selector grammar. Build a fresh binary and
run a local flow query fixture that returns only the requested owner’s rows.
Remove the temporary project root and assert it no longer exists.

## Required verification

- `go test -timeout 20m ./internal/flow`
- `go test -timeout 20m ./internal/app`
- `go test -timeout 20m ./internal/cli`
- targeted `go test -race` for changed flow/app packages where applicable
- `go vet ./...`; focused lint/static checks; docs and golden checks for
  changed surfaces
- `go build -o <temporary-path>/pm ./cmd/pm` plus a fresh local binary
  flow-query proof that checks returned row IDs
- `go run ./cmd/agentcontractgen check`
- individual `make verify` gates appropriate to the change; CI carries the
  full all-package suite when its duration exceeds the command harness limit
- `scripts/gsd prompt verify-work ...`; plan/execute gaps if verification
  finds any; `scripts/gsd prompt code-review ...` with every finding
  dispositioned
- `no-mistakes axi run --intent <complete issue intent> --skip=push,pr,ci`
  without `--yes`, then push/open the draft child PR manually after green.

## CLI parity checklist

- [ ] `pm help flow` checked.
- [ ] `pm flow` bare namespace behaviour checked.
- [ ] `pm flow run --help` checked.
- [ ] `docs/cli/flow.md` updated or explicitly not applicable.
- [ ] matching `website/**` flow documentation updated or explicitly not
  applicable.
- [ ] golden/help fixtures updated or explicitly not applicable.

## Commit and PR record

Every checkpoint commit includes `Refs #3897` and `Refs #3988`. The final
draft PR title uses Conventional Commits and its body records RED/GREEN,
manual-GSD fallback, skills, parity, no-mistakes, Shepherd-equivalent evidence,
review routing, cleanup, and exact branch topology.
