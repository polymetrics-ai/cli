# TDD ledger — issue #4015 GitHub declared-command parity

## Slice 1 — aliases and existing reads

- Red: `go test -timeout 20m ./internal/connectors/commandrunner -run
  '^TestGitHubDeclaredParityVerdicts$' -count=1` failed as intended: all 23 promotion candidates
  still reported `unsupported_api`/`unsupported_local`.
- Green: `TestGitHubDeclaredParityVerdicts` now passes through the real runtime preflight for 23
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
  record schema, and the legacy autolink create action exposed no request fields.
- Green: workflow aliases now require numeric IDs; autolink create has a closed required schema;
  REST aliases enter reverse-ETL planning and `release download` enters the bounded binary executor.
- Refactor: REST mutations reuse existing write actions and the download reuses
  `github.release.download_assets` with extraction disabled.

## Slice 4 — retained-command evidence and exact count

- Red: the same focused test asserted 50 unique verdicts and failed on retained rows whose legacy
  note did not name the actual missing media-type, composite, local, upload, verification, or
  capability boundary.
- Green: the 50-row test enforces concrete evidence fragments for all 27 retained declarations and
  passes with the exact 23 + 27 sum.
- Refactor: evidence lives with each command and is also explained provider-by-provider in
  `RESEARCH.md`.

## Live provider proof

- Red: pending — record the pre-change provider state or expected 404 for every disposable write.
- Green: pending — assert the mutation's independent observable state.
- Cleanup: pending — delete/revert and independently assert absence or restored state.

No test receives a secret literal. Provider credentials are environment-only and are never emitted
by test output or stored in evidence.
