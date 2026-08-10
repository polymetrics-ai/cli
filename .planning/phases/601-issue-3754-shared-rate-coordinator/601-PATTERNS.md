# Phase 601 pattern map — inline/manual fallback

| Target role | Closest current implementation | Reuse / constraint |
| --- | --- | --- |
| Budget state | `internal/coordination/rate_limits.go` | Reuse fixed/sliding/token/leaky algorithms, injected `RateLimitClock`, and tighten-only observations; add batch locking without changing ordinary `Limiter` behavior. |
| Opaque scope | `internal/connectors/coordination_identity.go` | Convert only the returned `RateLimitScopeKey` at the engine boundary; never derive a second identity. |
| Pre-send gate | `internal/connectors/connsdk/http.go` and `stream.go` | Keep admission immediately before send/retry/redirect and finish a granted opaque lease from the same internal requester path. |
| Typed response facts | `internal/connectors/connsdk/rate_limit_requester.go` | Reuse `RateLimitObservation`; do not retain a header map, URL, body, variable, or credential. |
| Runtime injection | `internal/connectors/connectors.go` and `internal/connectors/engine/rate_limit_runtime.go` | Add only internal config fields and a closed local/shared selection; no command/config parser or provider changes. |
| Ephemeral owner cleanup | Go `net.UnixListener` plus `os.MkdirTemp` | Fresh short owner directory, explicit modes, bounded frame protocol, idempotent close, no stale-owner recovery. |

The plan has no frontend, database schema, public CLI, generated-surface, or
external dependency work. UI, CLI parity, and runtime-service guidance are not
applicable to this bounded internal foundation.
