---
status: clean
phase: "3863"
depth: standard
files_reviewed: 14
findings:
  critical: 0
  warning: 0
  info: 0
  total: 0
reviewer: inline_manual
---

# Code review — issue #3863

The configured review workflow defaults to standard depth. Its normal reviewer
agent was not spawned because the canonical single-worker contract and current
Codex runtime prohibit subagent delegation; this report records the required
inline/manual fallback.

## Scope

- Typed identity implementation and runtime handoff.
- Protected credential-state migration, creation-time linking, and explicit
  existing-credential linking.
- Credential CLI behavior, help/manual/golden parity, and website reference
  generation.

## Review checks

- `CoordinationIdentity` accepts only the explicit non-secret binding and
  declaration metadata. It retains opaque HMAC projections only, uses separate
  domains for authentication and rate scopes, and returns distinct Go key types
  so an authentication cohort cannot be used as a rate budget by accident.
- App construction derives the identity before the vault read; legacy migration
  creates an isolated binding without a vault read; binding preimages remain in
  protected state rather than `CredentialMeta` output.
- Linking requires exact declared provider-family/auth-profile compatibility;
  cross-connector creation-time linking requires explicit declarations; an
  unlinked credential receives a fresh binding.
- New declaration and scope validation errors identify only a field/constraint.
  No new error, JSON, log, or runtime field exposes a binding preimage,
  credential revision, or raw rate scope subject.
- The change does not touch command-runner, connector schemas, requester
  dispatch, rate registries, fencing, parking, transport behavior, or provider
  I/O. No capability declaration was added.
- CLI help, generated manual, website reference/data, bare namespace behavior,
  and golden transcripts match the new `credentials link` surface.

## Findings

None. The review found no in-scope correctness, security, or maintainability
finding requiring a production change.

## Evidence

- Focused identity/app/CLI suites pass after the review.
- Package suites, build, scoped vet, docs check, website typecheck, lint,
  contract validation, connector validation/surface sync/boundary, and release
  workflow checks are recorded in `VERIFICATION.md` once final verification is
  complete.
