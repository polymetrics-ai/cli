# Phase 425 Summary

Status: native version implementation focused-green; broad verification pending.

Invocation session `issue-425-pi-openai-codex-gpt-5.6-sol-high-20260718T095316Z` explicitly uses `openai-codex/gpt-5.6-sol` with `thinking=high` from exact starting HEAD `479a62f930e7c8a9a51ba0b3deb088bf3aad3ecc` on `refactor/425-version-native-cobra`.

The repo-local GSD adapter is healthy (`doctor` and 69-command `list` passed); explicit plan-phase prompt generation passed. The required programming-loop command is absent with exact error `scripts/gsd: unknown GSD command: programming-loop`, so the manual universal-loop fallback is recorded and strict TDD remains required.

Focused tests cover native registration, deterministic plain/JSON output, flag and positional help, JSON manual, unknown flags, and invalid actions. Exact RED proved the router registration mismatch and `DisableFlagParsing=true`. The smallest implementation now registers a native version leaf, preserves hidden positional help, removes the legacy wrapper, and removes the obsolete handler argument check. Focused version/router/golden tests pass with no fixture delta; broad parity and repository gates remain pending. Help text, version output, JSON envelopes, docs, website, generated artifacts, and golden fixtures are expected unchanged and will be proven by tests/diffs. No secrets, services, reverse ETL, dependencies, unrelated namespaces, external review requests, or PR creation are allowed.
