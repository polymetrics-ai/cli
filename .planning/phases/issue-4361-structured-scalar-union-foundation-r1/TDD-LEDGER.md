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

### Empty-string baseline proof

`origin/main`'s `validateRequiredCommandFlags` decoded every required JSON flag and then called `commandValueEmpty`; its string branch treats `strings.TrimSpace("") == ""` as missing. The pre-change JSON decoder returned `""` for the literal JSON value `""`, so a required named JSON flag already failed as `missing required flag`. Xero's source-declared `delete_payment.Status` schema is exactly `{"type":["string","null"]}` with no `minLength`, but the command-required policy was independently stricter and remains unchanged. The green test asserts that existing policy; the nullable repair changes only the formerly-empty `nil` result for explicit JSON `null`, which the declaration schema then authorizes or rejects.

### Full generator regression classification

The captured branch run `go test -timeout 20m ./cmd/connectorgen` failed after 190.301s in exactly two pre-existing GitHub projection tests: `TestInstalledReverseActions_CoverProviderRequestContract` (three missing source-field rows) and `TestSourceProjectionSourceCitedMutationDispositionLeavesExistingProjectionByteIdentical` (`CLI:32`). A clean detached worktree at the same `origin/main` SHA `2165619ec8f5f9d4141b491b7a5a64bc460d0c71` passed the identical command after 190.128s. This was a patch regression, not a baseline blocker. The repair records `StrictJSONFields` only when the source descriptor has the exact `cli-structured-scalar-union-foundation-r1` field-location gap; existing admitted source projections retain their prior byte-identical flag representations. The repaired full run passed after 181.143s.

## Green

The implementation accepts only a two-member `string|null` (or `null|string`) type array with one of each member. It leaves all other scalar shapes subject to the prior refusal and preserves the closed top-level-field schema boundary.

```sh
go test -timeout 20m ./cmd/connectorgen -run TestSourceProjectionNullableStringUsesStrictNamedJSONFlag
# ok   polymetrics.ai/cmd/connectorgen  1.118s
go test -timeout 20m ./internal/connectors/engine -run 'TestValidateStructuredJSONRecordFieldAcceptsSourceDeclaredNullableString|TestValidateOperationStructuredJSONBodyFieldAcceptsExactNullableString'
# ok   polymetrics.ai/internal/connectors/engine  (cached)
go test -timeout 20m ./internal/connectors/commandrunner -run TestBuildWriteCommandValidatesSourceDeclaredNullableStringJSON
# ok   polymetrics.ai/internal/connectors/commandrunner  (cached)
go test -timeout 20m ./cmd/connectorgen
# ok   polymetrics.ai/cmd/connectorgen  181.143s
```

The commandrunner test uses the source-backed Xero `delete_payment.Status` field and proves string and explicit JSON `null` materialize, while a number, boolean, array, object, unknown field, missing field, `""`, and `null` against a string-only field fail during closed-command validation.

## Refactor

The reusable type decoder removes duplication between structured-field admission and the string-arm check. Generator projection keeps the same closed action/record path and has no Twilio/Xero-name exception, raw body, route, method, action, provider I/O, or credential change.
