# Verification checklist — source-import identity-bearing artifact query

## Security and behavior

- [x] The identity query is explicitly declared by a v3 source lock and never
      accepted from command input or a replacement URL.
- [x] The declared query reaches a GET only after the real importer validates
      the fixed lock and destination.
- [x] Capture/provenance query remains non-fetchable; omitted/false identity
      declarations keep identical projection behavior.
- [x] Credential-shaped keys, oversized query, and excess key count reject.
- [x] HTTPS, userinfo, fragment, ordinary-host, literal-IP, DNS, redirect,
      response/digest, and byte-limit protections remain exercised.

## Local gates

- [x] Targeted `cmd/connectorgen` behavioral tests.
- [x] `go test -timeout 20m ./cmd/connectorgen`.
- [x] `go test -timeout 20m ./internal/cli`.
- [x] `gofmt -w cmd/connectorgen`.
- [x] `go vet ./...`.
- [x] `go build ./cmd/connectorgen` and `go build ./cmd/pm`.
- [x] `make tidy-check`, `make lint`, `make docs-check`, `make smoke-no-build`,
      `make agent-contract-check`, `make connectorgen-validate`,
      `make connectorgen-surface-sync`, `make connector-boundary`, and
      `make release-workflow-check`.
- [x] `scripts/gsd prompt execute-phase 4333`, `verify-work 4333`, and
      `code-review 4333` executed inline and recorded.

## Merge revalidation

`origin/main` advanced with #4327 while this branch was in progress. The
branch merged `e338cd301` without conflict, then passed:

```sh
go test -timeout 20m ./cmd/connectorgen -count=1  # 206.088s
go vet ./cmd/connectorgen
go build ./cmd/connectorgen
git diff --check origin/main...HEAD
```

## CLI help/manual/website parity

Not applicable: no public runtime command, flag, help topic, output contract,
or generated manual changes. The source-lock authoring rule is documented in
`docs/migration/conventions.md`; existing source-import command/documentation
contract tests cover the relevant command text.
