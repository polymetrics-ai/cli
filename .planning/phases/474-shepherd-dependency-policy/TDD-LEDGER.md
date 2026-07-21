# Issue #474 TDD Ledger

GSD mode: `manual_gsd_fallback` because the repository adapter does not register
`programming-loop`.

| Slice | RED evidence | GREEN evidence | Refactor/broader evidence | Status |
|---|---|---|---|---|
| Lifecycle and retry policy | `node --test ...` failed because `autonomy-policy.ts` did not exist | focused suite 23/23 pass | strict production typecheck pass | green |
| DAG/scopes/maximum ready queue | same run failed because `dependency-graph.ts` did not exist | focused suite 23/23 pass | strict production typecheck pass | green |
| Pure idempotent reconciler | same run failed because `reconciler.ts`/its imports did not exist | focused suite 23/23 pass | strict production typecheck pass | green |

Rules:

- Production code is not written until the matching focused tests fail for the expected missing
  module/export behavior.
- Tests are not weakened to fit implementation.
- Every command, exit result, and exact failure/pass count is recorded after it runs.

## RED checkpoint

Command:

```bash
node --test .pi/extensions/shepherd/autonomy-policy.test.ts \
  .pi/extensions/shepherd/dependency-graph.test.ts \
  .pi/extensions/shepherd/reconciler.test.ts
```

Observed: 3 file-level tests, 0 pass, 3 fail. Each failed with `ERR_MODULE_NOT_FOUND` for the
intentionally absent production modules. The surrounding evidence wrapper then also reported
`zsh: read-only variable: status`; this shell-wrapper mistake did not cause or conceal the three
expected Node failures and will not be reused.

## GREEN checkpoint

After the minimum three pure modules were added, the first focused run exposed one Node strip-mode
syntax incompatibility (a TypeScript constructor parameter property): 6 tests passed and 2 test
files failed before loading. Replacing that syntax with an explicit readonly field produced:

```text
tests 23
pass 23
fail 0
```

The existing TypeScript 5.9.3 compiler available in the environment then ran with `--noEmit
--strict --target ES2022 --module NodeNext --moduleResolution NodeNext
--allowImportingTsExtensions --skipLibCheck` over the three production modules. It first found one
implicit-any callback introduced by runtime `Array.isArray` narrowing; after the minimal annotation,
the same strict command exited 0. The focused 23/23 suite and `git diff --check` also exited 0.

## Refactor gap loop

An adversarial pass added four fail-closed expectations before editing production code:

- runtime-invalid lifecycle/failure vocabulary;
- empty DAG completion;
- terminal completion independent of capabilities used only for future spawns; and
- invalid concurrency data returning a typed repository blocker instead of throwing.

Gap RED: 26 tests, 22 pass, 4 fail with the expected uncaught/incorrect decisions. Gap GREEN after
the minimal validation and decision-order changes: 26/26 pass. Strict production TypeScript and
`git diff --check` pass. The compact full Shepherd suite then passed 163/163.

## Exact-head correction loop

Status: RED captured against reviewed `28f165412de4c8165ba7717a1690c36b00af8857`; production remained
untouched through this checkpoint.

| Correction slice | Expected RED signal | Status |
|---|---|---|
| Authenticated resumable human decision and terminal abort | missing stage/result/guards | red |
| Bounded conflict-component scheduling | 64-cycle subprocess killed by `SIGTERM` at 1,004 ms | red |
| Code-unit ordering and Darwin/Git aliases | locale/case/NFD expectations differ | red |
| Dependency/status coherence and exact DTO validation | incoherent/hostile values accepted; `null` throws | red |
| Reconciliation precedence and selected-only isolation | capability blocker masks dependency; reader suppressed | red |
| Correction-required and ordinary evidence handling | unsafe advancement and human-gate misclassification | red |

Correction RED command: the focused three-file Node test command. Observed: **36 tests, 21 pass,
15 fail**, including all six correction slices. The isolated 64-item cycle reproduced the scheduler
DoS safely: its subprocess exceeded the one-second deadline and was terminated while the main test
runner completed in 1.09 seconds. No production file changed before this evidence was captured.

Correction GREEN: the minimum lifecycle, graph, and reconciliation changes produced **36/36 pass**.
The hostile 64-item cycle is now rejected with typed `conflict_component_too_large` in 80 ms in the
final focused run. The first production-only strict TypeScript pass found three `unknown` narrowing
errors at the graph DTO boundary; explicit validated locals fixed them and the exact strict command
then exited 0.

Audit gap RED added composed Unicode case mappings (`Straße` / `STRASSE`) and failed work with an
unsatisfied dependency. Targeted result: 2 tests, 0 pass, 2 fail. Minimal NFKC + upper/lower aliases
and terminal-status coherence extended the final focused GREEN to 36/36. `git diff --check` passes.

Correction refactor precomputes one canonical collision index per ready queue, then reuses it for
component discovery and bounded exact selection. This removes repeated scope normalization and
pairwise collision evaluation from recursive search without changing decisions. Focused tests
remain 36/36 and strict production TypeScript remains clean.

Final correction verification: focused 36/36, full Shepherd 173/173, strict production TypeScript
pass, offline Pi 0.80.6 RPC pass, and diff/ownership pass. The implementation head verified before
this evidence-only checkpoint was `ef2fd1e280128ccb2a0e46b749f9638472fad865`.

## Exact-head correction loop 2

Status: RED captured against reviewed `82ec59c0b3161639893ff2bce7a40dbafe7745df`;
production remained untouched.

| Slice | Expected reviewed-head failure | Status |
|---|---|---|
| Conservative sharp-S aliases | `ẞ` does not collide with `ß`/`ss` | red |
| Lifecycle/queue cross-validation | asserted facts bypass pending or dependency-blocked work | red |
| Caller-owned Proxy snapshot | hostile iterator mutates pending work into false completion | red |
| Terminal BLOCKED | returns ordinary `await_stage_evidence` | red |
| Primary run-state truth | primary fields remain at historical 26/163 | pending |

Focused RED command: 40 tests, 35 pass, **5 expected failures**, exit 1. The failures separately
prove sharp-S alias incompleteness, pending-work completion bypass, dependency-blocked ready-work
bypass, caller iterator mutation, and non-terminal BLOCKED reconciliation.

Minimum GREEN: 40/40 pass. The implementation uses a bounded NFKC + ECMAScript case-mapping alias
closure, clone/validate/freeze DTO isolation, one authoritative SCHEDULE selection before lifecycle
advancement, dependency/collision precedence, and an explicit terminal BLOCKED result. The first
strict production TypeScript run exposed only an overly broad helper return type; narrowing
`noSpawn`/`invalidGraphDecision` to their discriminated union members produced a clean strict pass
without changing runtime behavior.

Refactor renamed canonical-looking helpers to explicit alias-set/overlap vocabulary and extracted
the SCHEDULE-stage agreement predicate. This removes any implication of complete case folding and
makes lifecycle cross-validation auditable without changing behavior: focused 40/40 and strict
production TypeScript remain green.

Final review-2 verification: focused 40/40, full Shepherd 177/177, strict production TypeScript
pass, offline Pi RPC pass, and diff/ownership pass. Verified implementation head before the final
evidence-only commit: `55a8f8a5482311e9aa7a38a2bd2382ba4d9393b7`.

## Exact-head correction loop 3

Status: planned against reviewed `f461f9c811cf9d1d0e6804a82dd1201aab41f0a6`; production remains
untouched. The next test adds accessor-bearing root/nested/array DTOs whose getters count or vary
their output. Expected RED: reconciliation executes caller getters and/or returns non-typed,
non-idempotent decisions. GREEN requires zero getter calls and the same typed `invalid_snapshot`
result on repeated reconciliation.
