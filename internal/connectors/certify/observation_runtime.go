package certify

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
	"sync"
)

// certificationHTTPObservationMu guards the process-global default transport.
// Certification runs are intentionally serialized while live traffic is being
// observed, so an unrelated request can never be attributed to this proof.
var certificationHTTPObservationMu sync.Mutex

func installCertificationHTTPObserver(maxBodyBytes int) (*ObservedTransport, func(), error) {
	certificationHTTPObservationMu.Lock()
	previous := http.DefaultTransport
	underlying, err := certificationObservationTransport(previous)
	if err != nil {
		certificationHTTPObservationMu.Unlock()
		return nil, nil, err
	}
	observer := NewObservedTransport(underlying, maxBodyBytes)
	http.DefaultTransport = observer
	return observer, func() {
		http.DefaultTransport = previous
		certificationHTTPObservationMu.Unlock()
	}, nil
}

// certificationObservationTransport honors an explicit SSL_CERT_FILE inside
// the short-lived observed child process. macOS's system trust path does not
// consistently consume that standard Go environment setting; cloning the
// transport here keeps test and enterprise CA trust bounded to proof capture
// without changing ordinary CLI transport behavior.
func certificationObservationTransport(previous http.RoundTripper) (http.RoundTripper, error) {
	path := os.Getenv("SSL_CERT_FILE")
	if path == "" {
		return previous, nil
	}
	base, ok := previous.(*http.Transport)
	if !ok {
		return nil, fmt.Errorf("certification observer cannot apply SSL_CERT_FILE to transport %T", previous)
	}
	payload, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("certification observer read SSL_CERT_FILE: %w", err)
	}
	roots, err := x509.SystemCertPool()
	if err != nil || roots == nil {
		roots = x509.NewCertPool()
	}
	if !roots.AppendCertsFromPEM(payload) {
		return nil, fmt.Errorf("certification observer SSL_CERT_FILE %q contains no certificates", path)
	}
	clone := base.Clone()
	if clone.TLSClientConfig == nil {
		clone.TLSClientConfig = &tls.Config{}
	} else {
		clone.TLSClientConfig = clone.TLSClientConfig.Clone()
	}
	clone.TLSClientConfig.RootCAs = roots
	return clone, nil
}
