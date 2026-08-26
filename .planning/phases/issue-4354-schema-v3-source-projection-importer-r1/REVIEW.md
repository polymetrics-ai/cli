# Code review — source-reference projection foundation

## Scope and audit target

- Reviewed commit: `b0449adb035475de23f706be7ad6ae889c659bcb`
- Changed production paths: `cmd/connectorgen/sourceimport.go`,
  `sourceprojection.go`, and `operationevidence.go`.
- Reviewed documentation and GSD/TDD evidence with the production change.

## Findings

No blocking correctness, provenance, or generic-escape finding.

### Checked invariants

1. **Byte-backed source import remains strict.** Only an explicit v3
   `kind: source_reference` document or exact legacy Outreach
   `source_kind` selects the cited-only path. Ordinary v1/v2/v3 locks still
   use the retained-artifact reader and byte/canonical identity verification.
2. **No provider/network or credential route.** Reference-only locks bypass
   `fetchSourceImportArtifact`; source-import's normal retained reader is used
   only for a non-reference document. No user URL, token, request body, generic
   HTTP, shell, or SQL control was added.
3. **Citations are operation-specific.** The legacy adapter validates method
   counts, IDs, routes, per-operation source URL, and each supplement's
   identity/count. Operation evidence resolves a supplement by that exact URL,
   rather than attributing it to the primary OpenAPI citation.
4. **No false implementation.** A source-reference descriptor has no inferred
   execution contract and its runtime is merge-blocked with
   `source_contract_unavailable`; source projection skips materialization.
   Evidence may show an existing lane as declared but forces every lane's
   `enabled` value false and emits the named gap.
5. **Shared not Outreach-specific.** The v3 source-reference fixture exercises
   the same descriptor constructor and evidence path as the retained Outreach
   legacy adapter.

## Required automated review routing

This direct PR will be opened normally against `main`. GitHub's trusted-author
Claude review is expected to run on open/ready. Its status and any actionable
finding disposition remain a human/remote follow-up, not a local pass claim.
