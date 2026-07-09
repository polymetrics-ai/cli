# Verification — #103 Linear sensitive/admin policy

## Focused evidence

- `go run ./cmd/connectorgen validate internal/connectors/defs --json`, `./pm help linear`, `grep` evidence for `api graphql` and `Raw arbitrary GraphQL is disallowed`.

## Parent gates also covering this slice

- `go vet ./...` — pass
- `go test ./...` — pass
- `go build ./cmd/pm` — pass
- `./pm docs validate --connectors-dir docs/connectors` — pass
- `npm --prefix website run gen:website-data` — pass
- `make verify` — pass
- `git diff --check` — pass

## Safety

No secrets were requested or printed. No live Linear credentials or network checks were used. Fixed GraphQL documents only; raw arbitrary GraphQL remains disallowed.
