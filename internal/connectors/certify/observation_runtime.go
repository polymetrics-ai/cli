package certify

import (
	"net/http"
	"sync"
)

// certificationHTTPObservationMu guards the process-global default transport.
// Certification runs are intentionally serialized while live traffic is being
// observed, so an unrelated request can never be attributed to this proof.
var certificationHTTPObservationMu sync.Mutex

func installCertificationHTTPObserver(maxBodyBytes int) (*ObservedTransport, func()) {
	certificationHTTPObservationMu.Lock()
	previous := http.DefaultTransport
	observer := NewObservedTransport(previous, maxBodyBytes)
	http.DefaultTransport = observer
	return observer, func() {
		http.DefaultTransport = previous
		certificationHTTPObservationMu.Unlock()
	}
}
