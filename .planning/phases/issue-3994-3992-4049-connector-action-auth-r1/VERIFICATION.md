# Verification checklist

- [ ] R1-R7 have retained RED and observable GREEN evidence.
- [ ] Every named refusal asserts its typed error and zero sends/writes/checkpoint/grant consumption.
- [ ] Production call chains are exercised from a fresh `cmd/pm` binary.
- [ ] Available GitHub/PostgreSQL live evidence is recorded without credentials; unavailable R2
      destination proof names the source-only base contract and excluded owner issues.
- [ ] Focused package and selected race tests pass with `-timeout 20m`.
- [ ] `gofmt`, scoped `go vet`, `go build ./cmd/pm`, lint, diff check, and required non-suite gates pass.
- [ ] CLI help/manual/website parity checks pass.
- [ ] Derived artifacts are regenerated once and drift checks leave `git status` clean.
- [ ] Inline `verify-work` and `code-review` complete with no unresolved actionable findings.
- [ ] PR is open and API reports base `integration/4015-mvp-flat-r1`.

