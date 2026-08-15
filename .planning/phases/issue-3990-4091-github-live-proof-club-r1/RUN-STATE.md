# Run state — Issues #3990 and #4091 live proof

- Branch: `fm/cli-github-live-proof-club-r1`
- Base: `integration/4015-mvp-flat-r1`
- Current stage: both credentialed proofs, cleanup, one-pass generation, focused validation, and inline review passed; commit and PR delivery pending
- GSD mode: inline/manual fallback; combined issue slug is absent from ROADMAP.
- Secrets observed: none
- Provider mutations: run-owned labels and one GraphQL star were created only through approved production paths; all were subsequently removed through separately approved production paths
- Cleanup required: none; independent PM reads show empty issue labels, zero run-prefix repository labels, the certification tag absent, and star count restored to zero; the temporary remote branch is deleted
- #3990 result: current external child proof passed with 1,370 stages, zero failures, 83 fingerprinted GitHub exchanges, no 429/abuse response, and `leaks=null`
- #4091 result: set-replace and keyed first runs plus identical-scope unattended reruns passed live; both token replays and a real 401 left provider/checkpoint state unchanged
- Issue evidence: exact secret-free commands and verbatim output posted to https://github.com/polymetrics-ai/cli/issues/4091
- Validation: broad changed-package `internal/app/...` and `internal/cli` runs passed; focused certification/CLI tests, vet, lint, docs, smoke, connector drift, canon, and release gates passed
