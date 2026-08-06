package email

import (
	"errors"
	"fmt"
	"net"
	"net/mail"
	"strconv"
	"strings"
	"time"
	"unicode"

	"polymetrics.ai/internal/connectors"
)

const (
	defaultTimeout = 30 * time.Second
)

type transportSecurity string

const (
	securityTLS      transportSecurity = "tls"
	securitySTARTTLS transportSecurity = "starttls"
	securityNone     transportSecurity = "none"
)

// connectionConfig keeps credentials in memory only. Its Error and String
// paths intentionally never include password, and no protocol DebugWriter is
// configured by this connector.
type connectionConfig struct {
	imapHost     string
	imapPort     string
	imapSecurity transportSecurity
	smtpHost     string
	smtpPort     string
	smtpSecurity transportSecurity
	username     string
	password     string
	smtpUsername string
	fromAddress  string
	timeout      time.Duration
}

func (c connectionConfig) imapAddress() string { return net.JoinHostPort(c.imapHost, c.imapPort) }
func (c connectionConfig) smtpAddress() string { return net.JoinHostPort(c.smtpHost, c.smtpPort) }

func resolveConnectionConfig(cfg connectors.RuntimeConfig) (connectionConfig, error) {
	imapHost, err := normalizeHost(cfg.Config["imap_host"], "imap_host")
	if err != nil {
		return connectionConfig{}, err
	}
	imapPort, err := allowedPort(cfg.Config["imap_port"], "imap_port", "143", "993")
	if err != nil {
		return connectionConfig{}, err
	}
	imapSecurity, err := allowedSecurity(cfg.Config["imap_security"], "imap_security")
	if err != nil {
		return connectionConfig{}, err
	}
	smtpHost, err := normalizeHost(cfg.Config["smtp_host"], "smtp_host")
	if err != nil {
		return connectionConfig{}, err
	}
	smtpPort, err := allowedPort(cfg.Config["smtp_port"], "smtp_port", "25", "465", "587")
	if err != nil {
		return connectionConfig{}, err
	}
	smtpSecurity, err := allowedSecurity(cfg.Config["smtp_security"], "smtp_security")
	if err != nil {
		return connectionConfig{}, err
	}
	username := strings.TrimSpace(cfg.Config["username"])
	if username == "" || containsControl(username) {
		return connectionConfig{}, errors.New("email connector requires a non-empty username")
	}
	password := cfg.Secrets["password"]
	if strings.TrimSpace(password) == "" {
		return connectionConfig{}, errors.New("email connector requires secret password")
	}
	smtpUsernameRaw := cfg.Config["smtp_username"]
	if containsControl(smtpUsernameRaw) {
		return connectionConfig{}, errors.New("email config smtp_username must not contain control characters")
	}
	smtpUsername := strings.TrimSpace(smtpUsernameRaw)
	if smtpUsername == "" {
		smtpUsername = username
	}
	fromAddress, err := resolveFromAddress(cfg.Config["from_address"], username)
	if err != nil {
		return connectionConfig{}, err
	}
	timeout, err := connectionTimeout(cfg.Config["connection_timeout_seconds"])
	if err != nil {
		return connectionConfig{}, err
	}
	if imapSecurity == securityNone && !isLoopbackHost(imapHost) {
		return connectionConfig{}, errors.New("email config imap_security without transport encryption is allowed only for a loopback host")
	}
	if smtpSecurity == securityNone && !isLoopbackHost(smtpHost) {
		return connectionConfig{}, errors.New("email config smtp_security without transport encryption is allowed only for a loopback host")
	}
	return connectionConfig{
		imapHost: imapHost, imapPort: imapPort, imapSecurity: imapSecurity,
		smtpHost: smtpHost, smtpPort: smtpPort, smtpSecurity: smtpSecurity,
		username: username, password: password, smtpUsername: smtpUsername,
		fromAddress: fromAddress, timeout: timeout,
	}, nil
}

func (Connector) ValidateCredential(cfg connectors.RuntimeConfig) error {
	_, err := resolveConnectionConfig(cfg)
	return err
}

func normalizeHost(raw, field string) (string, error) {
	host := strings.TrimSpace(raw)
	host = strings.TrimSuffix(host, ".")
	if host == "" {
		return "", fmt.Errorf("email connector requires config %s", field)
	}
	if containsControl(host) || strings.ContainsAny(host, " @/\\?#[]'\"") {
		return "", fmt.Errorf("email config %s must be a host name or IP address", field)
	}
	if net.ParseIP(host) != nil {
		return host, nil
	}
	if strings.Contains(host, ":") || !validHostname(host) {
		return "", fmt.Errorf("email config %s must be a host name or IP address", field)
	}
	return host, nil
}

func validHostname(host string) bool {
	if len(host) > 253 {
		return false
	}
	for _, label := range strings.Split(host, ".") {
		if len(label) == 0 || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}
		for i := range len(label) {
			character := label[i]
			if !((character >= 'a' && character <= 'z') || (character >= 'A' && character <= 'Z') || (character >= '0' && character <= '9') || character == '-') {
				return false
			}
		}
	}
	return true
}

func allowedPort(raw, field string, allowed ...string) (string, error) {
	port := strings.TrimSpace(raw)
	for _, candidate := range allowed {
		if port == candidate {
			return port, nil
		}
	}
	return "", fmt.Errorf("email config %s must satisfy its declared enum constraint", field)
}

func allowedSecurity(raw, field string) (transportSecurity, error) {
	security := transportSecurity(strings.TrimSpace(raw))
	switch security {
	case securityTLS, securitySTARTTLS, securityNone:
		return security, nil
	default:
		return "", fmt.Errorf("email config %s must satisfy its declared enum constraint", field)
	}
}

func resolveFromAddress(raw, username string) (string, error) {
	if containsControl(raw) {
		return "", errors.New("email config from_address must be an email address (or username must be an email address)")
	}
	from := strings.TrimSpace(raw)
	if from == "" {
		from = username
	}
	address, err := mail.ParseAddress(from)
	if err != nil || address.Address == "" || containsControl(address.Address) {
		return "", errors.New("email config from_address must be an email address (or username must be an email address)")
	}
	return address.Address, nil
}

func connectionTimeout(raw string) (time.Duration, error) {
	if containsControl(raw) {
		return 0, errors.New("email config connection_timeout_seconds must satisfy its declared enum constraint")
	}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return defaultTimeout, nil
	}
	seconds, err := strconv.Atoi(raw)
	if err != nil || (seconds != 5 && seconds != 10 && seconds != 15 && seconds != 30 && seconds != 60) {
		return 0, errors.New("email config connection_timeout_seconds must satisfy its declared enum constraint")
	}
	return time.Duration(seconds) * time.Second, nil
}

func containsControl(value string) bool {
	return strings.IndexFunc(value, unicode.IsControl) >= 0
}

func isLoopbackHost(host string) bool {
	if strings.EqualFold(strings.TrimSuffix(host, "."), "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}
