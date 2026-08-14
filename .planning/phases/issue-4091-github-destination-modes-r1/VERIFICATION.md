# Issue #4091 verification

**Status:** local verification passed.

| Must-have | Result | Evidence |
| --- | --- | --- |
| Non-additive disabled/unauthorized execution sends zero provider writes | pass | `TestIssueLabelTransportNonAdditiveModesRequireExplicitConnectionConsent` plus its apply-after-switch-disable assertion record zero POST and PUT requests |
| Changed scope fails before provider request | pass | The stateful recorder remains at one PUT after its label configuration is changed and a no-token rerun is attempted |
| Single-use token cannot replay | pass | The first approved write stores an authorization; replaying its non-empty token fails while the recorder stays at one PUT |
| Identical authorized scope runs unattended | pass | A second run with the token removed reaches exactly a second PUT and read-back observes an exact label set |
| Scope revocation fails before provider request | pass | `RevokeAuthorization` causes the subsequent no-token run to return `AuthorizationRevokedError` with no third PUT |
| Modes are definition-owned and generated surface is current | pass | `go run ./cmd/connectorgen validate` (552 connectors, 0 findings); `go run ./cmd/connectorgen surface-sync --check`; fresh `./pm docs generate --dir docs/cli` produced only the intended GitHub mode/binding catalog metadata |
| GitHub inspect transcript is regenerated and scoped | pass | Sanctioned fresh-binary `POLYMETRICS_UPDATE_GOLDEN_TRANSCRIPTS=1 go test -count=1 -timeout 20m ./internal/cli -run '^TestGoldenTranscripts$' -v`; decoded diff adds only destructive confirmation to `set_issue_labels`; non-GitHub SHA-256 is `f66292f04a9487bc586aa381c80b156f087fbafd3f9ef6cc38c54124cc202100` before and after |
| CLI help/manual/website parity | pass | No command or flag text changed. Regenerated `docs/connectors/catalog/all-connectors.json` carries only #4091's GitHub metadata; the golden transcript is the applicable generated inspection surface. |
| Target packages and repository gates pass | pass | `go test -count=1 -timeout 20m ./internal/app/...`; `go test -count=1 -timeout 20m ./internal/connectors/hooks/github/...`; `go vet ./internal/app/... ./internal/connectors/hooks/github/...`; `go run ./cmd/agentcontractgen check`; `git diff --check` |
| Explicit PR base read-back matches integration | pass | `gh api /repos/polymetrics-ai/cli/pulls/4141 --jq .base.ref` printed `integration/4015-mvp-flat-r1` |

## Live-evidence gap

No GitHub credential or private runbook access was supplied. This worker will not request or expose either. Deterministic in-process GitHub provider tests prove state-changing allowed cases and zero-send refused cases. A separately authorized operator must append credentialed live evidence to #4091 before merge if it remains required.
