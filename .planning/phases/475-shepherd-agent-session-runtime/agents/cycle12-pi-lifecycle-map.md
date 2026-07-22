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
