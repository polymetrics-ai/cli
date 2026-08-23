# Verification checklist — source-import identity-bearing artifact query

## Security and behavior

- [ ] The identity query is explicitly declared by a v3 source lock and never
      accepted from command input or a replacement URL.
- [ ] The declared query reaches a GET only after the real importer validates
      the fixed lock and destination.
- [ ] Capture/provenance query remains non-fetchable; omitted/false identity
      declarations keep identical projection behavior.
- [ ] Credential-shaped keys, oversized query, and excess key count reject.
- [ ] HTTPS, userinfo, fragment, ordinary-host, literal-IP, DNS, redirect,
      response/digest, and byte-limit protections remain exercised.

## Local gates

- [ ] Targeted `cmd/connectorgen` behavioral tests.
- [ ] `go test -timeout 20m ./cmd/connectorgen`.
- [ ] `go test -timeout 20m ./internal/cli`.
- [ ] `gofmt -w cmd/connectorgen`.
- [ ] `go vet ./...`.
- [ ] `go build ./cmd/connectorgen` and `go build ./cmd/pm`.
- [ ] `make tidy-check`, `make lint`, `make docs-check`, `make smoke-no-build`,
      `make agent-contract-check`, `make connectorgen-validate`,
      `make connectorgen-surface-sync`, `make connector-boundary`, and
      `make release-workflow-check`.
- [ ] `scripts/gsd prompt execute-phase 4333`, `verify-work 4333`, and
      `code-review 4333` executed inline and recorded.

## CLI help/manual/website parity

Not applicable: no public runtime command, flag, help topic, output contract,
or generated manual changes. The source-lock authoring rule is documented in
`docs/migration/conventions.md`; existing source-import command/documentation
contract tests cover the relevant command text.
