# TDD Ledger: Foundation public-output repair r1

| Finding | Red evidence to add | Green behavior | Edge and regression evidence |
| --- | --- | --- | --- |
| FND-B10 | Name-shaped SQS/provider output is wrongly redacted. | Only configured material and its concrete encodings are masked. | `id`, `occurrence_id`, token-shaped values, headers, and key names remain byte-for-byte intact. |
| FND-B11 | Public cursor/receipt serialization contains configured material. | Every public cursor and receipt projection masks it while retaining ordinary identifiers. | JSON escaped and printable/encoded values cannot bypass masking. |
| FND-B12 | Invalid GitHub App restrictions continue to an authenticated request. | Invalid syntax/semantics returns a validation error before I/O. | Valid restriction succeeds; malformed empty/duplicate/out-of-range forms retain a zero request counter. |
| FND-B13 | Non-JSON public diagnostics expose configured material or erase ordinary text. | Printable text forms mask only configured material/representations. | Exact harmless words, provider IDs, and occurrence IDs remain unchanged. |
| FND-B14 | Binary download binds undeclared/unsafe parameters. | The declared parameter authority gate rejects before I/O. | Declared parameters make one exact request; unknown, cross-operation, invalid, and unsafe values make none. |
| FND-W02 | Status execution admits bindings outside the declaration. | Status uses the same authority gate as operation requests. | Valid declared parameters preserve request fidelity and invalid bindings make no request. |

## Actual evidence

### Slice 1 — FND-B10, FND-B11, FND-B13 public output

- Red: `go test -count=1 -timeout 20m ./internal/connectors/native/amazon-sqs -run 'TestOperationDirectReadListQueuesPreservesOrdinaryNameShapedProviderValues'` failed because the native field-name heuristic replaced the ordinary `Policy` value.
- Red: `go test -count=1 -timeout 20m ./internal/connectors/engine -run 'TestOperationDirectReadMasksConfiguredCursorAtThePublicBoundary'` failed because `next_cursor` was returned unchanged.
- Red: the receipt test failed against the pre-fix printable/substring sanitizer, demonstrating that a configured short value corrupts ordinary `occurrence_id` output instead of retaining the provider response faithfully.
- Green: `go test -count=1 -timeout 20m ./internal/connectors -run 'Test(PublicReceiptSanitizationMasksConcreteSecretRepresentationsWithoutChangingProviderNames|WriteResultOutputMasksConfiguredSecretsAndPreservesOrdinaryProviderTruth|OperationDirectWriteResultOutputMasksConfiguredAndDeclaredSecrets|SanitizeWriteErrorForOutputKeepsSystemDiagnosticsSecretFree)'`.
- Green: `go test -count=1 -timeout 20m ./internal/connectors/native/amazon-sqs -run 'TestOperationDirectReadListQueuesPreservesOrdinaryNameShapedProviderValues'`.
- Green: `go test -count=1 -timeout 20m ./internal/connectors/engine -run 'TestOperationDirectReadMasksConfiguredCursorAtThePublicBoundary'`.

At this checkpoint, FND-B12, FND-B14, and FND-W02 remained pending their separate red-green slices.

### Slice 2 — FND-B12 GitHub App restrictions

- Red: `go test -count=1 -timeout 20m ./internal/connectors/hooks/github -run 'TestAuthenticatorGithubAppRestrictionParsingFailsClosedBeforeDeclaredRouteIO'` failed for empty and unsafe repository entries, malformed or duplicate repository IDs, and malformed permissions because each silently fell through to the declared-route exchange.
- Green: `go test -count=1 -timeout 20m ./internal/connectors/hooks/github -run 'TestAuthenticatorGithubAppRestrictionParsingFailsClosedBeforeDeclaredRouteIO'` proves valid restrictions make one declared-route POST while every malformed restriction returns before authenticated provider I/O.

FND-B14 and FND-W02 remain pending their shared parameter-authority slice.
