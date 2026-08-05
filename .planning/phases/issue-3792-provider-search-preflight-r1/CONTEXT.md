# Context — issue #3792 provider-search runtime preflight

## Fixed decisions

- Deliver only #3792's runtime admission defect. #3788 (declaration/evidence) and #3797
  (fleet enforcement) remain deferred behind their stated dependencies.
- The change must not wait for #3788: `provider_search` and its executor already exist in
  `internal/connectors/engine/bundle.go:1499` and `internal/connectors/engine/direct_read.go:32`.
  No production provider declaration is needed to prove a runtime preflight contract.
- Avoid PR #3740's overlap by not touching `internal/connectors/connectors.go`,
  `internal/connectors/engine/connector.go`, `internal/connectors/engine/bundle.go`, any schema,
  connector definition, generated surface, or documentation. Its published diff includes the first
  two files but not `engine/direct_read.go` or commandrunner.
- The consumer-owned no-network contract may be an unexported commandrunner interface whose
  implementation is an engine `Connector` method defined in `engine/direct_read.go`. It must inspect
  the loaded operation rather than mirror operation rules in `cmd/connectorgen`.
- A command must be rejected before dispatch when its referenced operation is missing, unsupported,
  has a different method/path/policy, or lacks a positive response cap.
- #3771 owns commandrunner content-preservation and #3852 owns the declarable output-policy enum.
  This phase does not add, select, remove, or transform an output policy or `redact_fields`; it only
  requires exact command-to-loaded-operation policy agreement.

## Non-goals

- No provider query kind, Freshchat/YouTube declaration, citation/evidence schema, validator,
  `connectorgen` change, capability claim, `connectors.Querier` change, or `pm query` change.
- No provider calls, credentials, writes, dependencies, redaction/masking, or generic HTTP/SQL/body
  escape hatch.

## Discussion outcome

The issue and firstmate instruction resolve every material design choice. The GSD discussion was
performed inline because the canonical single-worker contract forbids role spawning. No question was
sent to a human and no product choice was reopened.
