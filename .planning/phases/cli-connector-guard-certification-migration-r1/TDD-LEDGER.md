# TDD Ledger — Connector Guard Issue C Certification Migration

## Red/green slices

| Slice | Red evidence | Green evidence | Status |
|---|---|---|---|
| Certification metadata parsing | `TestBundleLoadParsesCertification`, `TestBundleLoadRejectsUnknownCertificationKey`, and `TestBundleLoadRejectsCertificationUnknownStream` cover optional valid metadata plus malformed/unknown metadata failures. | `go test ./internal/connectors/engine -run 'TestBundleLoad(ParsesCertification|RejectsUnknownCertificationKey|RejectsCertificationUnknownStream|EmbeddedGitHubCertification)' -count=1` passed. | Green |
| GitHub source defaults/default stream | `TestEffectiveCredentialConfigAddsGitHubBaseURL`, `TestLiveStreamUnavailableClassifiesGitHubUnavailableErrors`, and `TestDefaultStreamName` assert the prior GitHub defaults/classifiers through definition-owned metadata and unknown connector false/no-default behavior. | Focused certify tests passed; production certify Go has no GitHub literal/source-default switch. | Green |
| Direct-read candidates | `TestDirectReadCandidatesForGitHub`, `TestDirectReadCandidateForGitHub`, and `TestDirectReadCandidateForUnknownConnector` assert GitHub args are rendered from metadata and unknown connectors have no candidates. | Focused certify tests passed. | Green |
| Binary candidates | `TestBinaryDownloadCandidateForGitHub` and `TestBinaryDownloadCandidateForUnknownConnector` assert release-download candidate metadata and unknown safe skip. | Focused certify tests passed. | Green |
| Pairings and sweeper safety | `TestDefinitionOwnedPairingGithubCreateLabel`, `TestPairingsForUnknownConnectorIsNoop`, `TestSweepPairingsForGithubHasMultiple`, and `TestSweeperUnknownConnectorDoesNotInventCleanup` assert pairings load from defs and unknown cleanup is not invented. | Focused and full certify tests passed. | Green |
| Record schemas | `TestGenerateRecordForGitHubLabelIncludesColor` now asserts `create_label` required fields include `color` from `defs/github/writes.json`, proving certification write schemas are no longer the old shared table. | Focused and full certify tests passed. | Green |
| Boundary exceptions | Boundary guard initially reported six stale GitHub `provider_certify_contract` exceptions after hard-coded certify contracts were removed. | Deleted only those six rows; `go run ./cmd/connectorgen boundary . --json` passed with `findings=0`, `warnings=0`, `exceptions=6`, and no provider-certify exceptions. | Green |

## Actual evidence

```bash
go test ./internal/connectors/engine ./internal/connectors/certify ./cmd/connectorgen ./internal/connectors/boundary
# ok polymetrics.ai/internal/connectors/engine 0.951s
# ok polymetrics.ai/internal/connectors/certify 345.884s
# ok polymetrics.ai/cmd/connectorgen 4.506s
# ok polymetrics.ai/internal/connectors/boundary 47.144s

go run ./cmd/connectorgen validate internal/connectors/defs --json
# connectors_checked=548 findings=0 warnings=0

go run ./cmd/connectorgen boundary . --json
# outcome=clean findings=0 warnings=0 exceptions=6 provider_certify_exceptions=[]

make connector-boundary
# pass

make verify
# pass

git diff --check
# pass
```

## Notes

- GSD programming-loop command was unavailable in the adapter registry; manual-GSD fallback recorded in `PLAN.md` and `RUN-STATE.json`.
- `go run ./cmd/connectorgen validate internal/connectors/defs/github --json` treats `schemas/` and `fixtures/` as sibling bundle directories because `validate` expects a parent defs root; focused GitHub validation is covered by `TestBundleLoadEmbeddedGitHubCertification` plus full defs validation.
- No live connector credentials or provider calls are part of this ledger.
