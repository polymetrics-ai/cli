# Batch R1 vNext cutover context

## Fixed decisions

- The sole executable pipeline is `schema-4 source.lock.json -> canonical descriptor -> deterministic execution JSON -> existing PM-CLI runtime`.
- Source locks are immutable authoring/provenance input. They are not embedded, loaded, admitted, or consulted by runtime dispatch.
- Remove rather than bridge importer, certification, retention, fixture, compatibility, feature-flag, and second-executor execution/admission paths.
- Preserve bounded provider encoding, credential boundaries, approval gates, and provider-I/O safeguards.
- Do not recover or inspect the ruled-out `/private/tmp/cli-batch1-vnext-legacy-cutover-r1` worktree.
- No new shared runtime foundation, credential persistence, live GitHub proof, provider I/O, merge, force push, rebase, or new PR.

## Ordered delivery

1. Remove native fixture-mode execution and stale execution/admission shims.
2. Remove stale explanatory/generated material and classify the exact residual scan.
3. Regenerate CLI docs, connector manuals, skills, and website-derived outputs.
4. Verify the cleanup, review the exact SHA, commit, and normally push it to the existing Batch R1 ref.
5. Migrate Bitbucket, CircleCI, Docker Hub, Jira, Notion, Sentry, Stripe, and Vercel in that order. Each connector is an independent source-lock render, seven-lane declaration, credential-boundary proof, review, green commit, and normal push.

## Known GSD limitation and manual fallback

The repo-local GSD adapter resolves all lifecycle commands and its canonical contract projection is current. Its generated workflows require `.planning/ROADMAP.md`; this established custom phase has no ROADMAP, and `scripts/gsd doctor` also reports the unrelated absent `.gsd/prompts/issue-122-rebootstrap.md`. The continuation records the generated command path and executes the equivalent discussion, TDD planning, implementation, verification, and review procedures inline here. This preserves the lifecycle evidence without representing unavailable generated output as executed.

## Required review map

The cleanup review begins from the immutable pre-cleanup SHA. It maps source locks and renderer inputs, embedded execution-only files, engine/commandrunner dispatch, hook/native registrations, credential/approval boundaries, generated documentation, and the exact residual scan. It records all mandatory Firstmate review lenses in `REVIEW-CONVERGENCE.md` before production changes and repeats fresh-context review on the final code SHA.

## 2026-09-04 — A1 entry-capacity correction

- Firstmate instruction `118.msg` reports an independent exact-SHA BLOCK against `701a0b45175f308400c938322fd1634a28efdaef`: `BundleStore` charges bytes to an in-flight load but not its entry slot, so distinct simultaneous flights can exceed `Limits.Entries` when byte capacity remains.
- Foundation Atlas classification: **constrained_extension** of `definitions.bundle-loader.v1`; the existing owner is `internal/connectors/manifeststore/bundle_store.go` (`BundleStore.AcquireEntry`, `reserveLocked`, `load`, and `releaseWaiter`). This is not a new foundation or execution route. Its retained-cache/concurrent-flight capacity guarantee and proof inventory must be updated with the implementation.
- The one allowed behavior change is atomic count reservation per distinct flight. Under the store lock, cached entries plus pending entry reservations must never exceed `Limits.Entries`; a canceled flight retains its reservation until its loader terminally releases it, so a distinct retry cannot overlap it.
- The proof is a deterministic two-identity barrier test with `Limits{Entries: 1, Bytes: 2}`. It must cover an initial blocked flight, second-identity rejection, cancellation while the first loader remains blocked, post-completion retry, and retained-cache state without inspecting provider, credential, rate-declaration, connector, or CLI behavior.
- CP09 remains blocked. No source lock, rendered execution JSON, rate declaration, generation/digest identity contract, connector/factory/CLI behavior, provider request, credential, release artifact, `.cache`, or certification residue is in scope.
- Inline GSD fallback: `scripts/gsd doctor` detects the pre-existing absent `.gsd/prompts/issue-122-rebootstrap.md`, while every required lifecycle command resolves through `scripts/gsd sources`. The configured runtime has no `golang-how-to` or named task-specific Go skills; their absence is recorded rather than faked. The available `go-engineering`, `diagnose`, `tdd`, `gsd-ns-workflow`, and `connector-migration-exact-sha-review` skills were loaded, and the generated discussion/planning prompts are executed inline with `--auto` because Firstmate supplied the closed correction contract.
