---
coverage:
  - id: D1
    description: Declared v3 identity query retrieves its locked document.
    verification:
      - kind: integration
        ref: TestSourceImportVersion3FetchesDeclaredIdentityQuery
        status: pass
    human_judgment: false
  - id: D2
    description: Capture query remains provenance-only and default projection is stable.
    verification:
      - kind: integration
        ref: TestSourceImportVersion3LeavesCaptureQueryAsProvenanceOnly; TestSourceImportVersion3AbsentOrFalseIdentityQueryProjectsIdentically
        status: pass
    human_judgment: false
  - id: D3
    description: Query and SSRF guards reject unsafe identity artifacts.
    verification:
      - kind: integration
        ref: TestSourceImportVersion3RejectsUnsafeIdentityArtifactQueries; TestSourceImportIdentityQueryRetainsArtifactURLGuards
        status: pass
    human_judgment: false
---

# Summary — source-import identity-bearing artifact query

Implemented `artifact.identity_query:true` for v3 REST source documents only.
The immutable source-lock URL supplies the exact fixed query; capture citation
queries remain provenance-only. The cache now preserves the declared artifact
through to the HTTP fetcher, whose default remains no-query for all other
callers.

No connector-specific path, credential input, generic HTTP capability, or
runtime query argument was added. The source-lock convention now documents the
declaration and its retained URL/DNS security limits.

## GSD lifecycle

The generated `discuss-phase`, `plan-phase --tdd`, `execute-phase`,
`verify-work`, and `code-review` prompts were resolved and executed inline.
This repository task is unnumbered in the roadmap and compatible isolated
Pi-role execution was unavailable; the TDD ledger, verification checklist, and
review record supply the manual fallback evidence.
