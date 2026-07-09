# Verification: Bitbucket full-surface implementation

## Results

- [x] `jq . internal/connectors/defs/bitbucket/*.json internal/connectors/defs/bitbucket/schemas/*.json`
- [x] `go test ./cmd/connectorgen -run Bitbucket -count=1`
- [x] `go test ./internal/cli -run Bitbucket -count=1`
- [x] `go test ./internal/connectors/engine -run DirectRead -count=1`
- [x] `go test ./internal/connectors/engine -run 'TestWriteQueryTemplatesUseRecordFieldsAndStayOutOfJSONBody|TestDirectReadReturnsBoundedBitbucketBinaryAsBase64JSON' -count=1`
- [x] `go test ./internal/connectors/engine -count=1`
- [x] `go test ./internal/connectors/conformance -run 'TestConformance/bitbucket' -count=1`
- [x] `go run ./cmd/connectorgen validate internal/connectors/defs`
- [x] `gofmt -w cmd internal`
- [x] `go vet ./...`
- [x] `go test ./...`
- [x] `go build ./cmd/pm`
- [x] `make verify`
- [x] `./pm docs validate --connectors-dir docs/connectors`
- [x] `./pm help bitbucket`, `./pm bitbucket`, `./pm bitbucket --help`
- [x] `npm --prefix website run gen:website-data`
- [x] `git diff --check`

## Coverage result

```json
{
  "covered": 331,
  "blocked": 0,
  "direct": 179,
  "writes": 152,
  "implemented_cli_commands": 342
}
```

## Safety notes

- No credentialed Bitbucket checks were run.
- No secrets were requested, printed, or stored.
- No new dependencies were added.
- No raw generic HTTP write, shell, SQL write, browser, local git, or arbitrary local filesystem executor was added.
- Binary/text GET operations are bounded direct reads returning JSON/base64 only; they do not write files, overwrite paths, or extract archives.
- Mutations remain named reverse-ETL write actions and still require plan → preview → approval → execute.
