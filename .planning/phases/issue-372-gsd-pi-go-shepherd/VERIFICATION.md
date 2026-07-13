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
- [x] Scoped Podman image/build-cache prune reclaimed approximately 18.14 GB without touching
  volumes.
- [x] Pinned image builds and runs as non-root `shepherd` UID/GID 1000 with a writable governed
  home and exact GSD 1.11.0 admission.
- [ ] Pinned image includes GSD's required `git` executable; dependency addition awaits explicit
  human approval.

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
non-root identity. The integrated governed query then proved that the slim Node base lacks GSD's
required `git` executable. Installing that image dependency is the current explicit human gate;
issue 372 remains blocked and was not resumed.
