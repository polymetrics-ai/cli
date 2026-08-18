# TDD ledger — issue #4015 GitHub declared-command parity

## Slice 1 — aliases and existing reads

- Red: `go test -timeout 20m ./internal/connectors/commandrunner -run
  '^TestGitHubDeclaredParityVerdicts$' -count=1` failed as intended: all 23 promotion candidates
  still reported `unsupported_api`/`unsupported_local`.
- Green: `TestGitHubDeclaredParityVerdicts` now passes through the real runtime preflight for 25
  commands; fixed GraphQL aliases and existing REST read bindings are executable.
- Refactor: aliases reuse the original operation/write contracts instead of duplicating provider
  behavior.

## Slice 2 — new fixed reads

- Red: `connectorgen validate` initially rejected aliases not named in `api_surface.covered_by` and
  the new RepositoryOwner operation was absent from the generated runtime endpoint ledger.
- Green: the authoritative endpoint coverage now names each alias; `surface-sync` regenerated the
  ledger and focused runtime/connectorgen tests pass. The fixed `github.repo.list` operation has a
  closed variables schema and cursor pagination.
- Refactor: global REST aliases reuse their already-generated provider declarations; only the
  polymorphic user/organization repository list needed a new fixed GraphQL operation.

## Slice 3 — REST writes and binary download

- Red: validation rejected workflow alias string flags against the existing numeric workflow-ID
  record schema, and the legacy autolink create action exposed no request fields. Live autolink
  deletion then proved that persisting provider IDs as numbers changes their spelling to scientific
  notation. `TestGitHubDeclaredParityWriteContracts` failed until those IDs and the variable request
  schemas were corrected.
- Green: workflow and autolink IDs are opaque strings; autolink create and variable writes have
  closed required schemas; REST aliases enter reverse-ETL planning and downloads enter the bounded
  binary executor.
- Refactor: REST mutations reuse existing write actions and downloads reuse fixed operations with
  extraction disabled.

## Slice 3b — fixed binary representation and two additional promotions

- Red: `TestDoStreamSendsDeclaredAcceptMediaType` did not compile because `StreamOptions` exposed no
  declaration-owned media type, while `TestGitHubDeclaredParityVerdicts` failed on `pr diff` and
  `run download` as unsupported.
- Green: `binary_download` now accepts one schema-validated, declaration-owned media type; `pr diff`
  sends `application/vnd.github.diff`, and `run download` reuses the single-artifact bounded ZIP
  operation. Focused connsdk, engine, commandrunner, and bundle validation tests pass.
- Refactor: no caller-supplied header or archive-extraction capability was introduced.

## Slice 4 — retained-command evidence and exact count

- Red: the same focused test asserted 50 unique verdicts and failed on retained rows whose legacy
  note did not name the actual missing media-type, composite, local, upload, verification, or
  capability boundary.
- Green: the 50-row test enforces concrete evidence fragments for all 25 retained declarations and
  passes with the exact 25 + 25 sum.
- Refactor: evidence lives with each command and is also explained provider-by-provider in
  `RESEARCH.md`.

## Live provider proof

- Red: the disposable variable and autolink did not exist; the workflow fixture file returned 404;
  issue and PR state were recorded before mutation; binary destination roots were empty.
- Green: safe reads returned their declared shapes; variable get observed the created value;
  issue pin/unpin and PR draft/ready state were independently queried through fixed GraphQL reads;
  workflow disable/enable states were independently observed; `pr diff` downloaded 217 bytes whose
  content began `diff --git`.
- Cleanup: variable, autolink, issue, workflow file, and temporary PR branch were independently 404;
  PR #56 was restored to closed/non-draft; all temporary binary destination roots were removed.
- Fixture limitations: codespace create reached GitHub but the fixture rejected creation and no
  codespace exists; the fixture has no release assets or workflow artifacts, so their implemented
  binary routes were provider-404 checks with zero filesystem residue. A gist mutation was not run
  because it would violate the authorized organization/repository-only fixture boundary.

No test receives a secret literal. Provider credentials are environment-only and are never emitted
by test output or stored in evidence.
