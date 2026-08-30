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
