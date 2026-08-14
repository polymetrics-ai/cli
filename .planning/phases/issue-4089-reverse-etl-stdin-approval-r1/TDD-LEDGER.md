# #4089 — TDD ledger

**Status:** green; implementation and focused regression evidence complete.

| Checkpoint | Evidence | Result |
| --- | --- | --- |
| Plan | GSD prompts resolved inline; issue decisions, scope, parity, and safety checks recorded before production edits. | recorded |
| Red: bounded stdin carrier | `go test -count=1 -timeout 20m -run '^TestReverseETLApprovalUsesBoundedStdin$' -v ./internal/cli` built a fresh binary, observed a live argv with only the stdin marker, then failed the valid run with exit status 1 because the unchanged CLI ignored stdin. | reproduced |
| Green: generic wiring | Both request construction paths call `reverseApprovalTokenFromStdin`, which only validates the bare marker then delegates to `readApprovalTokenFromStdin(os.Stdin)` before the plan lookup or write dispatch. | passed |
| Green: binary regression | `go test -count=1 -timeout 20m -run '^TestReverseETLApprovalUsesBoundedStdin$' -v ./internal/cli` passed: it logged the live safe argv, independently checked argv/environment/project files/logs/receipt/evidence, rejected empty/oversized/multiline/valued/argv inputs with no receipt, and rejected replay. | passed |
| Refactor: docs and safety | Generated manuals, website data, transcripts, explicit stale-syntax scan, focused certification harness tests, and smoke run are green. | passed |

## Required red/green record

- Red: observed on 2026-08-14. The fresh binary process argv was recorded without the token; the valid stdin run failed at the unchanged argv-only request construction (exit 1). No production file had changed.
- Green: observed on 2026-08-14. The same selector exited 0. Its logged live command line was `pm reverse run <plan-id> --approval-token-stdin --root <temp-root> --json`; it contained no token. The test then asserted the token absent from argv, the command environment, durable project files, captured logs, the outbox receipt, and its emitted evidence record, while replay was rejected without another receipt.
