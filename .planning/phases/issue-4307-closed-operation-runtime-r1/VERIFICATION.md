# Verification: Issue #4307

## Required checks

- [ ] Focused red/green engine, commandrunner, generated CLI, and App dispatch tests with `-timeout 20m`.
- [ ] Affected package suites for engine, commandrunner, CLI/App, definitions, and `cmd/connectorgen`.
- [ ] Two synthetic connector identities covering header and F4 request/response fidelity.
- [ ] Zero-I/O counters for every header, declaration, approval, multipart, text, status, and output rejection class.
- [ ] Result-preservation tests prove every ordinary item admitted by each exact declared response contract remains present; credential/transport-secret masking retains field presence with its explicit marker.
- [ ] Existing GraphQL, scalar/form/SCIM, structured-body, credential/auth, and no-credential preflight regressions.
- [ ] `go run ./cmd/connectorgen validate internal/connectors/defs`.
- [ ] `go run ./cmd/connectorgen surface-sync --check` plus relevant goldens/generated docs checks.
- [ ] `go vet ./...` and `go build ./cmd/pm`.
- [ ] `git diff --check`.
- [ ] Completion-tracked `make connector-boundary`.
- [ ] `make verify`.
- [ ] Standard source/security review; resolve or explicitly disposition every actionable finding.

## CLI help/manual/website parity assessment

The expected CLI change is generated connector-operation help, not a new hand-authored top-level
namespace. Before completion, run and record the applicable generated help/manual/golden and docs
checks. Inspect `docs/cli/**` and `website/**` for an operation declaration/adoption page and update
the authoritative page when the public contract gains typed headers or transfer input/output flags.
Run `pm help <topic>`, `pm <namespace>`, and `pm <command> --help` for any changed generated surface;
record a precise not-applicable result only when no corresponding public command or namespace exists.

## Results

Pending implementation.
