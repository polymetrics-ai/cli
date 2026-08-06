# Verification checklist — issue-3755-rate-limit-operator-output-r1

## Behavior

- [ ] A declared test-only bundle shows policy ID, subject kind, selection reason, selected policy count, local pacing duration, provider remaining budget when reported, provider 429 status/wait, and separate request latency.
- [ ] An absent bundle declaration reports `undeclared`, never unlimited, and does not attach a policy.
- [ ] A valid provider reset/Retry-After remains honored exactly; the observation path only reports it.
- [ ] A long run has a bounded summary: coalesced policies and scalar counters/durations, no per-request output.
- [ ] A report never contains credentials, secret map values, token-derived values, raw bindings, opaque scope key, runtime subject, selector runtime values, raw headers/bodies/URLs, or `CredentialRevision`.

## CLI parity

- [ ] `pm help etl`, `pm etl`, and `pm etl run --help` are accurate; no new command or flag exists.
- [ ] `pm etl run` shows a concise rate-limit breakdown after the existing completion line.
- [ ] `pm etl run --json` carries the same summary as `run.rate_limit`.
- [ ] `docs/cli/**`, relevant website docs, and generated help/manual output explain the structured summary and `undeclared` state.

## Local gates

- [ ] focused core, engine, app, and CLI tests pass
- [ ] focused race test passes
- [ ] `gofmt -w internal cmd`
- [ ] targeted `go vet` and `go build ./cmd/pm` pass
- [ ] `go run ./cmd/connectorgen validate` and `surface-sync --check` pass
- [ ] individual project gates (`tidy-check`, lint, docs-check, smoke-no-build, agent-contract-check, connector-boundary, release-workflow-check) pass
- [ ] generated `verify-work` and `code-review` prompts are applied inline; any gaps use the GSD gap loop
