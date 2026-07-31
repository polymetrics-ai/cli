# Mailchimp parity wave03-r1 verification checklist

Required fixture-only gates:

```bash
go run ./cmd/connectorgen validate internal/connectors/defs/mailchimp
go test ./internal/connectors/conformance -run 'TestConformance/mailchimp' -count=1
go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1
go build ./cmd/pm
make connector-boundary
make verify
git diff --check
```

Additional docs/help checks:

```bash
pm help connectors
pm connectors inspect mailchimp --json
rg -n "Mailchimp|mailchimp" docs/connectors website internal/connectors/defs/mailchimp
```

## Results

- PASS `go run ./cmd/connectorgen validate internal/connectors/defs/mailchimp` (`traces/validate-final.log`): `connectorgen validate: 1 connector(s) checked, 0 findings`.
- PASS `go test ./internal/connectors/conformance -run 'TestConformance/mailchimp' -count=1` (`traces/conformance-final.log`): `ok polymetrics.ai/internal/connectors/conformance 2.289s`.
- PASS `go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1` (`traces/cli-test-final.log`): `ok polymetrics.ai/internal/cli 14.879s`.
- PASS `go build ./cmd/pm` (`traces/go-build-pm-1.log`).
- PASS `make connector-boundary` (`traces/connector-boundary-1.log`): clean boundary report, no findings.
- PASS `make verify` (`traces/make-verify-pipefail.log`): rerun with shell `pipefail` completed in ~436s; includes gofmt, tidy-check, go vet, `go test -timeout 20m ./...`, build, docs validate, smoke, lint, connectorgen validate, connector boundary, and release workflow check.
- PASS `git diff --check` (`traces/git-diff-check-1.log`): no whitespace errors.

Docs/help parity checks performed:

- `./pm help connectors` saved to `traces/pm-help-connectors-final.txt`.
- `./pm connectors inspect mailchimp --json` saved to `traces/pm-inspect-mailchimp-final.json`; metadata-only, no credentials read.
- `go run ./cmd/pm docs generate --dir docs/cli` updated connector docs/catalog surfaces, with only Mailchimp and catalog outputs retained.
- `npm run gen:website-data` under `website/` updated generated website connector data/catalog surfaces.
- `POLYMETRICS_UPDATE_GOLDEN_TRANSCRIPTS=1 go test ./internal/cli -run '^TestGoldenTranscripts$' -count=1` refreshed CLI golden transcripts.
