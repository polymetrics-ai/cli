# Verification — Descriptor-Free Retention Admission

## Required proof

- [x] Jira’s 617 retained operations reconcile through exact source IDs.
- [x] Sentry’s 223 retained operations reconcile while a provider ID containing
  an ordinary space remains valid.
- [x] Vercel’s 400 retained operations reconcile through exact source IDs.
- [x] The source-lock digest/byte identity is checked through the existing
  enabled-contract bridge.
- [x] Missing descriptor is waived only for a source-only, nonimplemented
  `retention_only` contract whose actual loaded bundle has no operations,
  writes, streams, selected sync transport, or implemented CLI command.
- [x] Implemented, malformed, empty/control-character-ID, incomplete, or duplicate-ID contracts continue to
  fail without the canonical descriptor.
- [x] No source lock, matrix, connector artifact, runtime, transport,
  certification, or credential change appears in the diff.

## Commands

The final ledger will record exact outputs for focused normal/race tests,
affected package tests, `go vet`, JSON validation, `agentcontractgen check`,
and `git diff --check`.
