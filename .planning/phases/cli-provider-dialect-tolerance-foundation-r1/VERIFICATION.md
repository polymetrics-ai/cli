# Verification — Stripe provider-dialect tolerance foundation

## Source descriptor and regression proof

- `go test -timeout 20m ./cmd/connectorgen -run '^(TestSourceImportDoesNotTreatDepthAsDocumentSafetyExemption|TestSourceImportRetainedStripeCorpus|TestSourceImportLockReferenceDepthLimitOnlyNarrows|TestSourceImportRetainsUnusedDepthDisposition|TestSourceImportGapLockRejectsUnusedSchemaResourceExhaustion|TestSourceProjectionRetainedStripeDepthGapStopsAtRegistryPreflight)$' -count=1` — PASS (`2.289s`). It proves the immutable 7,967,776-byte Stripe artifact emits all 589 locked source descriptors, including the cited `GetAccount` GET and `DeleteAccountsAccount` DELETE identities. Each gets one exact operation-local response-reference-depth condition without invented responses/output; unseen over-depth components get a source-cited condition; hidden external, dynamic, and cyclic inputs remain terminal.
- `go run ./cmd/connectorgen source-import stripe` — PASS: `stripe, 589 operation(s), 0 inbound event(s) imported; source projection updated writes=0 cli=0`.
- `go run ./cmd/connectorgen source-import stripe --check` — PASS: `stripe, 589 operation(s), 0 inbound event(s) verified`.
- `go test -timeout 20m ./cmd/connectorgen -count=1` — PASS (`228.103s`).
- `go test -timeout 20m ./internal/connectors/engine -count=1` — PASS (`11.697s`).
- `go test -timeout 20m ./internal/connectors/commandrunner -count=1` — PASS (`21.603s`). The registry-to-`commandrunner.Preflight` assertion is in the full generator suite; it reaches the exact `missing_foundation=cli-source-descriptor-foundation-r1` condition before credentials, executor dispatch, or provider I/O.
- `go vet ./...`; `go build ./cmd/connectorgen`; `go build ./cmd/pm`; `git diff --check` — PASS.

The focused red command and the earlier current-main failure are retained in
[`TDD-LEDGER.md`](TDD-LEDGER.md). The synthetic Stripe fixture proves an exact
complete source descriptor where the retained contract fits the bound; the
immutable real artifact proves the source-cited operation-local incomplete
case, not a false executable claim.

## Repository verification gates

- `make tidy-check lint docs-check smoke-no-build agent-contract-check connectorgen-validate connectorgen-surface-sync connectorgen-declaration-admission connectorgen-operation-evidence github-parity-artifacts-check` — PASS. `connectorgen validate` checked 553 connectors with zero findings; operation evidence is current at 2,363 rows and five rollups.
- `go run ./cmd/connectorgen certification-subject` — regenerated the required derived subject after Stripe source projection changed its input fingerprint.
- `make connectorgen-certification-subject connectorgen-certification-matrix connectorgen-certification-candidates connectorgen-certification-sweep connector-canon-check release-workflow-check` — PASS. The installed GitHub certification archive proof passed.
- `make connector-boundary` — PASS: whole-tree `outcome: clean`, 325 files checked, 553 connectors loaded, no findings or warnings.

## Generated evidence and scope

- Generated from the exact retained source: `stripe-operation-descriptor.json` (589 descriptors), the Stripe `api_surface.json` source-foundation annotations, `internal/connectors/operation-evidence.json` (589 source-cited Stripe rows across the declared lane classifications), and `certifications/current-subject.json`.
- No Stripe command, action, route, generic HTTP/body/shell escape hatch, credential, or provider I/O was added. Therefore no command is described as credential-bound runnable; all unresolved Stripe contracts remain precise source-descriptor missing-foundation conditions.
- Full `go test -timeout 20m ./...` and monolithic `make verify` are intentionally not run in this per-command worker: project guidance directs changed-package testing plus each applicable make gate individually, while CI runs the monolithic suite.
