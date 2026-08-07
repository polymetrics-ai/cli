# DISCUSSION LOG — issue #3614 webhook receiver exposure modes

The captain supplied the material design decisions in the task and required
foundation report, so no interactive product question remains for this slice.

| Area | Decision |
| --- | --- |
| Public ingress | Three separate modes: `operator_endpoint`, `external_tunnel`, `provider_pull_or_stream`. |
| Tunnel implementation | Only accept an operator-supplied callback from named `tailscale_funnel`; never install, spawn, or shell out to a tunnel. |
| Real-tool proof | Tailscale Funnel is verified as zero-byte external tooling with stable node DNS and entitled ports observed locally. Existing port-443/8443 mappings remain untouched. |
| Receiver security | Loopback-only listener in tunnel mode; raw-body verification before parse/persist; bounded body, deadline, and in-flight admission; generic failures without secret/body/header logging. |
| Durability/recovery | Durable receipt/dedupe transaction precedes acknowledgement. A URL generation change or heartbeat loss marks the subscription degraded and requests #3810 reconciliation; a provider lane owns the actual re-registration/replay call. |
| Provider pull/stream | Explicit no-callback declaration only; #3855 owns polling implementation. |
