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
