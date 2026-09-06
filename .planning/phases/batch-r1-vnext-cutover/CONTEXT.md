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

## 2026-09-04 — CP09 strict source parsing and canonical graph

- Firstmate instruction `120.msg` closes A1-04 with an independent exact-SHA PASS, authorizes the already-completed normal parent update, and starts CP09. The parent now heads at `988dd16c3d206a28d3e7b16f8a0d805c4163f7ca`; no reset or rebase to `main` is permitted.
- CP09 is the strict-parser half of #4426/N2. It owns the one source-lock decoder and typed canonical graph only. CP10 owns the later source-to-execution/resolver/loader/preflight semantic admission; B1 owns transactional publication.
- Preserve the sole authoring-to-runtime route: schema-4 `source.lock.json` → canonical graph → deterministic execution JSON → existing runtime. The parser must reject trailing roots, duplicate members, unknown execution fields, wrong schema roles, invalid encoders, alias collisions, and missing structural bindings before a renderer can replace any file.
- Valid inputs retain every source identity and yield equal canonical/rendered bytes when source ordering is irrelevant. Do not fetch or re-pin provider sources, add an executor, construct runtime execution, publish rendered files, invent provider facts, or add a second parser.
- Foundation Atlas disposition is a constrained extension of `authoring.source-lock-vnext.v1`; CP09 may update its existing owner/guarantee/constraint/proof declarations only as required by the extended canonical graph contract. No shared runtime foundation is proposed.
- Inline/manual GSD fallback remains required: the known adapter doctor blocker is the missing `.gsd/prompts/issue-122-rebootstrap.md`, while lifecycle command sources resolve. `go-engineering`, `tdd`, and `connector-lane-build-order` are available and loaded; the required named `golang-how-to` skill remains unavailable and is explicitly not claimed.

## 2026-09-04 — CP10 semantic source-execution admission

- Firstmate instruction `122.msg` accepts CP09 at parent head `85c28e70e4c8f811ea342a1f1054e09759cde1c1` and authorizes immediate serial CP10 in #4426/N2. No reset or rebase to `main` is permitted; the coherent independent review remains later.
- CP10 extends the existing authoring-only `authoring.source-lock-vnext.v1` seam. It may build only an in-memory staged generation from the CP09 graph: deterministic execution bytes, loader identity, manifest/index input, source-to-execution provenance, and supplied synchronization/Atlas admission facts. CP11 owns all physical staging and publication.
- Exact runtime authority is mandatory: use the existing loader, engine connector, commandrunner preflight, native selection inventory, manifest index, and syncplan resolver. Do not copy encoder, route, executor, or resolver rules; do not fabricate a source/destination sync pairing where a source lock declares only one endpoint.
- Provider facts remain raw authoring facts. Source operation IDs and exact JSON paths identify every failed semantic join; shared physical HTTP/GraphQL routes never establish equivalent operation identity.
- No connector source lock, rendered execution file, generated global index, provider source, credential, runtime network call, publication/recovery code, `.cache`, or certification residue is in scope.
- Inline/manual GSD remains required. `scripts/gsd sources` resolved every lifecycle command and the generated discussion prompt is available; the known ROADMAP/doctor blocker and unavailable `golang-how-to` skill are recorded rather than bypassed or claimed.

## 2026-09-04 — N2 boundary-review CP09/CP10 correction

- Firstmate instruction `124.msg` accepts the independent exact-SHA N2 BLOCK for published parent `56ec3d9d7dc1d726203b0ef0c03ddec3209b8dde`. Resume only the three named admission corrections; CP11 remains prohibited pending a new Firstmate-managed exact-SHA review and explicit authorization.
- The existing `authoring.source-lock-vnext.v1` seam is a constrained extension. A request reference must join its actual loaded write/direct-operation request schema; a response reference must join its actual loaded stream response schema. A source role with no runtime schema consumer is rejected rather than treated as provenance-only.
- Manifest staging must use the closed production selection authorities: native executor inventory plus `hookset.Factories`. GitHub must stage the production `hook/github.v1` extension and construct its in-memory engine connector with that selected hook. No global manifest scan, connector-name inference, provider call, or credential path is permitted.
- Canonical source-ID order governs staged provenance identity. The original source-array position remains only for errors so a failed input still names the author-visible pointer; reordering semantically equivalent operations must not change stage provenance, manifest, index, rendered bytes, or digest.
- No source lock, rendered execution bundle, generated global output, provider/credential I/O, certification path, filesystem staging, activation, `.cache`, or certification residue is in scope. Publish one normal non-force correction only after real RED/GREEN tests, scoped verification, and inline/manual self-review.
