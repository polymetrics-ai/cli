# VERIFICATION — issue #3614 webhook receiver exposure modes

Status: implementation verified with one inherited #3810-stack package-suite
failure recorded below. No provider registration, provider credential, public
Funnel, or polling executor was invoked.

## Required evidence

## Passed implementation evidence

- `go test -race ./internal/webhook -count=1 -v`: all ingress tests passed,
  including loopback-only receiver, raw-body verification, durable-receipt
  acknowledgement ordering, duplicates, out-of-order delivery, oversized
  payload, receipt-capacity, in-flight, timeout, heartbeat, and generation
  rotation paths.
- `go test ./internal/app -run '^TestWebhookReceiver' -count=1 -v`: passed.
  It proves ordinary project state excludes a callback URL, raw body, and event
  identifier while the durable receipt body is encrypted in the vault.
- `go test ./internal/cli -count=1`: passed, including updated transcript and
  documentation-generator assertions. `go vet ./internal/webhook
  ./internal/app ./internal/cli`, website `pnpm --dir website run typecheck`,
  `pm help webhooks`, bare `pm webhooks`, and `pm webhooks --help` passed.
- `make tidy-check`, `make lint`, `make docs-check`, `make smoke-no-build`,
  `make agent-contract-check`, `make connectorgen-validate`, `make
  connectorgen-surface-sync`, `make connector-boundary`, and `make
  release-workflow-check` all passed. `tidy-check` confirms no `go.mod` or
  `go.sum` delta: Funnel adds zero bytes/dependencies to `pm`.
- Generated parity outputs are `docs/cli/webhooks.md`, CLI golden transcripts,
  and `website/lib/docs.generated.ts`. `pm docs validate` passed.

## Inherited package-suite failure

`go test ./internal/app -count=1 -v` fails two tests outside this ingress
diff: `TestGithubPullRequestsETLSupportsAllSyncModes/incremental_append_filters_older_cursor_and_appends_inclusive_cursor` and `TestRunETLWritesBoundedBatches`. Both expect generic ETL admission that the stacked #3810 durable-sync contract now rejects. This lane consumes #3810 and does not change ETL admission; the focused ingress app tests pass.

## Tailscale Funnel evidence and limit

- Read-only inspection on 2026-08-06 observed Tailscale client `1.92.5`, daemon
  `1.98.10`, the stable node MagicDNS name in its tailnet domain, and Funnel
  entitlement ports `443`, `8443`, and `10000`.
- `tailscale serve status` and `tailscale funnel status` showed existing
  tailnet-only Serve mappings on `443` and `8443`. They were not changed,
  reset, removed, or made public. No firewall or non-Tailscale configuration
  was touched.
- Official Funnel documentation confirms that names stay within the tailnet
  `*.ts.net` domain, allowed ports are `443`, `8443`, and `10000`, and a
  persistent `-bg` Funnel resumes after Tailscale/device restart. The website
  docs link the authoritative Funnel examples and CLI reference.
- A real public Funnel was intentionally not added on unused port `10000`: the
  safety contract forbids removing it afterward, and public exposure must not
  be used as a transient test. `TestStartLoopbackServesOnlyExternalTunnelMode`
  is the end-to-end local loopback proof with a supplied Funnel-shaped URL.
  Public Funnel reachability remains operator-run production validation.

## Inline GSD verification and review

The single-worker fallback recorded in `PLAN.md` was used because the canonical
delivery contract forbids role spawning. The implementation was checked
goal-backward against R1–R7 and reviewed inline for: secret/callback leakage,
provider mutation or shell execution, incorrect non-loopback binding,
acknowledgement before storage, ordering assumptions, unbounded ingress, and
new dependencies. No ingress finding remained after the empty-receipt-map fix.
