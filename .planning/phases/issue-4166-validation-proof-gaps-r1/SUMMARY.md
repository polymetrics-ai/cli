---
coverage:
  - id: D1
    description: Full certification fails when any declared GitHub write action is deliberately broken and prepares all 607 actions through the real engine.
    verification:
      - kind: integration
        ref: internal/connectors/certify/stages_write_coverage_internal_test.go — three TestFullWriteSweep tests
        status: pass
    human_judgment: false
  - id: D2
    description: Full certification resolves and durably executes the declared issue-label transport pair and fails for an absent declaration or unregistered executor.
    verification:
      - kind: integration
        ref: internal/app/issue_label_transport_certification_test.go and internal/connectors/certify/stages_transport_internal_test.go
        status: pass
    human_judgment: false
  - id: D3
    description: A fresh pm binary drives a real GitHub-definition issues ETL job and comment_issue reverse job as one flow through durable Parquet with independent provider read-back.
    verification:
      - kind: e2e
        ref: internal/cli/github_flow_roundtrip_test.go — TestFreshBinaryDeclarativeGitHubWarehouseFlowRoundTrip
        status: pass
    human_judgment: false
  - id: D4
    description: The same fresh-binary flow completes against a dedicated private GitHub repository and deletes it with a post-delete 404 zero-residue assertion.
    verification:
      - kind: e2e
        ref: internal/cli/github_flow_roundtrip_test.go — TestLiveFreshBinaryGitHubWarehouseFlowRoundTrip
        status: pass
    human_judgment: false
---

# Summary — Issue #4166 certification proof gaps

## Outcome

All three audited gaps are closed by fail-sensitive certification proofs. Gap 3 passed against real GitHub through a freshly built `pm`, a run-owned private repository under `Polymetrics-Cert`, durable Parquet, an independently read provider mutation, and a post-delete 404 zero-residue assertion.

The Gap 3 action is `comment_issue`: the flow extracts issue 1 into durable Parquet, maps its `number` and `title` back to `issue_number` and `body`, and independently reads the new comment from the provider. This is safe and reversible by deleting the dedicated repository, and it proves a same-object GitHub → warehouse → GitHub loop. `merge_pull_request` and `delete_file` are preview/refusal-only cases.

## Exact automated evidence

- Gap 1: 607 declared actions, 607 real engine preparations, 2 selected live-pair actions, 605 explicit non-live actions, and 3 curated lifecycle pairings. Corrupting either `create_label` or formerly blocked `update_issue` makes the report fail and exit 2.
- Gap 2: 3 provider reads, 1 provider write, 1 record read/loaded, 1 committed checkpoint, 1 transport manifest, and 1 Parquet artifact. Removing `sync_transport` or every executor factory fails before provider I/O.
- Gap 3 faithful control: binary SHA-256 `0849ee30f2b782684080ac1222837b5c76f19a5c46a5332cc85b36b93bb110c3`, size 151,259,538 bytes; flow sync/action 1 record each; provider comments 1→2; independent warehouse query 1 row; persisted checkpoint and flow receipt; replay, unapproved, and invalid-auth paths add zero comments. The invalid-auth checkpoint remains unchanged.
- Gap 3 live GitHub: binary SHA-256 `0849ee30f2b782684080ac1222837b5c76f19a5c46a5332cc85b36b93bb110c3`, size 151,259,538 bytes; flow sync/action 1 record each; 2 provider comments after flow; independent warehouse query 1 row; committed checkpoint and flow receipt; replay/unapproved/auth refusals preserve provider state; invalid-auth checkpoint unchanged; repository delete followed by provider 404; `zero_provider_residue=true`.
- The provider-verified 401 currently returns typed `internal/internal_error`. Issue #4169 owns correcting that product classification and adjacent 403/404/5xx handling; this validation PR records the behavior without fixing it.
- Runbook finding: the resource-owner labels `polymetrics-cli-cert` and `polymetrics-cli-cert-1` are stale. The live certification organization is `Polymetrics-Cert`; the line-277 token must be extracted from the labeled line as its 93-character token substring. No token value or repository identifier is retained.

## Declarative-path audit

The primary loop is generic after registry construction: `flow.Engine` → `connectorFlowActionRunner` → `app.ExecuteAuthorizedFlowAction` → registered `engine.Connector` → definition-owned `issues`, `comment_issue`, and `issue_comments`. `comment_issue` does not enter a GitHub write hook.

Non-generalizable code is limited to the existing issue-label transport adapter/factories and its certification fixture, GitHub App/compound-action hooks that the selected action does not call, and GitHub-owned definition data for paths, schemas, pagination, and rate policy. The CLI error-classification finding is provider-neutral and tracked by #4169.

## Scope and workflow

This branch adds tests, bounded certification wiring, and GSD evidence only; it adds no connector capability and does not fix #4125, #4158, or #4169. The required GSD lifecycle ran through the documented inline/manual fallback because the canonical issue-worker contract forbids workflow-role spawning.
