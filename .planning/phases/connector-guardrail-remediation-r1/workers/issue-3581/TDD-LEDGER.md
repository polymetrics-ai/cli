# TDD LEDGER — issue 3581 target-scope core validator

## Skill / GSD evidence

- GSD command path: `scripts/gsd prompt execute-phase connector-guardrail-remediation-r1 --dry-run`.
- Manual-GSD fallback: `scripts/gsd prompt programming-loop init --phase connector-guardrail-remediation-r1 --dry-run` returned `scripts/gsd: unknown GSD command: programming-loop`.
- Loaded skills: `gsd-core`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-documentation`, `golang-lint`, `no-mistakes`.
- Missing local skill note: `.pi/skills/go-implementation/SKILL.md` not present; repository-routed Go skills loaded instead.

## Planned red tests before production edits

| Case | Test target | Expected red reason | Status |
| --- | --- | --- | --- |
| exactly-one slug | `internal/connectors/boundary` | scope contract API missing | Planned |
| auto-detect without label/scope marker | `internal/connectors/boundary` | inference API missing | Planned |
| shared runtime rejection | `internal/connectors/boundary` | ownership validator missing | Planned |
| unrelated connector rejection | `internal/connectors/boundary` | target-aware validation missing | Planned |
| unrelated generated rejection | `internal/connectors/boundary` | generated/docs classes missing | Planned |
| allowed target defs/fixtures/docs | `internal/connectors/boundary` | allowlist classes missing | Planned |
| narrow shared indexes/goldens | `internal/connectors/boundary` | closed shared class missing | Planned |
| connector-lane exception/config cannot weaken gate | `internal/connectors/boundary` | validator config protection missing | Planned |
| CLI JSON/help | `cmd/connectorgen` | `ownership` command missing | Planned |

## Actual evidence log

- Red 2026-08-02: `go test ./internal/connectors/boundary ./cmd/connectorgen` failed before production edits with missing ownership API/CLI:
  - `internal/connectors/boundary/ownership_test.go:15:12: undefined: ValidateOwnership`
  - `internal/connectors/boundary/ownership_test.go:15:36: undefined: OwnershipOptions`
  - `internal/connectors/boundary/ownership_test.go:66:37: undefined: RuleOwnershipSharedPath`
  - `cmd/connectorgen/ownership_test.go:28:22: undefined: boundary.OwnershipReport`
  - package result: `FAIL polymetrics.ai/internal/connectors/boundary [build failed]`; `FAIL polymetrics.ai/cmd/connectorgen [build failed]`.
- Green 2026-08-02: focused implementation gates passed after production edits.
  - `gofmt -w cmd internal`
  - `go test ./internal/connectors/boundary ./cmd/connectorgen`: `ok   polymetrics.ai/internal/connectors/boundary 53.875s`; `ok   polymetrics.ai/cmd/connectorgen 10.227s`
  - `go run ./cmd/connectorgen ownership . --help`: printed usage and exit status contract
  - `go build ./cmd/pm`: exited 0
- JSON smoke 2026-08-02: `go run ./cmd/connectorgen ownership . --changed-path internal/connectors/defs/github/metadata.json --json` returned `kind: ConnectorOwnershipReport`, `outcome: clean`, `target_connector: github`, and empty `findings`/`warnings` arrays.
- Broader gate 2026-08-02: `go vet ./...` exited 0.
- Boundary guard 2026-08-02: `make connector-boundary` exited 0 with `outcome: clean`.
