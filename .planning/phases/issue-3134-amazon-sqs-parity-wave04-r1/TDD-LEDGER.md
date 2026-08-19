# TDD Ledger — Amazon SQS parity wave04 r1

## Red phase targets

- [x] Operation ledger test fails until `api_surface.json` covers all 23 official SQS operations exactly once.
- [x] Native action/direct-read coverage test fails until every operation has a typed native executor or the messages read stream.
- [x] Destructive manifest test fails until delete/purge/cancel/admin destructive operations carry `confirm: "destructive"` and redaction fields.
- [x] Write validation tests fail for missing required closed-schema fields and unknown actions.
- [x] Direct-read tests fail until XML response bodies are decoded to bounded redacted JSON result bodies.
- [x] Write shape tests fail until form-encoded SQS Query API requests use fixed Action names and required parameters.

## Green/refactor evidence

- Added native `amazon_sqs_test.go` coverage for 23-operation ledger parity, manifest write action count, destructive confirmations, closed-schema validation, batch chunking, direct reads, XML decoding, and redaction behavior.
- Implemented fixed SQS Query API direct-read and write dispatchers without generic AWS Action/body escape hatches.
- Focused green runs before compaction: `go test ./internal/connectors/native/amazon-sqs -count=1`, focused `go run ./cmd/connectorgen validate`, `go test ./internal/connectors/conformance -run 'TestConformance/amazon-sqs' -count=1`, and after compaction `go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1` after golden refresh.
- Post-review hardening tests cover unsafe URL rejection, required `QueueUrl` form injection for reads/destructive queue writes, whitespace-normalized write actions, destructive preview warnings, and blank direct-read redaction fields.

## Notes

No live AWS calls. All tests use fixture mode or `httptest.Server` with synthetic credentials and sanitized payloads.

## 2026-08-02 corrective TDD ledger — issueguard delta removal

Red/guard targets before production edits:

- [x] Restored working tree content has no diff from `origin/main` under `internal/coordination/issueguard/**`.
- [x] Restored working tree content has no diff from `origin/main` under `cmd/prissueguard/**`.
- [x] `cmd/prissueguard` accepts the corrected PR body through the established `Refs #3134` contract without shared issueguard weakening.
- [x] PR #3541 body contains accepted `Refs #3134` linkage and no longer contains the forbidden canonical raw issue URL workaround.

Green evidence:

- Restored forbidden shared paths from `origin/main` in the working tree for a forward corrective commit.
- `go test ./internal/coordination/issueguard ./cmd/prissueguard -count=1` passed after restoration.
- `go run ./cmd/prissueguard --title "chore(amazon-sqs): stage unvalidated parity checkpoint" --body-file /tmp/pr3541-body-refs-3134.md` passed with `issueguard: ok (1 linked issue)`.
- Focused Amazon SQS connector gates passed after restoration.
- Fresh native-Codex `gpt-5.6-sol` `xhigh` no-mistakes validation remains required after the corrective commit and push.

## 2026-08-02 review-finding fix round

Red evidence was the fresh review reproduction of four SQS-local gaps before production edits: reverse CLI examples omitted mandatory arguments, the 23-operation proof did not directly execute ten typed paths, certification classifiers could not match the emitted error kind or discarded AWS codes, and message-attribute preview validation admitted provider-invalid shapes.

Regression coverage authored before the fixes:

- CLI surface test requires every reverse command's mandatory flags and runnable synthetic example arguments while preserving zero-input delete/purge commands.
- Focused native tests execute GetQueueUrl and the nine previously uncovered write/admin actions through synthetic `httptest` endpoints.
- Connection test requires bounded sanitized AWS error-code reporting without exposing provider messages or bodies.
- Dry-run tests reject missing/incompatible message-attribute scalars, reserved list variants, invalid ordinary names, and unsupported system attributes.

The isolated no-mistakes review phase uses a local critical path because all fixes and tests collide in the SQS-owned bundle/native package; no subagents were used.

Green evidence: `go test ./internal/connectors/native/amazon-sqs -count=1` passed after the complete fix round; no live or credentialed AWS calls were made.

## 2026-08-02 message-attribute semantic follow-up

Red evidence is the reproduced boundary behavior from the fresh review: malformed Number strings, out-of-range/over-precision numbers, invalid base64 BinaryValue data, and malformed AWSTraceHeader strings all passed `DryRunWrite` because only type/value presence was checked.

Regression coverage authored before production edits adds table-driven rejection cases for each invalid semantic class and one acceptance case spanning numeric boundaries, 38-digit precision, binary base64, and a synthetic X-Ray trace header. Per the review-phase contract, the tests are run once after the complete fix round rather than between edits.

Green evidence: `go test ./internal/connectors/native/amazon-sqs -count=1` passed after semantic validation was applied; all cases are local and synthetic.

## 2026-08-02 decoded-empty binary follow-up

Red evidence is the reproduced boundary behavior from the fresh review: a newline-only BinaryValue is raw-nonempty, but Go's strict base64 decoder ignores the newline and produces zero bytes without error, so `DryRunWrite` accepts an effectively empty provider value.

Regression coverage authored before production edits adds a table-driven newline-only BinaryValue rejection case at the public dry-run boundary. Per the review-phase contract, the focused package test runs once after the complete fix round.

Green evidence: `go test ./internal/connectors/native/amazon-sqs -count=1` passed after the decoded-byte invariant was enforced; all inputs and endpoints remain local and synthetic.

## 2026-08-03 CI verify staticcheck lint-fix

Red evidence is the reproduced remote gate: `golangci-lint run ./internal/connectors/native/...` on the pre-fix head (`1f7971684`) reports `S1003: should use strings.ContainsAny(exponentText, "eE") instead (staticcheck)` at `message_attribute_values.go:17:28`, matching the PR #3541 `verify` failure annotation exactly.

Regression coverage is the idempotent one-line correction itself: `strings.IndexAny(exponentText, "eE") >= 0` -> `strings.ContainsAny(exponentText, "eE")` preserves the exact predicate semantics, so the existing table-driven Number validation cases continue to hold without new test lines.

Green evidence: `golangci-lint run ./internal/connectors/native/...` returns `0 issues` after the fix, `go test ./internal/connectors/native/amazon-sqs -count=1` still passes, `go run ./cmd/connectorgen validate internal/connectors/defs/amazon-sqs --json` reports no findings, and `pm docs validate --connectors-dir docs/connectors` passes; all inputs remain local and synthetic.
