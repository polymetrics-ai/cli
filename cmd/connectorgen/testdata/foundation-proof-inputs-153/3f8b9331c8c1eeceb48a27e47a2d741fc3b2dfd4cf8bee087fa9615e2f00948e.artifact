package safety

import (
	"fmt"
	"net/netip"
	"net/url"
	"strconv"
	"strings"
)

const (
	maxProviderCitationURLBytes   = 4096
	maxProviderCitationQueryBytes = 1024
	maxProviderCitationQueryKeys  = 16
	maxProviderCitationQueryKey   = 64
	maxProviderCitationQueryValue = 256
)

var providerCitationNonPublicPrefixes = []netip.Prefix{
	netip.MustParsePrefix("0.0.0.0/8"),
	netip.MustParsePrefix("100.64.0.0/10"),
	netip.MustParsePrefix("192.0.0.0/24"),
	netip.MustParsePrefix("192.0.2.0/24"),
	netip.MustParsePrefix("192.88.99.0/24"),
	netip.MustParsePrefix("198.18.0.0/15"),
	netip.MustParsePrefix("198.51.100.0/24"),
	netip.MustParsePrefix("203.0.113.0/24"),
	netip.MustParsePrefix("240.0.0.0/4"),
	netip.MustParsePrefix("100::/64"),
	netip.MustParsePrefix("100:0:0:1::/64"),
	netip.MustParsePrefix("2001::/23"),
	netip.MustParsePrefix("2001:2::/48"),
	netip.MustParsePrefix("2001:db8::/32"),
	netip.MustParsePrefix("2002::/16"),
	netip.MustParsePrefix("fec0::/10"),
}

// CanonicalProviderCitationURL validates a source citation without resolving
// or fetching it and returns its one canonical identity spelling. Callers that
// persist authored evidence must compare the result with the stored value and
// reject a mismatch instead of silently rewriting provenance.
func CanonicalProviderCitationURL(raw string) (string, error) {
	if raw == "" || raw != strings.TrimSpace(raw) || len(raw) > maxProviderCitationURLBytes {
		return "", fmt.Errorf("provider citation URL must be a bounded absolute HTTPS URL")
	}
	if err := RejectDangerousChars(raw, "provider citation URL"); err != nil {
		return "", err
	}
	parsed, err := url.Parse(raw)
	if err != nil || !parsed.IsAbs() || !strings.EqualFold(parsed.Scheme, "https") || parsed.Host == "" || parsed.Opaque != "" {
		return "", fmt.Errorf("provider citation URL must be an absolute HTTPS URL")
	}
	if parsed.User != nil {
		return "", fmt.Errorf("provider citation URL must not include userinfo")
	}
	if parsed.Fragment != "" || parsed.RawFragment != "" || strings.Contains(raw, "#") {
		return "", fmt.Errorf("provider citation URL must not include a fragment")
	}
	if parsed.ForceQuery {
		return "", fmt.Errorf("provider citation URL must not include an empty query marker")
	}

	host, err := canonicalProviderCitationHost(parsed)
	if err != nil {
		return "", err
	}
	path, err := canonicalProviderCitationEscapedPath(parsed.EscapedPath())
	if err != nil {
		return "", err
	}
	query, err := canonicalProviderCitationQuery(parsed.RawQuery)
	if err != nil {
		return "", err
	}

	canonical := "https://" + host + path
	if query != "" {
		canonical += "?" + query
	}
	if len(canonical) > maxProviderCitationURLBytes {
		return "", fmt.Errorf("provider citation URL exceeds the canonical length limit")
	}
	return canonical, nil
}

func canonicalProviderCitationHost(parsed *url.URL) (string, error) {
	hostname := parsed.Hostname()
	if hostname == "" || strings.HasSuffix(hostname, ".") || strings.Contains(hostname, "%") {
		return "", fmt.Errorf("provider citation URL must name an unambiguous host without a trailing dot")
	}

	port := parsed.Port()
	canonicalPort := ""
	if port != "" {
		number, err := strconv.Atoi(port)
		if err != nil || number <= 0 || number > 65535 {
			return "", fmt.Errorf("provider citation URL has an invalid HTTPS port")
		}
		if number != 443 {
			canonicalPort = ":" + strconv.Itoa(number)
		}
	}

	if literal, err := netip.ParseAddr(hostname); err == nil {
		literal = literal.Unmap()
		if !providerCitationIPIsPublic(literal) {
			return "", fmt.Errorf("provider citation URL destination must be public")
		}
		canonicalHost := literal.String()
		if literal.Is6() {
			canonicalHost = "[" + canonicalHost + "]"
		}
		return canonicalHost + canonicalPort, nil
	}

	if !providerCitationDNSHost(hostname) {
		return "", fmt.Errorf("provider citation URL must name an unambiguous DNS host")
	}
	return strings.ToLower(hostname) + canonicalPort, nil
}

func providerCitationDNSHost(host string) bool {
	if len(host) > 253 || !strings.Contains(host, ".") || strings.EqualFold(host, "localhost") || strings.HasSuffix(strings.ToLower(host), ".localhost") {
		return false
	}
	numeric := true
	for _, label := range strings.Split(host, ".") {
		if len(label) == 0 || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}
		for index := 0; index < len(label); index++ {
			character := label[index]
			if character < '0' || character > '9' {
				numeric = false
			}
			if (character >= 'a' && character <= 'z') || (character >= 'A' && character <= 'Z') || (character >= '0' && character <= '9') || character == '-' {
				continue
			}
			return false
		}
	}
	return !numeric
}

func providerCitationIPIsPublic(address netip.Addr) bool {
	if !address.IsValid() || !address.IsGlobalUnicast() || address.IsPrivate() || address.IsLoopback() || address.IsLinkLocalUnicast() || address.IsLinkLocalMulticast() || address.IsMulticast() || address.IsUnspecified() {
		return false
	}
	for _, prefix := range providerCitationNonPublicPrefixes {
		if prefix.Contains(address) {
			return false
		}
	}
	return true
}

func canonicalProviderCitationEscapedPath(path string) (string, error) {
	if path == "" {
		return "/", nil
	}
	var canonical strings.Builder
	canonical.Grow(len(path))
	for index := 0; index < len(path); index++ {
		if path[index] != '%' {
			canonical.WriteByte(path[index])
			continue
		}
		if index+2 >= len(path) {
			return "", fmt.Errorf("provider citation URL has an invalid escaped path")
		}
		high, highOK := providerCitationHexValue(path[index+1])
		low, lowOK := providerCitationHexValue(path[index+2])
		if !highOK || !lowOK {
			return "", fmt.Errorf("provider citation URL has an invalid escaped path")
		}
		decoded := high<<4 | low
		if providerCitationUnreserved(decoded) {
			canonical.WriteByte(decoded)
		} else {
			const upperHex = "0123456789ABCDEF"
			canonical.WriteByte('%')
			canonical.WriteByte(upperHex[decoded>>4])
			canonical.WriteByte(upperHex[decoded&0x0f])
		}
		index += 2
	}
	normalized := canonical.String()
	for _, segment := range strings.Split(normalized, "/") {
		if segment == "." || segment == ".." {
			return "", fmt.Errorf("provider citation URL path must not contain dot segments")
		}
	}
	return normalized, nil
}

func providerCitationHexValue(character byte) (byte, bool) {
	switch {
	case character >= '0' && character <= '9':
		return character - '0', true
	case character >= 'a' && character <= 'f':
		return character - 'a' + 10, true
	case character >= 'A' && character <= 'F':
		return character - 'A' + 10, true
	default:
		return 0, false
	}
}

func providerCitationUnreserved(character byte) bool {
	return (character >= 'a' && character <= 'z') ||
		(character >= 'A' && character <= 'Z') ||
		(character >= '0' && character <= '9') ||
		character == '-' || character == '.' || character == '_' || character == '~'
}

func canonicalProviderCitationQuery(raw string) (string, error) {
	if raw == "" {
		return "", nil
	}
	if len(raw) > maxProviderCitationQueryBytes {
		return "", fmt.Errorf("provider citation URL query exceeds the citation policy")
	}
	query, err := url.ParseQuery(raw)
	if err != nil || len(query) == 0 || len(query) > maxProviderCitationQueryKeys {
		return "", fmt.Errorf("provider citation URL query violates the citation policy")
	}
	for key, values := range query {
		if key == "" || len(key) > maxProviderCitationQueryKey || len(values) != 1 || len(values[0]) > maxProviderCitationQueryValue || providerCitationCredentialLikeQueryKey(strings.ToLower(key)) {
			return "", fmt.Errorf("provider citation URL query violates the citation policy")
		}
		if err := RejectDangerousChars(key, "provider citation URL query key"); err != nil {
			return "", err
		}
		if err := RejectDangerousChars(values[0], "provider citation URL query value"); err != nil {
			return "", err
		}
	}
	return query.Encode(), nil
}

func providerCitationCredentialLikeQueryKey(key string) bool {
	key = strings.ReplaceAll(key, "-", "_")
	for _, prohibited := range []string{"token", "secret", "password", "credential", "authorization", "api_key", "apikey", "signature", "sig", "key"} {
		if key == prohibited || strings.HasSuffix(key, "_"+prohibited) || strings.HasPrefix(key, prohibited+"_") {
			return true
		}
	}
	return false
}
