# Review — Issue #3982 PostgreSQL managed-table write driver

## Mapped write-driver re-review

`scripts/gsd prompt code-review cli-3982-postgres-write-driver-r1` was again
resolved through the documented inline/manual fallback after #4144 landed. The
review covered the optional shared mapping attachment, PostgreSQL DDL/type
renderer, target reassertion, session lock lifetime, each DML mode, tombstone
decoder, unknown-commit path, definition admission, and tagged dbtest evidence.

| Severity | Finding | Disposition |
| --- | --- | --- |
| Critical | None | — |
| Warning | A target assertion initially compared mapped type/nullability but not a declared string collation. | Fixed: physical target assertion now reads `pg_collation` and refuses a declared-collation mismatch. |
| Warning | A typed tombstone could encode `null` for a non-null key, turning the delete into a silent no-op. | Fixed: decoder now refuses a null key where the shared mapping says non-null. |
| Info | PostgreSQL has no native unsigned integer or int8 type. | DDL uses bounded `NUMERIC` / `SMALLINT` checks, mapping values are range checked before binding, and type/value unit tests cover the closed conversion table. |

No unresolved correctness, security, lifecycle, or scope finding remains. The
private definition admits only the five phase-one modes; public `write=false`,
legacy `Connector.Write`, registration, generic SQL, and history mode remain
unchanged.

## Method

`scripts/gsd prompt code-review cli-3982-postgres-write-driver-r1` was
resolved using the repository's documented inline/manual fallback: the
canonical contract forbids role spawning in this issue worker. Review scope was
the PostgreSQL managed-target driver, control ledger, tagged dbtest proof, and
the issue planning/TDD evidence through commit `0d35ba215`.

## Findings

| Severity | Finding | Disposition |
| --- | --- | --- |
| Critical | None | — |
| Warning | PostgreSQL permission failures raised while `pgx.Rows` iterated were originally collapsed to an unverifiable state. | Fixed in `loadNamespaceOwner`; live restricted-role proof now observes `ManagedTargetNamespaceOwnerUnreadable` / `ManagedTargetControlUnreadable`, and the provisioner refuses without mutation. |
| Info | First-create DDL, physical relation layout, five modes, tombstones, and receipts remain unimplemented. | Intentional hold: #3973 owns `MappingContractV1` / `DeliveryReceiptV1`. This driver returns before DDL, and the live test proves no namespace/relation/control state is created by the incomplete mapping path. |

## Security and correctness checks

- Every non-constant PostgreSQL identifier comes from a `ManagedTargetRef` and
  is passed through the native closed identifier quoter; all record/control
  values remain positional parameters.
- The driver takes only a pinned `pgx.Conn`, never a DSN or credential. Its
  mutex serializes pgx operations, while PostgreSQL advisory locks provide the
  cross-process namespace guard.
- Database, namespace, and relation OIDs are read from PostgreSQL catalog
  state. A changed physical OID, control schema fingerprint, unreadable
  control, or malformed foreign/colliding control assertion fails closed.
- The public connector capability remains `write=false`; this slice does not
  register a writer or admit a destination mode.

## Verification reviewed

- `go test -timeout 20m -count=1 ./internal/connectors/native/postgres/... ./internal/connectors/database/...`
- `go test -race -timeout 20m -count=1 ./internal/connectors/native/postgres ./internal/connectors/database`
- `go vet ./internal/connectors/native/postgres/... ./internal/connectors/database/...`
- Explicit direct-local Docker dbtest command from
  `internal/connectors/native/dbtest/README.md`, which passed the complete
  tagged PostgreSQL package in 22.719 seconds.

No unresolved review finding exists within the mapping-independent #3982
scope. A fresh review is required after #3973 lands and this branch adds the
write-session path.
