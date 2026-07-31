# VERIFICATION — issue #204 Crisp parity wave02 r1

## Local gates

| Gate | Result | Evidence |
| --- | --- | --- |
| `tmp=$(mktemp -d); cp -R internal/connectors/defs/crisp "$tmp/crisp"; go run ./cmd/connectorgen validate "$tmp" --json` | PASS, 1 connector checked, 0 findings | `connectorgen-validate-crisp.json` |
| `go run ./cmd/connectorgen validate internal/connectors/defs` | PASS, 550 connectors checked, 0 findings | `connectorgen-validate-root.txt`, `connectorgen-validate-root.json` |
| `go test ./internal/connectors/conformance -run 'TestConformance/crisp' -count=1 -v` | PASS | `conformance-crisp.log` |
| `go test ./internal/connectors/bundleregistry -count=1` | PASS | `bundleregistry-test.log` |
| `go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1 -timeout 10m` | PASS | `cli-targeted-test.log` |
| `go vet ./internal/connectors/... ./internal/cli/...` | PASS | `go-vet-targeted.log` |
| `go build ./cmd/pm` | PASS | `go-build-pm.log` |
| `make connector-boundary` | PASS, outcome clean | `connector-boundary.log` |
| `./pm docs validate --connectors-dir docs/connectors` | PASS | `docs-validate.log` |
| Website connector data generation | PASS, 550 connectors emitted | `website-gen-connectors.log` |
| `go test -timeout 20m ./...` | PASS | `go-test-all-timeout20m.log` |
| `make verify` | PASS | `make-verify.log` |
| `git diff --check` | PASS | `git-diff-check.log` |

Historical note: the first full `go test ./...` attempt used the default 10m package timeout and failed in slow existing packages plus the bundle count expectation before `internal/connectors/bundleregistry/registry_test.go` was updated. Evidence: `go-test-all-default-timeout-failed.log`. It was superseded by `go test -timeout 20m ./...` and `make verify`, both green.

## CLI/help/docs/website parity checklist

- Runtime connector-specific command execution remains planned only; `cli_surface.json` documents fixed-target planned commands and does not expose a raw API/path/body escape hatch.
- `pm crisp --help`, `pm crisp reverse delete-a-website --help`, and `pm crisp direct delete-suggested-conversation-segment --help` were captured under this phase and show planned availability plus approval/destructive confirmation where applicable.
- `pm connectors inspect crisp --json` confirms catalog-only capability: `check=false`, `read=false`, `write=false`, `query=false`, zero streams, and zero write actions.
- Connector docs were generated/validated under `docs/connectors/crisp/**`, `docs/connectors/README.md`, and `docs/connectors/catalog/all-connectors.*`.
- Website generated connector data was refreshed with `website/scripts/gen-connector-bundles.mjs`, `gen-connector-catalog.mjs`, and `gen-connectors.mjs`.
- No shared CLI router/help implementation was changed; only generated goldens/count expectations were updated for the additional connector.

## Safety verification

- No live Crisp provider calls, credentials, write execution, certification, VPS/Thaalam work, push, PR, or merge.
- `token_id` and `token_key` are secret fields only; no token values are present in fixtures, docs, logs, or issue addenda.
- DELETE/destructive/admin operations are inventoried and planned/blocked; future execution requires named typed actions, redaction, plan -> preview -> explicit approval -> execute, and typed destructive confirmation.
