# Verification — Issue #205 Crisp CLI surface metadata

## Completed commands

Red validation before scaffold:

```bash
go run ./cmd/connectorgen validate internal/connectors/defs/crisp
```

Result: failed as expected because the bundle path did not exist yet.

```text
connectorgen validate: validate: read root: open .: no such file or directory
exit status 1
```

Targeted validation after scaffold (connectorgen validates bundle directories under a root, so the single connector is copied under a temporary root):

```bash
tmp=$(mktemp -d); cp -R internal/connectors/defs/crisp "$tmp/crisp"; go run ./cmd/connectorgen validate "$tmp"
```

Result: pass.

```text
connectorgen validate: 1 connector(s) checked, 0 findings
```

Fleet validation:

```bash
go run ./cmd/connectorgen validate internal/connectors/defs
```

Result: pass.

```text
connectorgen validate: 548 connector(s) checked, 0 findings
```

Conformance smoke for the metadata-only bundle:

```bash
go test ./internal/connectors/conformance -run 'TestConformance/crisp'
```

Result: pass.

```text
ok  	polymetrics.ai/internal/connectors/conformance	3.733s
```

Connector help and inspect smoke:

```bash
go run ./cmd/pm help connectors >/tmp/pm-help-connectors.txt
```

Result: pass.

```bash
tmp=$(mktemp); go run ./cmd/pm connectors inspect crisp --json > "$tmp" && python3 -c 'import json,sys; obj=json.load(open(sys.argv[1])); print(obj["connector"]["name"])' "$tmp"; rm -f "$tmp"
```

Result: pass; output includes `crisp`.

Full local gates after updating catalog-count tests/docs:

```bash
gofmt -w cmd internal
go vet ./...
go test -timeout 20m ./...
go build ./cmd/pm
./pm docs validate --connectors-dir docs/connectors
make verify
```

Result: pass. `make verify` completed fmt, tidy-check, vet, test, build, docs validate, smoke, lint, and connectorgen validate. Re-run after CodeRabbit fixes also passed.

## CLI help/docs/website parity

#205 adds docs/help metadata and generated connector catalog/manual artifacts so `pm docs validate` remains green for the new connector. Runtime provider-specific `pm crisp ...` command dispatch is still pending #206/#207/#209/#211.

Completed applicable checks:

```bash
go run ./cmd/pm help connectors
go run ./cmd/pm connectors inspect crisp --json
./pm docs validate --connectors-dir docs/connectors
rg -n "crisp|Crisp" docs/cli docs/connectors
```

Website docs remain deferred to #206 because #205 exposes no provider-specific runtime command page.

## Current result

Targeted #205 validation and full local `make verify` passed on the stacked branch. Stacked PR #235 is open/ready against `feat/204-crisp-cli-parity`; CI checks passed before the review-fix commit. CodeRabbit reviewed the full range through `767056d` and reported two actionable findings; both were fixed locally, and `make verify` passed again. A follow-up push/incremental review remains pending. Full parent verification will need to be re-run after later implementation slices.
