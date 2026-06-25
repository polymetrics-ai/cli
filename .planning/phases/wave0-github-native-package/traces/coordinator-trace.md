# Agent Trace: coordinator

## Rendered Prompt Or Prompt Reference
GSD programming-loop `run --phase wave0-github-native-package`. Scope: migrate GitHub connector to
its own per-system package + data-driven self-registration (Wave 0 item 1). DuckDB (item 2) gated.

## Files Inspected
- internal/connectors/connectors.go, catalog.go, native_port.go, native_catalog_connector.go
- internal/connectors/github.go / github_auth.go / github_streams.go (pre-move)
- internal/connectors/connsdk/* (foundation, prior step)
- internal/app/app.go:99, internal/cli/cli.go:903 (registry construction sites)

## Actions Taken
1. PRD coverage: authored PRD/SPEC/PLAN/TEST-PLAN/ADR/THREAT-MODEL/RUNBOOK; marked UI/api/data/eval/
   release/postmortem not-applicable. prd-coverage gate → passed.
2. TDD gate: wrote red-first tests (registry_factory_test.go, github/github_test.go); confirmed red
   (undefined RegisterFactory; "no non-test Go files"); recorded red-confirmed in TDD-LEDGER for
   b-registry-factory, b-github-package, b-registry-wiring. tdd-gate.mjs → passed.
3. Researched the import-cycle constraint + helper ownership + which connectors-internal tests
   construct Github{} directly (must move). Handed a precise brief to the backend role.
4. EXECUTE: spawned gsd-loop-backend (Task) → implemented RegisterFactory + factory store, migrated
   github package, created registryset wiring, relocated tests, deleted old files.
5. TEST gate: independently ran gofmt/vet/`go test ./...`/`make verify` → green; built pm and
   checked inspect parity.
6. VERIFY: spawned gsd-loop-reviewer (read-only, security-inclusive) → GO, no must-fix.
7. Wrote VERIFICATION.md (real Go results; noted JS-verifier limitation), SUMMARY.md, RUN-STATE.json=completed.

## Commands Run
- node programming-loop.mjs {profile, run --phase ...}; prd-coverage.mjs; tdd-gate.mjs
- gofmt -l; go vet ./...; go test ./... (10 ok); make verify (exit 0)
- pm connectors inspect github/source-github --json (parity)

## Findings
- Backend agent corrected one inexact brief claim: github's Manifest() lived in manifest.go (not
  "no Manifest"); relocated cleanly into the package. No behavior impact.
- Bundled node verifier is JS-centric (install hardcoded to npm/pnpm) and cannot gate a Go repo;
  used `make verify` as the authoritative gate.

## Handoff Summary
Wave 0 item 1 complete and green. github is the reference per-system package; registryset is the
codegen target for future connectors. Item 2 (DuckDB) is blocked on a human gate (dependency+CGO).

## Verification Evidence
See VERIFICATION.md + RUN-STATE.json. make verify exit 0; reviewer GO; TDD red-before-code enforced.

## Unresolved Risks
- DuckDB dependency/CGO decision pending (human gate).
- connsdk adoption for github deferred (parity-preserving).
