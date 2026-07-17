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
`git diff --check` passes. The temporary worker lacked the packaged loader, but canonical deletion and
full integration suites now pass in normal/race modes.

## Canonical review-fix verification

Canonical focused/full Shepherd normal and race tests, focused deletion and full integration normal
and race tests, vet/build/nested `make verify`, root `make verify`, module boundary, root package list,
planning JSON, formatting/diff hygiene, and binary cleanup pass. One new test-only `errcheck` from the
first lint run was fixed; default and integration-tagged lint reruns exactly match the accepted
25 `errcheck`, 2 `staticcheck`, 1 `unused` baseline with zero differential. `verificationPassed=true`.
Two fresh independent read-only GPT-5.6 Sol/high correctness and security reviews at
`c72778def85ddccdee91bd648d7c0d569eb5fa94` against the exact parent base pass with no findings.
A replacement review follows this final docs-only evidence commit before push. No live external
operation, credential access, canary, cleanup/migration, PR merge, or `main` mutation occurred.

## Draft PR #456 nested-module CI repair

Draft PR #456 opened cleanly at `7432f0a5da90f255b74307d12c26863b61c1a16f`; nine checks passed and
`nested-module` failed in workflow `29578379908`, job `87877981167`. Human authorization now permits
only GSD test fixture/assertion changes and issue-389/PR evidence.

The first failure class is test portability: registry/snapshot fixtures hash the lexical Node path.
Production intentionally opens qualified runtime files with `O_NOFOLLOW`, so a symlinked `node` is
rejected. The ordinary local NVM Node is already canonical and its unedited test passes; a temporary
symlinked PATH reproduces the exact CI error, while the canonical target directory passes. The test-only
repair will resolve the complete symlink chain before existing ownership/type/mode/hash qualification.
Production `resolveQualifiedNode` remains fail-closed for symlinks.

The second failure is assertion timing: one runner test performs an immediate PID probe after cleanup,
while adjacent GSD process tests use bounded eventual `ESRCH` verification. The test-only repair will
share a strict five-second bounded helper without accepting a running descendant or changing production
cleanup. Implementation commit `20540e79bf8929390e64fcd165046d6704199e6b` changes only five
`internal/gsd/*_test.go` files: all fixture Node commands/hashes use one complete-chain canonical helper,
production symlink rejection has a direct regression, and runner/process cleanup assertions use one
strict bounded `ESRCH` helper that cannot accept a live descendant.

Targeted Node/descendant/security sets pass at count 10; full GSD race count 3 and normal count 5 pass.
Full nested normal/race/integration/race-integration/vet/build/make and root verify/boundary/list/diff/
JSON/hygiene gates pass. Lint remains exactly 25 `errcheck`, 2 `staticcheck`, 1 `unused`, with zero
changed-file finding. Generated binaries are absent and `verificationPassed=true`. Fresh independent
read-only GPT-5.6 Sol/high review of `7432f0a5...2b0c5ea7` passes with no findings and confirms the six
required scope/security/portability/assertion/evidence/gate conditions. Replacement review at
`0ab3651cfa437f035b2bfc81ce7a82483c37f5d5` also passed before push.

Fresh CI run `29587523261`, job `87908092713`, confirmed the descendant assertion fix and exposed the
remaining cross-owner portability condition: GitHub's canonical host Node is not owned by the runner,
so unchanged production bounded/no-follow/owned-regular qualification correctly rejects it. Human
authorization now permits one test-process-owned Node copy only. The fixture will resolve host Node only
for source bytes, stream at most 256 MiB plus one into an exclusively created executable under a private
0700 temp directory, verify ownership/type/mode without following symlinks, share it once per test
process, and remove the exact directory through package lifecycle. Production policy and code remain
immutable.

The implemented fixture is lazy and package-scoped: `sync.Once` copies Node only for test processes that
request it, while `TestMain` removes the exact private directory. The copy is streamed through a 256 MiB
+ 1 limiter into an exclusively created file, synced, closed, chmodded 0500, and `Lstat`-verified as an
owned regular non-symlink executable. Repeated calls return one path; unchanged production qualification
accepts it and rejects a symlink to it.

Focused count-10, GSD race count-3/normal count-5, full nested normal/race/integration/race-integration/
vet/build/make, root verify/boundary/list/diff/JSON/hygiene, exact lint baseline, fixture-temp cleanup,
and generated-binary absence all pass. `verificationPassed=true`; state is
`stacked_pr_ci_recheck_pending`. Exact-head review, normal push, and fresh CI remain; CI success is not
claimed. No canary or merge is authorized.
