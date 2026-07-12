# TDD Ledger — Issue #153 Chatwoot direct read

| Cycle | Red test / evidence | Expected failure | Green change | Status |
|---|---|---|---|---|
| 1 | `go test ./internal/connectors/engine -run TestDirectReadBoundedJSONPolicyRedactsSecretKeys -count=1` | Captured red failure: `direct read output policy "bounded_json" is not supported`. | Added bounded JSON policy with recursive secret-key redaction. | passed |
| 2 | `go test ./internal/connectors/engine -run TestDirectReadScopedBasePathEndpointDoesNotDuplicatePrefix -count=1` | Captured red failure: request path duplicated `/api/v1/accounts/{account_id}`. | Stripped the already-resolved base path prefix before dispatching through the requester. | passed |
| 3 | `go test ./internal/cli -run TestChatwootCommandSurfaceRunsDirectReadContactView -count=1` | Chatwoot direct-read command would remain planned/blocked or lack supported output policy before metadata/policy changes. | Marked selected commands/endpoints implemented and covered by direct_read. | passed |

## Notes

- All tests use synthetic local HTTP servers and synthetic env-backed tokens; no live Chatwoot credentials are used.
- The direct-read runner remains endpoint-fixed by command metadata; no generic URL/path command is introduced.
