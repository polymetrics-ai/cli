package connectors

import (
	"fmt"
	"net/url"
	stdpath "path"
	"strings"

	"polymetrics.ai/internal/safety"
)

func EncodePathParameter(name, value string) (string, error) {
	if name == "path" {
		if strings.Contains(value, "\\") {
			return "", fmt.Errorf("path variable %q must use forward slashes", name)
		}
		if err := safety.ValidateRelativePath(value, "path variable "+name); err != nil {
			return "", err
		}
		clean := stdpath.Clean(value)
		if clean == "." {
			return "", fmt.Errorf("path variable %q is required", name)
		}
		parts := strings.Split(clean, "/")
		for i, part := range parts {
			parts[i] = url.PathEscape(part)
		}
		return strings.Join(parts, "/"), nil
	}
	if err := safety.ValidateIdentifier(value, "path variable "+name); err != nil {
		return "", err
	}
	return url.PathEscape(value), nil
}

func CanonicalRequestTargetPath(path string) string {
	for decodedPasses := 0; decodedPasses < 8; decodedPasses++ {
		decoded, err := url.PathUnescape(path)
		if err != nil || decoded == path {
			break
		}
		path = decoded
	}
	return canonicalRequestTargetPercentEscapes(strings.ReplaceAll(path, "\\", "/"))
}

func CanonicalRequestTargetPathIdentity(path string) string {
	return stdpath.Clean("/" + strings.TrimLeft(CanonicalRequestTargetPath(path), "/"))
}

func canonicalRequestTargetPercentEscapes(path string) string {
	var canonical strings.Builder
	canonical.Grow(len(path))
	for i := 0; i < len(path); i++ {
		if path[i] != '%' || i+2 >= len(path) {
			canonical.WriteByte(path[i])
			continue
		}
		high, highOK := requestTargetHexValue(path[i+1])
		low, lowOK := requestTargetHexValue(path[i+2])
		if !highOK || !lowOK {
			canonical.WriteByte(path[i])
			continue
		}
		value := high<<4 | low
		if isRequestTargetUnreserved(value) {
			canonical.WriteByte(value)
		} else {
			canonical.WriteByte('%')
			canonical.WriteByte(uppercaseRequestTargetHex(path[i+1]))
			canonical.WriteByte(uppercaseRequestTargetHex(path[i+2]))
		}
		i += 2
	}
	return canonical.String()
}

func requestTargetHexValue(value byte) (byte, bool) {
	switch {
	case value >= '0' && value <= '9':
		return value - '0', true
	case value >= 'a' && value <= 'f':
		return value - 'a' + 10, true
	case value >= 'A' && value <= 'F':
		return value - 'A' + 10, true
	default:
		return 0, false
	}
}

func isRequestTargetUnreserved(value byte) bool {
	return (value >= 'a' && value <= 'z') || (value >= 'A' && value <= 'Z') || (value >= '0' && value <= '9') || strings.ContainsRune("-._~", rune(value))
}

func uppercaseRequestTargetHex(value byte) byte {
	if value >= 'a' && value <= 'f' {
		return value - ('a' - 'A')
	}
	return value
}
