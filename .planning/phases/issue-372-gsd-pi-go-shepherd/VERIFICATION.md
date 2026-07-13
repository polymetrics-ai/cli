# Verification Checklist

- [x] Pinned GSD Pi version resolves to 1.11.0.
- [x] `gpt-5.6-sol` appears in the Codex provider catalog with high reasoning support.
- [x] A disposable governed session was observed as GPT-5.6 Sol/high after the initial `off` mismatch was made admission-fatal.
- [x] Supported headless query and filtered lifecycle events observed.
- [x] Unsupported/ambiguous headless behavior recorded fail-closed.
- [x] Workflow contract tests pass.
- [x] `go test ./...` passes inside `agent-runtime/shepherd/`.
- [x] `go test -race ./...` passes inside `agent-runtime/shepherd/`.
- [x] Root `go list ./...` excludes the nested module.
- [x] Root `go test ./...` and `go build ./cmd/pm` passed before the adversarial hardening slice.
- [x] Root `make verify` passed after Podman runtime qualification and package hardening.
- [x] Core named incident guard suite passes.
- [x] Governed intake canary emits <=15-second heartbeat and corrects premature upstream success to blocked.
- [x] A delayed human response keeps emitting heartbeats; the Go deadline cancels before GSD's
  fallback timer, so confirmation cannot auto-approve.
- [x] Stable issue identity, typed context binding, persistent attempts, explicit blocked resume,
  real `skip`/`stop` query shapes, tool errors, project setting overrides, and symlink escape have
  adversarial tests.
- [ ] Merge-disabled canary reaches exact-head human gate.
- [ ] Cleanup inventory reviewed before deletion.
- [x] Scoped Podman image/build-cache prune reclaimed approximately 18.14 GB without touching
  volumes.
- [x] Pinned image builds and runs as non-root `shepherd` UID/GID 1000 with a writable governed
  home and exact GSD 1.11.0 admission.
- [x] Pinned image includes the approved Go/research toolchain, excludes GitHub/publisher binaries,
  and completes a real governed query through Shepherd.
- [x] Runtime probes report Go 1.25.12, GSD Pi 1.11.0, agent-browser 0.31.1, Git, Make, jq, and
  ripgrep; the bounded curl shim completes GET and rejects POST.
- [x] Agent-browser opens and snapshots HTTPS with content boundaries and denies `eval`.
- [x] Context7 is written as a trusted HTTP MCP server into protected task state; worktree MCP
  configuration is not copied.
- [x] SearXNG runs by exact digest on `shepherd-research`, generates its secret at runtime, exposes
  no host port, and returns JSON search results to a governed worker container.
- [x] `scripts/gsd prompt programming-loop ...` renders and the Pi compatibility extension merges
  the same local command registry.
- [x] Every durable Shepherd decision is synchronized to one marker-owned PR summary comment, with
  actor/basis provenance intact and idempotent create/update tests passing.
- [x] Measured 1.7 MiB and 3.9 MiB nested-agent lifecycle returns fit the bounded 8 MiB scanner
  ceiling while the event projector continues to discard raw payloads.
- [x] Successful implementation units checkpoint only immutable issue-context `write_scope` paths;
  out-of-scope changes fail before staging, and subsequent dispatch still starts clean.
- [x] Dirty-tree finalization preserves the primary canonical reconciliation failure for actionable
  incomplete-task recovery.

## Current canary blocker

The protected controller adopted `M001-k9bwxs` without duplicating it, and a fenced `next` attempt
returned exit 10 with no agent/tool events. Targeted GSD audit sampling showed the authoritative DB
contains the unplanned milestone while markdown cannot represent its empty vision/slices; GSD asks
for an interactive DB-to-markdown rebuild/menu step and refuses headless dispatch. Shepherd recorded
the attempt as blocked and now requires a one-shot human resume decision. No recovery command,
authority reset, state-directory substitution, or legacy cleanup was used.

The root cause is GSD 1.11 planning compatibility: it auto-detected the repository's existing
multi-milestone `.planning` tree, while its DB-to-planning writer reports that only `flat-phases`
is supported. The replacement now has a Podman backend that overlays task-isolated `.gsd` and
`.planning` mounts and exposes only the code worktree plus explicit read-only auth/settings files.
The approved image/build-cache-only prune reduced Podman images from 42.83 GB to 24.69 GB without
touching volumes. The image now builds, passes exact-version admission, and runs as the expected
non-root identity. Real qualification found missing Git, ARM64 browser availability, transient npm
reliability, and login-shell PATH gaps. The approved image now contains the repository build and
read-oriented research surface with bounded retries and multi-architecture browser setup. SearXNG
remains a separate private sidecar and Context7 uses GSD Pi's native HTTP MCP client. Issue 372's
prior delivery generation remains blocked and was not resumed; a full AI canary still requires an
explicit resume decision or a fresh governed delivery that does not violate immutable milestone
binding.
