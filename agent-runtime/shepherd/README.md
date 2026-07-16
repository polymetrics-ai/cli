# Go Shepherd

This nested Go module governs pinned GSD Pi delivery without compiling into `pm`. GSD Pi owns
milestones and local work; Shepherd owns admission, liveness, human questions, authority, exact-head
ratification, external-effect intents, and privacy-safe telemetry.

## Install and configure

Install exact `@opengsd/gsd-pi@1.11.0` outside the repository. Copy `shepherd.example.json` to a
private local path and use absolute paths. `state_dir` must be protected and outside the worker
worktree. `attempt_root` must be outside both the canonical worktree and `state_dir`; neither may
contain the other. Provision the controlled `gsd_home` separately; never put
credentials in the config or repository. Its `agent/settings.json` must pin the configured provider,
coordinator model, and `defaultThinkingLevel: high`. Its `PREFERENCES.md` must use official GSD Pi
phase routing: research, planning, discussion, completion, validation, and UAT use the coordinator
model; execution, simple execution, and subagents use the implementation model. Each governed phase
must explicitly pin `thinking: high`. Project `.gsd/PREFERENCES.md` values may add unrelated policy,
but a conflicting phase override fails admission. GSD Pi 1.11.0 direct headless unit aliases do not
apply the auto/guided phase selector, so Shepherd explicitly launches direct execution on the
implementation model, atomically restores the coordinator default after the unit, and independently
verifies the effective phase model. Startup self-healing accepts only the exact governed
implementation-model drift; any other identity fails closed.

For governed delivery, use the host-local runtime by default (`"runtime": "host"`) with the exact
pinned GSD Pi loader in `gsd_command`, a separately allowlisted absolute Pi executable in `pi_command`,
protected `gsd_home`, and external-effect publishers disabled. Slice D qualifies host GSD execution
only on `darwin/arm64` with Node 24.13.1 and the exact complete GSD Pi 1.11.0 package-tree digest;
other host platforms fail closed until their Node/package manifests are reviewed and added. Independent candidate validation invokes
Pi directly in non-interactive JSON mode with GPT-5.6 Sol/high, a dedicated protected session directory,
an isolated credential-free home/environment, and only `read,grep,find,ls` (no shell or network-capable
tool); it does not dispatch a canonical GSD workflow unit. This is the
qualified #389 path and requires no Podman service for default verification.

The legacy Podman assets remain available for separately authorized qualification/debug runs; they
are not required by the default supervisor path and are not removed in this issue. Slice D resolves
local tags to immutable image IDs but deliberately fails runtime admission because no complete image
digest has yet been human-qualified. If a later human-approved qualification build is performed, build
the image explicitly so the agent sees
only the issue worktree, task-isolated GSD/planning mounts, and read-only auth/settings files:

```bash
podman build -t localhost/polymetrics-gsd-pi:1.11.0 \
  -f agent-runtime/shepherd/container/Containerfile .
```

The container does not mount host SSH, GitHub, cloud, or home-directory credentials. The governed
image contains Go 1.25.12, Make, Git, jq, ripgrep, official GSD Pi 1.11.0, Context7 MCP 3.2.3, and
agent-browser 0.31.1 with Chrome. `curl` is a read-only compatibility wrapper around `web-fetch`:
common GET/HEAD flags work, responses are capped at 2 MiB, and bodies, auth headers, uploads, and
non-HTTP(S) schemes are rejected. GitHub CLI and publisher credentials are absent.

Start the private search sidecar only before selecting `container_network: shepherd-research`:

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

## Supervise an issue milestone

Create a validated context file inside the isolated issue worktree, then run the one-command
supervisor:

```bash
cd agent-runtime/shepherd
go run ./cmd/shepherd supervise \
  --config /absolute/private/path/shepherd.json \
  --issue 389 \
  --context .planning/phases/issue-389-shepherd-hardening/CONTEXT.json
```

`supervise` materializes the protected issue context, adopts or bootstraps exactly one issue-local
GSD project identity, then loops over native `headless query`. It dispatches only the canonical
`next.unitType`: planning, research, validation, completion, and UAT units use the configured
GPT-5.6 Sol coordinator model with `high` thinking; `execute-task` and delegated execution use the
configured GPT-5.5 implementation model with `high` thinking. `discuss-milestone` is routed through
the targeted `discuss` path. Generic `auto`, concurrent dispatch, and parent-PR merge are never
performed. When canonical GSD reaches `phase=complete,next.action=stop`, Shepherd prints a bounded
`final_human_gate` status and exits without merging.

The controller persists immutable issue identity in both Shepherd SQLite and `.gsd/ISSUE.json`:
issue, parent issue, branch, base branch, worktree/project root, initial SHA, context hash, and GSD
version. Restarting with the exact same identity is idempotent; a mismatched issue or branch fails
closed. Unit attempts are durably reserved per `{delivery,generation,unit,head}`. The default and
maximum-configured retry budget is bounded by `max_unit_attempts` (default 3); Shepherd only
automatically retries reversible runtime, artifact, or interruption failures while budget remains.
Contract, model/thinking, authority, scope, stale-head, identity, and orphaned-child failures stop at
the typed blocked/human-gate boundary.

Canonical-unit work runs in an external attempt worktree with a durable SQLite lifecycle. Ratified
promotion is journaled before canonical mutation and binds the exact attempt, base/candidate heads,
proof, validator attestation, governance version, and bounded deterministic `.gsd` manifests. The
candidate `.gsd/gsd.db` is staged with SQLite's online backup API and integrity checked. Git advances
only from the exact base; `.gsd` is then installed with same-filesystem stage/backup renames and
parent-directory fsyncs. Startup recovery runs before GSD query or dispatch, finishes forward after
Git reaches the candidate, and blocks on moved Git, dirty state, expired pre-promotion authority, or
resource ownership/hash ambiguity. Direct in-place `.gsd` adoption is disabled.

The process prints normalized lifecycle events and at most one heartbeat every 15 seconds. Heartbeats
include bounded operational child metadata only: status, child count, in-flight tool, and child turn
count. They never include prompts, model reasoning, output bodies, or credentials. Native GSD
questions are forwarded to the terminal and require an explicit response. The Go deadline always
precedes GSD's fallback response timer. Answer files, inline context, chained `--auto`, and generic
`recover` are rejected.

## Start or run one fenced issue unit

For manual qualification, run one unit at a time:

```bash
cd agent-runtime/shepherd
go run ./cmd/shepherd start \
  --config /absolute/private/path/shepherd.json \
  --issue 373 \
  --context .planning/phases/issue-372-gsd-pi-go-shepherd/CONTEXT.json
```

Continue one fenced unit at a time with `run --issue 372 --command next`. Governed units now use
disposable attempt worktrees, so `--continue-unit` fails closed until prior-attempt session identity
and worktree continuity are independently qualified; start a fresh canonical discussion unit instead.
If a prior qualification run already created the correct active milestone, `start --adopt-existing` binds it
explicitly instead of silently creating a second milestone.

`start` also binds an immutable copy of the validated issue context under protected controller
state. A successful `execute-task` may edit source files, but Shepherd verifies every changed path
against that context's `write_scope` and creates a local, hook-disabled checkpoint commit before
the next clean dispatch. The worker runtime still cannot push.

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

Every governed config binds a repository and pull request. After each answered gate, Shepherd
immediately synchronizes the fsynced ledger to one marker-owned PR comment. The comment preserves
the actor and concise basis; a Shepherd or contract answer is never presented as human. Publication
failure fails the governed unit after retaining the durable local decision for reconciliation.

Render the protected ledger locally, or explicitly reconcile the PR comment after an interrupted
publication:

```json
{"repository":"polymetrics-ai/cli","pull_request":388}
```

```bash
go run ./cmd/shepherd decisions --config /absolute/private/path/shepherd.json --issue 372
go run ./cmd/shepherd decisions --config /absolute/private/path/shepherd.json --issue 372 --publish
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
