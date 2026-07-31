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
