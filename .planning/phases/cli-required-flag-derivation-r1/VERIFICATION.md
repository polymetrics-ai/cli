# Verification — CLI required-flag derivation r1

## Checklist

- [x] Required-path invariant enumerates every bundle and returns zero violations.
- [x] Focused derivation test covers required path and optional query parameters.
- [x] Command-runner test asserts typed usage error before provider I/O.
- [x] GitHub sweep falls from 92 to zero; cross-connector before/after count is recorded: 104 fields in 92 GitHub commands, zero violations in every other connector.
- [x] All 50 unsupported declarations are listed with verification result; no silent reclassification. The audit reports 26 source-supported `unsupported_api` contradictions, one retained `unsupported_api`, and 23 valid `unsupported_local` entries.
- [ ] Generated connector surfaces are regenerated twice and the second run is byte-stable.
- [ ] `go test -timeout 20m ./cmd/connectorgen` passes.
- [ ] Connector validate/surface-sync/boundary/runtime-preflight gates pass.
- [ ] Website docs generator is byte-stable and relevant runtime help is accurate.
- [ ] Gofmt, vet, changed packages plus consumers, lint/docs/smoke/agent-contract/release checks pass or an exact base-branch blocker is recorded.
- [ ] Review completed with findings/dispositions recorded.
