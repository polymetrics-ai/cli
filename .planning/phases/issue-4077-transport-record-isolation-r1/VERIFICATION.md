# #4077 — verification checklist

## Goal-backward acceptance checks

- [ ] Exact accepted parent-head defect reproduced before project code changes.
- [x] Direct issue #4077 created under #3864; #3864 body visibly records the follow-on without
  changing #4047 history.
- [ ] RED committed before production code.
- [ ] `json.RawMessage` and `map[string]string` source storage is independent after stage/destination
  mutation.
- [ ] Nested supported combinations and existing mutable containers are preserved.
- [ ] Unknown mutable values fail closed before boundary crossing.
- [ ] Checkpoint, acknowledgement, CAS, and all seven canonical modes remain green.
- [ ] Focused normal/race, relevant broader gates, and real `pm` build are green.
- [ ] No credentialed provider/database test claimed; why it is inapplicable is recorded.
- [ ] no-mistakes local-only child run completed within five loops; unsafe stacked delivery gap is
  surfaced as `needs-decision` before a manual exception.

## Initial gate state

- `scripts/gsd doctor`: passed.
- `scripts/gsd sources discuss-phase|plan-phase|execute-phase|verify-work|code-review`: resolved.
- `go run ./cmd/agentcontractgen check`: blocked by pre-existing duplicate project agent inventory;
  no related file will be changed in this issue.
