# Verification checklist — Vercel Track A

- [x] `go run ./cmd/agentcontractgen check` passed.
- [x] Red focused Vercel matrix test failed because the matrix was absent.
- [x] Green focused Vercel matrix test passed, including hidden-row, missing-cell, backlink, boundary, and invalid-disposition edge assertions.
- [x] `go test ./internal/connectors/defs/vercel -count=1` passed.
- [x] `go vet ./internal/connectors/defs/vercel` passed.
- [x] `jq empty internal/connectors/defs/vercel/sources/vercel-source-lane-matrix.json` passed.
- [x] Check-only commands passed: `go run ./cmd/connectorgen source-import vercel --check --read-projection-only`; `go run ./cmd/connectorgen source-materialize vercel --check`; `go run ./cmd/connectorgen surface-sync internal/connectors/defs --connector vercel --check`; and `go run ./cmd/connectorgen validate internal/connectors/defs/vercel`.
- [x] `git diff --cached --check` and staged-path review passed; exactly the seven scoped Track A artifacts are staged.
- [ ] Scoped commit must be pushed and `git ls-remote` must confirm the immutable remote SHA.
- [ ] A no-checkbox completion proof comment must be posted to #4421 after the push.
