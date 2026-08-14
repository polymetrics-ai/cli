# Verification — issue #3754

Status: planned.

- [ ] Local scope registry applies a real shared-in-process budget and output labels it `process-local`.
- [ ] `require_shared` engages the external registry when available.
- [ ] `require_shared` returns a typed no-fallback reason when unavailable.
- [ ] Opaque coordination identity is absent from argv, environment, files, logs, receipts, and evidence.
- [ ] Two separate processes under one opaque scope obey the external budget.
- [ ] Context cancellation, atomicy, server-time TTL/reset, and supported declared models are covered.
- [ ] No connector-specific literal/branch, no production bundle edit, no parking/resumption, and no generic execution surface.
- [ ] Focused package tests, race test, targeted vet/build, formatting, docs parity if applicable, and individual repository gates pass.
