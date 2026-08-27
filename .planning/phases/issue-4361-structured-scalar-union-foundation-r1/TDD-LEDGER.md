# TDD ledger — issue 4361

## Red

Planned red tests (before implementation):

1. Definition-backed engine preflight loads Twilio `create_address.City` and Xero `delete_payment.Status`, each declared as exact `string|null`, and expects the shared structured-JSON gate to admit them. **Observed RED:** both were rejected with `must declare type object or array`.
2. Commandrunner uses strict `type:"json"` named record flags with a `string|null` field. It expects a JSON string and `null` to build the record, while number, boolean, array, object, unknown field, missing required field, and null against a string-only field fail before I/O. **Observed RED:** the source-backed Xero `delete_payment.Status` command is blocked at the same declaration preflight before it can materialize any input.
3. Source projection of exact `string|null` expects a bounded `json` flag with no `allow_bare_string`, preserving a closed JSON grammar. **Observed RED:** it emits `type:"string"` with `max_bytes:256`, which cannot represent null distinctly.
4. Operation structured-body preflight (synthetic fixed `rest_write`) expects the same exact union admission while preserving fixed method/path/body and full structured-schema validation. **Observed RED:** it rejects `label` with `must be an object or array`.

### Red commands

```sh
go test -timeout 20m ./internal/connectors/engine -run 'TestValidateStructuredJSONRecordFieldAcceptsSourceDeclaredNullableString|TestValidateOperationStructuredJSONBodyFieldAcceptsExactNullableString'
go test -timeout 20m ./internal/connectors/commandrunner -run TestBuildWriteCommandValidatesSourceDeclaredNullableStringJSON
go test -timeout 20m ./cmd/connectorgen -run TestSourceProjectionNullableStringUsesStrictNamedJSONFlag
```

Both commands failed only at the expected nullable-scalar declaration boundary; no provider I/O or credentials were used.

## Green

Pending implementation and exact results.

## Refactor

Pending; no broad union, generic body, connector exception, or raw transport input is permitted.
