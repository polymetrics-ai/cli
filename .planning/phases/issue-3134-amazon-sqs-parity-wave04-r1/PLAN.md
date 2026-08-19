# GSD Plan — Amazon SQS official API parity wave04 r1

## Scope

Parent issue #3134 and subissues #3135-#3141. Implement complete documented Amazon SQS (`amazon-sqs`) operation-level parity in the connector-local Tier-3 native connector and definition bundle only.

Allowed paths: `internal/connectors/defs/amazon-sqs/**`, `internal/connectors/native/amazon-sqs/**`, owned connector docs/generated catalog/help surfaces, and this GSD artifact directory.

Explicit non-goals/safety gates: no live AWS calls, no credential requests, no generic AWS action/raw HTTP/body escape hatch, no shared runtime behavior changes, no certification claim, no push/PR/no-mistakes, no Herdr lifecycle commands.

## GSD command and fallback evidence

- `scripts/gsd doctor` passed in this worktree.
- `scripts/gsd prompt programming-loop init --phase issue-3134-amazon-sqs-parity --dry-run` failed because this repo-local adapter does not expose `programming-loop` in `scripts/gsd list`.
- Fallback used: `scripts/gsd prompt plan-phase 3134 --skip-research` generated `/tmp/gsd-plan-phase-3134.prompt`; this manual GSD/TDD plan records the required plan, TDD ledger, and verification checklist before production edits.

## Required skills loaded

- `gsd-core`
- `golang-how-to`
- `golang-design-patterns`
- `golang-structs-interfaces`
- `golang-error-handling`
- `golang-security`
- `golang-safety`
- `golang-testing`
- `golang-context`
- `golang-concurrency`
- `golang-cli`
- `golang-documentation`
- CLI help/docs/website parity reference

## Official-source re-audit ledger

Sources re-audited before production edits:

- AWS SQS Smithy model: `aws_sqs_smithy_model` (`https://raw.githubusercontent.com/aws/aws-sdk-go-v2/main/codegen/sdk-codegen/aws-models/sqs.json`), API version 2012-11-05.
- AWS SQS API Reference operations page: `aws_sqs_operations_docs` (`https://docs.aws.amazon.com/AWSSimpleQueueService/latest/APIReference/API_Operations.html`).
- Landed audit record referenced by issues is not present in this worktree at `cli-official-api-parity-audit-r2/audit.json`; issue bodies provide the landed r2 counts and provenance. Official source re-audit confirms the same 23 documented operations.

Confirmed official operations (23): AddPermission, CancelMessageMoveTask, ChangeMessageVisibility, ChangeMessageVisibilityBatch, CreateQueue, DeleteMessage, DeleteMessageBatch, DeleteQueue, GetQueueAttributes, GetQueueUrl, ListDeadLetterSourceQueues, ListMessageMoveTasks, ListQueues, ListQueueTags, PurgeQueue, ReceiveMessage, RemovePermission, SendMessage, SendMessageBatch, SetQueueAttributes, StartMessageMoveTask, TagQueue, UntagQueue.

Planned count disposition after this slice: total 23, implemented 23, fixture/native-tested 23, blocked/planned 0, excluded/not-applicable 0, certified 0.

The `fixture/native-tested 23` count means aggregate focused native tests directly execute every fixed typed read, direct-read, and write path with fixture data or synthetic `httptest` endpoints. Bundle conformance dynamic replay remains skipped by design and is not counted as operation execution evidence.

## Implementation slices

Resume checkpoint after context compaction: connector-local bundle and native implementation are in place, focused native/conformance validation passed once, and the CLI golden transcript fixture was regenerated to reflect the new SQS dynamic command surface. Remaining work is targeted inspection, documentation/generated-surface cleanup, full local gates, issue addendum comments if available, and a clean local commit.

Post-resume safety hardening added URL canonicalization (`queue_url`/`endpoint_url` reject userinfo/query/fragment and non-local HTTP), sanitized `QueueUrl` injection, normalized write action lookup, and blank-safe direct-read redaction. A follow-up security-auditor pass reported no actionable findings. Captain-policy addendum comments were posted idempotently to #3134-#3141.

1. Red tests/native coverage:
   - Add operation-level native tests that require all 23 operations to be ledger-covered and executable through fixed typed actions/direct-read/read paths.
   - Add tests for destructive confirmation metadata, redaction of receipt handles/message bodies, bounded ReceiveMessage forms, direct read XML-to-JSON mapping, write form shapes, batch chunking, and provider-supported idempotency metadata where applicable.
2. Definition bundle:
   - Update `metadata.json` capabilities/risk to read+write, with explicit native dynamic skip evidence.
   - Add static `streams.json`/`schemas/messages.json` only to make the native messages stream visible to the v2 ledger and generated surfaces; dynamic execution remains native and conformance dynamic replay is explicitly skipped with native test evidence.
   - Add `writes.json`, `cli_surface.json`, `certification.json`, fixtures for messages/direct reads/writes, and complete `api_surface.json` covering every official operation exactly once.
   - Update `docs.md` with operation ledger, auth, bounded read, typed actions/risks, destructive gates, and known native limits.
3. Native connector:
   - Preserve SigV4 signing and fixture mode; add typed operation dispatcher helpers without generic arbitrary AWS actions.
   - Implement `OperationDirectRead` for fixed read operations.
   - Implement `ValidateWrite`, `DryRunWrite`, `Write`, and `Manifest` for fixed action set.
   - Ensure destructive queue/message operations require `confirm: "destructive"` in the manifest and rely on app plan -> preview -> approval -> execute.
4. Generated/docs surfaces:
   - Regenerate connector docs/catalog/help surfaces with project generators if available; otherwise update owned generated files exactly.
5. Issue addendum:
   - Append idempotent captain-policy addendum comments to #3134-#3141 with truthful post-change counts and gate list.
6. Verification:
   - Run focused connectorgen validation, conformance, native tests, focused CLI tests, build, connector-boundary, make verify, and git diff check.
   - Commit only after clean local gates.

## 2026-08-02 forward corrective slice — remove unrelated issueguard deltas

Authority: captain decision `sqs-ci-issueguard-removal-20260802` mirrored by Alfred to issue #3134 and PR #3541. The prior no-mistakes validation run `01KZ13QR5T4Z4FDAF6QR25GQFT` was synchronized to published head `66bc2d0f4a70550649e95c821d07fec9b04c796a`; its completed CI-monitoring phase was ended with `no-mistakes axi abort --run 01KZ13QR5T4Z4FDAF6QR25GQFT` before local production edits.

GSD evidence for this corrective slice:

- `scripts/gsd doctor` passed.
- `scripts/gsd prompt programming-loop init --phase issue-3134-amazon-sqs-ci-issueguard-removal --dry-run` is unavailable in this adapter (`unknown GSD command: programming-loop`).
- Fallback used: `scripts/gsd prompt gsd-quick "issue-3134 amazon-sqs ci issueguard removal"`, then manual GSD/TDD execution recorded in this plan, `TDD-LEDGER.md`, and `VERIFICATION.md` before production edits.
- Required skills loaded/active: `gsd-core`, `no-mistakes`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-security`, `golang-safety`, `golang-error-handling`, `golang-lint`.

Scope and constraints:

- Remove every branch delta under `internal/coordination/issueguard/**` and `cmd/prissueguard/**` by a forward corrective commit, restoring those paths to content-equivalent approved base behavior.
- Correct PR #3541 body through Alfred to the established `Refs #3134` contract rather than weakening shared issueguard behavior.
- Do not reset, rebase, force-push, discard SQS validation history, retain or relocate the issueguard behavior change, add dependencies, run credentialed connector checks, or edit unrelated shared behavior.
- After the corrective commit, start fresh comprehensive native-Codex `gpt-5.6-sol` validation at `xhigh`; do not claim readiness until that fresh run and remote checks pass.

Implementation steps:

1. Compare branch changes under `internal/coordination/issueguard/**` and `cmd/prissueguard/**` against the approved base.
2. Restore those paths from base into the working tree, preserving all SQS-owned commits in history.
3. Inspect PR #3541 body and replace forbidden/canonical-link workaround content with the accepted `Refs #3134` issue linkage through `gh-axi pr edit` as Alfred.
4. Run focused verification for issueguard/prissueguard and SQS gates, commit the forward correction, push normally, then launch the required fresh 5.6 SOL no-mistakes validation.

## 2026-08-02 review-finding corrective slice

- Reproduced and accepted all four SQS-local findings before edits; no shared engine, runner, issueguard, prissueguard, unrelated connector, or generic runtime changes are in scope.
- Root causes: incomplete CLI command metadata, incomplete operation-execution proof, unreachable certification classifiers plus discarded provider codes, and permissive shared message-attribute boundary validation.
- Required skills loaded: `gsd-programming-loop`, `no-mistakes`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-lint`, and `golang-documentation`.
- `scripts/gsd doctor` passed; `scripts/gsd prompt programming-loop init --phase issue-3134-amazon-sqs-parity-wave04-r1 --dry-run` returned `unknown GSD command: programming-loop`, so this plan, the TDD ledger, and verification checklist record the required manual GSD/TDD fallback.
- Execution is local critical path: the review phase is isolated and the fixes/tests share the same SQS-owned bundle and native package, so no subagents are used.
- Fix forward with required flags plus synthetic examples, exact missing-path tests, safe reachable `Error` classifiers, bounded code-only AWS errors, and strict scalar/system message-attribute validation before preview.

## 2026-08-02 message-attribute semantic validation follow-up

- Reproduced the remaining SQS-local gap before edits: Number, Binary, and AWSTraceHeader values reached plan/preview after only nonempty scalar checks.
- Root cause is shared semantic validation at the native SQS write boundary, not individual action serialization; the fix remains confined to `internal/connectors/native/amazon-sqs` and this phase trace.
- Official AWS constraints applied: Number syntax with at most 38 digits of precision and nonzero magnitude from `1e-128` through `1e126`, standard base64 BinaryValue encoding, and an AWS X-Ray trace header with a valid Root plus validated optional Parent/Sampled fields.
- `scripts/gsd doctor` passed; the required `scripts/gsd prompt programming-loop init --phase issue-3134-amazon-sqs-parity-wave04-r1 --dry-run` remains unavailable as `unknown GSD command: programming-loop`, so the manual GSD/TDD fallback continues here.
- Required skills loaded: `gsd-programming-loop`, `no-mistakes`, `golang-how-to`, `golang-testing`, `golang-safety`, `golang-security`, `golang-error-handling`, `golang-design-patterns`, `golang-structs-interfaces`, and `golang-lint`.
- Execution decision: `local_critical_path`; the single validator and its focused tests share one SQS-owned package in this isolated review phase.

## 2026-08-02 decoded-empty binary follow-up

- Reproduced the SQS-local gap before production edits: Go's strict standard-base64 decoder still ignores CR/LF, so a raw nonempty newline-only BinaryValue decodes successfully to zero bytes and reaches preview.
- Root cause is the decoded-value invariant at the shared native SQS binary-attribute boundary, not an individual caller or serializer.
- `scripts/gsd doctor` passed; the required `scripts/gsd prompt programming-loop init --phase issue-3134-amazon-sqs-parity-wave04-r1 --dry-run` remains unavailable as `unknown GSD command: programming-loop`, so the manual GSD/TDD fallback continues here.
- Required skills loaded: `gsd-programming-loop`, `no-mistakes`, `golang-how-to`, `golang-testing`, `golang-safety`, `golang-security`, `golang-error-handling`, `golang-design-patterns`, `golang-structs-interfaces`, and `golang-lint`.
- Execution decision: `local_critical_path`; this assigned review fix is one validator plus its collocated regression case, and subagent delegation is not authorized for this turn.

## 2026-08-03 CI verify staticcheck lint-fix slice

- Reproduced the failing remote gate before production edits: PR #3541 `verify` failed with `S1003: should use strings.ContainsAny(exponentText, "eE") instead (staticcheck)` at `internal/connectors/native/amazon-sqs/message_attribute_values.go:17:28`, introduced by the 2026-08-02 message-attribute semantic validation commit (`6d996aad1`), which made `golangci-lint` exit 2 and failed the workflow.
- Root cause is a single non-idiomatic membership check at the shared native SQS Number-attribute boundary; no behavior change is intended or made. The one-line fix stays inside `internal/connectors/native/amazon-sqs` and this phase trace.
- `scripts/gsd doctor` passed; the required `scripts/gsd prompt programming-loop init --phase issue-3134-amazon-sqs-parity-wave04-r1 --dry-run` remains unavailable as `unknown GSD command: programming-loop`, so the manual GSD/TDD fallback continues here.
- Required skills loaded: `gsd-programming-loop`, `no-mistakes`, `golang-how-to`, `golang-lint`, `golang-safety`, `golang-error-handling`, `golang-testing`, and `golang-code-style`.
- Execution decision: `local_critical_path`; the fix is one idempotent lint correction with no new runtime behavior and no subagent delegation needed.
