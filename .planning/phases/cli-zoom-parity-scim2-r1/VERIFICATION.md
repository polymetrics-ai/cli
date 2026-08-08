# Verification Checklist — Zoom SCIM2 parity, R1

## Lifecycle

- [x] GSD command provenance was resolved with `scripts/gsd sources`.
- [x] Required skills and canonical connector/CLI references are recorded in `PLAN.md`.
- [x] Live provider artifact URL, retrieval date, HTTP result, byte count, SHA-256, and exact
  eleven-operation audit are recorded before RED.
- [x] RED failure captured verbatim before production changes; this test-only checkpoint is staged
  for its required commit/push.
- [x] Required reusable foundations red/green tested and separated from connector authoring:
  operation-scoped direct-read origin/auth is pushed; named root-object mapping is green in this
  pending foundation checkpoint.
- [ ] Connector declaration, generated output, docs, and website catalog committed/pushed.
- [ ] Inline verify-work and code-review evidence complete.

## Source parity

- [ ] All eleven SCIM2 ledger rows have exactly one executable disposition.
- [ ] Zero Zoom rows are `unsafe_or_disallowed`.
- [ ] Every command accepts only documented typed path/body members; no paging flags are invented.
- [ ] Each 204 action proves status-only success and no request body when none is declared.
- [ ] User/group data is redacted from previews, errors, and results according to declared policy.

## Runtime/docs checks

- [ ] Focused engine/commandrunner/app/Zoom/conformance/certify tests, vet, and lint pass.
- [ ] Fresh binary passes base/group/command help and every isolated SCIM2 lifecycle.
- [ ] Isolated fixtures prove declared root origin/auth, exact method/path/body/status, confirmation,
  no paging input, and redaction.
- [ ] Surface sync/reconciliation/validation, endpoint-ledger scope, docs/website validation, and
  CLI golden checks pass.
