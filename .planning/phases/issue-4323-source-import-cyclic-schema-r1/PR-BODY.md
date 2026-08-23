## Intent

Refs #4323

Retain ordinary cyclic OpenAPI schema references as source-bound missing-foundation
evidence instead of aborting source import during grammar preflight.

## What changed

- Preserve the canonical source `$ref` at a recursive schema edge; no cycle is
  flattened, truncated, or expanded to an arbitrary depth.
- Emit `cli-recursive-schema-foundation-r1` gaps that name both the schema JSON
  Pointer and the response/request location, leaving affected operations present
  but `merge_blocked`.
- Use the existing top-level descriptor gap aggregation when a recursive schema
  is otherwise unreferenced.
- Cover direct, mutual, deeply nested, canonical-pointer, unused-schema, and
  non-cyclic cases through the real import path.

## TDD and GSD lifecycle

- Red: the focused recursive-schema test failed before production edits with
  `preflight source grammar ... reference cycle at "#/components/schemas/Folder"`.
- Green: the same importer-path suite now retains the operation/source mapping
  and raw `$ref`, records the explicit missing-foundation gap, and keeps the
  finite control gap-free.
- Lifecycle commands were resolved through `scripts/gsd doctor`, `scripts/gsd
  sources`, and `scripts/gsd prompt` for `discuss-phase`, `plan-phase --tdd`,
  `execute-phase`, `verify-work`, and `code-review`. The generated prompts were
  completed inline because this direct-PR worker has no compatible isolated GSD
  role runtime; plan, TDD, verification, summary, and review evidence are in
  `.planning/phases/issue-4323-source-import-cyclic-schema-r1/`.

## Testing

- `go test -count=1 -v -timeout 20m ./cmd/connectorgen -run '^TestSourceImport(RetainsRecursiveSchemaReferencesAsSourceBoundGaps|PreflightsUnusedGrammarObjects|NormalizesDirectReferenceFragments)$'`
- Pinned public Grafana import: 314 operations retained and 52 explicit
  recursive-schema gaps, including `#/components/schemas/Folder`.
- `go vet ./...`
- `go build ./cmd/pm`
- `make verify` (full repository suite; lint and generated/snapshot checks included)

## Frozen GitHub artifacts

- Source lock: `3,420,025` bytes,
  `281b1cfcc67eb63e19ef83daf06197bf3d3b23db0b6bc9b73e02fc18ee278fb6`
- Descriptor: `43,354,021` bytes,
  `d1978c0c6fd0eb66e9fcd4d78d637864a6e486f558aaad1e51550bc43758b899`
- `internal/connectors/defs/github/rate_limits.json` is unchanged.

## Safety and delivery record

- No credentials, tokens, runtime state, dependencies, generic write surface,
  or generated connector artifacts were added.
- Commit checkpoints: planning/TDD `c49e6f8de`; red test `8e7be4565`; this
  implementation/review checkpoint follows full local verification.
- Direct-PR mode: no no-mistakes pipeline was run, as instructed.

## Automated review

- Primary route: `claude_auto`, pending the non-draft PR-open trigger for this
  trusted author and the final implementation commit.
- Fallback: none. No manual Claude or Copilot request was sent, per review
  routing policy.
- Dispositions at open: no automated findings yet; inline review found no
  unresolved actionable findings.
