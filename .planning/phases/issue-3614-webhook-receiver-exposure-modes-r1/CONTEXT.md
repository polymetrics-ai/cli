# CONTEXT — issue #3614 webhook receiver exposure modes

## Captain-directed scope

This slice amends the ingress contract recommended by
`data/cli-plan-webhook-receiver-foundation-r1/report.md`. It owns only the local
receiver, its exposure-mode state, durable ingress receipt, endpoint-generation
tracking, and reconciliation request seam. It does not register or tear down a
provider subscription, implement a provider verifier, add a generic poller, or
claim connector change-capture capability.

The current task direction authorizes this implementation slice despite the
earlier scout report's queue-gate recommendation. The report's exposure policy
and security/recovery invariants remain binding.

## Dependency and stack boundary

`fm/cli-found-database-sync-contract-r1` (issue #3810) is the immediate stack
base. Its `synccontract.CheckpointEnvelope`, durable acknowledgement rule, and
closed `RecoveryOutcome` vocabulary are consumed unchanged. This slice must
not add another cursor, checkpoint, delivery-mode, tombstone, or recovery enum.

`connectors.CoordinationIdentity` is the only credential-scoped identity input.
The receiver stores no raw credential, credential revision, binding preimage,
or signing secret. #3855 remains the owner of polling execution; this slice
only declares `provider_pull_or_stream` as no-callback configuration.

## Observed Tailscale Funnel evidence (2026-08-06)

- Installed client: `tailscale 1.92.5`; daemon: `1.98.10`.
- `tailscale status --json` reported the stable node DNS name
  `burny---m2-max.bigscale-inconnu.ts.net` and Funnel entitlement.
- The installed daemon capability advertised Funnel ports `443,8443,10000`.
- `tailscale serve status` and `tailscale funnel status` both showed existing
  tailnet-only mappings on 443 and 8443; no public Funnel mapping is active.
- Port 10000 is the only entitled Funnel port not occupied by the observed
  mappings. No existing mapping will be edited, reset, removed, or made public.

The real Funnel test, if performed, may only add a port-10000 mapping after a
fresh status capture and local-listener proof. `pm` itself never invokes
Tailscale or any external binary; `tailscale_funnel` is an operator-declared
external tunnel name and its callback URL is supplied to `pm`.

## Locked behavioral distinctions

- `operator_endpoint`: accepts an operator-owned HTTPS callback and records
  only an opaque endpoint generation. It starts no local listener.
- `external_tunnel`: validates an HTTPS callback from the fixed
  `tailscale_funnel` tool and starts only a loopback listener. It records an
  opaque endpoint generation plus heartbeat state; a stale heartbeat or URL
  rotation degrades the subscription and opens a #3810 recovery request.
- `provider_pull_or_stream`: accepts a documented adapter reference, starts no
  listener, has no callback URL, and never presents itself as a webhook.

## Inline GSD fallback

`scripts/gsd doctor`, all required `sources` commands, and
`go run ./cmd/agentcontractgen check` passed. The generated discuss/plan
prompts are executed inline because this worker has no compatible Pi runtime
and the canonical contract forbids role spawning. This preserves TDD,
verification, review, and human gates.
