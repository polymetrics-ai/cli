# Verification — Zoom Auto Dialer documented-operation parity, R1

## Planned checks

- [ ] Test-only RED captured before production declaration changes.
- [ ] All sixteen command paths pass real commandrunner preflight.
- [ ] Eight direct reads run against isolated fixtures with exact fixed method/path/auth, bounded
  response handling, redaction, and no invented query/paging input.
- [ ] Eight direct writes run through plan/preview/approval/execute fixtures with exact typed bodies;
  every DELETE/204 action is status-only with destructive confirmation where documented.
- [ ] Endpoint ledger reconciliation is confined to `provider_module=auto-dialer`.
- [ ] Generated CLI docs and website catalog retain Zoom-only output after whole-file generation.
- [ ] Fresh `pm` binary reaches base, namespace, provider group, and every command help route.
- [ ] Scoped local gates and code review complete.
