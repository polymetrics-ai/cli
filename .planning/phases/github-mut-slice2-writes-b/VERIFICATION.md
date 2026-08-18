## Verification Checklist

- [x] All 146 assigned paths attempted serially or explicitly captain-gated and classified exactly once. `FINAL.md` proves the disjoint bucket split and total.
- [x] Every certified result has plan, preview, token-stdin execution, independent produced-value read-back, and independently proven terminal containment.
- [x] Every product defect includes a raw GitHub API control; later `integer_id_scientific_notation` instances reuse the fleet-proven exact-integer control as directed.
- [x] Every retained schema-v2 record passes `go run ./cmd/connectorgen certification-matrix --check`. Twenty-nine records are added over the integration base and each comes from captured live traffic.
- [x] `git diff --check` passes.
- [x] Targeted Go tests pass.
- [x] PR [#4230](https://github.com/polymetrics-ai/cli/pull/4230) is open; direct GitHub API read-back reports base `integration/4015-mvp-flat-r1`.

## Batch 1 checkpoint — 2026-08-18

The per-command receipt and corrected bucket split are in `BATCH-1.md`. The empty-collection audit converted commands 11, 19, and 20 from `no_object` to certified by creating contained fixtures and proving both the mutation and cleanup. Command 58 was subsequently attempted against a nonexistent enterprise scope under supervisor direction and classified as entitlement.

## Batches 2 and 3 checkpoint — 2026-08-18

`BATCH-2.md` records the completed 50-command execution batch and the completed real-fixture re-audit. `BATCH-3.md` records commands 109–146, including raw provider controls, explicit captain escapes, the valid chained-PR stack fixture, and terminal cleanup. No pending row remains.

## Live attempt receipt — 2026-08-18

`issue create` used the contained fixture title `pm-cert-mut2-issue-create-20260818-0320`.

- Produced-value assertion (`agent_derived`): the independent issue collection contained exactly one issue with that unique title; a missing title or a different title is rejected.
- The command completed the connector-command plan → preview → bare `--approval-token-stdin` run lifecycle.
- The mandated REST provider DELETE returned HTTP 404 while the independent item read-back returned HTTP 200: this reproduces the stated cleanup defect.
- A direct GitHub GraphQL `deleteIssue` cleanup mutation then succeeded; subsequent item read-back returned HTTP 410 and collection read-back contained zero title matches. The cleanup is therefore proven absent, but it did not meet the task's literal REST-DELETE requirement.
- The importer limitation is a surface finding only. The fleet ruling authorizes direct schema-v2 evidence authoring from the captured real command and provider traffic; it never authorizes invented exchanges.

## Local verification — 2026-08-18

- `go test -timeout 20m ./cmd/connectorgen` — passed (`ok polymetrics.ai/cmd/connectorgen`, 127.738s).
- `go vet ./cmd/connectorgen` — passed.
- `go run ./cmd/connectorgen certification-matrix --check` — passed; certification shards current.
- `go run ./cmd/connectorgen surface-sync --check` — passed; 552 connectors scanned with zero drift.
- `go run ./cmd/agentcontractgen check` — passed; canonical contract and projections current.
- `scripts/verify-gsd-workflow` — passed; GSD/TDD evidence recognized.
- `git diff --check` — passed.

The full repository suite and `make verify` were not run locally because `AGENTS.md` directs per-command-timeout agents to scope tests and let CI carry the 550+ connector suite. This PR changes only schema-v2 certification evidence and its planning ledger; no Go production or test source is changed by this branch.
