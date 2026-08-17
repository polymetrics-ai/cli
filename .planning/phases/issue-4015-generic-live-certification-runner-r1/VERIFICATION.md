# Verification — generic live certification runner

## Planned checks

- `node --check scripts/certify-connector-live.mjs`
- `node scripts/certify-connector-live.mjs <connector> --definition-check`
- `node scripts/certify-connector-live.mjs <authorized-connector> --credential-env <environment-name>`
- `node scripts/certify-connector-live.mjs <different-connector> --credential-env <environment-name>`
- `go run ./cmd/connectorgen certification-matrix --check` after every accepted record and once after the run
- `git diff --check`

The full Go suite and `make verify` are deliberately deferred until the end of the live run, as authorized by the captain; this task has no Go production changes.

## Executed checks

- `node --check scripts/certify-connector-live.mjs` — passed.
- `node scripts/certify-connector-live.mjs github --definition-check` — passed: `commands=1571 candidates=122 eligible=122 credential_configs=2`.
- Definition-only invocation for every one of the 36 command-surface connectors — passed.
- `node scripts/certify-connector-live.mjs freshchat ...` — passed unchanged: `executed=0 certified=0 provider_refused=0 missing_fixture=1 product_defect=0`.
- Authorized live run: `executed=122 certified=38 provider_refused=80 missing_fixture=4 product_defect=0`; every accepted record was immediately checked by `go run ./cmd/connectorgen certification-matrix --check`, which also passed after the run.
- `go test -timeout 20m ./cmd/connectorgen` — passed (82.667s).
- `go test -timeout 20m ./internal/connectors/commandrunner` — passed (12.541s).
- `go test -timeout 20m ./internal/cli` — passed (511.153s).
- End-of-run repository gates passed: `make fmt tidy-check vet build docs-check-no-build smoke-no-build lint agent-contract-check connectorgen-validate connectorgen-surface-sync github-parity-artifacts-check connectorgen-certification-matrix connectorgen-certification-candidates connectorgen-certification-sweep connector-boundary connector-canon-check release-workflow-check`.
- `git diff --check` — pending immediately before commit.
- Post-PR App retry: selected 16 captain-probed App-200 rows, executed all 16, certified 15, and retained one HTTP 400 product-defect receipt. Final `certification-matrix --check` passed with 53 GitHub accepted records on disk.
- Classic credential batch: `GITHUB-CLASSIC-REACHABLE-STAGES.json` selected the 12 `CLASSIC` rows from the authoritative credential matrix. An initial 12-command local preflight exposed the missing `rate_limit_account` configuration and produced no provider request; the corrected run executed 12 live commands, certified 11, and retained `dependabot list-alerts-for-enterprise` as a reproducible HTTP 400 product defect against the direct classic-PAT 200 probe. Local credential cleanup was verified and `certification-matrix --check` passed with 64 accepted GitHub records on disk.
- Real-ID fixture avoidance: the already-certified classic collection identifies enterprise code-security configuration `17`; the two sibling reads ran with that identifier, both certified, and `certification-matrix --check` passed with 66 accepted GitHub records on disk.
- Enterprise-team fixture lifecycle: created one `pm-cert-` team in the captain-authorized disposable enterprise boundary, attempted the sibling read, then deleted and independently checked absence. The read stopped locally because the provider's valid `ent:`-prefixed slug violates runtime path validation. The first confirmed delete reported success but left the team present; a second delete using the separately resolved plain fixture slug removed it. Both product defects and the cleanup proof are retained in sanitized fixture receipts; no accepted evidence was written.
- Copilot Space resource fixture: the classic credential first read an empty collection, created one `pm-cert-` free-text resource, then certified `copilot-spaces get-resource-for-org` with an immediate matrix check. Two approved PM delete runs each reported one succeeded and zero failed records, but the resource remained through six read-back polls. A credential-scoped `gh-axi api DELETE` cleanup returned success and the specific-resource read-back returned 404. The accepted-record count is 67.
- Controlled enterprise-team timing probe: a fresh disposable team was created, then PM deleted it using the provider-returned `ent:` slug. The run returned exit 0, `records_succeeded=1`, and `records_failed=0`; six bounded read-back polls did not observe absence. This proves a false-success cleanup defect rather than eventual consistency. Final direct read-back verified absence. The unambiguous direct-read `ent:` path-validation defect is tracked in issue #4220.
- P0 containment: issue #4221 tracks the independently reproduced false-success destroy-path defect. Direct provider cleanup is now the only trusted fixture cleanup path. The pre-existing `pm-cert-space-1786995413` fixture at Space 1 was removed through `gh-axi api DELETE`; the first independent GET read-back returned 404 in 491 ms.
