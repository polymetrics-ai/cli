# TDD ledger — GitLab R3 bodyless POST source-output policy

## Preconditions

- Branch: `codex/4394-gitlab-bodyless-post-output-r3`.
- Frozen parent: `ceaae873aef0dd19aa23c036b9cb598f9b3eacc8`.
- R2 baseline to preserve: `a87908418b4bf69fa7b49bd64ae9ac8fa6a574bd`.
- No GitLab credential, provider I/O, provider write, source-lock rewrite, or
  runtime executor change is authorized.

## Red test specification

| ID | Red assertion | Initial expected failure | Green assertion | Edge assertion |
| --- | --- | --- | --- | --- |
| R3-1 | Source-backed status-only operation/CLI pair cannot use `json_redacted`. | R2 declares all eight bodyless POST reads `json_redacted`; the four status source rows fail the expected policy table. | Those four exact pairs use `none`; every source ID/path/method/binding remains unchanged. | Missing or mismatched operation/CLI policy is rejected. |
| R3-2 | Source-backed JSON Conan operation/CLI pair cannot use `none`. | Validator test constructs JSON source fact + `none`; closed admission rejects it. | All four exact Conan pairs retain `json_redacted`. | Malformed policy, undeclared binding, and a source/policy mismatch remain rejected. |
| R3-3 | Status-only source response is not decoded as JSON. | R2 engine fixture supplies JSON for every row, hiding the source mismatch. | Empty status body through `none` returns no decoded body after an exact zero-byte POST. | A nonempty status-only response errors; caller `Body` or `RawBody` errors before the fixture observes I/O. |
| R3-4 | Public generated commands remain safely reachable. | N/A: existing generated command reaches the credential boundary but does not cover the eight-bodyless cohort. | Each exact command with valid required flags returns typed missing credential; spy sees zero provider requests. | Incomplete/invalid flags fail before credential lookup/I/O; a normal implemented command is not converted to a deferred command. |

## Evidence status

| Step | Command/test | Status | Evidence |
| --- | --- | --- | --- |
| Adapter preflight | `scripts/gsd doctor`; canonical source resolution; `go run ./cmd/agentcontractgen check` | green | Completed before edits; exact output recorded in PLAN.md. |
| Red R3-1 | `go test -count=1 -timeout 180s ./internal/connectors/defs/gitlab -run '^TestGitLabSourceBoundMaterializationCohort$'` | red | `postApiV4AiThirdPartyAgentsDirectAccess` declared `json_redacted`, want source-backed `none` (exit 1). |
| Red R3-2 | `go test -count=1 -timeout 180s ./cmd/connectorgen -run '^TestSourceProjectionNonExecutableMutationDispositionAllowsOnlyClosedBodylessPOSTRead/closed_bodyless_semantic_direct_read_preserves_source_output_policy$'` | red | Status with `json_redacted`, JSON with `none`, and operation/CLI disagreement all admitted before the narrow predicate correction (exit 1). |
| Green R3-1/R3-2 | Focused definition and source-projection tests in `VERIFICATION.md` | green | Four source-status rows now select `none`; four JSON Conan rows retain `json_redacted`; mismatches are rejected. |
| Green R3-3/R3-4 | Focused engine and public CLI tests in `VERIFICATION.md` | green | Exact eight-row wire/policy cohort, status-body edge, pre-I/O body rejection, and all eight missing-credential/no-I/O paths pass. |
| Refactor | `gofmt`, `go vet`, narrow race, JSON/schema/diff | green | Narrow engine race passed; only unrelated broader suite/validator failures are recorded in `VERIFICATION.md`. |

## Test isolation rationale

- `httptest` is the only provider-wire fixture. It is necessary because a
  credentialed GitLab environment is not part of this task; it observes an
  actual `net/http` request and source-derived response policy, not a mock of
  an internal method.
- CLI boundary tests use the real public parser and commandrunner with no
  credential. They explicitly assert the negative observable: zero provider
  requests before the typed missing-credential boundary.
- No race or full-suite run starts if free disk space is below the shared
  validation floor. Normal scoped tests remain separate from any other lane.
