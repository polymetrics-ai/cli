# Issue 389 Verification

## Focused gates

- [ ] Prompt-advertised tools are a subset of the active unit registry.
- [ ] Two issues cannot share one canonical Shepherd/GSD project identity.
- [ ] Same-issue restart adopts the exact existing identity.
- [ ] Unit attempt budget survives process restart.
- [ ] Signal reconciliation interrupts orphaned nested runs.
- [ ] Nested activity is visible through bounded heartbeats.
- [ ] Success rejects missing artifacts, stale heads, unchanged canonical state, and live children.
- [ ] `supervise` dispatches the canonical sequence and stops at the final human gate.
- [ ] Planning and validation observe GPT-5.6 Sol/high; execution observes GPT-5.5/high.

## Module gates

- [ ] `go test ./...`
- [ ] `go test -race ./...`
- [ ] `go vet ./...`
- [ ] `go build ./cmd/shepherd`
- [ ] `make verify`

## Repository boundary

- [ ] Root `go list ./...` excludes `agent-runtime/shepherd`.
- [ ] No parent PR merge to `main`.
- [ ] No secrets, raw prompts, chain-of-thought, or unrestricted tool output in logs.
