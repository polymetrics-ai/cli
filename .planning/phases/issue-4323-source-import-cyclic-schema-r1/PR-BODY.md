## Intent

Refs #4323
Refs #4326

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
- Retain the closed OpenAPI 3.0 reference-sibling set: `description`,
  `summary`, schema `readOnly`, and schema `type`. A non-equivalent `type`
  remains exact source data but emits
  `cli-openapi30-reference-sibling-foundation-r1` with its canonical target
  pointer and keeps the operation merge-blocked; other semantic siblings (for
  example response `content`) still reject.
- Preserve preflight-only type-sibling evidence through the existing top-level
  descriptor gap aggregation.

## TDD and GSD lifecycle

- Red: the focused recursive-schema test failed before production edits with
  `preflight source grammar ... reference cycle at "#/components/schemas/Folder"`.
- Green: the same importer-path suite now retains the operation/source mapping
  and raw `$ref`, records the explicit missing-foundation gap, and keeps the
  finite control gap-free.
- Red/green for #4326: OpenAPI 3.0 response `description`/`summary`, schema
  `readOnly`, and non-equivalent `type` initially failed grammar preflight or
  emitted no unused-schema evidence. The bounded sibling suite now preserves
  the source declaration, pointer-named type gap, and semantic-sibling
  rejection.
- Lifecycle commands were resolved through `scripts/gsd doctor`, `scripts/gsd
  sources`, and `scripts/gsd prompt` for `discuss-phase`, `plan-phase --tdd`,
  `execute-phase`, `verify-work`, and `code-review`. The generated prompts were
  completed inline because this direct-PR worker has no compatible isolated GSD
  role runtime; plan, TDD, verification, summary, and review evidence are in
  `.planning/phases/issue-4323-source-import-cyclic-schema-r1/`.

## Testing

- `go test -count=1 -v -timeout 20m ./cmd/connectorgen -run '^TestSourceImport(RetainsRecursiveSchemaReferencesAsSourceBoundGaps|PreflightsUnusedGrammarObjects|NormalizesDirectReferenceFragments)$'`
- `go test -count=1 -v -timeout 20m ./cmd/connectorgen -run '^TestSourceImport(RetainsBoundedOpenAPI30ReferenceSiblings|PreflightsUnusedGrammarObjects|RetainsRecursiveSchemaReferencesAsSourceBoundGaps)$'`
- Pinned public Grafana import: 314 operations retained and 52 explicit
  recursive-schema gaps, including `#/components/schemas/Folder`.
- Pinned public Asana import: 249 locked operations retained and `--check`
  verified; its two non-equivalent type overlays are explicit source-bound
  gaps. GitLab live source drift and Docker Hub's separate dangling schema
  reference are recorded in the GSD ledger without source mutation.
- `go vet ./...`
- `go build ./cmd/pm`
- `make verify` (full repository suite; lint and generated/snapshot checks included)
- Final low-load `make verify` exit 0: `cmd/connectorgen` 189.173s,
  `internal/cli` 528.045s, connector-boundary 406.747s; the shared 20-minute
  package budget was not changed.

## Frozen GitHub artifacts

- Source lock: `3,420,025` bytes,
  `281b1cfcc67eb63e19ef83daf06197bf3d3b23db0b6bc9b73e02fc18ee278fb6`
- Descriptor: `43,354,021` bytes,
  `d1978c0c6fd0eb66e9fcd4d78d637864a6e486f558aaad1e51550bc43758b899`
- `internal/connectors/defs/github/rate_limits.json` is unchanged.
- `operation-evidence --check` passes and neither operation-evidence generated
  artifact differs from `origin/main`.

## Safety and delivery record

- No credentials, tokens, runtime state, dependencies, generic write surface,
  or generated connector artifacts were added.
- Commit checkpoints: planning/TDD `c49e6f8de`; cyclic red test `8e7be4565`;
  bounded sibling tests `cd537d329`; implementation `030e48a9c`; unreferenced
  gap handling `374ab7828`; this evidence checkpoint follows final local
  verification.
- Direct-PR mode: no no-mistakes pipeline was run, as instructed.

## Automated review

- Primary route: `claude_auto`, pending the non-draft PR-open trigger for this
  trusted author and the final implementation commit.
- Fallback: none. No manual Claude or Copilot request was sent, per review
  routing policy.
- Dispositions at open: no automated findings yet; inline review found no
  unresolved actionable findings.
