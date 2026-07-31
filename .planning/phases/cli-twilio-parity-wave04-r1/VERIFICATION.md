# VERIFICATION — Twilio parity wave04 r1

Required local gates:

```bash
go run ./cmd/connectorgen validate internal/connectors/defs --json  # focused by inspecting twilio findings/warnings
go test ./internal/connectors/conformance -run 'TestTwilioOfficialParityLedgerAndFixtureCoverage|TestConformance/twilio' -count=1
go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1
go vet ./internal/connectors/... ./internal/cli/...
go build ./cmd/pm
make connector-boundary
make verify
git diff --check
```

Additional metadata/doc checks:

```bash
go run ./cmd/pm docs generate --dir docs/cli --connectors-dir docs/connectors
go run ./cmd/pm docs validate --connectors-dir docs/connectors
pnpm --dir website gen:website-data
```

Safety verification:

- No Twilio credential values requested or printed.
- No live provider calls.
- No new dependencies.
- No shared runtime behavior edits.
- `direct_read_query_search` remains 0 and no generic query/search escape hatch is introduced.
- Certification remains `certified=0`; no live-safe executor/certification run was invoked.

## Results

### Passed

- Focused Twilio connectorgen validation via temporary one-connector root: `connectorgen validate: 1 connector(s) checked, 0 findings`.
- Root connectorgen validation: `go run ./cmd/connectorgen validate internal/connectors/defs --json` reported 549 connectors, 0 findings, 0 warnings, and 0 Twilio findings/warnings.
- Focused conformance: `go test ./internal/connectors/conformance -run 'TestTwilioOfficialParityLedgerAndFixtureCoverage|TestConformance/twilio' -count=1` passed.
- Focused CLI tests: `go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1` passed.
- Focused vet/build/boundary/diff checks passed:
  - `go vet ./internal/connectors/... ./internal/cli/...`
  - `go build ./cmd/pm`
  - `make connector-boundary`
  - `git diff --check`
- Docs/website generation and validation passed:
  - `go run ./cmd/pm docs generate --dir docs/cli --connectors-dir docs/connectors`
  - `go run ./cmd/pm docs validate --connectors-dir docs/connectors`
  - `pnpm --dir website gen:website-data`
- Slow packages implicated by the aggregate timeout pass in isolation with extended/focused runs:
  - `go test -timeout 20m ./internal/connectors/certify -run TestWriteStagesLedgerWrittenBeforeCreate -count=1 -v` passed in 108s.
  - `go test -timeout 35m ./internal/cli -count=1` passed in 1088s.

### Not fully green

- `make verify` did not complete green in this local harness. Plain attempts reached `go test -timeout 20m ./...` and hit the package-level 20m timeout in slow, pre-existing aggregate packages (`internal/connectors/certify` on the first attempt, `internal/cli` on rerun with `TestBahmniDeclaredCommandMatrixIsRecognizedOrExplicitlyBlocked`). A later `GOFLAGS=-p=1 make verify` attempt was terminated by the command timeout while still in the serial `go test -timeout 20m ./...` phase. No Twilio-specific failures were reported by these attempts.
