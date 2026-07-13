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
  -f agent-runtime/shepherd/container/Containerfile .
```

The host runtime remains a qualification/debug fallback and keeps all external-effect publishers
disabled. The container does not mount host SSH, GitHub, cloud, or home-directory credentials.

The governed image contains Go 1.25.12, Make, Git, jq, ripgrep, official GSD Pi 1.11.0, Context7
MCP 3.2.3, and agent-browser 0.31.1 with Chrome. `curl` is a read-only compatibility wrapper around
`web-fetch`: common GET/HEAD flags work, responses are capped at 2 MiB, and bodies, auth headers,
uploads, and non-HTTP(S) schemes are rejected. GitHub CLI and publisher credentials are absent.

Start the private search sidecar before selecting `container_network: shepherd-research`:

```bash
podman compose -f agent-runtime/shepherd/research/compose.yaml up -d
```

SearXNG is available only to containers on that network at `http://searxng:8080` and has no host
port. Agent-browser is deny-by-default: it permits navigation, snapshots, scrolling, waits, and
reads; content boundaries are enabled and output is capped at 50,000 characters. Shepherd writes
the Context7 HTTP MCP entry into protected task state rather than trusting worktree MCP config.

Run the pinned official `gsd` binary directly for local interactive work. `scripts/gsd` remains a
compatibility prompt renderer for shell and legacy Pi callers; its local registry now supports:

```bash
scripts/gsd prompt programming-loop init --phase issue-372 --dry-run
```

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
For a multi-turn local milestone interview, use `--command discuss --continue-unit` after the first
round. Shepherd resumes only the newest Pi session whose header is bound to the exact configured
worktree.
If a prior qualification run already created the correct active milestone, `start --adopt-existing`
binds it explicitly instead of silently creating a second milestone.

For official GSD Pi 1.11.0 on the host, runtime admission idempotently applies one qualified
compatibility patch inside the pinned package: the upstream headless idle timer must wait until all
tool calls finish. Shepherd refuses a different version, symlinked runtime file, partially patched
file, or unexpected upstream source shape instead of guessing.

Every answered gate has explicit provenance. Direct terminal answers default to `human`; an agent
answering from approved issue context must identify itself instead of impersonating a human:

```bash
go run ./cmd/shepherd run \
  --config /absolute/private/path/shepherd.json \
  --issue 372 \
  --command discuss \
  --decision-actor shepherd \
  --decision-basis "explicit user-approved issue context"
```

Render the protected ledger as the deterministic section that must be copied into the PR summary
or a PR comment before handoff:

```bash
go run ./cmd/shepherd decisions --config /absolute/private/path/shepherd.json --issue 372
```

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
tool output, or credentials. The decision ledger stores only bounded question/answer labels and
their actor, basis, canonical unit, execution identity, and timestamp.
