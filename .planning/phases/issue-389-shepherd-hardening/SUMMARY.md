# Issue #389 Shepherd Hardening Summary

## Accepted checkpoints

- Slice A — real independent Sol/high validation and ratification:
  `95a17f18274c87ed0e3fde825b41257039b757de`.
- Slice B — durable attempt worktree lifecycle and crash recovery:
  `1a050692f9e47b5b4d3d74cfb38e56c67d461399`.
- Slice C — crash-safe Git/GSD-state promotion:
  `f0fbf47f54c688792a5d53edfa4b680b38b39eed`.

## Completed local slice

Slice D only is locally complete: regex/hard-coded registry compatibility is replaced by bounded strict
normalized metadata exported from the validated pinned official GSD Pi 1.11.0 runtime. Full official
tool contracts and reasons are preserved, models route from phase metadata, versioned sidecars remain
separate, and fallback/partial/unknown metadata fails closed.

Workflow: `scripts/gsd prompt programming-loop init --phase issue-389-shepherd-hardening --dry-run`.
Execution decision: `local_critical_path`; read-only recon/review sidecars are allowed, no overlapping
mutating workers.

Skills loaded: `gsd-core`, `polymetrics-issue-delivery`, `gsd-programming-loop`, `golang-how-to`,
`golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`,
`golang-context`, `golang-concurrency`, `golang-design-patterns`, `golang-structs-interfaces`,
`golang-observability`, and `golang-lint`.

Host execution is qualified only for `darwin/arm64` with the exact Node 24.13.1 binary and complete
v3 source/copied/sealed manifests. Registry import uses verified immutable bytes; prompts, settings,
project preferences, model/thinking transitions, current-run sessions, unit IDs, workflow tools, and
durable attempt identity are checked fail-closed. State-only units receive fresh hook-disabled
checkpoints. Exporter/runner/validator process trees are synchronously cleaned. Podman and disposable
unit continuation remain retained but fail closed until separately qualified.

Focused and installed-runtime tests, full/race tests, vet, build, nested/root `make verify`, module
boundary, root package listing, formatting, and diff checks pass. Lint reports 28 known findings, below
the 29-finding baseline with zero Slice D differential production findings. Four independent local
review/security cycles are dispositioned; the selected same-UID host model is explicitly documented as
an architecture trust assumption rather than an isolation claim.

The coherent Slice D checkpoint is pushed at the branch head on
`fix/389-shepherd-proof-recovery`; local and remote heads match and the worktree is clean. Slice E
onward, PR creation, final Sol review, canaries, GitHub mutation, and parent PR #390 merge remain
blocked/human-gated.
