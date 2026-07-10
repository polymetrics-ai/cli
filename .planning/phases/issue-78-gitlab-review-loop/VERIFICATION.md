# Verification: GitLab CLI Parity Review Loop (#78 / PR #127)

## Required gates

- [ ] `gofmt -w cmd internal`
- [ ] `go vet ./...`
- [ ] `go build ./cmd/pm`
- [ ] `go test -timeout 20m ./...`
- [ ] `make connectorgen-validate`
- [ ] `golangci-lint run`
- [ ] `make verify` if time allows
- [ ] CLI parity help checks:
  - [ ] `./pm help gitlab`
  - [ ] `./pm connectors`
  - [ ] `./pm gitlab --help`
  - [ ] `./pm gitlab project list --help`
  - [ ] `./pm gitlab repo branches check --help`
  - [ ] docs/website grep or generator checks

## Results

Pending.
