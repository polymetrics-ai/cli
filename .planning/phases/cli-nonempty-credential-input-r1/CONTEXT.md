# Context — non-empty credential input foundation

## Decision record

The shared credential contract is provider-neutral. It applies to every
persisted credential field, regardless of connector, and to the declarative
engine's selected required authentication route. It deliberately does not
modify `internal/connectors/defs/twenty/` or add a Twenty-specific branch.

- Standard-input transport normalization removes at most one documented
  terminal delimiter: `\n` or the preceding `\r\n` sequence.
- All other bytes, including leading/trailing whitespace and additional
  newlines, are preserved through encrypted storage and retrieval.
- A value empty after that transport-only normalization is rejected before
  `vault.Put`, with a typed, actionable, non-secret error.
- App persistence repeats the non-empty validation so non-CLI callers cannot
  persist an empty secret field.
- When a declared auth route is selected, blank required bearer, basic,
  API-key header/query, and OAuth credential material fails before a request
  can emit an empty credential form. A connector-declared optional route that
  does not match remains a supported no-auth case.

## Scope and safety fence

The implementation uses only generated synthetic canaries. Tests report only
byte lengths and SHA-256 fingerprints; neither production code nor planning
evidence logs, serializes, or prints credential values. Long inputs remain
in-memory stdin-to-vault data and never move through argv, temporary plaintext
files, Keychain assumptions, or process listings.

## Inline lifecycle fallback

The canonical GSD worker contract forbids role spawning in this delivery
context. The generated `discuss-phase`, `plan-phase --tdd`, `execute-phase`,
`verify-work`, and `code-review` prompts are executed inline in this isolated
worktree. Their evidence is recorded in this phase directory.
