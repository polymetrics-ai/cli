# TDD LEDGER — issue #3614 webhook receiver exposure modes

| ID | Contract | RED evidence | GREEN evidence |
| --- | --- | --- | --- |
| R1 | The three exposure modes are closed and change behavior | `go test ./internal/webhook` failed against the intentionally absent receiver API | `TestConfigureExposureModesRemainDistinctAndOpaque`; `TestAtLeastOnceDeliveryStatesWebhookOrderingTruthfully` |
| R2 | Tunnel mode binds loopback only and accepts only the named external Tailscale Funnel callback | Same missing-API test run | `TestStartLoopbackServesOnlyExternalTunnelMode`; named tool and `.ts.net` boundary cases |
| R3 | Raw request bytes are verified before parse or durable write; failures persist nothing | Same missing-API test run | `TestReceiverVerifiesRawBodyAndPersistsBeforeAcknowledging`; unverified rejection coverage |
| R4 | A receipt is durable before success; duplicate/out-of-order inputs are safe | Same missing-API test run | Receipt-before-ack writer assertion, `TestReceiverAcceptsOutOfOrderAndDuplicateDeliveries`, and encrypted app-store coverage |
| R5 | Size, timeout, and in-flight limits reject excess work instead of buffering | Same missing-API test run | Oversize, durable-capacity, in-flight, and verifier-timeout rejection tests |
| R6 | URL rotation or missing heartbeat degrades and yields #3810 recovery work until explicit re-registration/reconciliation | Same missing-API test run | `TestSubscriptionLifecycleDegradesOnGenerationRotationAndHeartbeatLoss`; app rotation test checks `source_generation_changed` |
| R7 | CLI/help/docs expose the active mode without leaking callback URLs or secret material | `TestWebhookCommandsDeclareModesWithoutLeakingCallbacks` failed before the namespace existed | CLI test, goldens, generated `docs/cli/webhooks.md`, and website typecheck |
