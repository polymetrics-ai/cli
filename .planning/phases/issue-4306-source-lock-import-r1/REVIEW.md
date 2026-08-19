# Code review — source-lock operation import

## Scope

- `cmd/connectorgen/sourceimport.go`, command registration, source-import tests and synthetic fixtures
- migration adoption documentation and issue #4306 GSD/TDD evidence

## Inline review result

No remaining Critical, Warning, or Info findings.

- Trust boundary: source artifact location is read only from a connector-owned HTTPS lock; user-provided generic transport fields and credentials are absent.
- Integrity: fetched bytes must match both lock byte count and SHA-256 before source parsing; redirect, size, malformed lock, and out-of-sources symlink paths fail closed.
- Schema/identity: only local JSON Pointers resolve under explicit count/depth/cycle controls; dynamic/ambiguous request contracts and duplicate identities are rejected.
- Response preservation: response declarations are retained as resolved provider data; output classification cannot filter fields.
- Cleanup: `make verify` initially caught an unchecked HTTP response-body close. The source fetcher now returns read and close errors, and focused, full-package, and full repository gates passed after the repair.

## Automated review route

No PR exists in this Firstmate handoff lane, so GitHub-hosted Claude/Copilot review cannot yet run. The later PR owner must follow the repository automated-review routing contract; this inline review is not represented as GitHub review coverage.
