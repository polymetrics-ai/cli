```
NAME
  pm webhooks - declare webhook receiver exposure modes

SYNOPSIS
  pm webhooks configure <name> --mode operator_endpoint --callback-url https://operator.example/receiver --receipt-capacity N [--credential credential]
  pm webhooks configure <name> --mode external_tunnel --tunnel-tool tailscale_funnel --callback-url https://node.tailnet.ts.net/receiver --heartbeat-ttl duration --receipt-capacity N [--credential credential]
  pm webhooks configure <name> --mode provider_pull_or_stream --adapter documented-adapter --receipt-capacity N [--credential credential]
  pm webhooks status <name> [--json]

DESCRIPTION
  Webhook ingress has three separate modes. The selected mode changes the
  recorded listener scope and recovery behavior; it is not a provider
  subscription command.

  operator_endpoint records an operator-owned stable HTTPS callback generation
  and starts no local listener. pm stores and renders only an opaque generation,
  never the supplied callback URL. An enabled stable Funnel endpoint may be
  selected as operator_endpoint, but that does not make pm own the Funnel.

  external_tunnel accepts a callback URL from the already-running named
  tailscale_funnel tool. It requires a loopback-only receiver contract and a
  positive heartbeat timeout. pm never installs, starts, probes, or shells out
  to Tailscale or any other tunnel. A heartbeat lapse or callback generation
  rotation marks the subscription degraded and requires provider-lane
  re-registration plus reconciliation before it can be active again.

  provider_pull_or_stream accepts no callback URL and starts no receiver. It
  describes a documented polling or outbound event-stream adapter, not a
  webhook. The polling executor is owned separately and is not started here.

  --receipt-capacity is an explicit bound for durable queued receipts. Provider
  adapters must verify the raw request body before parsing, durably record a
  receipt before success, handle duplicates and out-of-order delivery, and use
  a provider-specific reconciliation path after downtime. This command does
  not register, modify, or remove a provider subscription.

SECURITY
  Callback URLs, signing secrets, signature headers, raw event bodies,
  credentials, and credential revisions are never rendered in status output.
  Supplying --credential creates only an opaque credential-cohort association.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
