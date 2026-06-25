# THREAT-MODEL — GitHub Native Package + Data-Driven Registry

## Assets
- GitHub credentials (PAT / OAuth token / GitHub App private key) held in the pm vault.
- Reverse-ETL write capability against GitHub (issues/PRs/comments).

## Trust boundaries
- Connector code ↔ GitHub API (network). Connector code ↔ pm vault (secrets).

## Risks & mitigations (unchanged from current GitHub connector; preserve, don't regress)
- **Secret leakage**: secrets never logged/printed; only secret field *names* surface in docs.
  Migration must not add logging of `cfg.Secrets`. (security review checks this.)
- **GitHub App JWT**: RS256 signing of short-lived JWT → installation token; keep key parsing and
  expiry windows intact; do not weaken to longer-lived tokens.
- **Unauthorized writes**: reverse-ETL stays plan→preview→approve→execute; write actions remain an
  allow-listed set; `ValidateWrite` rejects unknown actions.
- **SSRF / base_url**: `base_url` override remains validated as today; no new unvalidated URL sinks.

## Migration-specific risk
- Moving helpers across packages could accidentally change visibility/behavior of auth or write
  validation → covered by parity tests (auth modes, write action accept/reject) and `make verify`.

## Non-applicable
No new network surface, no new secret types, no new dependency in this phase.
