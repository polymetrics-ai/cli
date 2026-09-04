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
