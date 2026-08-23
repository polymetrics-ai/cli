# TDD ledger — source-import identity-bearing artifact query

| ID | Guarantee | Required red assertion | Green proof |
| --- | --- | --- | --- |
| Q1 | A declared v3 identity query reaches the fixed artifact fetch | A raw v3 lock with `identity_query:true` and `?version=` fails to parse/fetch under the current no-query gate. | The real importer retrieves its locked bytes through the HTTP fetch path, the transport observes `version`, and the projected operation is present. |
| Q2 | Capture queries and omitted declarations do not alter behavior | A v3 provenance `?slug=` and absent/false artifact declaration are exercised together. | No provenance citation request occurs, the artifact request is queryless, and absent/false descriptors are byte-identical. |
| Q3 | Identity query cannot carry credentials or exceed bounded policy | Raw v3 locks with a credential-shaped key, 17 keys, and a >1024-byte query must be rejected. | Every malicious lock is rejected before Fetch. |
| Q4 | Existing URL/SSRF guards survive the opt-in | Identity-declared URLs exercise HTTP, userinfo, fragments, localhost/private literal, and private DNS resolution. | Each error is rejected at lock validation or request destination validation; no request reaches the transport. |

## Red evidence

The declared-identity behavioral test was run before production edits and
failed at the strict v3 lock boundary, proving the locked query cannot reach
the importer:

```sh
go test -timeout 20m ./cmd/connectorgen -run '^TestSourceImportVersion3FetchesDeclaredIdentityQuery$' -count=1
```

The failure is retained in `traces/red-run.txt`.

## Green evidence

Pending implementation.
