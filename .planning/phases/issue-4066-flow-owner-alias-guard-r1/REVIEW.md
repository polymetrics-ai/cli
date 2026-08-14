# #4066 Manual Code Review

**Status:** Pass — no unresolved finding

## Review boundary

The inline/manual code-review fallback covers the #4066 query-origin policy,
its call sites, and its real-Parquet regression suite. It is paired with the
no-mistakes read-only review cycle because the issue is not a numbered roadmap
phase and the parent contract forbids role spawning.

## Dispositioned findings

1. `flow-omitted-selector-alias-bypass` — fixed by carrying a flow-only origin
   from the production flow adapter and suppressing generated owner aliases for
   an unscoped flow.
2. `generated-alias-real-table-collision` — fixed from the immutable resolver
   snapshot before bare-view registration. Generic SQL retains the legitimate
   real table; the unscoped flow resolves through the typed base ambiguity.
3. `case-insensitive-generated-alias-collision` — fixed by applying
   ASCII-only DuckDB identifier keys to policy maps while retaining original
   catalog names for registration and resolver errors.
4. `casefolded-bare-table-ambiguity-bypass` — fixed with a flow-only canonical
   bare-name policy that routes case-equivalent replacement scans through the
   original ambiguous resolver name.

## Review checks

- Policy is constructed once from the resolver snapshot already used by the
  DuckDB query; it does not parse, filter, or rewrite SQL text.
- `QuerySQLOriginFlow` is set only by the flow adapter. Generic CLI and extract
  callers retain the zero-value generic origin.
- The policy is empty for an explicit connection, preserving selected-source
  query behavior.
- Generated aliases remain available to generic SQL when non-colliding; real
  user tables take precedence on collisions without table renaming or a new
  reserved namespace.
- Focused race coverage validates quoted/unquoted aliases, ASCII case variants,
  bare-name collisions, generic real-table access, and existing flow selectors.

## Non-actionable observations

- `lint-preexisting-unused-planerr` is an existing unused test-only field in
  `internal/flow/action_test.go`, outside this correction's changed paths. It
  was recorded as no-op and not changed.
- `target-scope-drift` correctly observed that #4066 is behavioral while #4063
  is metadata-only. It was approved because #4066 is the separately
  user-authorized final 5/5 correction on the same existing stacked PR.
