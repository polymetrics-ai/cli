# Plan — PR #49 destructive confirmation gate correction

## Scope

Primary correction for GitHub CLI parity parent PR #49: prevent critical/destructive GitHub write actions from executing through connector-command or generic reverse ETL flows unless a typed confirmation gate is satisfied.

Secondary hardening in the same safety slice:
- make full-cert write inventory read/parse failures fail loudly;
- make GitHub live-unavailable matching robust;
- keep direct-read full-cert defaults explicit/safe if touched by tests.

## Non-goals

- No live GitHub writes.
- No credential scope changes or secret handling.
- No new dependencies.
- No parent PR merge to `main`.
- No generic raw API, generic HTTP write, shell, or SQL write tool exposure.

## GSD mode

- Runtime: Pi/local tools.
- `scripts/gsd`: unavailable in this checkout; using manual GSD loop.
- Orchestration decision: `local_critical_path` — one tightly coupled safety defect spanning app, CLI, commandrunner, and tests; no subagent tool is available in this harness.

## TDD slices

### Slice 1 — destructive confirmation red tests
1. Add tests that prove implemented GitHub destructive commands cannot execute without typed confirmation.
2. Add tests for the happy path with matching confirmation.
3. Add CLI JSON/argument tests for `--confirm` propagation if needed.

### Slice 2 — implementation
1. Add confirmation metadata to manifest/write-action plumbing as needed.
2. Add plan-time challenge fields to `ReversePlan`.
3. Add `Confirmation` to `RunReverseETLRequest` and `--confirm` to CLI run paths.
4. Reject missing/mismatched confirmations before connector dispatch.

### Slice 3 — certification robustness
1. Change write action inventory helper to return errors.
2. Make full write sweep fail on inventory read/parse errors.
3. Normalize GitHub unavailable classifier.

### Slice 4 — verification and docs
1. Run focused tests.
2. Run gofmt.
3. Run broader local gates as time permits.
4. Update phase summary/verification.

## Verification targets

Focused:
```bash
go test ./internal/app ./internal/cli ./internal/connectors/commandrunner ./internal/connectors/certify ./internal/connectors/engine ./internal/connectors/conformance
```

Broader:
```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
```
