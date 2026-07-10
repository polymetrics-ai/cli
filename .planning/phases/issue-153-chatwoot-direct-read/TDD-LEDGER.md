# TDD Ledger — Issue #153 Chatwoot direct read

| Cycle | Red test / evidence | Expected failure | Green change | Status |
|---|---|---|---|---|
| 1 | `go test ./internal/connectors/engine -run TestDirectReadBoundedJSONPolicyRedactsSecretKeys -count=1` | `bounded_json` output policy unsupported. | Add bounded JSON policy with recursive secret-key redaction. | planned |
| 2 | `go test ./internal/connectors/engine -run TestDirectReadScopedBasePathEndpointDoesNotDuplicatePrefix -count=1` | Chatwoot-like direct read duplicates `/api/v1/accounts/{account_id}` when surface endpoint and base URL both include the account prefix. | Strip the already-resolved base path prefix before dispatching through the requester. | planned |
| 3 | `go test ./internal/cli -run TestChatwootCommandSurfaceRunsDirectReadContactView -count=1` | Chatwoot direct-read command remains planned/blocked or lacks supported output policy. | Mark selected commands/endpoints implemented and covered by direct_read. | planned |

## Notes

- All tests use synthetic local HTTP servers and synthetic env-backed tokens; no live Chatwoot credentials are used.
- The direct-read runner remains endpoint-fixed by command metadata; no generic URL/path command is introduced.
