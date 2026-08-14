# #4089 — verification checklist

## Lifecycle

- [x] Issue #4089 and parent #3988 read.
- [x] GSD adapter doctor, command-source resolution, prompt generation, and agent contract check passed.
- [x] Required Go CLI, security, testing, documentation, and website skills loaded; inline/manual GSD fallback recorded.
- [x] Existing bounded stdin carrier and both current argv call sites verified.

## Before completion

- [ ] Red test recorded before production edit.
- [ ] Both command paths require a bare stdin marker and use the shared carrier.
- [ ] Empty, oversized, malformed, retired-argv, and replay inputs reject before a write-side effect.
- [ ] Six independent secret-surface checks pass: argv, environment, files, logs, receipts, and evidence.
- [ ] Plan → preview → run and exact-once replay behavior remain proven.
- [ ] Runtime help, CLI manual, generated skills, website source/data, and CLI transcript are updated.
- [ ] Repository stale-syntax scan finds no retired argv approval examples in active source, docs, website, generated data, or tests.
- [ ] Targeted package tests, vet, docs, connector boundary, and split local gates are green.
- [ ] Inline `verify-work` and `code-review` records contain no unresolved finding.
