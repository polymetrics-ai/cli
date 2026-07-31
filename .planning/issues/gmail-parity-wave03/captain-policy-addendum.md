<!-- gmail-parity-wave03-r1:captain-policy-addendum -->
## Captain-policy addendum — Gmail parity wave03 r1

Local fixture-only branch `fm/cli-gmail-parity-wave03-r1` completed the Gmail parity implementation slice for #3046/#3047-#3053 without live Gmail credentials, provider calls, provider writes, VPS/Thaalam checks, push, PR creation, or certification claims.

Post-change Gmail API ledger counts from Gmail Discovery revision `20260727`:

- total official operations: 79
- executable in the existing connector contract: 61
  - streams: 10
  - bounded `json_redacted` direct reads: 11
  - typed reverse-ETL writes: 40
- blocked/planned with explicit evidence/dependency: 18
  - 5 email-address path direct reads blocked on the shared direct-read identifier-safe path-variable validator
  - 2 `watch`/`stop` CDC controls blocked on shared CDC/changefeed foundations (#2986/#2988)
  - 11 Google Workspace Client-Side Encryption enterprise/admin operations tracked as not-applicable/admin/destructive operation rows
- live certified operations: 0 (fixture-only; no provider credentials/calls)

Safety notes:

- Gmail mutations remain behind reverse ETL plan → preview → explicit approval → execute.
- New destructive writes (`batch_delete_messages`, `delete_smime_info`) require destructive confirmation; delete semantics use idempotent missing-ok handling where the contract/provider supports it.
- Sensitive fields are redacted in previews/errors/output metadata where declared, including raw message/draft bodies, S/MIME PKCS#12/password fields, forwarding/delegate/send-as emails, SMTP password/signature fields, vacation response bodies, attachment data, and certificate material.
- No generic raw HTTP/path/body, shell, SQL write, or credential escape hatch was introduced.

Local verification completed:

```bash
no-mistakes doctor
go run ./cmd/connectorgen validate internal/connectors/defs/gmail
go test ./internal/connectors/conformance -run 'TestConformance/gmail' -count=1
go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1
go build ./cmd/pm
make connector-boundary
make verify
git diff --check
```

The implementation also adds a focused `connectorgen validate` tooling regression fix so the documented single-bundle gate validates `internal/connectors/defs/gmail` directly instead of treating `fixtures/` and `schemas/` as connectors.
