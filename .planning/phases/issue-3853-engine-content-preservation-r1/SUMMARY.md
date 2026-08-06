---
coverage:
  - id: D1
    description: Dry-run write previews preserve the complete resolved config, secret, top-level record, and nested record content that execution will use.
    verification:
      - kind: unit
        ref: internal/connectors/engine/write_test.go:TestDryRunWritePreviewResolvedMethodPathPreservesSecretValues; TestDryRunWritePreviewResolvedPathPreservesConfiguredRecordFields; TestDryRunWritePreviewResolvedPathPreservesNestedRecordFields
        status: pass
    human_judgment: false
  - id: D2
    description: Direct-read, operation-direct-read, and binary-download errors preserve bounded HTTP URL/query/body diagnostics.
    verification:
      - kind: integration
        ref: internal/connectors/engine/direct_read_test.go:TestDirectReadPreservesHTTPErrorText; TestOperationDirectReadPreservesHTTPErrorText; internal/connectors/engine/binary_read_test.go:TestBinaryDownloadPreservesHTTPErrorTextAndLeavesNoFile
        status: pass
    human_judgment: false
  - id: D3
    description: Runtime help, checked-in manual, golden transcripts, and website documentation accurately distinguish complete connector-engine content from omitted approval tokens and the separate source-table summary.
    verification:
      - kind: unit
        ref: internal/cli:TestGoldenTranscripts; TestGoldenDocsGenerateMatchesTrackedCLIManuals
        status: pass
      - kind: other
        ref: ./pm reverse; ./pm help reverse; ./pm reverse --help; ./pm docs validate --connectors-dir docs/connectors
        status: pass
    human_judgment: false
---

# SUMMARY — issue #3853 engine content preservation

Implemented parent issue #3853 with no child issues.

- Dry-run write previews now interpolate the same full runtime config, secrets, and record values
  that execution uses. Existing `redact_fields` declarations remain load-compatible metadata, but
  no longer replace preview values; no connector bundles were rewritten.
- Direct read, operation direct read, and binary download now render bounded `connsdk.HTTPError`
  URL/body fields directly, preserving provider diagnostics while retaining existing error-map class
  and hint order. Generic non-HTTP errors retain their original error text.
- CLI help, the checked-in manual, golden transcripts, and website documentation state the
  complete-content engine guarantee. Approval tokens remain time-bounded, single-use authorization
  capabilities omitted from JSON output; that omission is not described as connector-content
  redaction. The generic source-table sample remains an explicitly separate app-level summary.
- The old masking tests were reversed rather than removed, recorded red before the production
  change, and are permanent content-preservation coverage.
- No command-runner function from #3771, #3852 policy enum, connector declaration, capability,
  provider request, credential, dependency, live check, or reverse-ETL execution was changed.

GSD discuss, plan, execute, verify, and code-review prompts were run inline under the required
no-spawn fallback. Detailed red/green and local-gate evidence is in `TDD-LEDGER.md`,
`VERIFICATION.md`, and `REVIEW.md`.
