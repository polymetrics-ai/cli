// Package sqltls is the one transport-security option shape shared by every
// native SQL connector, so MySQL and PostgreSQL cannot drift into two
// different spellings of the same choice.
//
// The user chooses; the connector enforces. Both TLS and non-TLS are
// supported for local and remote servers, and a mode is never silently
// downgraded: only "preferred" may fall back to plaintext, and only when the
// server itself advertises no TLS support.
package sqltls

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// Mode is the canonical transport-security vocabulary. libpq and MySQL
// spellings are accepted as aliases and normalise to these values.
type Mode string

const (
	// ModeDisabled never negotiates TLS and never upgrades.
	ModeDisabled Mode = "disabled"
	// ModePreferred encrypts when the server advertises TLS and falls back to
	// plaintext when it does not. This is the only downgrading mode.
	ModePreferred Mode = "preferred"
	// ModeRequired encrypts and fails closed when the server offers no TLS.
	// It does not prove the server's identity.
	ModeRequired Mode = "required"
	// ModeVerifyCA encrypts, fails closed, and verifies the server chain, but
	// does not check the hostname.
	ModeVerifyCA Mode = "verify-ca"
	// ModeVerifyIdentity encrypts, fails closed, and verifies both the chain
	// and the server name.
	ModeVerifyIdentity Mode = "verify-identity"
)

// Config keys are shared verbatim by every SQL connector.
const (
	KeyMode       = "sslmode"
	KeyRootCert   = "sslrootcert"
	KeyServerName = "sslservername"
)

// modeAliases maps every accepted spelling to a canonical mode. The libpq
// names exist so a value that works on the PostgreSQL connector means the
// same thing on MySQL.
var modeAliases = map[string]Mode{
	"disabled":        ModeDisabled,
	"disable":         ModeDisabled,
	"preferred":       ModePreferred,
	"prefer":          ModePreferred,
	"allow":           ModePreferred,
	"required":        ModeRequired,
	"require":         ModeRequired,
	"verify-ca":       ModeVerifyCA,
	"verify_ca":       ModeVerifyCA,
	"verify-identity": ModeVerifyIdentity,
	"verify_identity": ModeVerifyIdentity,
	"verify-full":     ModeVerifyIdentity,
}

// AcceptedModes is the documented vocabulary, in increasing strictness.
const AcceptedModes = "disabled/preferred/required/verify-ca/verify-identity"

// Options is the resolved transport-security choice for one connection.
type Options struct {
	Mode       Mode
	RootCAPath string
	ServerName string
}

// ParseMode normalises one accepted spelling. An unrecognised value is an
// error rather than a fallback, so a typo can never quietly disable TLS.
func ParseMode(raw string) (Mode, error) {
	mode, ok := modeAliases[strings.ToLower(strings.TrimSpace(raw))]
	if !ok {
		// Do not echo the value: a mistyped configuration value can itself
		// carry connection material.
		return "", errors.New("sslmode is not one of " + AcceptedModes)
	}
	return mode, nil
}

// Resolve reads the shared keys through the caller's accessor. An absent
// sslmode takes defaultMode, which each connector states in its own spec.
func Resolve(get func(string) string, defaultMode Mode) (Options, error) {
	options := Options{Mode: defaultMode}
	if raw := strings.TrimSpace(get(KeyMode)); raw != "" {
		mode, err := ParseMode(raw)
		if err != nil {
			return Options{}, err
		}
		options.Mode = mode
	}
	if raw := strings.TrimSpace(get(KeyRootCert)); raw != "" {
		if err := validateRootCAPath(raw); err != nil {
			return Options{}, err
		}
		options.RootCAPath = raw
	}
	if raw := strings.TrimSpace(get(KeyServerName)); raw != "" {
		if strings.ContainsAny(raw, "/\\?#@ \t\r\n\x00") {
			return Options{}, errors.New("sslservername must be a bare hostname")
		}
		options.ServerName = raw
	}
	if options.RootCAPath != "" && !options.Verifies() {
		// Naming a CA that will never be consulted reads as verification that
		// is not happening. Refuse rather than accept a misleading config.
		return Options{}, errors.New("sslrootcert requires sslmode verify-ca or verify-identity")
	}
	if options.ServerName != "" && options.Mode != ModeVerifyIdentity {
		return Options{}, errors.New("sslservername requires sslmode verify-identity")
	}
	return options, nil
}

// Encrypted reports whether this mode negotiates TLS at all.
func (o Options) Encrypted() bool { return o.Mode != ModeDisabled }

// Verifies reports whether the server's certificate chain is checked.
func (o Options) Verifies() bool {
	return o.Mode == ModeVerifyCA || o.Mode == ModeVerifyIdentity
}

// MayFallBackToPlaintext reports whether a plaintext retry is permitted when
// the server advertises no TLS. Only preferred may; every stricter mode must
// fail closed.
func (o Options) MayFallBackToPlaintext() bool { return o.Mode == ModePreferred }

// TLSConfig builds the client configuration for host, or nil when the mode
// negotiates no TLS. Unlike the driver's own helper it returns an error
// rather than panicking on an unreadable or malformed CA file.
func (o Options) TLSConfig(host string) (*tls.Config, error) {
	if !o.Encrypted() {
		return nil, nil
	}
	config := &tls.Config{MinVersion: tls.VersionTLS12}
	if !o.Verifies() {
		// required and preferred encrypt without proving identity. This is
		// stated in the connector docs rather than implied.
		config.InsecureSkipVerify = true
		return config, nil
	}

	roots, err := o.rootPool()
	if err != nil {
		return nil, err
	}
	config.RootCAs = roots
	if o.Mode == ModeVerifyIdentity {
		config.ServerName = host
		if o.ServerName != "" {
			config.ServerName = o.ServerName
		}
		return config, nil
	}

	// verify-ca checks the chain but not the name. Go has no direct switch
	// for that, so verification is done by hand with the hostname omitted.
	// InsecureSkipVerify only suppresses the built-in check; VerifyConnection
	// supplies the explicit one below for every handshake, including a resumed
	// TLS session. VerifyPeerCertificate would not run on resumptions.
	config.InsecureSkipVerify = true
	config.VerifyConnection = func(state tls.ConnectionState) error {
		return verifyChainWithoutHostname(state.PeerCertificates, roots)
	}
	return config, nil
}

func (o Options) rootPool() (*x509.CertPool, error) {
	if o.RootCAPath == "" {
		// No CA named: verify against the host trust store. That is real
		// verification, not a downgrade, and is the usual case for a managed
		// server with a publicly-trusted certificate.
		pool, err := x509.SystemCertPool()
		if err != nil {
			return nil, errors.New("read system certificate pool for sslmode " + string(o.Mode))
		}
		return pool, nil
	}
	pem, err := os.ReadFile(o.RootCAPath)
	if err != nil {
		// Do not echo the path: it is caller-supplied configuration.
		return nil, errors.New("read sslrootcert file")
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(pem) {
		return nil, errors.New("sslrootcert file contains no PEM certificate")
	}
	return pool, nil
}

func verifyChainWithoutHostname(certs []*x509.Certificate, roots *x509.CertPool) error {
	if len(certs) == 0 {
		return errors.New("server presented no certificate")
	}
	intermediates := x509.NewCertPool()
	for _, cert := range certs[1:] {
		intermediates.AddCert(cert)
	}
	// DNSName is deliberately empty: verify-ca proves the chain, not the name.
	if _, err := certs[0].Verify(x509.VerifyOptions{
		Roots:         roots,
		Intermediates: intermediates,
	}); err != nil {
		return errors.New("server certificate chain is not trusted")
	}
	return nil
}

// LibpqSSLMode renders a canonical mode in libpq's vocabulary for drivers
// that take a keyword/value connection string.
func (m Mode) LibpqSSLMode() string {
	switch m {
	case ModeDisabled:
		return "disable"
	case ModePreferred:
		return "prefer"
	case ModeRequired:
		return "require"
	case ModeVerifyCA:
		return "verify-ca"
	case ModeVerifyIdentity:
		return "verify-full"
	default:
		return ""
	}
}

// validateRootCAPath bounds the file a configuration value may open to an
// absolute, traversal-free path.
func validateRootCAPath(path string) error {
	if strings.ContainsAny(path, "\r\n\x00") {
		return errors.New("sslrootcert must be a safe absolute path")
	}
	if !filepath.IsAbs(path) || filepath.Clean(path) != path {
		return errors.New("sslrootcert must be a clean absolute path")
	}
	for _, element := range strings.Split(path, string(filepath.Separator)) {
		if element == ".." {
			return errors.New("sslrootcert must not traverse parent directories")
		}
	}
	return nil
}
