# Verification checklist — certify-harness cost foundation (#3795)

## Planned local gates

- [ ] `go test -count=1 ./internal/connectors/certify`
- [ ] `go test -count=1 ./internal/cli`
- [ ] `go test -race ./internal/connectors/certify`
- [ ] `go test -race ./internal/cli`
- [ ] timing parser/runner pass and controlled failure modes
- [ ] two cold local `-count=1 -json` samples for each target package
- [ ] `go vet ./...`
- [ ] `go build ./cmd/pm`
- [ ] `make tidy-check`
- [ ] `make lint`
- [ ] `make docs-check`
- [ ] `make smoke-no-build`
- [ ] `make agent-contract-check`
- [ ] `make connectorgen-validate`
- [ ] `make connectorgen-surface-sync`
- [ ] `make connector-boundary`
- [ ] `make release-workflow-check`
- [ ] `git diff --check`

## CI evidence required before measured threshold

- [ ] At least two post-refactor GitHub-hosted Verify samples from the timing
  target, each cold (`-count=1`) and carrying raw `go test -json` events.
- [ ] Record run/job URL, retrieval date, runner environment if exposed,
  certify elapsed time, CLI elapsed time, slowest tests, and sample maximum.
- [ ] Record the arithmetic used to derive the explicit threshold and explain
  its margin. A local laptop measurement is not sufficient for #3807.

## Deliberate exclusions

No provider/credential/runtime checks are run. Aggregate `go test ./...` and
`make verify` remain CI gates because repository guidance documents that their
single-command duration exceeds the agent timeout. No CLI documentation parity
change is required because production CLI surface remains unchanged.
