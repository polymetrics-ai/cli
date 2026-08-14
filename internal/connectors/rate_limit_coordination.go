package connectors

// RateLimitCoordinationMode is the safe user-facing provenance of a
// connector's declared rate-limit policy. It never exposes a scope, binding,
// credential, subject, or coordinator address.
type RateLimitCoordinationMode string

const (
	RateLimitCoordinationProcessLocal  RateLimitCoordinationMode = "process_local"
	RateLimitCoordinationRequireShared RateLimitCoordinationMode = "require_shared"
)

// RateLimitCoordination describes only the protection a connector declaration
// requests. Process-local means one pm process coordinates its own requests;
// it deliberately makes no cross-process claim.
type RateLimitCoordination struct {
	Mode    RateLimitCoordinationMode `json:"mode"`
	Message string                    `json:"message"`
}

// RateLimitCoordinationProvider is implemented by connectors that expose a
// declared rate-limit provenance for safe inspection.
type RateLimitCoordinationProvider interface {
	RateLimitCoordination() RateLimitCoordination
}

// RateLimitCoordinationOf returns an inspection-safe policy provenance when
// a connector declares one.
func RateLimitCoordinationOf(connector Connector) (RateLimitCoordination, bool) {
	provider, ok := connector.(RateLimitCoordinationProvider)
	if !ok {
		return RateLimitCoordination{}, false
	}
	return provider.RateLimitCoordination(), true
}
