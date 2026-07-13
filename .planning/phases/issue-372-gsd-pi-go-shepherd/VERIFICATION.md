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
- [ ] Root gates rerun after the adversarial hardening slice.
- [x] Core named incident guard suite passes.
- [x] Governed intake canary emits <=15-second heartbeat and corrects premature upstream success to blocked.
- [x] A delayed human response keeps emitting heartbeats; the Go deadline cancels before GSD's
  fallback timer, so confirmation cannot auto-approve.
- [x] Stable issue identity, typed context binding, persistent attempts, explicit blocked resume,
  real `skip`/`stop` query shapes, tool errors, project setting overrides, and symlink escape have
  adversarial tests.
- [ ] Merge-disabled canary reaches exact-head human gate.
- [ ] Cleanup inventory reviewed before deletion.

## Current canary blocker

The protected controller adopted `M001-k9bwxs` without duplicating it, and a fenced `next` attempt
returned exit 10 with no agent/tool events. Targeted GSD audit sampling showed the authoritative DB
contains the unplanned milestone while markdown cannot represent its empty vision/slices; GSD asks
for an interactive DB-to-markdown rebuild/menu step and refuses headless dispatch. Shepherd recorded
the attempt as blocked and now requires a one-shot human resume decision. No recovery command,
authority reset, state-directory substitution, or legacy cleanup was used.

The root cause is GSD 1.11 planning compatibility: it auto-detected the repository's existing
multi-milestone `.planning` tree, while its DB-to-planning writer reports that only `flat-phases`
is supported. The replacement now has a Podman backend design that overlays task-isolated `.gsd`
and `.planning` mounts and exposes only the code worktree plus explicit read-only auth/settings
files. Its image build is blocked by Podman VM capacity: 42.83 GB of images (41.19 GB reclaimable)
and 44.74 GB of volumes (17.64 GB reclaimable). Per the cleanup gate, no prune/removal was run.
