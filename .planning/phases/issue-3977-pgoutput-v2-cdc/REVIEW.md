# Code review — #3977 pgoutput v2 CDC

## Scope reviewed

- PostgreSQL 14+ preflight and pgoutput protocol-v2 `streaming=on` startup.
- Transaction framing, bounded `CommittedTransactionStage` use, durable receipt,
  checkpoint, and standby-status ordering.
- Restart/rebootstrap, publication/slot lifecycle, capability projection, generated
  connector documentation, and the direct Unix-socket `dbtest` proof.

## Findings

The first CI run found two stale promotion-fence assertions outside the native
package: `internal/cli/changefeed_cli_test.go` still expected PostgreSQL to be
absent from the CDC catalog, and `internal/connectors/engine/database_definition_test.go`
still expected `metadata.cdc=false`. Both are part of this capability promotion,
so they now assert the exact executable descriptor and bundle projection. Their
focused CLI and engine suites pass locally.

The review also checked that no failure path can acknowledge an un-receipted
staged transaction, that an abort cannot reach the event/receipt/checkpoint/
acknowledgement sequence, and that promotion happens only after the recorded live
proof.

## Follow-up hardening

The review added a source + XID + WAL-position stage identity so an eventual XID
reuse cannot collide with a prior receipt. It also rejects a streamed data or
metadata frame whose embedded XID differs from the active stream. The focused
unit suite and the final live PostgreSQL container proof pass after this change.

## Review routing

The direct-PR task declares automated review unnecessary for this integration
branch. The normal automated-review routing was therefore recorded as not required;
local review, required local gates, and CI remain the delivery evidence.
