# REVIEW — issue #3995 shared connector-certification Shepherd gate

## Method

The issue-named phase has no registered numbered GSD adapter phase, so `code-review` was executed
inline under the documented fallback. Review covered the canonical contract, strict decoder,
evaluator, evidence binding, projection renderer/checker, generated Claude/Codex/Pi/OpenCode
files, tests, and changed-path scope. No provider, credential, connector-bundle, transport, or
`cmd/connectorgen/certification*.go` path is changed.

Automated review evidence before this report:

- `go test -timeout 20m ./internal/agentcontract -count=1`
- `go test -timeout 20m ./cmd/agentcontractgen -count=1`
- `go vet ./...`, `go build ./cmd/pm`, `make lint`, and
  `go run ./cmd/agentcontractgen check`
- generated-projection equivalence/drift tests and the individual repository gates recorded in
  `VERIFICATION.md`

## Findings and dispositions

| ID | Severity | Finding | Disposition | Evidence |
| --- | --- | --- | --- | --- |
| R-1 | medium | The first proof consumer only required `body.value` to be valid JSON. A forged accepted sidecar could therefore retain a raw scalar rather than a repository-salted fingerprint. | Fixed. `validateHTTPBody` now mirrors the proof boundary: `none`, `opaque`, and `json` forms are versioned/typed and every JSON scalar must be a fingerprint (or `null`). Header/query/statement values also accept only fingerprint sequences. | `TestCertificationGateRejectsUnredactedProofBody` passes; `TestCertificationGateMatchesSemanticallyEquivalentEvidenceProof` preserves safe formatting tolerance. |
| R-2 | low | Raw `json.RawMessage` whitespace could make a semantically identical matrix pointer and sidecar appear unequal. | Fixed. Pointer/record proof comparison serializes the typed proof into canonical JSON before comparison. | `TestCertificationGateMatchesSemanticallyEquivalentEvidenceProof` passes. |

## Verdict

PASS after correction round 1 of 5. There are no open actionable review findings. The remaining
external dependency is #3989: its eventual proof-schema version must be integrated explicitly;
the present gate halts unknown proof versions rather than inventing fields.

## Correction tracking

The proof-consumer fixes are tracked as child issue [#4024](https://github.com/polymetrics-ai/cli/issues/4024)
under #3995, with `Refs #3988`. It owns both R-1 fingerprint-only proof consumption and R-2
canonical semantic JSON comparison. The production remediation is commit `842f1c271`; the
verification record is `d511186bc`. The issue is deliberately open pending this child branch's PR
review and the parent acceptance flow.
