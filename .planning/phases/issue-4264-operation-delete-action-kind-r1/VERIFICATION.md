# Verification — issue 4264

## Planned gates

- `go test -timeout 20m ./cmd/connectorgen -run 'TestCertification'`
- focused action-kind and GitHub/Zoom sweep tests
- `go run ./cmd/connectorgen certification-sweep . --connector github --check`
- `go run ./cmd/connectorgen certification-sweep . --connector zoom --check`
- `make connector-boundary`
- the individual `make verify` gates where available, then full `make verify`
  before any push, as required by the task brief

## Acceptance review

- PASS — red/green classifier tests cover operation-backed delete, non-delete
  create/update/upsert/custom classifications, the existing writes-backed
  delete, and an indeterminate operation that becomes a product defect.
- PASS — `go test -timeout 20m ./cmd/connectorgen -run 'TestCertification'`.
- PASS — `go test -timeout 20m ./cmd/connectorgen`.
- PASS — `go run ./cmd/connectorgen certification-sweep . --connector github
  --check` after regeneration; its package test provides the byte-drift check.
- PASS — `go run ./cmd/connectorgen surface-sync --check`.
- PASS — `make connector-boundary` and `go vet ./cmd/connectorgen/...`.
- PASS (declaration proof) — `origin/fm/cli-zoom-full-parity-r1` supplies 18
  REST DELETE operations, each referenced by a direct-write command. This
  branch only reads that concurrent lane; its definitions and generated sweep
  remain owned by the Zoom lane.
- PASS — full `make verify` (2026-08-19): format/tidy/vet; full `go test
  -timeout 20m ./...`; `go build ./cmd/pm`; connector docs validation;
  smoke; golangci-lint; contract/bundle/surface checks; GitHub generation,
  certification matrix/candidates/sweep; connector boundary; canon; pinned
  dependency; Homebrew notification; and release-target parity gates all exited 0.
- Pending — final diff/base review, GSD review record, push, PR/API base
  confirmation, and hosted check/review observations.
