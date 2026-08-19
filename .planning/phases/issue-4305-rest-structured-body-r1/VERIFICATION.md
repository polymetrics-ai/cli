# Verification: Issue #4305

## Required checks

- [ ] Targeted engine, commandrunner, and installed CLI tests with -timeout 20m.
- [ ] Existing scalar, form, SCIM, binary, and specialized GitHub focused regressions.
- [ ] go vet ./...
- [ ] go build ./cmd/pm
- [ ] go run ./cmd/connectorgen validate internal/connectors/defs
- [ ] go run ./cmd/connectorgen surface-sync --check
- [ ] Generated help/manual/schema checks and applicable website parity check.
- [ ] Completion-tracked make connector-boundary.
- [ ] make verify.
- [ ] git diff --check.
- [ ] Code-review evidence with actionable findings resolved.

## Results

Pending execution.
