# Verification: Issue #92 bitbucket-stream-runner

Date: 2026-07-09

## Required checks

```bash
jq . internal/connectors/defs/bitbucket/*.json internal/connectors/defs/bitbucket/schemas/*.json
go test ./cmd/connectorgen -run Bitbucket -count=1
go test ./internal/cli ./internal/connectors/commandrunner ./internal/connectors/engine ./cmd/connectorgen -count=1
go run ./cmd/connectorgen validate internal/connectors/defs --json
go vet ./...
go test ./...
go build ./cmd/pm
make verify
```

## Results

Pending.

## Final verification results

Stream-backed Bitbucket commands are implemented for repositories, branches, commits, tags, pull requests, issues, pipelines, deployments, downloads metadata, webhooks, branch restrictions, projects, and snippets with conformance fixtures.

- [x] `jq . internal/connectors/defs/bitbucket/*.json internal/connectors/defs/bitbucket/schemas/*.json`
- [x] `go test ./cmd/connectorgen -run Bitbucket -count=1`
- [x] `go test ./internal/cli -run Bitbucket -count=1`
- [x] `go test ./internal/connectors/conformance -run 'TestConformance/bitbucket' -count=1`
- [x] `go vet ./...`
- [x] `go test ./...`
- [x] `go build ./cmd/pm`
- [x] `make verify`
- [x] `go run ./cmd/connectorgen validate internal/connectors/defs`
- [x] `./pm help bitbucket`
- [x] `./pm bitbucket`
- [x] `./pm bitbucket --help`
- [x] `npm --prefix website run gen:website-data`

Safety notes: no credentialed Bitbucket checks were run; fixtures use synthetic data only; no secret values were requested, printed, or stored.
