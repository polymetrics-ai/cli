# #4091 code review

**Mode:** inline/manual standard review. The generated canonical delivery contract forbids spawning a reviewer for this isolated lane.

## Reviewed surface

- GitHub definition and transport action binding schema
- closed issue-label source/destination dispatch and per-connection consent
- plan/preview/proceed transition to the landed durable `AuthorizationRecord`
- scope re-derivation, replay, revocation, disabled-consent, and changed-scope refusal ordering
- stateful HTTP recorder assertions for PUT/POST counts and exact label read-back
- generated `connectors inspect github --json` transcript audit

## Result

CodeQL reported one actionable error-propagation finding in the first-use authorization branch: `consumePlanApproval` assigned its error to an inner shadow. Commit `fix(github): propagate transport approval consumption errors` replaces that shadow with `authorizationErr`, leaving the outer `err` to be checked before the writer. No findings remain after the fix.

The review specifically confirmed that a later non-additive run re-derives the durable scope before creating the writer evidence; any changed configuration or revoked record returns before the provider writer is called. The writer's declaration-owned action remains the `WriteAction` used by its approval gate, while the canonical sync mode remains separately bound in `EnabledOperations`; this preserves both gate compatibility and mode-specific authorization identity.

The golden transcript was regenerated through its test-owned environment switch, never hand-edited. Its decoded GitHub entry adds only the destructive confirmation for the pre-existing `set_issue_labels` action. A SHA-256 comparison proved every non-GitHub entry, including sample, unknown, and unsafe entries, is byte-for-byte unchanged.

## Commands

- `git diff --check`
- `go test -count=1 -timeout 20m ./internal/app/...`
- `go test -count=1 -timeout 20m ./internal/connectors/hooks/github/...`
- `go vet ./internal/app/... ./internal/connectors/hooks/github/...`
- `go run ./cmd/agentcontractgen check`
- `go run ./cmd/connectorgen validate`
- `go run ./cmd/connectorgen surface-sync --check`
- `POLYMETRICS_UPDATE_GOLDEN_TRANSCRIPTS=1 go test -count=1 -timeout 20m ./internal/cli -run '^TestGoldenTranscripts$' -v`
