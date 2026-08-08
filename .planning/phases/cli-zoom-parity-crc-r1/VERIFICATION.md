# Verification — Zoom CRC documented-operation parity, R1

## Status

Planned. This file will record the `verify-work` evidence after implementation.

## Required checks

- [ ] Fresh artifact URL, timestamp, HTTP status, bytes, hash, and source/ledger delta recorded.
- [ ] RED test committed and its failure recorded before production changes.
- [ ] All 20 documented source rows covered; no `unsafe_or_disallowed` row introduced.
- [ ] All 20 commands reached by a freshly built `pm` binary.
- [ ] Status-only and destructive actions exercised through the real lifecycle with fixtures.
- [ ] Secret-returning private-key routes output only redacted values.
- [ ] `surface-sync`, scoped reconciliation, docs/site output, and non-Zoom locality checks pass.
- [ ] Scoped local gates and manual code review completed.
