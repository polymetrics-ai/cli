# Verification: Issue #4307

## Required checks

- [x] Focused red/green engine, commandrunner, generated CLI, and App dispatch tests with `-timeout 20m` (commands and Red/Green record in `TDD-LEDGER.md`).
- [x] Affected package suites for engine, commandrunner, CLI/App, definitions, and `cmd/connectorgen`.
- [x] Two synthetic connector identities covering header and F4 request/response fidelity (`synthetic-alpha` and `synthetic-beta` engine subtests).
- [x] Zero-I/O counters for header rejection classes; existing multipart and bounded-file suites cover their established pre-I/O/no-artifact classes.
- [x] Result-preservation tests prove every ordinary item admitted by each exact declared response contract remains present; credential/transport-secret masking retains field presence with its explicit marker.
- [x] Generic multipart conformance fixtures stage only declaration-owned local assets, bind bounded SHA-256 approval evidence, and reject missing/oversized/tampered/stale inputs before provider I/O.
- [x] Existing GraphQL, scalar/form/SCIM, structured-body, credential/auth, and no-credential preflight regressions.
- [x] `go run ./cmd/connectorgen validate internal/connectors/defs` (552 checked, 0 findings).
- [x] `go run ./cmd/connectorgen surface-sync --check` plus relevant goldens/generated docs checks (no drift).
- [x] `go vet ./...` and `go build ./cmd/pm`.
- [x] `git diff --check`.
- [x] Completion-tracked `make connector-boundary`.
- [x] `make verify`.
- [x] Standard source/security review; no actionable findings.

## CLI help/manual/website parity assessment

The expected CLI change is generated connector-operation help, not a new hand-authored top-level
namespace. Before completion, run and record the applicable generated help/manual/golden and docs
checks. Inspect `docs/cli/**` and `website/**` for an operation declaration/adoption page and update
the authoritative page when the public contract gains typed headers or transfer input/output flags.
Run `pm help <topic>`, `pm <namespace>`, and `pm <command> --help` for any changed generated surface;
record a precise not-applicable result only when no corresponding public command or namespace exists.

## Results

- Focused multipart fixture approval: passed
  `go test -timeout 20m ./internal/connectors/conformance -run 'Test(WriteRequestShape_MultipartFixture|ApprovedFixtureWriteRequestMultipartPayloadDigest)' -count=1`.
- Full conformance package: passed
  `go test -timeout 20m ./internal/connectors/conformance -count=1`.
- Full aggregate verification: passed `make verify`, including the complete Go
  suite, `go vet ./...`, `go build ./cmd/pm`, documentation, help/manual,
  generator, lint, and release-workflow checks.
- Connector boundary: passed `make connector-boundary` and the completion-tracked
  boundary pass inside `make verify` (552 loaded connectors; clean report).
- Source/security review: declaration-bound response/header and multipart paths
  were reviewed for unbounded output, header injection, secret disclosure, and
  approval-digest bypasses; no actionable finding remained. `git diff --check`
  passed after the final lint simplification in `cmd/connectorgen/validate.go`.

## CI repair: generated connector manuals

- PR #4311's `verify` check found the Gmail connector manual stale after its
  attachment download was correctly reclassified as `availability=partial`.
  The same generation pass also reconciled Gorgias's already-declared partial
  binary download.
- Regenerated the canonical output with
  `go run ./cmd/pm docs generate --dir docs/cli`; the only generated changes
  are the Gmail and Gorgias `MANUAL.md` and `SKILL.md` command entries.
- Targeted recheck passed:
  `go run ./cmd/pm docs validate --connectors-dir docs/connectors`.
  `git diff --check` also passed. Full aggregate verification remains owned by
  the outer CI phase.
