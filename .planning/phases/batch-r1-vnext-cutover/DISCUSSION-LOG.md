# Batch R1 vNext cutover discussion log

## 2026-09-01 restart decision record

The captain has already made every product and architecture decision needed for this continuation:

- Finish legacy cleanup before migrating any additional connector.
- Use only `source.lock.json -> canonical descriptor -> deterministic execution JSON -> existing PM-CLI runtime`.
- Delete alternate execution/admission paths; do not preserve them as fixtures, compatibility adapters, feature flags, importers, certification, or retention gates.
- Do not introduce a shared runtime foundation, recover excluded local work, create credentials, use provider I/O, merge, force push, rebase, or open another PR.
- Migrate the named remaining connectors in fixed order and push every independently green cohort normally to the established Batch R1 remote branch.

No interactive product question remains. The only operational constraint discovered at restart is the missing GSD ROADMAP/prompt prerequisite described in `CONTEXT.md`; it is handled through the documented inline/manual-GSD evidence fallback, not by changing scope or weakening TDD/review requirements.

## 2026-09-04 A1 entry-capacity review disposition

Firstmate's exact-SHA review supplies the closed product decision: correct only `BundleStore` entry-capacity accounting, add the named barrier regression, preserve every listed boundary, and obtain one new independent review before any parent publication. `--auto` is safe because no product, architecture, scope, or testing choice remains open.

Ranked diagnosis record:

1. `reserveLocked` compares only `len(cache)` to `Limits.Entries`, so a second distinct key starts while the first key is only byte-reserved. Prediction: `Entries: 1, Bytes: 2` permits two barrier-held loaders before either completion.
2. A flight's entry reservation must outlive cancellation until the loader exits. Prediction: if cancellation releases it early, the second identity can begin while the first loader is still live; if it never releases, the retry remains capacity-blocked after completion.
3. Same-key joining must remain unchanged. Prediction: the existing concurrent same-identity test still invokes one loader after count reservation is added.

The regression will use the existing package-private store state only to assert the bounded resource invariant; its acceptance behavior is distinct-identity admission/rejection/retry. No public debug API or capacity knob is added.

## 2026-09-04 CP09 strict source parsing decision record

Firstmate instruction `120.msg` leaves no product choice open: after recording and normally publishing the A1-04 exact-SHA PASS, begin CP09 as the narrow first #4426 checkpoint. The existing parent head is `988dd16c3d206a28d3e7b16f8a0d805c4163f7ca`; `main` remains only the eventual PR merge base.

Resolved scope:

1. Reuse and extend the sole `decodeVNextSourceLock` → canonical-descriptor route. A second parser, runtime reader, provider call, re-pin, generated-file publication, executor construction, or connector migration is out of scope.
2. Promote raw execution fragments into a typed canonical graph sufficiently to reject unknown execution fields, role mismatches, invalid encoders, alias collisions, and missing structural bindings before render. CP10 alone will join that valid graph to the S1A resolver and runtime preflight.
3. Canonicalize irrelevant object/list ordering without discarding source identities; a semantically equivalent valid lock must render the same closed byte set.
4. The Atlas extension stays authoring-only. Its owner, guarantees, constraints, and proof list change in the same commit only if this graph contract changes.

Ranked hypotheses to test before production edits:

1. Existing `decodeStrictJSON` already handles one root and duplicate JSON tokens but `json.RawMessage` lets unknown nested execution fields pass. A no-write sentinel should currently survive only because the malformed fragment is accepted or fails later.
2. `canonicalizeVNextSourceLock` verifies referenced schema existence but does not verify schema role, encoder vocabulary, command aliases, or all required structural cross-references. A compact in-memory lock should expose each admission gap without a provider or runtime executor.
3. Existing map serialization sorts object keys, but operation/command input order can still influence output. Equivalent deliberately reordered locks should reveal whether canonical ordering is missing while preserving IDs.

## 2026-09-04 CP10 semantic source-execution admission decision record

Firstmate instruction `122.msg` closes the CP09 gate and leaves no product choice open. CP10 is the second #4426/N2 half: admit one canonical graph into an in-memory staged generation, then publish one normal parent checkpoint only after its focused RED/GREEN proof and inline/manual review.

Resolved scope:

1. The CP09 graph remains the sole source-lock representation. Semantic admission consumes its retained immutable source operation IDs and raw provider facts; it neither reparses a lock nor normalizes away unknown provider facts.
2. Runtime authority stays in existing packages. The staging layer must call the real loader and commandrunner preflight, construct only a no-I/O engine connector, and use the sealed native selection/index and syncplan resolver APIs rather than reproduce their validation.
3. A command's source operation binding is exact. Matching REST/GraphQL paths is supplementary evidence only; two GraphQL operations sharing `/graphql` remain different nodes and a cross-operation command target must fail at its source field.
4. An incomplete single-connector sync declaration is not a saved plan. The staged admission API accepts complete externally supplied source/destination/Atlas facts for real resolver validation, but does not invent a target, executor, foundation, or digest.
5. CP10 produces no filesystem stage or global output. CP11 owns filesystem durability/activation; connector migrations and schema-4 source authoring remain out of scope.

Ranked hypotheses to test before production edits:

1. The CP09 loader check can accept a command bound to a different existing GraphQL operation because both have the same valid route. A two-operation `/graphql` fixture should expose that identity gap without provider I/O.
2. Rendering can produce correct runtime JSON while losing the immutable source-to-runtime correspondence needed by CP11. A staged provenance/index comparison should detect a missing, duplicate, cross-connector, or stale identity before any write.
3. A rate declaration can be structurally rendered but omitted from identity/index data. Adding, changing, and removing only `rate_limits.json` should alter and restore the closed staged identity through the real loader, not a copied rate validator.
4. A source-only transport declaration lacks the destination/foundation facts required by `syncplan.Resolve`. Synthesizing those facts would be a false executable claim; explicit supplied facts must be the only resolver input.

## 2026-09-04 N2 boundary-review correction decision record

Firstmate instruction `124.msg` accepts the independent N2 BLOCK as a closed three-defect correction set. No product or architecture choice remains; CP11 remains blocked.

Resolved scope:

1. Schema references are runtime claims, not inventory labels. Request and response references must resolve to the exact schema consumed by a loaded write, direct operation, or stream. A direct form with no typed runtime counterpart cannot retain that role as a provenance-only claim.
2. The in-memory stage must match production selection. Reuse the closed generated hook factory and native selection authorities; GitHub's `hook/github.v1` is behavior-bearing and must be retained in both staged entry and in-memory engine construction.
3. Canonical staged provenance represents the semantic descriptor, so it uses canonical source-ID order. Original operation positions remain error-only diagnostics and must not influence staged identity.
4. Tests are hermetic local authoring tests: source-lock temp roots/no-write sentinels, in-memory loader/connector/preflight, and the generated production manifest entry. No provider, credential, global-output, or CP11 staging action is allowed.

Ranked hypotheses to test before production edits:

1. Replacing a request or response reference with a second valid registry schema leaves current admission green because it checks existence rather than the loaded consumer. Prediction: a sentinel `lock-render` test reaches the existing write/stream output unless semantic equivalence is added.
2. GitHub's staged entry lacks its required extension because admission selects only native executors. Prediction: the staged entry differs from `manifestindex.GeneratedEntries()` at `Extension`, and supplying the real entry fails the current binding check.
3. Reordering operations preserves bytes/digest but changes provenance because paths retain authored source indexes. Prediction: the existing CP09 reorder fixture has equal rendered bytes and different `Staged.Provenance`; using a canonical index only for provenance removes that difference without changing diagnostic paths.

## 2026-09-04 — CP11 B1 discussion resolution

Firstmate instruction `125.msg` fixes the product and architectural choices:
CP11 is a connector-local authoring publication foundation, not a runtime
materialization checkpoint. The admitted CP10 stage is the sole input; no new
source-lock reader, renderer, executor selector, provider request, credential
path, database path, connector-specific branch, or distributed transaction is
allowed.

Resolved decisions:

1. A generation is a closed set: admitted execution bytes plus manifest,
   provenance, Atlas reference, compact index, proof, integrity, and a private
   lease. Optional execution files are set members, never inherited from an old
   generation.
2. Stage directly below one connector's `generations/` parent, validate those
   physical bytes through the existing loader/selection/preflight, fsync, then
   select only through an atomic typed `CURRENT`. No flat-artifact reader or
   fallback is allowed.
3. A durable `{old,new,state}` journal recovers only a complete old or new
   pointer. A reader holds a generation lease from pointer observation through
   handle release; pruning requires both the lease and valid publisher-owned
   integrity, preserving held and unowned directories.
4. `--check` is read-only: it uses an existing shared lock, rejects pending
   journal/stage state, and never recovers, prunes, creates, or rewrites.
5. Existing source locks and flat execution artifacts are a deterministic
   in-memory reference corpus only in B1. Tests use temporary roots; no real
   connector/source-lock materialization or author-owned-file deletion occurs.

No user decision remains. The inline manual-GSD fallback is necessary because
the custom phase has no adapter roadmap and no compatible isolated worker; it
does not waive RED/GREEN, scoped verification, read-only review, or the next
Firstmate exact-SHA gate.

## 2026-09-05 — CP11 Astra B-01/B-02 correction decision record

The captain's recovery authorization adopts the two Astra blockers as one bounded
CP11 correction wave. No product, connector, runtime, or shared-foundation
decision remains.

1. Reuse the existing descriptor-relative no-replace primitive for both
   quarantine restoration and final-generation activation. A second pathname
   absence check is not atomic and is not a permissible substitute.
2. Resume only a strict valid pre-marker base preparation: no predecessor, no
   phase, equal prior/intended logical state, validated private transaction and
   anchors, under the existing exclusive publisher lock. Finish its durable
   terminal phase, then complete missing base heads and marker creation.
3. `lock-render --check` remains read-only and refuses the pending markerless
   state. Malformed/missing prepared records and non-base graphs remain
   fail-closed; no manual unlock, cleanup, or legacy recovery route is added.
4. The first artifacts are deterministic RED witnesses on temporary roots using
   the real production descriptor paths. Green requires inode/byte retention
   and ordinary fresh recovery/retry, not an error code alone.

## 2026-09-05 — CP11-R3-01 successor discussion record

- Independent exact-range publication and bootstrap reviews found the same FIFO-before-type-check liveness defect. A separate Astra merged-ledger audit affirmed it and added the semantic-admission `vNextPublicationDirectoryFS.Open` bypass as a sibling of the same descriptor-reader cause.
- The accepted boundary is local and authoring-only: a malformed local nonregular member must refuse promptly without changing identity/bytes or holding the connector lock. This is neither a provider capability nor a generic runtime/filesystem foundation decision.
- The repair uses TDD with a subprocess RED control because the pre-fix behavior can hang. GREEN must preserve descriptor-relative no-follow semantics and validate the exact opened descriptor. A pathname precheck, broad retry, manual cleanup surface, or new runtime reader is rejected.
- B-01 remains the final no-clobber correction and B-02 remains the strict interrupted-bootstrap correction. Both are retained for final fresh exact-SHA review; L-01 cache cleanup remains optional and out of scope. CP12 remains prohibited.

## 2026-09-06 — CP11 F-01–F-08 coordinated repair discussion

The Firstmate-authored `CP11-F01-F08-coordinated-repair.md` is the complete
input to this discussion.  It closes all product and architecture questions:
this is one authoring-only publication repair, retaining existing descriptor
relative confinement, no-follow/type validation, authority history, no-replace
public restore, strict bootstrap, directory lifetime locking, independent lease
identity, and lock-render-only signal scope.  No provider/runtime/credential/
database action, source-lock materialization, new filesystem abstraction,
generic writer, new flag, receiver, or cancellation model is in scope.

The required `scripts/gsd` adapter was checked with `scripts/gsd doctor` and
all five lifecycle sources/prompts resolved.  Doctor retains the established
missing `.gsd/prompts/issue-122-rebootstrap.md` issue prompt, while
`agentcontractgen check` passed.  The phase has no ROADMAP-backed compatible Pi
worker and the Firstmate single-writer contract forbids role spawning, so this
is the adapter-documented inline/manual fallback: generated `discuss-phase`
and `plan-phase --tdd` prompts were read and their steps are recorded here and
in PLAN/TDD.  This records a runtime limitation, not a waiver of TDD,
verification, audit, or independent exact-SHA review.

The discussion ordering is fixed by the completed audit: first make the
snapshot/process witnesses safe and bounded (F-04/F-08), then obtain real
production REDs and repair the one capture/resource/error group (F-01/F-02/
F-03), then strengthen reader/durable-cut proof (F-05/F-06), then correct F-07
provenance canonically and verify the complete wave.  Existing correct reader,
lease and signal behavior is not intentionally broken merely to manufacture a
RED; oracle mutation controls are labeled as controls.  All source changes stay
frozen until the plan/TDD ledger below exists.
