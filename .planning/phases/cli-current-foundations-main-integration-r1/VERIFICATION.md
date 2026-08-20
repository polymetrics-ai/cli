# Verification Checklist — Current Foundations Main Integration r1

## Required local gates

- [ ] Build the actual installed `pm` binary from the final composed SHA.
- [ ] Run focused engine, commandrunner, App, CLI, sync-transport, source-import, generator, and regression packages with `-timeout 20m`.
- [ ] Run `go vet ./...` and `go build ./cmd/pm`.
- [ ] Run `go run ./cmd/connectorgen validate` and `go run ./cmd/connectorgen surface-sync --check`.
- [ ] Run generated CLI/help/manual/website and `connector-boundary` gates.
- [ ] Run full `make verify` once through a completion-tracked command.

## Required real-provider gates

- [ ] #4308: real-provider 200 HEAD, harmless 404 HEAD, exact 48,219-byte locked CSV with SHA-256, and missing-path binary GET error with no file.
- [ ] Closed header/multipart/structured-write route against a source-backed actual provider operation.
- [ ] Reverse ETL against persisted App and installed CLI: plan, apply, durable acknowledgement, and provider readback on an approved disposable connector.
- [ ] Source-lock import from immutable provider artifact with exact count and SHA-256.

## CLI parity

- [ ] `pm help <topic>`, each affected bare namespace, and each affected `pm <command> --help` are correct.
- [ ] `docs/cli/**`, `website/**`, generated help/manual/golden surfaces, and discovery metadata reflect the final declarations.
- [ ] JSON output preserves status, headers, body/result fields as declared; diagnostics remain on stderr.

## Hygiene

- [ ] Evidence contains no credentials or secret values.
- [ ] Temporary credentials, binaries, downloads, declarations, and output files are inventoried and removed recoverably.
- [ ] `git status --short` is empty except intended committed rollup evidence and integration changes.
