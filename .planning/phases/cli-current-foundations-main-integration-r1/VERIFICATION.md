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
- [x] I (`bbbfc670…`) / E (`25b2f844…`) are superseded by the confirmed RC-09 CodeQL repair; the source change is limited to safe allocation and two dead assignments, with no suppression or unrelated alert change.
- [x] Implementation SHA I3 `ad8bff65958a20924aff43d9aea3ea42301e9703` is frozen after the RC-09 focused behavioral suite and RC-10 direct binary adjudication: 15 disputed commands each issue one authenticated GET matching their declared path template and validate the declared output (39.16s); 97 generated candidates also pass after stdout/stderr decoder separation (70.30s). Full `internal/cli` (1061.223s), cold full `cmd/connectorgen` (143.786s), full `commandrunner` (21.816s), vet, build, and surface sync pass. Existing GitHub parity remains the independently confirmed exact release baseline.
- [x] Replacement evidence-only closure is prepared above I3; strict evidence gate is run only from the resulting clean E3 commit.
- [x] `gofmt`/diff check, targeted tests and races, `go vet ./...`, `go build ./cmd/pm`, generation, docs/help/skills/certification, and individually runnable release gates pass; predecessor aggregate test failures are explicitly documented rather than waived.
- [ ] PR #4314 is open and unmerged against the intended non-default integration parent; after E3 is immutable, GitHub API must report its head exactly E3 before a merge recommendation is made.
