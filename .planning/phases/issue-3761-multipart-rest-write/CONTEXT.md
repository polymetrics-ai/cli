# CONTEXT — issue #3761 typed multipart `rest_write`

## Phase mapping

Parent #3761 is delivered as one foundation branch covering its dependency-ordered
sub-issues: #3763 (closed contract), #3768 (preview binding), #3772 (bounded
dispatch), #3774 (author documentation), and #3777 (executable-claim
enforcement). No provider definition is adopted in this phase.

## Locked decisions

- Extend the existing declared `rest_write` path from #3739. Do not add an
  upload runner, raw request input, generic HTTP write, or a parallel approval
  lifecycle.
- A new opt-in `rest.multipart` declaration reuses `MultipartSpec` and its
  field/file part model. It is executable only when its exact, closed contract
  loads successfully; legacy `file_upload` rows remain non-executable.
- The method, connector-relative path, response capture cap, output policy,
  typed body schema, field names, file names, media-type allow-list, per-file
  caps, and aggregate cap are declaration-owned. Commands can only materialize
  their declared `body.*` fields.
- Multipart preview records canonical fields and source-path identities plus an
  approved SHA-256 for each file. Execution re-prepares through the current
  plan → preview → approval → execute path and lets the existing requester
  snapshot, digest, regular-file, root-containment, media-type, and byte-limit
  checks fail before network dispatch.
- Every direct write remains single-attempt (`DisableRetries=true`), including
  no 401 refresh and no redirect replay. Provider response and error content
  remain complete; this phase adds no redaction declaration or masking path.
- `rest.max_bytes` remains the bounded response-capture declaration. The new
  `multipart.max_bytes` bounds uploaded file bytes and each file part has its
  own positive cap.

## Scope fences

- No `internal/connectors/defs/**` adoption, generated artifact, provider
  request, credential, or live write.
- No changes to the `#3771`, `#3775`, or `#3769` owned functions in
  `internal/connectors/commandrunner/runner.go`; #3777 only adds a
  preflight/integration regression when the real runtime gap is demonstrated.
- Do not absorb active rate-limit work in `connsdk/http.go` or
  `engine/direct_write.go`. This branch starts from main and must be rebased on
  main before delivery; sibling foundation branches are not bases.
- CLI/manual/website parity is intentionally not applicable: no connector
  command becomes available in this shared foundation. A connector adoption
  lane owns its own runtime help, `docs/cli/**`, website, generated manual, and
  command evidence.

## GSD execution note

The generated `discuss-phase` and `plan-phase --tdd --skip-research` prompts
were executed inline. The issue and its child contracts already decide all
product and safety choices, and this worker is explicitly prohibited from
spawning roles. This is the repo-local GSD adapter's documented inline/manual
fallback; it does not relax TDD, verification, or review requirements.
