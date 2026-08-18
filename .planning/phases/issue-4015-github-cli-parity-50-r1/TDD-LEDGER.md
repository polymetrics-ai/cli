# TDD ledger — issue #4015 GitHub declared-command parity

## Slice 1 — aliases and existing reads

- Red: `go test -timeout 20m ./internal/connectors/commandrunner -run
  '^TestGitHubDeclaredParityVerdicts$' -count=1` failed as intended: all 23 promotion candidates
  still reported `unsupported_api`/`unsupported_local`.
- Green: pending.
- Refactor: pending.

## Slice 2 — new fixed reads

- Red: pending — operation lookup and direct-read tests will fail while the documented global REST
  endpoints and paginated RepositoryOwner query are absent.
- Green: pending.
- Refactor: pending.

## Slice 3 — REST writes and binary download

- Red: pending — tests will assert each command enters the plan lifecycle, rejects missing required
  flags before any network call, and binds to exactly one declared endpoint.
- Green: pending.
- Refactor: pending.

## Slice 4 — retained-command evidence and exact count

- Red: the same focused test asserted 50 unique verdicts and failed on retained rows whose legacy
  note did not name the actual missing media-type, composite, local, upload, verification, or
  capability boundary.
- Green: pending.
- Refactor: pending.

## Live provider proof

- Red: pending — record the pre-change provider state or expected 404 for every disposable write.
- Green: pending — assert the mutation's independent observable state.
- Cleanup: pending — delete/revert and independently assert absence or restored state.

No test receives a secret literal. Provider credentials are environment-only and are never emitted
by test output or stored in evidence.
