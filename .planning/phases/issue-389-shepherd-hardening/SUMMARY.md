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
- Slice E — real Sol/high recovery planning:
  `9556cb24412f3598b2b8a94a3089b61ef3d1dd91`.
- Slice F — authority-gated external effects:
  `ea88c92f5f3c0b1c5f3f434fa52efba24624f803`.
- Slice G — real built-CLI supervise integration coverage:
  `ee474811378edd604e1e86e413f0bcafeced452b`.

## Accepted Slice E checkpoint

Static recovery text and broad retry handling are replaced by exhaustive typed failure policy, a
separately governed no-tool GPT-5.6 Sol/high Pi planner, strict stream/current-session evidence, durable
globally ordered per-class budgets, deterministic restart-safe backoff, and controller-selected typed
actions/primitives. Unsafe or ambiguous joined failures dominate; GitHub/outbox uncertainty is typed
and blocks without another write. Reservations, results, expiry, and dispatch are bound to owner plus
lease epoch. Failed retries remain in fresh Slice B worktrees and cannot replay an older plan.

## Accepted Slice F checkpoint

The immutable base is `9556cb24412f3598b2b8a94a3089b61ef3d1dd91`; local and remote heads matched
and the worktree was clean before planning-only edits. Slice F makes a strict durable fenced outbox the
only production path for external write effects. It separates request, controller-derived authorization,
enqueue, claim, execution, reconciliation, and terminal persistence. Only the outbox GitHub executor
receives write capability; reply polling stays read-only. Strict immutable payloads, owner/epoch-fenced
claims, stable comment markers, uncertainty handling, monotonic summary revisions, question binding,
cross-store startup reconciliation, and bounded typed telemetry are required. Merge capabilities and
unsupported future effects fail closed.

Workflow: `scripts/gsd prompt programming-loop run --phase issue-389-shepherd-hardening --mode auto`.
Execution decision: `local_critical_path`; read-only tester/reliability/review/security sidecars are
allowed, no overlapping mutating worker. No live GitHub mutation is permitted.

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

Slice E remains the immutable Slice F base. Slice F RED was captured before production edits; the
implementation now centralizes writes in a durable fenced outbox, persists immutable target and effect
identity, reconciles exact GitHub markers and uncertain writes, atomically applies human replies, and
requires exact-head promotion proof for final-gate projection. Focused/full/race/vet/build, nested/root
verification, module boundary, root package listing, formatting, and diff checks pass. Lint remains the
accepted 28 findings with no Slice F package finding. Repeated independent correctness, security, and
restart/reliability review findings are fully dispositioned. No live GitHub mutation occurred.

Slice F is accepted at `ea88c92f5f3c0b1c5f3f434fa52efba24624f803`; exact local/remote equality
and a clean worktree were confirmed before Slice G activation.

## Accepted Slice G checkpoint

Slice G now adds build-tagged process-level coverage for the actual built `shepherd supervise` command.
The harness uses isolated real Git repositories and SQLite stores while replacing only GSD/Pi/GitHub
processes with strict bounded fakes. It proves official-metadata GPT-5.5/high execution and GPT-5.6
Sol/high planning, fresh implementation/validator/recovery sessions, exact diff/hash/phase/tool proof,
ratification, WAL-normalized staged Git/GSD promotion, legacy post-Git forward recovery, complete
canonical-state rejection oracles, outbox reconciliation/collision/uncertainty, exact human reply
binding, two-heartbeat cadence, and terminal `final_human_gate` without merge capability.

Strict RED exposed the missing compile-only process seam. Later REDs exposed artifact mutation after
validator return, empty governed deltas, validator deadline typing, unbound staged GSD state, transient
WAL/SHM hashing, legacy proof recovery, and final-gate GSD drift. Production now normalizes installable
GSD state in protected storage, binds it into evidence, verifies the exact stage, preserves Slice-F
post-Git recovery, and rechecks canonical GSD before every final-gate projection. Exact-head review fixes
also keep awaiting-decision supervision alive with fenced polling/expiry, enforce complete Pi/GSD lifecycle
and tool pairing, clean every GSD process group after ordinary exit, and cover SIGINT plus both pre-send
outbox boundaries. Release builds compile only inert seam implementations.

Normal/race integration (including race-built child binaries), full nested unit/race/vet/build and
`make verify`, root `make verify`, module boundary/root listing/diff/JSON checks, and default/tagged lint
all pass; lint is exactly the accepted 28 findings with zero differential. Independent GPT-5.6
Sol/high findings from exact-head reviews `45927348`, `b08c93cc`, `c1a34d23`, `ee8f1fa7`, and later
closure cycles are dispositioned, including complete fresh validator and implementation
turn/session/durable-proof provenance, strict case-fold-safe lifecycle JSON, canonical assistant rows,
and bounded detached output draining. Final immutable exact-head correctness, security, restart,
verification, and test-realism review at `ee474811378edd604e1e86e413f0bcafeced452b` reports no
findings. Local/remote equality, cleanliness, and generated-binary absence were confirmed.

## Post-Slice-G parent synchronization

Parent PR #390 was revalidated open/draft from `feat/372-gsd-pi-go-shepherd` to `main` at exact head
`d72e597e35b5104cf58936612053705c280fc2b1`. Pre-squash head
`c539b49bd767b0839f0989d52bd69da80c30843e` is a Slice G ancestor; its tree and the squashed parent
tree are both `9c9ffd9a0e0f6d76955cd048978662d57e888291`. The guarded `-s ours` merge
`17ca31f6d04def71d55137d25d8194feaea10829` repairs ancestry with no diff from accepted Slice G.
This planning reconciliation is intentionally separate and phase-directory-only.

Fresh full verification and exact-head GPT-5.6 Sol/high review follow this checkpoint. Only after they
pass may this branch be pushed and a draft stacked PR be opened. Execution remains
`local_critical_path`; canaries, cleanup/migration, every PR merge, and all `main` mutation remain
blocked.

## Post-Slice-G exact-head review fix

The authorized review fix now adds typed deletion artifacts and bounded Git execution. Git artifact
manifests parse `--name-status -z --no-renames`, bind deletions with `deleted:true` plus the exact
zero-hash sentinel, stream present object hashing through an 8 MiB maximum, and expose typed
`ErrOutputLimit` for bounded stdout/stderr/object overflow. Validator requests preserve non-deleted JSON
via `omitempty`, enforce flag/sentinel consistency, and revalidate deleted paths as absent through a
lexical no-follow component walk before and after independent validation. Promotion proof decoding and
process integration fixtures understand deleted artifacts.

Focused Git/validation/command tests, focused race, full nested normal/race, vet/build/nested
`make verify`, root `make verify`, module-boundary, diff, root package list, and default/tagged lint
baseline gates passed. The new built-CLI deletion integration tests could not execute in this isolated
checkout because the packaged official GSD loader is absent; `verificationPassed` intentionally remains
false and coordinator-owned exact-head review/full integration gates remain pending.

## Bounded Git / descriptor-root follow-up

Starting clean at `bfc937ef2bc523950c14929b73b00d9e054957d6`, the follow-up closes confirmed adjacent
bounded-Git gaps without amending or rebasing. Git status parsing is now strict before hashing and rejects
malformed/extra terminators plus more than 128 artifacts before any `cat-file` spawn. `hashGitObject`
uses bounded `cat-file -s`, rejects oversized immutable objects before streaming, accepts exact 8 MiB,
verifies streamed bytes match the declared size, and cancels/reaps the Git process group on stream or
stderr overflow. Generic Git execution uses internal `ErrOutputLimit` cancellation with parent context
precedence and finite process-tree cleanup. Validator artifact reads now use descriptor-relative
`os.Root` stable identity checks, and validation/promotion alias the Git deletion sentinel.

Focused normal/race Git/validation/cmd tests, full nested normal tests, vet, build, and diff checks pass.
Deletion integration is still blocked by the absent packaged official GSD loader, and exact-head review is
coordinator-owned; `verificationPassed=false` remains intentional. The shared repository Git
config/environment concern is declined as out of scope under the accepted same-UID host trust assumption.

## Cleanup / UTF-8 residual follow-up

Starting clean at `ec8c2dc523a2ce55c0d4a4bcbd9b5739df541fad`, this follow-up closes three adjacent
residuals without amending or rebasing. Git process groups are explicitly cleaned after ordinary parent
exit for both generic `run` and `hashGitObject`/`cat-file`, with parent context priority over output
limits and output limits over cleanup errors. Validator artifact revalidation now runs as per-artifact
work so descriptor roots/files close before the next bounded item. Git name-status status/path records
must be valid UTF-8 before conversion, hashing, deletion sentinel evidence, or JSON proof construction.

Focused normal and race gates for `./internal/git ./internal/validation ./cmd/shepherd` pass, and
`git diff --check` passes. `verificationPassed=false` remains intentional pending coordinator-owned
integration and exact-head review gates.
