# Verification checklist — Issue #3992

- [ ] Required GSD lifecycle prompts and inline fallback are recorded.
- [ ] R1–R7 have observable red/green evidence.
- [ ] Target schedule, flow, and app suites pass with `-timeout 20m`.
- [ ] Help, bare namespace, docs, website, and generated artifact parity is checked.
- [ ] `gofmt`, `go vet`, build, and the non-suite `make verify` gates pass.
- [ ] Inline `verify-work` and `code-review` evidence is recorded.
- [ ] PR base is API-confirmed as `integration/4015-mvp-flat-r1`.

## External live proof

No credentialed connection is authorized in this worktree. Hermetic isolated
connector proof will establish exact request, acknowledgement, read-back, and
cleanup ordering; it does not claim an external provider mutation. The PR
will name the captain-runbook live proof as a human-gated follow-up.
