# Go Shepherd

This nested Go module governs pinned GSD Pi delivery without compiling into `pm`. GSD Pi owns
milestones and local work; Shepherd owns admission, liveness, human questions, authority, exact-head
ratification, external-effect intents, and privacy-safe telemetry.

## Install and configure

Install exact `@opengsd/gsd-pi@1.11.0` outside the repository. Copy `shepherd.example.json` to a
private local path and use absolute paths. Provision the controlled `gsd_home` separately; never put
credentials in the config or repository. Its `agent/settings.json` must pin the configured provider,
model, and `defaultThinkingLevel: high`; admission fails on any mismatch.

## Start an issue milestone

Create a validated context file inside the isolated issue worktree, then run:

```bash
cd agent-runtime/shepherd
go run ./cmd/shepherd start \
  --config /absolute/private/path/shepherd.json \
  --issue 373 \
  --context .planning/phases/issue-373/CONTEXT.md \
  --auto
```

The process prints normalized lifecycle events and a heartbeat at least every 15 seconds. Native GSD
questions are forwarded to the terminal and require an explicit response. Answer files and inline
context are rejected.

Query workflow state without an LLM:

```bash
go run ./cmd/shepherd query --config /absolute/private/path/shepherd.json
```

## Verify

```bash
go vet ./...
go test ./...
go test -race ./...
```

The module does not autonomously merge and does not store raw prompts, reasoning, command arguments,
tool output, or credentials.
