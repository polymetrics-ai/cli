# Verification — Issue #475

Current result: Cycle 13 issue-owned verification is complete at implementation/refactor
checkpoints `48f546a5` / `e50b5f97`; see **Cycle 13 Exact-Head Verification** below. The
complete-suite status remains environment-blocked, not green, because the managed sandbox denies
the unchanged controller/state-store process-identity children with `spawn EPERM`.

## Cycle 13 Exact-Head Verification

Cycle 13 was executed against frozen reviewed candidate
`5dafc5725167bb74ce88a723073b8c4ceb8314e0` and immutable base
`e659d6f1b666f58748e2d8c86599ceb4bbc62ff8`. Both complete Cycle 12 reports were read before PLAN
and re-read after implementation. Their two-plus-five blocker union maps exactly to the seven
named C13 rows; no row is omitted or deferred.

The ordered checkpoints are PLAN `61a364e0`, read-only seam-map amendment `5e86520c`, test-only RED
`974d2e79`, RED evidence `6d3ce03a`, cohesive GREEN `48f546a5`, and classifier/lifecycle refactor
`e50b5f97`. RED executed all 131 focused rows: every one of the 124 retained rows passed and
exactly seven intended C13 rows failed while strict focused TypeScript passed and all three
production blobs stayed frozen. GREEN/refactor passes all 131 rows.

| Cycle 13 gate | Status | Evidence |
|---|---|---|
| RED integrity | pass | 131 executed; 124 retained pass; exactly 7 intended behavior failures; strict TS pass; all production blobs frozen |
| Focused GREEN | pass | 131 passed, 0 failed/skipped/cancelled/todo |
| Complete Shepherd | environment-blocked | 268 executed; 237 passed; unchanged 31 controller/state-store `spawn EPERM` failures; 0 skipped/cancelled/todo |
| Environment isolation | pass | excluding only controller/state-store, 194 passed, 0 failed/skipped/cancelled/todo |
| Strict TypeScript | pass | focused production/tests and all 12 non-test Shepherd `.ts`, TypeScript 5.9.3 with explicit Pi 0.80.6 roots |
| Pi offline boundary | pass | explicit binary reports package 0.80.6; RPC `get_commands` registers `pm-shepherd` and exits 0 |
| Repository integrity | pass | diff, JSON, immutable-base/frozen-start ancestry, exact same 20 paths, credential/dependency/Go/connector scans, clean worktree |
| External mutation | not attempted | no push, network, GitHub, review bot, merge, model/auth, credential, service, Go, connector, `make`, parent, or #478 action |

Final production blobs are runtime `cd5c05411933c1a1f1b239d8ac85112e47e10b8b`, tool policy
`5a7f91b863f3a3eba3b489e79944c17a6511a776`, and role prompts
`d4365dd2e32854589a7d1bee91439e5cb0a17fe0`. The bounded-array contract intentionally makes
non-index peers untouched, discarded, and non-authoritative because ECMAScript has no bounded
primitive that can prove arbitrary hidden/symbol-key absence. Parent orchestration owns the
process-capable complete-suite rerun, fresh independent exact-head review, integration, and
delivery.

## Cycle 11 Verification Result

Cycle 11 was executed against exact clean reviewed candidate
`1571dc4d4f45ad4285107d04f2d7c489a7f357ab` and immutable base
`e659d6f1b666f58748e2d8c86599ceb4bbc62ff8`. The complete
`/tmp/475-REVIEW-CYCLE10-1.md` and `/tmp/475-REVIEW-CYCLE10-2.md` reports are fully accepted and
mapped in `REVIEW.md` and `PRD-COVERAGE.md`.

The ordered checkpoints are artifact-only PLAN
`9366296dcde200bf1f21e74d3cd8dec321581155`, read-only installed-Pi trace
`a2a8b0e7da426f8c0c6fac91ead65d6a19c4534a`, comprehensive test-only RED
`c58865202623805f8877a583eecf5e301b589f3d`, first runtime/policy GREEN
`1e605675f8e021a14ed7f709451a2d3a8111c6ad`, and complete-envelope stream refactor
`d9b4eaee71907c662f87f737c9b1a901c35146f9`. RED executed all 114 focused tests: every one of
the 102 retained assertions passed and exactly 12 named Cycle 11 rows failed their intended
production assertions. Strict focused TypeScript passed and all three frozen production blobs
remained exact. GREEN/refactor passes all 114 focused tests.

| Cycle 11 gate | Status | Evidence |
|---|---|---|
| RED integrity | pass | 114 executed; 102 retained pass; exactly 12 intended behavior failures; strict TS pass; runtime/policy/role-prompt blobs unchanged |
| Focused GREEN | pass | 114 passed, 0 failed, 0 skipped/cancelled/todo |
| Complete Shepherd | environment-blocked | 251 executed: 220 passed; the unchanged 31 controller/state-store failures all originate from managed-sandbox `/bin/ps` `spawn EPERM` |
| Environment isolation | pass | excluding only `controller.test.ts` and `state-store.test.ts`, 177 passed, 0 failed |
| Strict TypeScript | pass | focused production/tests and all 12 non-test Shepherd `.ts`, TypeScript 5.9.3 with explicit Pi 0.80.6 roots |
| Pi offline boundary | pass | explicit binary `0.80.6`; writable temporary agent directory; RPC `get_commands` registered `pm-shepherd`; C11-01 exercised the real no-model factory/result path and cleaned the actual session |
| Repository integrity | pass | `git diff --check`, valid JSON, immutable-base/frozen-head ancestry, exact issue-owned path scope, no Go/connectors, and clean worktree |
| External mutation | not attempted | no push, network, GitHub, review bot, merge, model, credential, service, Go, connector, or `make` action was authorized in this lane |

The runtime now accepts Pi 0.80.6's required `extensionsResult.runtime` field as descriptor-checked
compatibility evidence without storing or invoking it. Native-only signal leases, admission/close
linearization, run-associated creation ownership, complete stateful assistant projections,
replacement-safe aggregate stream accounting, fixed-envelope/bounded-JSON adapters, and total
failure normalization close the runtime findings. The policy closes forbidden capability
compounds, AWS SSO/CLI caches, Cookie/Set-Cookie headers, and qualified sensitive assignment keys.
Public `HostCapability` and `ScopedWorkspace` interfaces and serialized handoff APIs are unchanged.

All twelve Cycle 11 top-level RED assertions remain intact. Post-RED test edits are additive or
fixture-only: signature and diagnostic-envelope accounting subcases were added; retained fixtures
now provide canonical Pi assistant `api`/`usage`, expect native signal operations to ignore shadow
hooks, and treat hidden/symbol peers as inert discarded source extras at bounded arbitrary DTO and
array boundaries.

The healthy GSD adapter still rejects `programming-loop`, so the required manual-GSD fallback is
recorded without weakening PLAN -> RED -> GREEN/refactor -> verify. The complete-suite result is
neither a pass nor a product regression; it reproduces the same parent-owned sandbox limitation,
while every other Shepherd test passes in isolation. Parent orchestration owns the
process-capable complete-suite rerun, fresh exact-head review, integration, and delivery.

## Cycle 10 Verification Result

Cycle 10 was executed against exact clean reviewed candidate
`f63957aed6fd1406eb3bd9a82adbd10b23b34c33` and immutable base
`e659d6f1b666f58748e2d8c86599ceb4bbc62ff8`. The complete
`/tmp/475-REVIEW-CYCLE9-1.md` and `/tmp/475-REVIEW-CYCLE9-2.md` reports, including WR-01, are fully
dispositioned in `REVIEW.md` and covered in `PRD-COVERAGE.md`.

The ordered checkpoints are artifact-only PLAN `0eb7999f29e538c5a15d9c10f37b167be19817de`,
one comprehensive test-only RED `6df77689d7bcd3a25d9028af258694e84d24f238`, and cohesive
GREEN/refactor `a88cbe5242f070059ea49446ffac6914716a8c5d`. RED executed all 102 focused
tests: every one of the 86 retained assertions passed and exactly 16 named Cycle 10 behavior tests
failed their intended assertions. Strict focused TypeScript passed and all three production blobs
remained exact. GREEN passes all 102 focused tests.

| Cycle 10 gate | Status | Evidence |
|---|---|---|
| RED integrity | pass | 102 executed; 86 retained pass; exactly 16 intended behavior failures; strict TS pass; runtime/policy/role-prompt blobs unchanged |
| Focused GREEN | pass | 102 passed, 0 failed, 0 skipped/cancelled/todo |
| Complete Shepherd | environment-blocked | 239 executed: 208 passed; the unchanged 31 controller/state-store failures all originate from managed-sandbox `/bin/ps` `spawn EPERM` |
| Environment isolation | pass | excluding only `controller.test.ts` and `state-store.test.ts`, 165 passed, 0 failed |
| Strict TypeScript | pass | focused production/tests and all 12 non-test Shepherd `.ts`, TypeScript 5.9.3 with explicit Pi 0.80.6 roots |
| Pi offline boundary | pass | explicit binary `0.80.6`; writable temporary agent directory; RPC `get_commands` registered `pm-shepherd`; retained real-validator/no-model tool exercise passed |
| Repository integrity | pass | `git diff --check`, immutable-base/frozen-head ancestry, issue-owned path scope, and clean worktree |
| External mutation | not attempted | no push, GitHub, review bot, merge, model, credential, service, Go, or connector action was authorized in this lane |

All sixteen Cycle 10 RED assertions are preserved byte-for-byte through GREEN. Post-RED test changes
only align older controls with the newly accepted interfaces: three legacy successful-handoff
fixtures render equivalent sensitive evidence on one terminal-safe line because WR-01 rejects the
original control-bearing string before redaction; the harmless documentary fixture uses the
allowed colon form, and an additive assertion proves the former equals form redacts under BL-04.

The healthy GSD adapter still rejects `programming-loop`, so the required manual-GSD fallback is
recorded without weakening PLAN -> RED -> GREEN/refactor -> verify. The complete-suite limitation
is neither called a pass nor a product regression: it is the same parent-owned sandbox denial, and
isolation proves every other Shepherd test passes. Parent orchestration owns a permitted-environment
complete-suite rerun, fresh exact-head review, integration, and any external delivery action.

## Cycle 9 Verification Result

Cycle 9 is planned against exact clean reviewed candidate
`0cdcda7e049b7ecfa2fdc52027c66c5de161f2c8` and immutable base
`e659d6f1b666f58748e2d8c86599ceb4bbc62ff8`. The two complete Cycle 8 reviews are dispositioned in
`REVIEW.md`; all deduplicated lifecycle, typed-boundary, data-snapshot, event, parser, path,
capability, and terminal-safety findings are accepted into one correction.

The ordered checkpoints are PLAN `b175cc4a`, read-only delegation trace `7047a8f4`, test-only RED
`dbf796b3`, and cohesive GREEN/refactor `94918f4e`. RED executed all 86 focused tests: all 70
retained assertions passed and exactly 16 Cycle 9 assertions failed, strict focused TypeScript
passed, and all three production blobs matched `0cdcda7e`. GREEN passes all 86 focused tests.

| Cycle 9 gate | Status | Evidence |
|---|---|---|
| RED integrity | pass | 86 executed; 70 retained pass; exactly 16 intended assertion failures; strict TS pass; production blobs unchanged |
| Focused GREEN | pass | 86 passed, 0 failed |
| Complete Shepherd | environment-blocked | 223 executed: 192 passed; the unchanged 31 controller/state-store failures all originate from managed-sandbox `/bin/ps` `spawn EPERM` |
| Strict TypeScript | pass | focused production/tests and every non-test Shepherd `.ts`, TypeScript 5.9.3 with explicit Pi 0.80.6 roots |
| Pi offline boundary | pass | explicit binary `0.80.6`; temporary writable agent directory; RPC `get_commands` registered `pm-shepherd`; focused real-validator/no-model tool exercise passed |
| Repository integrity | pass | `git diff --check`, immutable-base/frozen-head ancestry, and issue-owned path scope |

The healthy 69-command GSD adapter registry still rejects `programming-loop`; Cycle 9 uses the
recorded manual-GSD fallback. The complete-suite result is neither claimed as a product failure nor
as a pass; it reproduces the same parent-owned sandbox limitation as Cycle 8. DNS-deferred push and
fresh exact-head review remain parent-owned.

## Cycle 8 Verification Result

Cycle 8 was executed against frozen reviewed head `f219b730c63adc9188c93093a40511433a3d0110`
and immutable base `e659d6f1b666f58748e2d8c86599ceb4bbc62ff8`. PLAN
`9dd71a812795b7ac74b07db06c4fae03a3004871`, pre-RED amendment
`04dc72f31a3bdd461045a4ef12d92c260f8ffd3f`, test-only RED
`11aa221231a52fab91f41dfce9742b7dfe180c02`, and cohesive GREEN/refactor
`c4d34c377532c903238400c986a6b488fab3646d` preserve the required order.

The RED ran all 70 focused tests with 53 retained passes and exactly 17 intended assertion-level
failures; strict focused TypeScript passed and production stayed byte-identical to `f219b730`.
GREEN passes all 70 focused tests, both strict TypeScript scopes, the pinned Pi 0.80.6 offline RPC,
`git diff --check`, immutable-base/frozen-head ancestry, and issue-owned path checks.

| Cycle 8 gate | Status | Evidence |
|---|---|---|
| Focused AgentSession/tool-policy tests | pass | 70 passed, 0 failed, exit 0 |
| Focused strict TypeScript | pass | TypeScript 5.9.3, explicit Pi 0.80.6 package/type roots, exit 0 |
| All-production strict TypeScript | pass | every non-test Shepherd `.ts`, same compiler/type roots, exit 0 |
| Explicit Pi 0.80.6 offline RPC | pass | `get_commands` returned success and registered `pm-shepherd`, exit 0 |
| Diff/base/frozen-head/owned-scope | pass | clean diff check; both ancestors present; only issue #475 paths changed |
| Complete Shepherd suite | environment-blocked | 207 executed: 176 passed, 31 failed only where the managed sandbox denied the pre-existing `/bin/ps` identity probe with `spawn EPERM`; only controller/state-store test files are affected |
| Push / remote-head equality | environment-blocked | `ssh: Could not resolve hostname github.com: -65563` |

The complete-suite result is not represented as a product failure or a pass. Per-file isolation
shows every Shepherd test file except `controller.test.ts` and `state-store.test.ts` passes; those
two use the default Darwin process-identity probe in `state-store.ts`. Parent orchestration must
rerun the complete suite where `/bin/ps` child execution is permitted, push this local chain, and
then obtain fresh exact-head review.

Before RED, local #471 contract evidence and the #479 consumer audit added one acceptance seam:
bounded concurrent mutator leases for canonically disjoint isolated authorities, with overlap and
capacity denial plus per-lease cleanup. This remains inside issue #475's runtime/test scope; no
scheduler or parent artifact is changed.

Cycle 7 terminal verification is complete at implementation head
`5c638d7f21a3910f40e499dba5c82cb7646642ac`. The frozen candidate was
`a3cd85a5d0871dd1c4c99dd8b30bcd609a228c45`, the immutable base remains
`e659d6f1b666f58748e2d8c86599ceb4bbc62ff8`, and the combined 11-finding stable-head campaign is
<https://github.com/polymetrics-ai/cli/pull/486#issuecomment-5037079867>. PLAN `f40a08f1` preceded
the one test-only RED `3b7e886a`; all findings pass after GREEN `5c638d7f`.

The lifecycle verification accounts for throwing signal attachment/removal and abandoned
creation across late resolve, reject, hang, and malformed fulfillment. A successful close cannot
precede owned late work; an uncancellable create must instead reject/quarantine close within a
bound. Each row records timers, reservations, cleanup hooks, close outcome, and unhandled
rejections. The redaction verification spans multiline/indented/YAML continuation,
numeric/Authorization/alias/PKCS#8 forms, unmatched quote recovery, safe multiline quote
preservation, every direct/prompt/workspace/tool/handoff consumer, and deterministic total scanner
work for padded 25/50/100 KiB flows.

Deterministic padded-flow diagnostics report 76,465 / 152,774 / 305,505 total visits for 25,645 /
51,235 / 102,453-byte inputs, including 8,533 / 17,066 / 34,133 key-start visits. The ratios remain
near-linear without timing assertions.

PR #486 correction Cycle 6 revalidation is complete. The independent-review findings against
`d918617a19749cd16d6bfcf3d2fee3e5146e7380` are covered by PLAN checkpoint `4f9c5a96`, committed
test-only RED checkpoint `e8422d53`, and pass at implementation head
`93314a54302e84e053ad0d6ff44371fbf1a167e0`.

## Declared Phase Equivalent

The user/parent explicitly replaced generic repository-wide verification for this sub-worker with
the issue-focused and complete Shepherd TypeScript gates below. Therefore `verificationPassed`
means every declared command here passed; it does not claim parent-level Go/connector verification.

| Gate | Status | Evidence |
|---|---|---|
| Focused AgentSession/tool-policy tests | pass | 53 passed, 0 failed, exit 0 |
| Complete `.pi/extensions/shepherd/*.test.ts` suite | pass | 190 passed, 0 failed, exit 0 |
| Deterministic scanner scale | pass | total/key-start visits are bounded and near-linear for padded 25/50/100 KiB inputs |
| Strict no-emit TypeScript against installed Pi 0.80.6 types | pass | owned production/tests plus all Shepherd production `.ts`; exit 0 |
| Supported offline Pi extension/RPC smoke | pass | explicit 0.80.6 binary, `PI_OFFLINE=1`, RPC `get_commands`; `pm-shepherd` registered, exit 0 |
| `git diff --check` | pass | exit 0 |
| Owned-scope diff check | pass | every base diff path matches issue #475 production/test/phase scope |

## Explicit Boundary

Not run in this lane by policy: `go vet ./...`, `go test ./...`, `go build ./cmd/pm`, `make verify`,
connector certification, and all other repository-wide Go/connector gates. These remain parent
integration and GitHub CI responsibilities. This is an intentional verification partition, not a
skip, pass, or failure.

## Model / Dependency / Safety Baseline

- Installed Pi: `0.80.6`.
- Required route: `openai-codex/gpt-5.6-sol`; `high` implementation/correction, `xhigh` all other
  roles.
- New dependencies: none permitted or planned.
- Live auth/model/network/credential checks: not run and not needed.

## Commands And Results

```bash
node --test .pi/extensions/shepherd/agent-session-runtime.test.ts \
  .pi/extensions/shepherd/tool-policy.test.ts
# 53 passed, 0 failed

node --test .pi/extensions/shepherd/*.test.ts
# 190 passed, 0 failed
```

TypeScript used the already-installed TypeScript 5.9.3 compiler and explicit Pi 0.80.6 package/type
roots with:

```text
--noEmit --strict --target ES2024 --module NodeNext --moduleResolution NodeNext
--allowImportingTsExtensions --skipLibCheck
```

It passed both the owned production/test input list and every non-test
`.pi/extensions/shepherd/*.ts` production file. `--skipLibCheck` suppresses unrelated declaration
issues inside the globally installed SDK dependency graph; all Shepherd inputs remain under
`--strict`, and the new runtime imports Pi 0.80.6 `AgentSessionEvent` and
`CreateAgentSessionOptions` directly.

The supported smoke used the explicit pinned binary:

```bash
printf '%s\n' '{"type":"get_commands"}' |
  PI_OFFLINE=1 /Users/karthiksivadas/.nvm/versions/node/v24.13.1/bin/pi \
    --mode rpc --no-session \
    --no-extensions --no-skills --no-prompt-templates --no-context-files \
    --extension .pi/extensions/shepherd/index.ts
```

The explicit binary reports `0.80.6`; RPC `get_commands` returned success with the `pm-shepherd`
command registered.

Cycle 7 proves that signal-listener attach/remove exceptions cannot strand a reservation or
referenced deadline timer. Creation ownership remains visible to close through late fulfillment,
validation, and cleanup: late resolve and reject settle before successful close, while hung and
malformed creation boundedly quarantine and reject close without an unhandled rejection.

The structured redactor covers multiline outer flows, indented/key-only/continued YAML, numeric
secrets, credential-bearing Basic and non-Bearer Authorization values, unmatched-quote recovery,
repository secret aliases, and generic PKCS#8 at direct, prompt, workspace, typed-tool, and handoff
boundaries. Harmless structurally quoted multiline prose remains byte-identical.

## Retained Cycle 6 Baseline

Cycle 6 proves that a nested sensitive mapping value retains its value-local delimiter ownership
across a newline, so its close cannot remove the outer flow-map closer and hide a later same-line
`client_secret`. It also proves that `rock-'n-roll` remains one unquoted scalar: `-` is a quote
boundary only when it is a line-local YAML sequence marker. Direct probes, serialized role prompts,
`workspace_read`, typed capability output, and handoff summary/finding fields remove every synthetic
marker and retain `[REDACTED]`; the harmless apostrophe control remains byte-identical.

Assignment decisions reuse the scanner-owned current line end instead of scanning each remaining
suffix. The typed metrics sink reports exactly 25,618 / 51,218 / 102,418 line-boundary visits for
inputs of the same sizes, establishing linear line discovery without wall-clock assertions. The
overloaded entry point also preserves `Array.map(redactSensitiveText)` callback behavior by
ignoring its numeric index. Every earlier lifecycle and redaction regression remains green.

The Cycle 7 immutable-base check retained
`e659d6f1b666f58748e2d8c86599ceb4bbc62ff8`; every changed path is an issue-owned Shepherd
production/test file or one of the phase's durable planning artifacts. Implementation HEAD and the
local remote-tracking branch were identical after the GREEN push.

## Deviations

- `manual_gsd_fallback`: the healthy repo adapter has no `programming-loop` registry entry.
- Ambient PATH drift: default `pi` resolved to 0.80.6 initially and 0.80.10 at final verification.
  All authoritative pinned checks use the unchanged explicit 0.80.6 binary/package/types.
- No Shepherd prompt asset file was needed beyond the owned typed `role-prompts.ts` builder; it
  generates the trusted role prompt and schema envelope without widening the allowed scope.

## Cycle 12 Exact-Head Verification

Cycle 12 closes all ten accepted Pi lifecycle and boundary findings at implementation checkpoint
`3dc4de7114d5ee501fdc4ecfb4364244a58a3ab9`, descended from frozen start
`7882cd70c25971e889ec04f63b98c936d605003e` and immutable base
`e659d6f1b666f58748e2d8c86599ceb4bbc62ff8`.

| Gate | Status | Evidence |
|---|---|---|
| RED integrity | pass | 124 executed; all 114 retained tests passed; exactly 10 intended C12 behavior rows failed; strict focused TS passed; production blobs stayed frozen |
| Focused AgentSession/tool-policy | pass | 124 passed, 0 failed, 0 skipped/cancelled/todo |
| Actual pinned Pi sessions | pass | real no-tool and one-tool `createAgentSession` runs; 2 prompts, 3 offline provider turns, 1 scoped tool callback, complete settled traces, 0 fetches, 2 post-settlement disposals |
| Focused strict TypeScript | pass | TypeScript 5.9.3, explicit Pi 0.80.6 package/type roots, exit 0 |
| All-production strict TypeScript | pass | all 12 non-test Shepherd `.ts` files, same compiler/type roots, exit 0 |
| Explicit Pi 0.80.6 offline RPC | pass | `get_commands` returned success and registered `pm-shepherd`; only sandbox settings-lock warnings, exit 0 |
| Safe complete-suite isolation | pass | all test files except the process-identity-dependent controller/state-store pair: 187 passed, 0 failed |
| Complete serialized Shepherd suite | environment-blocked | 261 executed: 230 passed; all 31 failures are the unchanged controller/state-store `spawn EPERM` process-creation family |
| Repository integrity | pass | diff check, base ancestry, JSON, credential patterns, exact issue-owned path scope, no dependency/Go/connector paths |

The focused command was:

```bash
node --test .pi/extensions/shepherd/agent-session-runtime.test.ts \
  .pi/extensions/shepherd/tool-policy.test.ts
```

Both strict TypeScript scopes used the already-installed TypeScript 5.9.3 compiler with:

```text
--noEmit --strict --target ES2024 --module NodeNext --moduleResolution NodeNext
--allowImportingTsExtensions --skipLibCheck
--baseUrl <Node 24.13.1 global module root> --typeRoots <pinned Pi 0.80.6 @types root>
```

The authoritative offline smoke used the explicit pinned binary and no session, extensions,
skills, prompt templates, or context files beyond the issue-owned extension:

```bash
printf '{"id":"commands","type":"get_commands"}\n' |
  PI_OFFLINE=1 /Users/karthiksivadas/.nvm/versions/node/v24.13.1/bin/pi \
    --mode rpc --no-session --approve --no-extensions --no-skills \
    --no-prompt-templates --no-context-files -e .pi/extensions/shepherd/index.ts
```

The actual-session focused row additionally installs a temporary throwing `fetch` sentinel around
the two in-memory provider runs and asserts zero calls. It does not use a live credential, model,
service, or network. No push, GitHub, Go, connector, `make`, runtime-service, or parent integration
gate was attempted. The complete-suite classification is deliberately not called green; parent
orchestration owns its rerun where child process creation is permitted.
