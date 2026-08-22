# Verification Checklist — Current Foundations Main Integration r1

## Provisional-composite boundary (captain direction, 2026-08-20)

This checkpoint is intentionally limited to exact local-pipeline composition, the focused combined behavioral gates below, evidence, and independent deep review. It must remain unpushed. Full `make verify`, real-provider qualification, CI, pull-request creation, and no-mistakes are deferred to the later Firstmate-directed stage; they are not represented as passing here.

## Required local gates

- [x] Build the local `pm` binary from the provisional composed SHA; use it for documentation validation; move it recoverably to Trash afterward.
- [x] Run focused engine, commandrunner, App, CLI, sync-transport, source-import, generator, and regression packages with `-timeout 20m`.
- [ ] Run `go vet ./...` (deferred to the later full verifier); `go build ./cmd/pm` passed.
- [x] Run `go run ./cmd/connectorgen validate` and `go run ./cmd/connectorgen surface-sync --check`.
- [x] Run generated CLI/help/manual/website and `connector-boundary` gates.
- [ ] Run full `make verify` once through a completion-tracked command (deferred to the later Firstmate stage).

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

## Foundation 0.3.0 RC r1 release checklist

- [x] Isolated worktree verified; `no-mistakes doctor`, GSD command resolution, and `go run ./cmd/agentcontractgen check` pass.
- [x] Core and public private heads fetched, their predecessor origin heads proved ancestors, and both origin refs advanced non-force with exact `git ls-remote` read-back.
- [x] Reverse merge committed, pushed non-force, and read back; focused transport/App/Postgres/CDC boundary suites pass.
- [x] Public merge committed, pushed non-force, and read back; focused output/engine/connector/SQS boundary suites pass.
- [x] FND-B09 schema-2 failure reproduced before closure; FND-W01 disposition is evidence-backed and does not broaden scope.
- [x] Prior I/E (`a5005fae…` / `7c3d856…`) superseded by the confirmed Recurly path-flag regression; the unchanged binary regression test is green after the narrow declaration/generation repair.
- [x] Prior I/E (`1d83dd9…` / `133afe481…`) superseded by CI-proven generated website-data drift; generator output is now committed and locally idempotent.
- [x] Implementation SHA I `bbbfc67002363b35c6f9ac5ca340a1886523039c` frozen after all production/generated checks; no production/generated path changes afterwards. Its tag-versus-head GitHub parity is exact: endpoints=1,225, blocked=1 with the same retired POST identity, candidates=120.
- [x] Replacement evidence-only closure is prepared above I; strict evidence gate is run only from the resulting clean E commit.
- [x] `gofmt`/diff check, targeted tests and races, `go vet ./...`, `go build ./cmd/pm`, generation, docs/help/skills/certification, and individually runnable release gates pass; predecessor aggregate test failures are explicitly documented rather than waived.
- [ ] PR opened only after E is immutable; GitHub API reports its base exactly `fm/cli-current-foundations-main-integration-r1` and head exactly E.
