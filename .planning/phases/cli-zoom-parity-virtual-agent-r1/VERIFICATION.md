# Verification — Zoom Virtual Agent documented-operation parity, R1

## Planned checks

- [ ] Test-only RED captured before production declaration changes.
- [ ] All thirteen command paths pass real commandrunner preflight.
- [ ] Nine direct reads run against isolated fixtures with exact fixed method/path/auth, bounded
  response handling, redaction, and no invented query/paging input.
- [ ] Four direct writes run through plan/preview/approval/execute fixtures with exact typed bodies;
  create-sync sends no body and delete is status-only with destructive confirmation.
- [ ] Endpoint ledger reconciliation is confined to `provider_module=virtual-agent`.
- [ ] Generated CLI docs and website catalog retain Zoom-only output after whole-file generation.
- [ ] Fresh `pm` binary reaches base, namespace, provider group, and every command help route.
- [ ] Scoped local gates and code review complete.
