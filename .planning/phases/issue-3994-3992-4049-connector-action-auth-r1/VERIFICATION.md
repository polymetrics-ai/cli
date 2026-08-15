# Verification checklist

- [ ] Corrected R1-R8 have retained RED and observable GREEN evidence; superseded per-fire grant
      evidence is labelled and removed from production.
- [ ] Every named refusal asserts its typed error and zero sends/writes/files/sentinels/checkpoints.
- [ ] Production call chains are exercised from a fresh `cmd/pm` binary.
- [ ] Available GitHub/PostgreSQL live evidence is recorded without credentials; unavailable R2
      destination proof names the source-only base contract and excluded owner issues.
- [ ] Focused package and selected race tests pass with `-timeout 20m`.
- [ ] `gofmt`, scoped `go vet`, `go build ./cmd/pm`, lint, diff check, and required non-suite gates pass.
- [ ] CLI help/manual/website parity checks pass.
- [ ] Derived artifacts are regenerated in one final pass and drift checks leave only intentional
      committed task changes.
- [ ] Inline `verify-work` and `code-review` complete with no unresolved actionable findings.
- [ ] PR is open and API reports base `integration/4015-mvp-flat-r1`.
