# Issue #389 Shepherd Hardening Summary

## Accepted checkpoints

- Slice A — real independent Sol/high validation and ratification:
  `95a17f18274c87ed0e3fde825b41257039b757de`.
- Slice B — durable attempt worktree lifecycle and crash recovery:
  `1a050692f9e47b5b4d3d74cfb38e56c67d461399`.
- Slice C — crash-safe Git/GSD-state promotion:
  `f0fbf47f54c688792a5d53edfa4b680b38b39eed`.
- Slice D — official GSD 1.11 metadata/runtime hardening:
  `cacb32e8e16b7ba70742cc5365cb83fffd74ca35`.

## Completed active slice

Slice E is GREEN and checkpoint-ready: static recovery text and broad retry handling are replaced by
exhaustive typed failure policy, a separately governed no-tool GPT-5.6 Sol/high Pi planner, strict
stream/current-session evidence, durable globally ordered per-class budgets, deterministic restart-safe
backoff, and controller-selected typed actions/primitives. Unsafe or ambiguous joined failures dominate;
GitHub/outbox uncertainty is typed and blocks without another write. Reservations, results, expiry, and
dispatch are bound to owner plus lease epoch. Failed retries remain in fresh Slice B worktrees and cannot
replay an older plan. Production work stayed on the `local_critical_path`; only read-only sidecars
overlapped.

## Last accepted base slice

Slice D remains the accepted base: regex/hard-coded registry compatibility is replaced by bounded strict
normalized metadata exported from the validated pinned official GSD Pi 1.11.0 runtime. Full official
tool contracts and reasons are preserved, models route from phase metadata, versioned sidecars remain
separate, and fallback/partial/unknown metadata fails closed.

Workflow: `scripts/gsd prompt programming-loop init --phase issue-389-shepherd-hardening --dry-run`.
Execution decision: `local_critical_path`; read-only recon/review sidecars are allowed, no overlapping
mutating workers.

Skills loaded: `gsd-core`, `polymetrics-issue-delivery`, `gsd-programming-loop`, `golang-how-to`,
`golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`,
`golang-context`, `golang-concurrency`, `golang-database`, `golang-design-patterns`,
`golang-structs-interfaces`, `golang-observability`, `golang-lint`, `golang-code-style`,
`golang-naming`, and `golang-documentation`.

Host execution is qualified only for `darwin/arm64` with the exact Node 24.13.1 binary and complete
v3 source/copied/sealed manifests. Registry import uses verified immutable bytes; prompts, settings,
project preferences, model/thinking transitions, current-run sessions, unit IDs, workflow tools, and
durable attempt identity are checked fail-closed. State-only units receive fresh hook-disabled
checkpoints. Exporter/runner/validator process trees are synchronously cleaned. Podman and disposable
unit continuation remain retained but fail closed until separately qualified.

Slice E focused/full/race tests, vet, build, nested/root `make verify`, module boundary, root package
listing, formatting, and diff checks pass. The live recovery smoke proved a fresh no-tool
`openai-codex/gpt-5.6-sol`/`high` session and strict bound result. Lint remains exactly 28 accepted
findings with no `internal/recovery` or new Slice E production finding. Independent Sol/high
correctness/security review cycles are fully dispositioned. The selected same-UID host model remains
explicitly documented as an architecture trust assumption rather than an isolation claim.

Slice D is pushed and accepted at `cacb32e8e16b7ba70742cc5365cb83fffd74ca35`, the immutable Slice E
base. GSD activation used `scripts/gsd prompt programming-loop run --phase issue-389-shepherd-hardening
--mode auto`. Slice E is ready for one coherent checkpoint commit and push. Slice F onward, PR creation,
final issue review, canaries, general GitHub mutation, cleanup/migration, and parent PR #390 merge remain
blocked/human-gated.
