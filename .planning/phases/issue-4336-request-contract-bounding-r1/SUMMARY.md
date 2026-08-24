---
coverage:
  - id: D1
    description: Valid source-silent common bounds import without falsifying provider schema.
    verification:
      - kind: integration
        ref: "cmd/connectorgen/sourceimport_test.go::TestSourceImportVersion3RepresentsGongWorkspaceQueryWithPMExecutionEnvelope"
        status: pass
    human_judgment: false
  - id: D2
    description: Every promoted common input carries a canonical finite PM execution envelope.
    verification:
      - kind: unit
        ref: "cmd/connectorgen/sourceimport_test.go::body and parameter envelope fail-closed tests"
        status: pass
      - kind: integration
        ref: "GOFLAGS=-p=3 go test -timeout 20m ./cmd/connectorgen"
        status: pass
    human_judgment: false
  - id: D3
    description: Runtime rejects over-cap encoded inputs before provider I/O without truncation.
    verification:
      - kind: integration
        ref: "internal/connectors/engine/operation_direct_read_bindings_test.go::TestOperationParametersReportPMExecutionCapBeforeIO"
        status: pass
    human_judgment: false
  - id: D4
    description: Exact numeric bounds and caller lexemes are preserved without breaking accepted CLI syntax.
    verification:
      - kind: integration
        ref: "internal/connectors/engine/operation_direct_read_bindings_test.go::exact and previously accepted numeric lexeme tests"
        status: pass
      - kind: unit
        ref: "cmd/connectorgen/paramsimport_test.go::TestParamsImportPreservesExactOneSidedNumericBounds"
        status: pass
    human_judgment: false
  - id: D5
    description: Help and credential-free inspection identify the effective PM limit and policy provenance.
    verification:
      - kind: e2e
        ref: "pm help gong; pm gong; pm gong calls get --help; pm connectors inspect gong --json"
        status: pass
      - kind: unit
        ref: "internal/cli/agentic_contract_test.go::TestConnectorInspectJSONIncludesRequestExecutionLimits"
        status: pass
    human_judgment: false
  - id: D6
    description: Semantic serialization ambiguity and uncensused headers remain source-traced and merge-blocking.
    verification:
      - kind: unit
        ref: "cmd/connectorgen/sourceimport_test.go::header quarantine and schema disposition tests"
        status: pass
      - kind: integration
        ref: "GOFLAGS=-p=3 make connector-boundary"
        status: pass
    human_judgment: false
---

# Summary — issue-4336 request-contract execution envelopes

The implementation separates provider validation from PM resource policy. V3
source descriptors retain source schemas exactly and attach versioned execution
envelopes for supported path/query scalars and JSON bodies. Generation fails
closed if a required envelope is missing or altered. Unsupported serialization,
dynamic/composed/untyped schemas, and headers without a proved textual byte cap
remain operation-local merge blockers.

Runtime path/query enforcement still uses the shipped 4 KiB default and 64 KiB
hard ceiling, measured after exact encoding and before authentication or network
I/O. Direct-write and structured-body limits are exposed as the existing 1/16
MiB, item/property, depth, and node policies; no provider maximum is invented.
Exact provider numeric bounds now survive import and surface synchronization,
and arbitrary-precision validation preserves the caller's wire lexeme as well as
the numeric spellings accepted by the previous runtime.

Generated connector help, manuals, skills, and credential-free inspection expose
effective PM limits and policy version. No connector definition or provider
allowlist changed. PR #4339 remains the independent dependency for the four
malformed paths, two retained-media quota cases, and unrelated dialect
prerequisites before a real Batch-1 executable-operation recount.

All task-scoped packages and repository gates passed with `GOFLAGS=-p=3` and one
heavy suite at a time. The full `internal/cli` suite was not claimed green: two
fresh-child binary tests failed only when their builds ran concurrently, then
each passed serialized. Focused CLI policy, inspection, help, docs, skills, and
golden parity tests passed.

After the required final merge of `main` at `43cda8924`, GitHub Verify exposed
one generated-artifact integration gap rather than a runtime-policy defect:
Twenty's newly admitted command surface was absent from the tracked execution-
policy notice. Canonical skill regeneration changed only
`docs/skills/pm-twenty/SKILL.md`; a second pass was byte-stable and the exact
generated-tree test passed in 4.376s. PR #4345 must rerun its required checks on
that repair before handoff.

Firstmate then required the branch to absorb main `60f85ae94` after #4346.
The merge was conflict-free despite overlapping `sourceprojection.go` changes.
Canonical `surface-sync --check` found zero drift across 553 connectors, the
tracked-skills equality test remained green in 4.942s, and the complete
`cmd/connectorgen` package passed in 183.024s under `-p 2`. GitHub checks must
settle again on the resulting merge head before the final rollup is reported.
