# Verification checklist — Issue #3992

- [x] Required GSD lifecycle prompts and inline fallback are recorded.
- [x] R1–R7 have observable red/green evidence.
- [x] Target schedule, flow, and app suites pass with `-timeout 20m`.
- [x] Help, bare namespace, docs, website, and generated artifact parity is checked.
- [x] `gofmt`, `go vet`, build, and the non-suite `make verify` gates pass.
- [x] Inline `verify-work` and `code-review` evidence is recorded.
- [ ] PR base is API-confirmed as `integration/4015-mvp-flat-r1` after creation.

## Automated evidence

- Red: `traces/red-run.txt` records the pre-implementation compilation failure.
- Green: `go test -timeout 20m ./internal/schedule/... ./internal/flow/... ./internal/app/...`
  passed (`schedule`, `flow`, and `app`). Focused `internal/cli` certification covers
  unattended fire, drift, revocation, expiry, rate limiting, status, and safe outputs.
- Full aggregate check: `go test -timeout 20m ./internal/cli -count=1` passed
  in 261.334s after the certification compatibility update.
- `go test -timeout 20m ./internal/cli -run TestCertifyCLISingleConnectorPassExitsZero
  -count=1 -v` passed after the legacy non-action schedule stage was made
  authorization-aware (65 real CLI invocations, within the 92-call budget).
- Help/manual: `go run ./cmd/pm help schedule`, `go run ./cmd/pm schedule`, and
  `go run ./cmd/pm schedule --help` all rendered the authorized schedule manual.
  `POLYMETRICS_UPDATE_GOLDEN_TRANSCRIPTS=1 go test -timeout 20m ./internal/cli -run
  TestGoldenTranscripts -count=1` regenerated the approved transcript; the normal
  golden/manual comparison passed.
- Docs/website: `go run ./cmd/pm docs generate --dir docs/cli --connectors-dir
  docs/connectors`, `npm --prefix website run gen:docs`, `make docs-check-no-build`,
  and `npm --prefix website run typecheck` passed.
- Code/gates: scoped `go vet`, `go build ./cmd/pm`, `make tidy-check`, `make lint`,
  `make agent-contract-check`, `make connectorgen-validate`,
  `make connectorgen-surface-sync`, `make connector-boundary`,
  `make release-workflow-check`, and `make smoke-no-build` passed. The whole-tree
  boundary report was `clean` (242 files, 552 connectors, zero findings).

## Manual verify-work fallback

The official `verify-work` prompt cannot resolve this issue-only phase as a
roadmap-numbered phase. The inline acceptance test is
`TestInstalledScheduleFireRunsAuthorizedRoundTripAndRestoresBackend`: it installs
an isolated crontab payload, fires with its durable reference and no human token,
observes connector validate/preview/write/read-back ordering plus a durable receipt,
and removes the entry with byte-for-byte restoration. Its companion tests prove
scope drift, revocation, and expiry stop before any target event; crash/overlap and
rate-limit paths halt or park without replay. This is fully automated because the
deliverable has no unautomatable human judgment.

## Manual code-review fallback

The official `code-review` prompt also requires a numbered phase and an isolated
reviewer runtime. Inline review covered persistence/locking, error classification,
rendered scheduler arguments, flow receipt propagation, secret boundaries, and
CLI/docs parity. No actionable findings remain; `REVIEW.md` has the scoped record.

## External live proof

No credentialed connection is authorized in this worktree. Hermetic isolated
connector proof will establish exact request, acknowledgement, read-back, and
cleanup ordering; it does not claim an external provider mutation. The PR
will name the captain-runbook live proof as a human-gated follow-up.
