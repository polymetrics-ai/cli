# #3989 verification checklist

- [ ] Every acceptance row has an observable state-change assertion.
- [ ] External binary is freshly built and its SHA is recorded in sanitized evidence.
- [ ] No raw credential appears in any captured stream, argv, project file, vault/key, or artifact.
- [ ] Error/refusal paths write zero accepted-evidence artifacts.
- [ ] HTTPS transcript covers method, target/query, request/response headers and bodies, status, redirect/retry/error behavior, and explicit byte bounds.
- [ ] `go test -timeout 20m ./internal/connectors/certify/...` passes.
- [ ] Required local gates and code review complete.
- [ ] CLI help/manual/website parity checked or explicitly not applicable.
