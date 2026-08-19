# #4077 — verification checklist

## Goal-backward acceptance checks

- [ ] Exact accepted parent-head defect reproduced before project code changes.
- [x] Direct issue #4077 created under #3864; #3864 body visibly records the follow-on without
  changing #4047 history.
- [x] RED test failed for the expected aliasing and unknown-value pass-through behavior; the
  accompanying `test(4077-01)` commit precedes all production code.
- [x] Focused GREEN test proves `json.RawMessage` and `map[string]string` source storage is independent
  after stage/destination mutation.
- [x] Focused GREEN test covers nested supported combinations and the existing `[]byte` container path.
- [x] Focused GREEN test proves unknown mutable values fail closed before boundary crossing.
- [x] Checkpoint, acknowledgement, CAS, and all seven canonical modes remain green.
- [x] Focused normal/race, relevant broader gates, and real `pm` build are green.
- [x] No credentialed provider/database test is claimed; why it is inapplicable is recorded below.
- [x] no-mistakes local-only child run completed with zero loops; unsafe stacked delivery gap is surfaced
  as `needs-decision` before any manual exception.

## Initial gate state

- `scripts/gsd doctor`: passed.
- `scripts/gsd sources discuss-phase|plan-phase|execute-phase|verify-work|code-review`: resolved.
- `go run ./cmd/agentcontractgen check`: blocked by pre-existing duplicate project agent inventory;
  no related file will be changed in this issue.

## Automated evidence

| Check | Result |
|---|---|
| Focused normal regression | Pass |
| Focused race regression | Pass |
| `go test -timeout 20m ./internal/synctransport` | Pass |
| `go test -race -timeout 20m ./internal/synctransport` | Pass |
| `internal/app` canonical route plus all-seven-mode selector | Pass |
| `go vet ./...` and `go build ./cmd/pm` | Pass |
| `scripts/verify-gsd-workflow c67f40a5ff67a131950f3123e70527027dca8493` | Pass |
| Split repository gates | Pass except the two pre-existing `.claude/worktrees` inventory failures below |
| no-mistakes `01KZWMAV3JEKZ9GFK5REF0K2RV` | Pass: review/test/document/lint clean; push/PR/CI intentionally skipped |

The no-mistakes targeted behavior transcript additionally showed source raw JSON and string-map values
unchanged after downstream mutation, and `map[string]int` rejected before the next boundary.

## Scope and live-system boundary

Live GitHub or PostgreSQL credentials cannot strengthen this proof: the mutation-isolation and
fail-closed checks execute before a provider request, warehouse write, or database connection is made.
The evidence is unit/fake transport evidence only; it is not represented as provider E2E certification.

## Repository gate exceptions

`make agent-contract-check` and `make release-workflow-check` fail on generated `.claude/worktrees`
inventory that predates this child branch (duplicate project-agent names and Dockerfile digest findings).
`git diff` from the accepted parent head has no `.claude` path, so this issue neither caused nor changes
those failures.
