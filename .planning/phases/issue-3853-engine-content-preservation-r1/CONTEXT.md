# CONTEXT — issue #3853 engine content preservation

## Phase mapping

Issue #3853 is one shared-engine foundation with no GitHub sub-issues. It closes the final
engine-owned content-stripping paths after #3739, #3743, #3771, and #3852. It is not a connector
migration: no provider bundle, API-surface declaration, capability, availability claim, schema, or
generated connector index changes are in scope.

## Locked decisions

- Write previews must resolve the same complete method, URL, path parameters, and declared
  `redact_fields` values that the executable request uses. Existing declarations remain
  load-compatible; the runtime ignores them for preview replacement rather than rewriting bundles.
- The preview also retains configuration-secret substitutions in the resolved URL. Secret storage
  remains encrypted at rest; this change only stops operator-visible runtime content stripping.
- Direct-read, operation-direct-read, and binary-download failure messages preserve the captured
  HTTP URL/query/body text. The existing bounded transport capture and error-map class/hint behavior
  remain unchanged.
- This issue does not change successful direct-read response policies, binary result records,
  write-action error redaction, #3771-owned command-runner functions, or #3852's policy enum.
- The reverse plan → preview → approval → execute lifecycle, preview digest, typed destructive
  confirmation, single-use approval evidence, non-idempotent retry policy, bounds, endpoint
  preflight, and redirect protections remain intact.
- CLI/manual/website language may accurately promise complete connector-engine request, preview,
  response, and error content while preserving the special handling of authorization tokens. The
  separate generic source-table plan-sample behavior remains explicitly scoped as an app-layer
  follow-up; this foundation must not claim to change it.

## Scope and ownership guard

- Owned production code:
  - `internal/connectors/engine/write.go` preview-only resolution at the #3853-cited region.
  - `internal/connectors/engine/direct_read.go` direct-read and operation-direct-read error
    rendering.
  - `internal/connectors/engine/binary_read.go` binary-download error rendering.
- Owned focused tests:
  - `internal/connectors/engine/write_test.go`
  - `internal/connectors/engine/direct_read_test.go`
  - `internal/connectors/engine/binary_read_test.go`
- Owned operator surfaces:
  - `internal/cli/docs.go`
  - `docs/cli/reverse.md`
  - `internal/cli/testdata/golden_transcripts.json`
  - `website/content/docs/reverse-etl.mdx`
- Do not edit `internal/connectors/commandrunner/**`, #3852 enum/schema policy code, connector
  bundles, `internal/app` generic source-table masking, or `internal/connectors/connsdk`.

## Evidence inspected

- `internal/connectors/engine/write.go:109-175` replaces secrets and declared action fields before
  resolving preview lines.
- `internal/connectors/engine/direct_read.go:109-118` and `179-188` call
  `safety.RedactErrorText`; the operation-direct-read path has the same behavior.
- `internal/connectors/engine/binary_read.go:160-169` calls `safety.RedactErrorText`.
- `internal/connectors/connsdk.HTTPError` exposes bounded raw `URL` and `Body` fields, enabling the
  engine wrapper to preserve captured content without changing the shared transport layer.
- Existing write tests at `write_test.go:830`, `935`, and `965` assert the old redacted preview
  behavior and will be reversed, not deleted.

## GSD execution note

`scripts/gsd doctor` passed; every required command resolved through `scripts/gsd sources`; and
`go run ./cmd/agentcontractgen check` passed. The generated `discuss-phase` and `plan-phase --tdd`
prompts are executed inline because this task's single-worker contract forbids spawning GSD roles.
The issue itself resolves all product choices, so discussion records those decisions rather than
reopening them.

## CLI help/docs/website parity disposition

This changes the factual contract of the existing `pm reverse` help/manual/website documentation.
The plan therefore covers runtime help (`pm reverse`, `pm help reverse`, `pm reverse --help`), its
checked-in manual, golden transcripts, and the reverse-ETL website page. No command, flag,
completion, namespace behavior, or documentation dependency changes are expected.
