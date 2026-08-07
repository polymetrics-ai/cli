# TDD Ledger — Issue 3744

## Red / green evidence

| Slice | Red evidence | Green evidence |
| --- | --- | --- |
| Optional redaction fields | `TestBundleLoadAcceptsSecretOperationWithoutRedactFields` failed for both `secret_sensitive` and `mutation_class: "secret"` with `sensitive_policy must declare at least one redact_fields entry`. | Removing that branch makes both loader cases pass with no declared redaction fields. |
| Preserved policy validation | Existing loader tests preserve missing-policy, inline-input-mode, and typed-confirmation rejections; direct validator tests cover unknown input mode and unknown transform. | All sensitive-policy tests pass after the forced redaction removal. |
| Honest diagnostic | Automated review identified the stale missing-policy message as still naming `redact_fields` as mandatory. | The diagnostic now names only required policy components; its loader test asserts it does not mention `redact_fields`. |

## Commands

```sh
go test ./internal/connectors/engine -run '^TestBundleLoadAcceptsSecretOperationWithoutRedactFields$' -count=1 # red before implementation
go test ./internal/connectors/engine -count=1
```

No fixture or connector declaration changed; the red/green tests use in-memory bundle files.
