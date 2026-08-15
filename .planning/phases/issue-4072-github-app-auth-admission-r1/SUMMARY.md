---
phase: issue-4072-github-app-auth-admission-r1
status: complete
key_files:
  modified:
    - internal/connectors/engine/auth.go
    - internal/connectors/engine/hooks.go
    - internal/connectors/engine/read.go
    - internal/connectors/hooks/github/hooks.go
    - internal/connectors/hooks/github/hooks_test.go
  created:
    - internal/connectors/hooks/github/github_app_rate_admission_integration_test.go
---

# Issue #4072 summary

GitHub App installation-token minting now goes through an engine-owned,
rate-admitted requester instead of calling HTTP directly. The request retains
the #3754 physical-path boundary and fail-closed typed error behavior.

The integration proof uses two independently launched test processes sharing a
real Dragonfly budget of one. It asserts one minted token, one blocked child,
and one physical token POST, proving shared state changed.
