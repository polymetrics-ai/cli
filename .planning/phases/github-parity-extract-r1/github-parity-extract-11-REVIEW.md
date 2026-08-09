# Review — plan 11 generated fixed GraphQL contracts

**Mode:** inline/manual GSD code-review fallback. The canonical parent lane prohibits spawning a
review role; the generated `code-review --depth=standard` prompt was followed against this scoped
diff.

## Scope and safety

- PASS — the only new transport is the fixed, declaration-owned `POST /graphql` row. It carries an
  exact operation-ID list rather than granting a generic GraphQL request path.
- PASS — query variables are closed/bounded schemas; JSON flags can name only a declared top-level
  object/array variable. No raw document, selection, endpoint, header, cursor, or generic HTTP
  write interface was added.
- PASS — every generated mutation is routed through the existing shared write lifecycle; the
  recorded `deleteIssue` provider/product decision remains a terminal unavailable command.
- PASS — local source drift detection is hermetic; the scheduled workflow is read-only and has
  `contents: read` only.
- PASS — no private lab data is included or inspected; no provider request was made.

## Findings found and resolved during review

1. **Raw GraphQL HTTP errors bypassed GraphQL redaction.** Both direct-query and direct-mutation
   non-2xx paths could include a provider body outside the bounded `errors[]` sanitizer. Fixed by
   applying the existing safety redactor and pinning two loopback regressions.
2. **A prefix-only query check permitted an appended mutation selected by `operationName`.** Fixed
   with a comment/string-aware fixed-document scanner that requires exactly one named operation of
   the declared kind. The regression starts with a query and appends a mutation; preflight now
   rejects it before configuration or network setup.
3. **The static generator did not count a GraphQL query as a read because its transport is POST.**
   Fixed the capability accounting to match conformance and added a red/green test for
   `capabilities.read=false`.

## Result

No Critical, Warning, or Info findings remain in this checkpoint after the focused regressions and
package/generator/lab boundary suites passed. The remaining work is current-head PM-only live proof,
not an unresolved implementation review finding.
