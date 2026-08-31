# GitLab Unicode `origin` schema foundation R1

## Task Delivery Header

- Issue: Captain-approved runtime foundation #2; no GitHub issue number was supplied for this local-only candidate.
- Base branch: `fm/cli-top100-declaration-batch-r1@c9ae575a734514b728a5e6add7ff8b0e55233437`.
- Merges into: Batch R1 integration branch → `main`; this candidate does not open or push a pull request.
- Delivery: A local commit on the stated candidate branch after focused green verification and an independent review; no push.
- Working branch: `codex/4283-gitlab-unicode-origin-r1`.
- Task: Add one closed typed Unicode-scalar string policy for the exact GitLab AI-detection `origin` pattern, retain source numeric `confidence_score` bounds, and remove only the named schema-compilation blocker. Preserve the immutable upstream descriptor identity and declare the resulting lock-backed descriptor correction explicitly. The existing source-bound mutation must remain non-executable.
- Verification: Red/green focused Go tests; `go test -race`; `go vet`; connector JSON validation; Foundation Atlas JSON/uniqueness validation; agent-contract check; and `git diff --check`.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Exact Unicode-scalar `origin` semantics are compiled without a generic regex fallback | live | An engine test accepts ASCII, CR, astral Unicode and exactly 255 logical characters, while rejecting LF, invalid UTF-8, surrogate-byte representation, and 256 characters. |
| `confidence_score` retains source range `0..100` | live | The retained GitLab descriptor, materialized write schema, and both existing partial CLI intents agree on `0..100`; the compiled write schema accepts `0` and `100`, and rejects `-1` and `101`. |
| Projection clears only the exact pattern-compiler blocker | live | The source-import and GitLab contract tests prove the `origin` engine-gap disappears while the dynamic-root and source-bound non-executable mutation blockers remain. |
| Derived descriptor provenance stays truthful | live | The matrix retains the immutable `git_archive_byte_identical` base identity separately from the corrected descriptor identity, and generic validation verifies the source-lock operation, pointer, source schema, removed stale gaps, and retained blockers. |
| No generic runtime admission or connector-source rewrite occurs | live | Diff scope and lane checks show no changes to runtime source-bound mutation admission, source locks, lane disposition/counts, credentials, transport, approval, or webhook functionality. |
| Atlas stays accurate and authoring-only | live | Catalog validation and an exact Atlas-entry assertion show the policy's closed scope and explicit non-goals. |

## Discovery and design

- Foundation Atlas classification: constrained extension of `runtime.json-schema-surrogate-regex.v1`.
- Existing owner: `internal/connectors/engine/schema.go:compileNode`; the current use of Go's RE2 compiler rejects the retained JSON-Schema surrogate-pair expression.
- Approved closed contract: only the exact retained source expression `^(?:[\\uD800-\\uDBFF][\\uDC00-\\uDFFF]|[^\\n\\uD800-\\uDFFF]){1,255}$` becomes a typed policy. It validates UTF-8 scalar strings, rejects LF, counts logical scalar values (not bytes), allows CR, and enforces `1..255`.
- Deliberate non-goals: no generic JSON-Schema regex parser; no pattern rewrite; no connector-name branch; no source-lock change; no lane disposition/count change; no source-bound mutation admission/execution change; no credentials, transport, approval, or webhook change.
- Provenance correction: the GitLab source-lane matrix now declares the original immutable descriptor blob as its base snapshot identity and the corrected descriptor blob as its derived retained identity. `git_archive_byte_identical` is reserved for the upstream base; the derived bytes are never represented as raw archive bytes.
- Source-backed numeric field: the existing GitLab descriptor's `confidence_score` facts are retained as declared integer minimum `0` and maximum `100` in the affected materialized write schema and two existing partial CLI intent flags. This does not introduce a generic cross-connector source-projection rule.

## GSD / TDD execution record

- GSD adapter: `scripts/gsd doctor` and `go run ./cmd/agentcontractgen check` passed before implementation.
- Resolved workflow prompts: `discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and `code-review` through `scripts/gsd prompt`.
- Manual fallback: this non-interactive Codex worktree has no compatible isolated Pi GSD worker; the canonical workflow also forbids role spawning. The lifecycle is recorded inline here with test evidence.
- Loaded skills: `connector-lane-build-order`, `go-engineering`, `golang-how-to`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`, and `golang-testing`.

### Plan

1. Add red tests for the exact Unicode policy and source projection behavior.
2. Add the smallest closed engine policy and preserve source numeric bounds in projection.
3. Update only the affected GitLab descriptor/write artifact, correction provenance, matrix retained identity, and Atlas entry; retain all broader blockers.
4. Add generic provenance validation and red/green tests for missing/invalid provenance and incorrect source-operation/pointer bindings.
5. Run focused and broader static verification, record results, then locally commit the green slice.

### Red

- `go test ./internal/connectors/engine -run '^TestSchemaValidatesClosedUnicodeScalarNoLineFeedPolicy$' -count=1 -v` failed as intended before the engine change: Go RE2 rejected the retained JSON-Schema `\u` escape with `invalid escape sequence: \\u`.
- `go test ./cmd/connectorgen -run '^(TestSourceRequestGapsAdmitsClosedUnicodeScalarP5PatternButRetainsDynamicRootGap|TestGitLabUnicodeScalarSourceRowClearsOnlyExactSchemaGap)$' -count=1 -v` failed as intended before the descriptor/artifact changes: the `origin` compiler gap remained and the affected write/CLI artifacts lacked the retained numeric bounds.

### Green

- The engine-focused policy test passed after the closed equality-selected policy was added.
- The connectorgen-focused tests passed after the affected GitLab write/CLI artifacts retained the numeric bounds, the exact `origin` gap was removed, and the unrelated dynamic-root and mutation blockers remained.
- The correction-provenance test failed before its declaration existed, then passed after the connector-owned sidecar, matrix identity split, and generic validator were added. Synthetic negative cases reject missing/invalid provenance and wrong operation/pointer bindings; the live GitLab case binds the frozen lock to the exact descriptor correction.
- Final scoped commands and static checks are recorded below before commit.

### Refactor / review

- The implementation intentionally recognizes one literal source pattern by equality rather than parsing a family of regular expressions. A near pattern with `{1,254}` remains rejected by the ordinary engine compiler.
- A temporary generic source-projection numeric-bound change was removed during review because it would create newly enforced differences in unrelated GitLab rows. The final slice preserves only the exact source-backed GitLab field in its existing materialized artifacts; it adds no cross-connector admission or mapping rule.
- CLI-help/manual parity is not applicable: no command, command name, flag name, availability, help topic, output format, or executable command behavior changed. Existing partial GitLab command metadata gains the source's bounds while its independently cited non-executable mutation blocker remains.

## Verification results

| Check | Result |
| --- | --- |
| `go test ./internal/connectors/engine -run '^(TestSchemaValidatesClosedUnicodeScalarNoLineFeedPolicy|TestSchemaDoesNotGeneralizeUnicodeScalarNoLineFeedPolicy)$' -count=1 -v` | PASS — ASCII, CR, astral boundary, scalar count, LF, invalid UTF-8/surrogate bytes, bounds, and near-pattern non-admission. |
| `go test ./cmd/connectorgen -run '^(TestSourceRequestGapsAdmitsClosedUnicodeScalarP5PatternButRetainsDynamicRootGap|TestGitLabUnicodeScalarSourceRowClearsOnlyExactSchemaGap)$' -count=1 -v` | PASS — exact field gap clears; dynamic-root and source-bound mutation gaps remain; descriptor/write/CLI `0..100` agree. |
| `go test ./cmd/connectorgen -run '^TestValidateSourceDescriptorCorrectionProvenance$' -count=1 -v` | PASS — validates the live GitLab frozen-lock correction and rejects missing/invalid provenance plus wrong operation/pointer fixture evidence. |
| `go test ./internal/connectors/defs/gitlab -run '^(TestGitLabSourceLaneMatrixRetainsEveryLockedOperationAndLane|TestGitLabDescriptorCorrectionProvenanceIsDeclared)$' -count=1 -v` | PASS — base/derived descriptor identity split is bound to the sidecar, all uncorrected retained identities remain stable, and all 1,754 source rows / 12,278 lane cells retain their existing semantics. |
| matching two `go test -race` commands | PASS — focused engine and connectorgen coverage. |
| `go vet ./internal/connectors/engine ./cmd/connectorgen` | PASS. |
| `go run ./cmd/connectorgen validate internal/connectors/defs/gitlab --json` | PASS — one connector, no findings or warnings. |
| `jq` parse/unique-ID/exact-Atlas-entry checks | PASS. |
| `go run ./cmd/agentcontractgen check` | PASS — canonical contract and registered projections are current. |
| `git diff --check` and manual scope review | PASS — bounded engine schema support, generic authoring-only provenance validation, GitLab descriptor/artifact evidence, Atlas, focused tests, and this trace; no source-lock, lane disposition/count, transport, credential, approval, webhook, or source-bound mutation-admission change. |

No credentialed command, write dispatch, reverse-ETL execution, webhook path, or broad repository build is run by this task. Disk was 30 GiB before targeted race tests and 28 GiB after all scoped checks.
