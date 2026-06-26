# TDD Ledger

## Reverse ETL Approval

- Red: `TestRunReverseETLRejectsApprovalTokenReplay` failed because replayed approval tokens executed again.
- Red: `TestRunReverseETLRejectsPlanHashMismatchWhenRowsChange` failed because execution re-read changed rows without rejecting.
- Green: plan hashes now include mapped records; execution rejects non-planned plans and consumes valid tokens before writes.

## Error Redaction

- Red: `TestRedactErrorTextRemovesHTTPURLQueryAndBodySecrets` failed because no central redactor existed.
- Red: `TestHTTPErrorErrorRedactsURLQueryAndBody` and `TestRequesterDoJSONDecodeErrorDoesNotIncludeRequestURL` failed because HTTP errors exposed full request context.
- Green: safety redaction is applied to CLI output, persisted errors, and connsdk errors.

## CLI Validation

- Red: malformed `--config`, `--map`, and integer flags were silently ignored/defaulted.
- Green: parsers now return validation errors and agent JSON gets category `validation`.

## Path Policy

- Red: `TestWarehouseCredentialExternalPathRequiresOptIn` showed external write paths were accepted without opt-in.
- Green: warehouse/outbox external write paths require `allow_external_path=true`; file read paths remain allowed.

## Registry Semantics

- Green tests added for production live registry excluding staged connectors and staged registry including them.
- Generator updated so helper packages are skipped and live/staged constructors persist across regeneration.

## Materialization Seam

- Red: `TestRunETLUsesMaterializationInterfaceInsteadOfWarehouseName` failed because a non-`warehouse` materializer used `Write`.
- Green: `connectors.LocalWarehouseMaterializer` replaces `destination.Name() == "warehouse"`.

## Read Limit

- Red: `TestQueryTableStopsAtLimitBeforeLaterDecodeError` failed because reads scanned past the limit.
- Green: `connectors.LimitEmitter` stops reads with `ErrReadLimitReached`, ignored at caller seams.

## New Modules

- `internal/runtime`, `internal/state`, and `internal/connectors/httpsource` were implemented with package-level tests by subagents and verified with `GOTOOLCHAIN=auto`.
