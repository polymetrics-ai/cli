# Verification Checklist — Zoom Chatbot parity, R1

## Lifecycle

- [x] GSD command provenance was resolved with `scripts/gsd sources`.
- [x] Required skills and canonical connector/CLI references are recorded in `PLAN.md`.
- [x] Live provider artifact URL, retrieval date, HTTP result, byte count, SHA-256, and exact four-operation audit are recorded before RED.
- [ ] RED failure is captured, committed, and pushed before production changes.
- [ ] Reusable foundations are red/green tested, separately committed, and pushed.
- [ ] Connector declaration and generated output are committed and pushed.
- [ ] Inline verify-work and code-review evidence are complete.

## Source parity

- [ ] All four Chatbot ledger rows are executable direct writes with exactly one disposition.
- [ ] Zero Zoom rows are `unsafe_or_disallowed`.
- [ ] Each command accepts only documented typed path/body members; no paging flags are invented.
- [ ] Client-credentials Basic exchange and API Bearer application are declared and tested without raw secret output.
- [ ] Link Unfurls treats HTTP `204 No Content` as a successful action.

## Runtime/docs checks

- [ ] Focused engine/connsdk/commandrunner/app/Zoom/conformance/certify tests, vet, and lint pass.
- [ ] A fresh binary passes base/group/command help and every isolated Chatbot plan lifecycle.
- [ ] The isolated fixture proves token request Basic auth, action request Bearer auth, exact method/path/body/status, and redaction.
- [ ] Surface sync/reconciliation/validation, endpoint-ledger scope, docs/website validation, and CLI golden checks pass.
