# Issue #4289 — Verification Checklist

- [ ] Every listed bundle has a parsing source lock and a corrected declaration-disposition ledger.
- [ ] The source-lock/map-integrity assertion proves all nineteen source inventories, API-surface bindings, row shapes, six-class totals, and rejected-operation reasons.
- [ ] Every documented DELETE is declared or carries an explicit disabled disposition.
- [ ] `go run ./cmd/connectorgen validate` is green.
- [ ] `go run ./cmd/connectorgen surface-sync --check` is green.
- [ ] Targeted connector conformance/fixture and commandrunner preflight tests are green where applicable.
- [ ] Repository generated-file/snapshot checks applicable to changed definition data are green.
- [ ] `connector-boundary` is captured through a detached process and its result is recorded without weakening the check.
- [ ] `verify-work` and data-focused code review have no actionable map defects.
- [ ] No provider credential, live provider call, provider write, or live-certification claim appears in the diff or evidence.
- [ ] PR uses `Refs #4289` and its GitHub API base is read back as `main`.
