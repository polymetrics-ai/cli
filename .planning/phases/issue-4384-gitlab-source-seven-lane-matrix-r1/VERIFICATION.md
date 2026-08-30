# Verification checklist — GitLab Track A

- [x] Source snapshot artifacts match `fm/cli-top100-declaration-batch-r1@dc481bac` byte/digest evidence.
- [x] Red focused reconciliation test fails because the matrix is absent.
- [x] Green focused reconciliation test passes for the retained source denominator and every lane cell.
- [x] Edge variants prove no hidden rows, lane omissions, mutation collapse, boundary loss, or invalid promotion.
- [x] JSON syntax validation passes for source inputs and matrix.
- [x] Focused GitLab package test passes; no aggregate cache-heavy test is run.
- [x] Source-import/projection check was executed and has the known mapping-control blocker: `unknown field "path_bridge"`; no shared repair is in this scope.
- [ ] Staged path review and `git diff --cached --check` pass.
- [ ] Scoped commit is pushed and `git ls-remote` matches its SHA.
- [ ] No-checkbox completion proof is posted on #4384.

## Semantic repair continuation — 2026-08-31

- [x] Real source semantics classify bounded GET/HEAD and documented POST query/lookup reads without an operation-ID allow-list.
- [x] The focused test dynamically proves HEAD and POST candidates, rejects a mutation POST read promotion, and rejects semantic POST mutation promotion.
- [x] Retained request controls are separated from retained successful-response continuation facts: 2 ETL candidates, 257 explicit no-continuation non-candidates.
- [x] Sync remains limited to the three source-cited webhook registrations; pagination contributes none.
- [x] Focused test, full GitLab package test, race test, `go vet`, JSON syntax, agent-contract check, and diff check pass.
- [x] `connectorgen validate internal/connectors/defs/gitlab --json` was rerun and fails only at the unchanged shared parser rejection of `rest.path_bridge`; no shared repair is included.
- [x] Scoped repair commit is pushed and its remote branch SHA is read back before the #4384 proof comment.
- [ ] Fresh independent review is requested in the #4384 completion-proof comment.
