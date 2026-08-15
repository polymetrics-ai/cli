# Verification checklist — Issue 4095

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Explicit CDC delete reaches PostgreSQL history close | pending live | Query the real target’s versions after a CDC-derived tombstone; the latest is retained but closed, never physically removed. |
| Physical absence is never a deletion instruction | pending live | Query the real non-history target after a projection omits a prior row, then again after its explicit CDC-derived tombstone. |
| Shared mapping does not accept ambiguous delete keys | pending fake | Unit test verifies exact target-key projection and rejects malformed source tombstones before any write input. |
| Required package and live harness regression | pending | Run the launch-brief targeted packages and the complete tagged native PostgreSQL dbtest command. |

## Planned non-test gates

- `gofmt -w` on changed Go files
- `go vet` on changed packages
- `go build ./cmd/pm`
- relevant individual `make verify` non-test gates, per repository policy
