# Discussion log — #3979 PostgreSQL gap-free bootstrap

The Firstmate launch brief supplied the implementation decisions non-interactively: PostgreSQL only; use the landed bounded snapshot and pgoutput-v2 CDC foundations; preserve the existing delivery receipt-before-checkpoint-before-acknowledgement order; prove a concurrent mutation during the snapshot; and keep unrelated issues out of scope.

The active GSD adapter cannot resolve issue `3979` to a numbered roadmap phase. Its `--auto` prompts were generated and executed inline under the repository's single-worker manual-fallback rule. The decisions are recorded in `CONTEXT.md` rather than silently inferred during implementation.
