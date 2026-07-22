# Cycle 12 Pi Lifecycle Mapper

## Assignment

Read-only map of the explicitly installed `@earendil-works/pi-coding-agent` 0.80.6 contract for
issue #475, frozen start `7882cd70c25971e889ec04f63b98c936d605003e`.

Determine, without edits, network, credentials, services, or model calls:

1. exact no-tool and one-tool `AgentSessionEvent` order, including user/tool-result messages,
   intermediate assistants, tool execution, `agent_end.messages`, retries, and `agent_settled`;
2. the authoritative terminal/freeze boundary;
3. a supported public seam that lets the entire actual `createAgentSession` result and actual
   session prompt offline through a custom inert stream/provider;
4. exact factory-result/extension fields and installed assistant diagnostic shapes; and
5. the smallest relevant source/type locations.

Conclude once those facts are established. Return concise evidence and implementation guidance;
do not edit or expand into unrelated source.

## Authority

- Read-only filesystem/source inspection only.
- No writes, commits, tests that mutate state, network, credentials, live auth/model calls,
  services, GitHub, Go/connectors, or `make`.
- The issue worker owns every artifact, test, production, commit, and verification action.

## Cycle 13 Seam-Map Appendix

Assignment: read-only mapping of the seven blockers in `/tmp/475-REVIEW-CYCLE12-1.md` and
`/tmp/475-REVIEW-CYCLE12-2.md` against frozen candidate `5dafc572`, including exact Pi 0.80.6 tool
identity/result seams. No file, test, network, credential, service, model, or external state was
modified.

1. Runtime request arrays flow from `normalizeRunRequest` into `captureFreshDenseArray`, whose
   `Reflect.ownKeys` whole-source materialization is the defect. ECMAScript has no bounded own-key
   iterator: the compatible bounded contract is intrinsic brand + one own-data length descriptor +
   exactly indexed own-data descriptors into a fresh frozen array, with every non-index peer inert.
2. The pre-create SDK sequence is settings manager, session manager, first agent-dir lookup,
   resource loader, reload, second agent-dir lookup, then creation scheduling. The second lookup
   lacks the critical active-scope barrier; every re-entrant callback should be followed by one.
3. `createToolPolicy`, `validatePolicyInput`, and `normalizeScopedPrefixes` repeatedly consume raw
   authority/capability arrays. Snapshot once at the exported boundary and thread only captured
   arrays through validation, scope normalization, declaration, and tool construction.
4. Capability classification misses reversed generic operations and semantic synonyms such as
   process/run, web/send, database/modify, and vault/view. Classify normalized token categories,
   retaining explicit safe host controls.
5. `sensitiveAssignmentKind` considers only the final dotted segment. Classify final and joined
   bounded segments so `api.key`, `private.key`, `database.url`, and `aws.secret.access.key` share
   every current redaction consumer.
6. `buildRolePrompts` and `validateAuthority` call caller iterators and instance methods for context,
   prefixes, and tools. Capture all four arrays before validation or rendering and freeze the
   returned prompt record.
7. `newTerminalCapture()` drops expected tool authority; assistant tool calls are not registered;
   execution end discards result identity; tool-result messages and `turn_end.toolResults` are only
   shape-checked. Add a one-to-one per-turn ledger from authorized assistant call through execution,
   result message, turn result, next turn, and final handoff.

## Cycle 14 Boundary-Map Appendix

Assignment: read-only mapping of the unique correction families in the two complete Cycle 13
reports against frozen candidate `67050a4a`, including post-create Pi callbacks, the complete
current/#479 host-capability inventory, and every structured-redaction consumer. No file, test,
network, credential, service, model, or external state was modified.

1. The post-create sequence is `created.session`; independent operation getters (`abort`,
   `waitForIdle`, `dispose`, `prompt`, `subscribe`, `getActiveToolNames`); result/extension shape;
   `model` plus provider/id; thinking level; session file; `getActiveToolNames()`; claim;
   subscription; and prompt. The mandatory cleanup root must be retained before optional
   validation can fail, and every later re-entrant callback needs a closure-aware barrier.
2. A synchronous `close()` sets runtime closing state before its asynchronous cancellation work,
   so the existing scope-only assertion does not prevent prompt. Reuse the combined runtime/scope
   barrier after ownership acquisition and after subscription returns. A terminating subscription
   must return first so its unsubscriber can be captured; then captured failure/closure is checked
   before prompt.
3. Current source has no production-created arbitrary host capability. The two existing positive
   contract identities are `host_inspect` (non-mutating) and `host_verify` (mutating);
   `host_publish` is only an undeclared negative fixture. #479 scheduler, Git/worktree, GitHub,
   review, decision, and integration ports remain controller-owned and outside AgentSession.
4. `sensitiveAssignmentKind` is private to assignment parsing. All external consumers converge on
   `redactSensitiveText`: role task/context, workspace read/edit/write, host output/reference,
   handoff summary/finding/verification, and policy/runtime public errors. Qualified controls such
   as `api.key.version`, `private.key.algorithm`, and `database.url.scheme` expose the fuzzy
   ancestor-matching defect without weakening their paired terminal-secret cases.

## Cycle 15 Boundary-Map Appendix

Assignment: read-only consolidation of every unique blocker in the two complete Cycle 14 reports
against frozen candidate `f41cde91`, with exact scanner and post-create Pi seams. No file, test,
network, credential, service, model, or external state was modified.

1. Quoted unknown keys already classify as `unknown-sensitive`, but `quotedValueRedaction` routes
   them through the authorization parser. A quoted multi-component value can therefore return no
   range. Unknown-sensitive quoted, scalar, and block values must always use whole-value redaction.
2. Assignment recognition repeats a short ASCII/64-character grammar. Unsupported, spaced,
   extended, overlong, or malformed candidates collapse to “not assignment”; flow lookahead then
   fails to open a frame and can expose later protected siblings. Replace this with a total
   `none | public-certain | protected-certain | opaque-uncertain` result and monotonic resume cursor.
3. `captureCreatedSession` batches nested validation behind one barrier. Three `for...in` loops in
   result, empty-array, and active-tool validation traverse Proxy prototypes before the trailing
   lifecycle check. Decode only approved direct prototypes and own data descriptors/canonical
   indexes, with every unavoidable Pi callback followed immediately by the execution barrier.
4. Cleanup ownership remains first: capture the own session descriptor and mandatory operations
   before optional DTO validation. Split creation-result decoding, extension-result decoding,
   active-tool invocation/array decoding, subscription/unsubscriber capture, and prompt into
   distinct callback/barrier or callback-free decode steps so cancellation stops later effects.
5. Retain 134 rows and add exactly three aggregate RED rows: whole quoted unknown redaction through
   every shared consumer; uncertain key/cutoff/flow recovery with public controls and bounded work;
   and prototype-safe post-create decoding plus immediate lifecycle barriers. Expected frozen-head
   result is 137 executed, 134 retained passes, and exactly three intended failures.

## Cycle 16 Architecture-Map Appendix

Assignment: read-only consolidation of both complete Cycle 15 reports against frozen candidate
`df930d62`, with the exact scanner, untrusted DTO, prompt-settlement, and focused-test seams. No
file, test, network, credential, service, model, or external state was modified.

1. R1-01/R2-B1, R1-03/R2-B2, and R1-04/R2-B3 are duplicates. The six unique families are
   default-deny assignment admission, malformed quote bounds, monotonic scanner work, finite public
   scalar fidelity, zero whole-key capture, and prompt settlement ownership.
2. The scanner's `looksLikeFlowMapping`/`looksLikeFlowSequence`, line-end discovery, and global
   quote-end search rescan suffixes. Replace them with one cursor and explicit member/key/value/
   flow/comment/malformed modes. Unsupported key characters retain uncertainty rather than
   abandoning assignment admission.
3. Export one exact work record: source length, cursor advances/regressions, maximum main visits,
   key/boundary visits, and total work. Charge all helpers; require zero regressions, one main visit
   per offset, and `totalWork <= 8n + 64`.
4. Public fidelity is a finite exact field/scalar grammar: locator/query-`=`, time-like/RFC3339,
   and syntactically closed quoted scalar lines stay byte-identical only at reviewed public fields.
5. Eight runtime call sites use captured whole-key materialization for result fields, arrays,
   events, record fields, JSON, and handoff validation. Replace every untrusted site with a fixed
   field projector, dense length/index capture, or opaque/redacted summary. Instrument all seven
   JavaScript whole-key primitives with 4,096 hidden peers and require zero calls/influence.
6. Convert prompt return/throw synchronously into an always-fulfilled settlement record with both
   rejection handlers installed before the next barrier. Table native promises and foreign
   thenables against abort/close/shutdown and exact cleanup/unhandled accounting.
7. Retain 137 rows and add exactly the six named Cycle 16 aggregate rows. Frozen-head acceptance is
   143 executed, 137 retained passes, exactly six intended failures, and frozen production.

### Cycle 16 implementation disposition

The advisory map was followed without widening its boundary. PLAN `6e3b05c8` and test-only RED
`7ec93fcd` preceded production changes; cohesive GREEN `c30cfe88` and exact-metric REFACTOR
`af9222b1` close all six rows at 143/143. Fixed descriptor projectors/dense indexes replace every
whole-key capture, arbitrary records are opaque, prompt settlement is installed before the
lifecycle barrier, and the scanner exposes only the exact seven-field monotonic work record. Both
Cycle 15 reports were re-read after refactor. The mapper remained read-only throughout.
