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

## Independent-audit repair review — 2026-08-27

- Review target: `4fdb68f981a95d4ef13fbcf10d1a9a6ffd03d454` on top of current
  `origin/main@1324c52bab0b224ed8958858af7676b8b8e191b4`.
- GSD `code-review` source resolution was run. The local Pi runtime cannot
  provide its isolated reviewer worker, so this is the documented inline
  manual fallback. The subsequent Firstmate-requested independent Codex audit
  remains a separate, required exact-SHA review gate.

### Findings

No blocking finding after review of the repair diff.

1. **Projection ordering:** `source_contract_unavailable` now filters before
   any direct-read reachability, CLI, API-surface, flag, or write transform;
   the GET regression proves write and check modes preserve all declaration
   bytes.
2. **Exact proof:** the test fixture is a SHA-256-checked compression of the
   100124-byte, 259-row candidate lock. It asserts both real selected source
   IDs, the 253/6 citation split, current Outreach mapping, and no bundle
   mutation. It does not claim those rows were generated into production
   evidence.
3. **Decoder separation:** normal schema-v1/v2 locks decode through their
   original closed wire shape. Only a non-empty schema-v2 `source_kind`
   selects the distinct legacy reference wire contract, so source-only fields
   cannot be silently ignored by byte-backed import.
4. **Reference identity:** legacy and v3 reference rows share the exact REST
   protocol, method, textual identity, route, and citation-mixture checks.
   Digest use remains provenance-only; no credential, live request, generic
   HTTP/shell/SQL control, or executor was introduced.
