# Square parity #3191 — verification results

Focused gates:

```bash
go run ./cmd/connectorgen validate --json
# findings=0 warnings=0 connectors_checked=550

go test ./internal/connectors/conformance -run 'TestConformance/square' -count=1
# ok polymetrics.ai/internal/connectors/conformance

go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1
# ok polymetrics.ai/internal/cli

go vet ./internal/connectors/... ./internal/cli/...
# pass, no output

go build ./cmd/pm
# pass

./pm help connectors
./pm connectors --help
./pm connectors inspect square --json
# help rendered; inspect shows connector square write=true

rg -n "Square|square" internal/connectors/defs/square docs/cli website/data/connectors.generated.json internal/cli/testdata/golden_transcripts.json | wc -l
# 893 matches

make connector-boundary
# outcome=clean findings=0 warnings=0

git diff --check
# pass, no output
```

Makefile gates run individually during recovery:

```bash
make docs-check
# Validated connector docs in docs/connectors

make smoke
# smoke ok

make lint
# golangci-lint: 0 issues

make connectorgen-validate
# 550 connector(s) checked, 0 findings

make release-workflow-check
# homebrew release notification assertions passed

go test ./internal/connectors/certify -timeout 30m -count=1
# ok polymetrics.ai/internal/connectors/certify
```

Full gate:

```bash
make verify
# exit 0; includes gofmt, go mod tidy + diff check, go vet ./..., go test -timeout 20m ./..., go build ./cmd/pm, docs-check, smoke, lint, connectorgen-validate, connector-boundary, release-workflow-check
```

Timeout note: an earlier full `make verify` run timed out in `internal/connectors/certify` at the package 20m timeout under the concurrent whole-suite run. The certify package passed standalone and the final `make verify` passed, so this was not a Square connector failure.
