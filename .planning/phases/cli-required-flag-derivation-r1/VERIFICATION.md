# Verification — CLI required-flag derivation r1

## Checklist

- [ ] Required-path invariant enumerates every bundle and returns zero violations.
- [ ] Focused derivation test covers required path and optional query parameters.
- [ ] Command-runner test asserts typed usage error before provider I/O.
- [ ] GitHub sweep falls from 92 to zero; cross-connector before/after count is recorded.
- [ ] All 50 unsupported declarations are listed with verification result; no silent reclassification.
- [ ] Generated connector surfaces are regenerated twice and the second run is byte-stable.
- [ ] `go test -timeout 20m ./cmd/connectorgen` passes.
- [ ] Connector validate/surface-sync/boundary/runtime-preflight gates pass.
- [ ] Website docs generator is byte-stable and relevant runtime help is accurate.
- [ ] Gofmt, vet, changed packages plus consumers, lint/docs/smoke/agent-contract/release checks pass or an exact base-branch blocker is recorded.
- [ ] Review completed with findings/dispositions recorded.
