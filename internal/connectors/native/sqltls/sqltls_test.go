package sqltls

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

func getter(pairs map[string]string) func(string) string {
	return func(key string) string { return pairs[key] }
}

func TestParseModeAcceptsBothEngineVocabularies(t *testing.T) {
	for raw, want := range map[string]Mode{
		"disabled":        ModeDisabled,
		"disable":         ModeDisabled,
		"PREFERRED":       ModePreferred,
		"prefer":          ModePreferred,
		"allow":           ModePreferred,
		"required":        ModeRequired,
		"require":         ModeRequired,
		"verify-ca":       ModeVerifyCA,
		"verify_ca":       ModeVerifyCA,
		"verify-identity": ModeVerifyIdentity,
		"verify-full":     ModeVerifyIdentity,
		" Required ":      ModeRequired,
	} {
		got, err := ParseMode(raw)
		if err != nil || got != want {
			t.Fatalf("ParseMode(%q) = %q, %v, want %q", raw, got, err, want)
		}
	}
}

func TestParseModeRefusesUnknownValueRatherThanDisablingTLS(t *testing.T) {
	for _, raw := range []string{"", "on", "true", "verify", "vrify-ca", "none"} {
		if _, err := ParseMode(raw); err == nil {
			t.Fatalf("ParseMode(%q) succeeded, want refusal rather than a silent fallback", raw)
		}
	}
}

func TestResolveDefaultsOnlyWhenUnset(t *testing.T) {
	options, err := Resolve(getter(nil), ModeRequired)
	if err != nil || options.Mode != ModeRequired {
		t.Fatalf("Resolve(unset) = %+v, %v, want the connector default", options, err)
	}
	options, err = Resolve(getter(map[string]string{KeyMode: "disabled"}), ModeRequired)
	if err != nil || options.Mode != ModeDisabled {
		t.Fatalf("Resolve(explicit) = %+v, %v, want the explicit choice to win", options, err)
	}
}

func TestResolveRefusesOptionsThatWouldNeverBeApplied(t *testing.T) {
	for _, tc := range []struct {
		name  string
		pairs map[string]string
	}{
		{
			// A CA under "required" reads as verification that never happens.
			name:  "root certificate without a verifying mode",
			pairs: map[string]string{KeyMode: "required", KeyRootCert: "/etc/ssl/ca.pem"},
		},
		{
			name:  "server name without identity verification",
			pairs: map[string]string{KeyMode: "verify-ca", KeyServerName: "db.example.com"},
		},
		{
			name:  "relative root certificate path",
			pairs: map[string]string{KeyMode: "verify-ca", KeyRootCert: "ca.pem"},
		},
		{
			name:  "traversing root certificate path",
			pairs: map[string]string{KeyMode: "verify-ca", KeyRootCert: "/etc/../etc/ssl/ca.pem"},
		},
		{
			name:  "server name carrying a URL",
			pairs: map[string]string{KeyMode: "verify-identity", KeyServerName: "https://db.example.com/x"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := Resolve(getter(tc.pairs), ModeDisabled); err == nil {
				t.Fatal("Resolve() accepted a misleading transport-security configuration")
			}
		})
	}
}

func TestTLSConfigMatchesTheDeclaredStrictness(t *testing.T) {
	for _, tc := range []struct {
		mode           Mode
		wantNil        bool
		wantSkipVerify bool
		wantServerName string
		wantFallback   bool
	}{
		{mode: ModeDisabled, wantNil: true},
		{mode: ModePreferred, wantSkipVerify: true, wantFallback: true},
		{mode: ModeRequired, wantSkipVerify: true},
		// verify-ca leaves the built-in check off but installs an explicit
		// chain verifier, asserted separately below.
		{mode: ModeVerifyCA, wantSkipVerify: true},
		{mode: ModeVerifyIdentity, wantServerName: "db.internal"},
	} {
		t.Run(string(tc.mode), func(t *testing.T) {
			options := Options{Mode: tc.mode}
			config, err := options.TLSConfig("db.internal")
			if err != nil {
				t.Fatalf("TLSConfig(): %v", err)
			}
			if tc.wantNil {
				if config != nil {
					t.Fatal("TLSConfig() returned a config for a mode that negotiates no TLS")
				}
				if options.Encrypted() {
					t.Fatal("disabled reported itself as encrypted")
				}
				return
			}
			if config == nil {
				t.Fatal("TLSConfig() returned no config for an encrypting mode")
			}
			if config.MinVersion != 0x0303 {
				t.Fatalf("MinVersion = %x, want TLS 1.2", config.MinVersion)
			}
			if config.InsecureSkipVerify != tc.wantSkipVerify {
				t.Fatalf("InsecureSkipVerify = %t, want %t", config.InsecureSkipVerify, tc.wantSkipVerify)
			}
			if config.ServerName != tc.wantServerName {
				t.Fatalf("ServerName = %q, want %q", config.ServerName, tc.wantServerName)
			}
			if options.MayFallBackToPlaintext() != tc.wantFallback {
				t.Fatalf("MayFallBackToPlaintext() = %t, want %t", options.MayFallBackToPlaintext(), tc.wantFallback)
			}
		})
	}
}

func TestOnlyPreferredMayFallBackToPlaintext(t *testing.T) {
	for _, mode := range []Mode{ModeDisabled, ModeRequired, ModeVerifyCA, ModeVerifyIdentity} {
		if (Options{Mode: mode}).MayFallBackToPlaintext() {
			t.Fatalf("mode %q permitted a silent plaintext downgrade", mode)
		}
	}
	if !(Options{Mode: ModePreferred}).MayFallBackToPlaintext() {
		t.Fatal("preferred did not permit its documented plaintext fallback")
	}
}

func TestVerifyIdentityPrefersTheConfiguredServerName(t *testing.T) {
	config, err := Options{Mode: ModeVerifyIdentity, ServerName: "primary.db.example"}.TLSConfig("10.0.0.4")
	if err != nil {
		t.Fatalf("TLSConfig(): %v", err)
	}
	if config.ServerName != "primary.db.example" {
		t.Fatalf("ServerName = %q, want the configured name", config.ServerName)
	}
	if config.InsecureSkipVerify {
		t.Fatal("verify-identity disabled certificate verification")
	}
}

func TestVerifyCARejectsAChainOutsideTheConfiguredRoot(t *testing.T) {
	dir := t.TempDir()
	trustedCA, _ := newTestCA(t, "trusted")
	untrustedCA, untrustedKey := newTestCA(t, "untrusted")
	caPath := filepath.Join(dir, "ca.pem")
	if err := os.WriteFile(caPath, pemBytes(trustedCA.Raw), 0o600); err != nil {
		t.Fatalf("write CA: %v", err)
	}

	config, err := Options{Mode: ModeVerifyCA, RootCAPath: caPath}.TLSConfig("db.internal")
	if err != nil {
		t.Fatalf("TLSConfig(): %v", err)
	}
	if config.VerifyConnection == nil {
		t.Fatal("verify-ca installed no connection verifier")
	}
	if config.VerifyPeerCertificate != nil {
		t.Fatal("verify-ca used a verifier that skips resumed TLS sessions")
	}
	leaf, _ := newTestLeaf(t, untrustedCA, untrustedKey, "db.internal")
	if err := config.VerifyConnection(tls.ConnectionState{PeerCertificates: []*x509.Certificate{leaf}}); err == nil {
		t.Fatal("verify-ca accepted a certificate signed by an untrusted CA")
	}
	if err := config.VerifyConnection(tls.ConnectionState{}); err == nil {
		t.Fatal("verify-ca accepted an empty certificate chain")
	}
}

func TestVerifyCAAcceptsTheConfiguredRootIgnoringHostname(t *testing.T) {
	dir := t.TempDir()
	ca, key := newTestCA(t, "trusted")
	caPath := filepath.Join(dir, "ca.pem")
	if err := os.WriteFile(caPath, pemBytes(ca.Raw), 0o600); err != nil {
		t.Fatalf("write CA: %v", err)
	}
	config, err := Options{Mode: ModeVerifyCA, RootCAPath: caPath}.TLSConfig("db.internal")
	if err != nil {
		t.Fatalf("TLSConfig(): %v", err)
	}
	if config.VerifyConnection == nil {
		t.Fatal("verify-ca installed no connection verifier")
	}
	// The leaf names a different host on purpose: verify-ca proves the chain,
	// not the name.
	leaf, _ := newTestLeaf(t, ca, key, "someone-else.internal")
	if err := config.VerifyConnection(tls.ConnectionState{PeerCertificates: []*x509.Certificate{leaf}}); err != nil {
		t.Fatalf("verify-ca rejected a correctly chained certificate: %v", err)
	}
}

func TestVerifyCARevalidatesResumedSessions(t *testing.T) {
	dir := t.TempDir()
	ca, key := newTestCA(t, "trusted")
	caPath := filepath.Join(dir, "ca.pem")
	if err := os.WriteFile(caPath, pemBytes(ca.Raw), 0o600); err != nil {
		t.Fatalf("write CA: %v", err)
	}

	clientConfig, err := Options{Mode: ModeVerifyCA, RootCAPath: caPath}.TLSConfig("db.internal")
	if err != nil {
		t.Fatalf("TLSConfig(): %v", err)
	}
	if clientConfig.VerifyConnection == nil {
		t.Fatal("verify-ca installed no connection verifier")
	}
	clientConfig.MinVersion = tls.VersionTLS12
	clientConfig.MaxVersion = tls.VersionTLS12
	clientConfig.ClientSessionCache = tls.NewLRUClientSessionCache(1)
	verifyConnection := clientConfig.VerifyConnection
	var verificationCalls atomic.Int32
	clientConfig.VerifyConnection = func(state tls.ConnectionState) error {
		verificationCalls.Add(1)
		return verifyConnection(state)
	}

	serverLeaf, serverKey := newTestLeaf(t, ca, key, "someone-else.internal")
	serverConfig := &tls.Config{
		MinVersion:   tls.VersionTLS12,
		MaxVersion:   tls.VersionTLS12,
		Certificates: []tls.Certificate{{Certificate: [][]byte{serverLeaf.Raw}, PrivateKey: serverKey}},
	}
	if state, err := testTLSHandshake(clientConfig, serverConfig); err != nil {
		t.Fatalf("initial TLS handshake: %v", err)
	} else if state.DidResume {
		t.Fatal("initial TLS handshake unexpectedly resumed")
	}
	state, err := testTLSHandshake(clientConfig, serverConfig)
	if err != nil {
		t.Fatalf("resumed TLS handshake: %v", err)
	}
	if !state.DidResume {
		t.Fatal("second TLS handshake did not resume")
	}
	if got := verificationCalls.Load(); got != 2 {
		t.Fatalf("VerifyConnection calls = %d, want 2 including the resumed handshake", got)
	}
}

func TestTLSConfigReportsUnusableRootCertificateWithoutPanicking(t *testing.T) {
	dir := t.TempDir()
	junk := filepath.Join(dir, "junk.pem")
	if err := os.WriteFile(junk, []byte("not a certificate"), 0o600); err != nil {
		t.Fatalf("write junk: %v", err)
	}
	// The driver's own helper panics here; this package must not.
	if _, err := (Options{Mode: ModeVerifyCA, RootCAPath: junk}).TLSConfig("db.internal"); err == nil {
		t.Fatal("TLSConfig() accepted a file containing no PEM certificate")
	}
	missing := filepath.Join(dir, "absent.pem")
	if _, err := (Options{Mode: ModeVerifyIdentity, RootCAPath: missing}).TLSConfig("db.internal"); err == nil {
		t.Fatal("TLSConfig() accepted an unreadable root certificate path")
	}
}

func TestLibpqSSLModeRoundTripsEveryCanonicalMode(t *testing.T) {
	for canonical, libpq := range map[Mode]string{
		ModeDisabled:       "disable",
		ModePreferred:      "prefer",
		ModeRequired:       "require",
		ModeVerifyCA:       "verify-ca",
		ModeVerifyIdentity: "verify-full",
	} {
		if got := canonical.LibpqSSLMode(); got != libpq {
			t.Fatalf("LibpqSSLMode(%q) = %q, want %q", canonical, got, libpq)
		}
		back, err := ParseMode(libpq)
		if err != nil || back != canonical {
			t.Fatalf("ParseMode(%q) = %q, %v, want %q", libpq, back, err, canonical)
		}
	}
}

func newTestCA(t *testing.T, name string) (*x509.Certificate, *ecdsa.PrivateKey) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate CA key: %v", err)
	}
	template := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: name},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}
	raw, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("create CA: %v", err)
	}
	cert, err := x509.ParseCertificate(raw)
	if err != nil {
		t.Fatalf("parse CA: %v", err)
	}
	return cert, key
}

func newTestLeaf(t *testing.T, ca *x509.Certificate, caKey *ecdsa.PrivateKey, host string) (*x509.Certificate, *ecdsa.PrivateKey) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate leaf key: %v", err)
	}
	template := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject:      pkix.Name{CommonName: host},
		DNSNames:     []string{host},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	raw, err := x509.CreateCertificate(rand.Reader, template, ca, &key.PublicKey, caKey)
	if err != nil {
		t.Fatalf("create leaf: %v", err)
	}
	cert, err := x509.ParseCertificate(raw)
	if err != nil {
		t.Fatalf("parse leaf: %v", err)
	}
	return cert, key
}

func testTLSHandshake(clientConfig, serverConfig *tls.Config) (tls.ConnectionState, error) {
	serverRaw, clientRaw := net.Pipe()
	server := tls.Server(serverRaw, serverConfig)
	client := tls.Client(clientRaw, clientConfig)
	serverDone := make(chan error, 1)
	go func() {
		serverDone <- server.Handshake()
	}()
	clientErr := client.Handshake()
	state := client.ConnectionState()
	_ = clientRaw.Close()
	serverErr := <-serverDone
	_ = serverRaw.Close()
	if clientErr != nil {
		return state, clientErr
	}
	if serverErr != nil {
		return state, serverErr
	}
	return state, nil
}

func pemBytes(der []byte) []byte {
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
}
