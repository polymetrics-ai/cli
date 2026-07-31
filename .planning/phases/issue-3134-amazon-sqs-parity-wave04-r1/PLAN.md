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
