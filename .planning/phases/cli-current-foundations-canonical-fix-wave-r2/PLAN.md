# Canonical Foundation Fix Wave R2 — GSD TDD plan

## Task Delivery Header

- Issue: Refs #4305, #4306, and #4307 — complete canonical blocker repair from the immutable Foundation review.
- Base branch: `fm/cli-current-foundations-postfix-fix-wave-r1` at `c3f83cbf6eabbae00219566fb02719ca2d6c480d` (remote/object identity verified before branching).
- Merges into: `fm/cli-current-foundations-postfix-fix-wave-r1` → `main`.
- Delivery: A clean, non-force-pushed `fm/cli-current-foundations-canonical-fix-wave-r2` branch with every FND-B01…FND-B19 behavior repaired, generated artifacts synchronized, and the report-specified verification plus no-mistakes pipeline recorded. No PR, merge, tag, or release is opened.
- Working branch: `fm/cli-current-foundations-canonical-fix-wave-r2`.
- Task: Repair the complete 19-blocker canonical set without connector-name special cases, generic write surfaces, provider-value masking heuristics, credentialed/live-provider checks, or reverse-ETL approval bypasses.
- Verification: Execute the immutable report’s source/generator, runtime/security, orchestration, conformance, certification, generated-artifact, static, package, built-binary, and no-mistakes checks; record each observable Red/Green result in `TDD-LEDGER.md` and final outcome in `VERIFICATION.md`.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Source locks and executable projections are closed | live | Generic REST/GraphQL fixtures, source inventory validation, and generated projections distinguish omitted/unreachable operations from explicitly source-bound gaps. |
| Declared commands are actually reachable | live | Engine/conformance provider doubles observe prepared PATCH/GraphQL/REST requests and reject invalid records before I/O. |
| Public outputs mask only configured/declared secrets | live | Focused projections preserve ordinary provider keys and token-shaped identifiers byte-for-byte while removing configured secret values from each public output field. |
| Bounded typed requests remain declaration-owned | live | Binary/request tests prove unknown, invalid, and mismatched parameters make zero sends/files; valid declarations produce exact requests. |
| Reverse ETL stays durable and acknowledged | live | Fakes observe admission, deadlines, exact checkpoint equality, full acknowledgement, and a no-replay empty-publication witness across retry/restart paths. |
| Generated docs, skills, website, certification, and evidence agree | live | Check-mode generators and tracked-output tests pass; schema-3 evidence validation binds the exact repaired subject and closure artifacts. |
| Provider/target execution is safe | fake | Unit/provider-double fixtures are required: the task forbids credentials, live provider mutation, runtime daemon, Docker, and PostgreSQL integration. Each fake asserts request, receipt, state, or failure boundary. |

## Required skills and workflow

Loaded: `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-context`, `golang-concurrency`, `golang-database`, `golang-graphql`, `golang-documentation`, `golang-lint`, `vercel-react-best-practices`, `vercel-composition-patterns`, and `no-mistakes`.

The repository GSD adapter was healthy; its sources for `discuss-phase`, `plan-phase`, `execute-phase`, `verify-work`, and `code-review` resolved, and `agentcontractgen check` passed. The upstream executor only accepts numeric phase IDs, while this Foundation recovery is a named Firstmate phase. The generated prompts are executed inline as the documented manual fallback; no isolated roles are spawned. This does not waive the sequence: discussion is frozen in the immutable review, this document is the TDD plan, execution is recorded below, then verification and deep code review follow.

## Scope and invariants

- The immutable review is authoritative: `data/cli-current-foundations-final-review-r1/report.md`, final verdict `BLOCKED`, exact source SHA `c3f83cbf6eabbae00219566fb02719ca2d6c480d`.
- No generic behavior may branch on connector name. Source identity, declaration metadata, and provider contracts remain the authority.
- Public clones, not internal receipts, are sanitized. Exact configured secrets and declared sensitive locations may be masked; ordinary names, occurrence IDs, token-shaped identifiers, headers, bodies, and provider fields are preserved.
- Reverse ETL remains declaration-owned and executes only through plan → preview → approval → execute with provider readback/receipt evidence.
- Every listed ID receives a behavior-first Red test before its fix, then a focused Green and relevant aggregate regression. No blocker is deferred for test expense.

## Delivery groups

### Group 1 — source authority, generation, reachability, evidence

| IDs | Red contract | Green/aggregate proof |
| --- | --- | --- |
| FND-B01 | A REST-only lock with `total=2`, REST count 1 is accepted. | Same fixture fails; `total=1` preserves exact descriptor/hash. |
| FND-B02 | Locked non-mutation REST `GET /widgets` disappears from an empty bundle. | Generic coverage requires one reachable disposition/command or concrete source-bound gap. |
| FND-B03 | Locked GraphQL `Query.widgets` disappears generically. | Protocol-neutral identity validates typed operation/output/command or concrete gap. |
| FND-B09 | Checked-in Foundation manifest cannot decode/close through schema-3 gate. | Exact-subject manifest validates; each mutated identity/digest/category/worktree input fails. |

Commit and non-force push immediately after focused Green, generators, source/certification checks, and a current evidence-only closure are green.

### Group 2 — declaration reachability and honest proof inputs

| IDs | Red contract | Green/aggregate proof |
| --- | --- | --- |
| FND-B04 | GitHub PR close/reopen hit a nested hook guard before I/O. | Plan/preview/approval `DryRunWrite` captures one declared PATCH and preserves acknowledgement/readback. |
| FND-B05 | Eleven Google Ads witnesses fail closed schemas before capture. | Valid declared arms reach exact capture; wrong/missing/extra fields remain pre-I/O errors. |
| FND-B06 | GitHub proof hides row failures behind stale generic-route topology. | Candidate set equals reported routes; injected topology/row failures both appear. |
| FND-B07 | Synthesized GitHub records violate min length/items/pattern constraints. | Production validator accepts six witnesses; unsupported/excessive patterns fail explicitly. |
| FND-B08 | Dual-direction GraphQL witnesses send neither direction. | One direction reaches capture/page receipt; both/neither remain zero-send failures. |

Regenerate affected bundle, CLI/help/manual/docs/skills/website projection and run GitHub/Google Ads conformance plus exhaustive provider-double proof before atomic push.

### Group 3 — safe receipts, secrets, and typed transfer authority

| IDs | Red contract | Green/aggregate proof |
| --- | --- | --- |
| FND-B10 | SQS key-name heuristics replace ordinary provider values. | Ordinary fields round-trip; only declared/configured secrets are absent publicly. |
| FND-B11 | REST, GraphQL, and SQS cursors expose configured secrets. | Public continuation withholds exact secret, preserves internal receipt and truth. |
| FND-B12 | Malformed GitHub App restriction is dropped, widening a token request. | Invalid non-empty restriction makes zero sends; valid/absent restrictions are intentional. |
| FND-B13 | Public non-JSON/XML receipt embeds configured secret. | Bounded public text masks concrete secret without changing internal raw/ordinary fields. |
| FND-B14 | Binary read accepts undeclared/untyped parameters and creates request/file. | Validator rejects zero-I/O escape paths and preserves valid exact requests. |

Commit and non-force push after public-output, engine, native SQS, hook, commandrunner, and CLI focused suites prove preservation and masking boundaries.

### Group 4 — durable destination/reverse-ETL delivery

| IDs | Red contract | Green/aggregate proof |
| --- | --- | --- |
| FND-B15 | PostgreSQL public target route calls DB before auth admission. | Every route enters admission and respects fence through all phases; cohort isolation is observed. |
| FND-B16 | Full-overwrite Begin has no unit deadline/pre-effect fence. | Ordinary, Arrow serial/pipeline Begin stop on deadline/lease takeover before effects/checkpoint. |
| FND-B17 | Continuation-only checkpoint changes compare equal. | Nil/kind/token participate in parked/resume/CAS equality before I/O. |
| FND-B18 | Partial/invalid nil-error bulk acknowledgement completes plan. | Complete action-aware accounting is required before terminalization. |
| FND-B19 | Empty overwrite has no durable no-replay witness. | Empty publication persists receipt/plan witness; post-publication faults repair without second publish. |

Commit and non-force push after synctransport, App, coordination, PostgreSQL, and synccontract focused proofs pass.

## CLI/docs/website parity and final gates

For each generator-affecting group, run `surface-sync`, tracked skills/manual/docs checks, connector-guide checks, website generator tests, `pm help <topic>`, connector namespace help, and representative `pm <connector> <command> --help`; never execute a reverse command against a provider. After all groups run every immutable-ledger command, `gofmt`, `go vet`, changed-package tests, `go build ./cmd/pm`, individual `make verify` gates, bounded `connectorgen boundary`, the full no-mistakes pipeline, and inline deep GSD review.
