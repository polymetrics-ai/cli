# App and invocation lifetime

Each production `cli.Run` owns one lazy shared rate-limit scope and the Apps it
constructs. Ordinary local, help, metadata and missing-credential calls do not
construct a Redis client. An explicitly `require_shared` policy first requests
the scope through `RuntimeConfig.SharedRateLimits`; unavailable coordination
still refuses before provider I/O. Local policy, opaque keys, budgets and request
deadlines retain their existing owners.

Go callers can pass borrowed capabilities with `app.RuntimeOptions` through
`OpenWithRuntime` or `OpenWithRegistryAndRuntime` (and the corresponding reverse
execution constructors). These options are installed before durable parking
starts, including synchronous due-work resumption. Existing `Open` wrappers
remain service-free and have no ambient shared coordinator.

Call `App.Close` after foreground work finishes. It cancels and joins parking
callbacks and stops timers while retaining durable parked work. Close is
idempotent; it does not close supplied registries or shared capabilities. The
caller that owns a shared scope closes it after every borrowing App has closed.
Closing an unused scope opens nothing. App lifecycle does not promise general
concurrent mutation safety.

CLI custom/fixture App openers retain ownership of their supplied Apps. An
explicitly supplied coordinator is also borrowed. Production CLI teardown closes
owned Apps before the invocation scope and preserves the command/provider
result. Cleanup failure emits a bounded diagnostic without endpoint or credential
values; direct Go scope owners can inspect the returned `Close` error.

Proofs: `TestSharedRateLimitScopeOwnership218`,
`TestSharedRateLimitRealResourceChild218`, `TestAppParkingLifecycle218`,
`TestCLIConcurrentInvocationOwnership218`, and
`TestBorrowedSharedRateConsumers218`. Foundation Atlas integration and parent
release verification remain separate from these focused lifecycle checks.
