# Go Shepherd

This nested Go module governs pinned GSD Pi delivery without compiling into `pm`. GSD Pi owns
milestones and local work; Shepherd owns admission, liveness, human questions, authority, exact-head
ratification, external-effect intents, and privacy-safe telemetry.

## Install and configure

Install exact `@opengsd/gsd-pi@1.11.0` outside the repository. Copy `shepherd.example.json` to a
private local path and use absolute paths. `state_dir` must be protected and outside the worker
worktree. Provision the controlled `gsd_home` separately; never put
credentials in the config or repository. Its `agent/settings.json` must pin the configured provider,
model, and `defaultThinkingLevel: high`; admission fails on any mismatch.

For governed delivery, build and use the Podman image so the agent sees only the issue worktree,
task-isolated GSD/planning mounts, and explicit read-only auth/settings files:

```bash
podman build -t localhost/polymetrics-gsd-pi:1.11.0 \
  -f agent-runtime/shepherd/container/Containerfile \
  agent-runtime/shepherd/container
```

The host runtime remains a qualification/debug fallback and keeps all external-effect publishers
disabled. The container does not mount host SSH, GitHub, cloud, or home-directory credentials.

## Start an issue milestone

Create a validated context file inside the isolated issue worktree, then run:

```bash
cd agent-runtime/shepherd
go run ./cmd/shepherd start \
  --config /absolute/private/path/shepherd.json \
  --issue 373 \
  --context .planning/phases/issue-372-gsd-pi-go-shepherd/CONTEXT.json
```

The process prints normalized lifecycle events and a heartbeat at least every 15 seconds. Native GSD
questions are forwarded to the terminal and require an explicit response. The Go deadline always
precedes GSD's fallback response timer. Answer files, inline context, chained `--auto`, and generic
`recover` are rejected. Continue one fenced unit at a time with `run --issue 372 --command next`.
If a prior qualification run already created the correct active milestone, `start --adopt-existing`
binds it explicitly instead of silently creating a second milestone.

Query reconciled workflow state without an LLM. GSD 1.11 query can mutate reconciliation state, so
Shepherd requires the issue identity and holds the same delivery lease:

```bash
go run ./cmd/shepherd query --config /absolute/private/path/shepherd.json --issue 372
```

Compute reproducible eval counters from the normalized local spool:

```bash
go run ./cmd/shepherd eval --config /absolute/private/path/shepherd.json
```

The asynchronous `telemetry.Sink` interface is the extension boundary for future ClickHouse or
Parquet consumers; no such exporter is implemented yet. Shepherd never reads analytics sinks for
controller decisions.

## Verify

```bash
go vet ./...
go test ./...
go test -race ./...
```

The module does not autonomously merge and does not store raw prompts, reasoning, command arguments,
tool output, or credentials.
