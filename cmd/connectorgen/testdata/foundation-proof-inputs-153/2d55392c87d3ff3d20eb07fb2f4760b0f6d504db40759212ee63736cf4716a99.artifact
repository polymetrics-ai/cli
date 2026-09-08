package connectors

import (
	"fmt"
	"strings"
)

func CanonicalOperationHeaderName(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" || trimmed != raw {
		return "", fmt.Errorf("header parameter is required")
	}
	for _, r := range trimmed {
		if strings.ContainsRune("!#$%&'*+-.^_`|~", r) || (r >= '0' && r <= '9') || (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
			continue
		}
		return "", fmt.Errorf("header parameter contains invalid character %q", r)
	}
	return strings.ToLower(strings.ReplaceAll(trimmed, "_", "-")), nil
}

func IsProtectedOperationHeaderName(name string) bool {
	canonical, err := CanonicalOperationHeaderName(name)
	if err != nil {
		return true
	}
	if _, ok := protectedOperationHeaderNames[canonical]; ok {
		return true
	}
	if strings.HasPrefix(canonical, "content-") ||
		strings.HasPrefix(canonical, "forwarded-") ||
		strings.HasPrefix(canonical, "if-") ||
		strings.HasPrefix(canonical, "proxy-") ||
		strings.HasPrefix(canonical, "sec-") ||
		strings.HasPrefix(canonical, "x-forwarded-") ||
		strings.HasPrefix(canonical, "x-proxy-") {
		return true
	}
	return false
}

var protectedOperationHeaderNames = map[string]struct{}{
	"accept": {}, "accept-charset": {}, "accept-encoding": {}, "accept-language": {},
	"access-control-request-headers": {}, "access-control-request-method": {},
	"alt-svc": {}, "alt-used": {}, "api-key": {}, "api-token": {}, "authorization": {},
	"connection": {}, "cookie": {}, "dnt": {}, "early-data": {}, "expect": {},
	// Retry/idempotency is owned by the sealed runtime request plan. A
	// declaration may not publish either provider-conventional spelling until
	// it can prove the exact preview-bound value survives every retry policy.
	"idempotency-key": {}, "x-idempotency-key": {},
	"forwarded": {}, "host": {}, "keep-alive": {}, "max-forwards": {}, "origin": {},
	"priority": {}, "proxy-authorization": {}, "proxy-connection": {}, "range": {},
	"referer": {}, "set-cookie": {}, "te": {}, "trailer": {}, "transfer-encoding": {},
	"upgrade": {}, "user-agent": {}, "via": {}, "x-access-token": {}, "x-api-key": {},
	"x-api-token": {}, "x-auth-token": {}, "x-authentication-token": {},
	"x-authorization": {}, "x-client-secret": {}, "x-original-url": {}, "x-real-ip": {},
	"x-rewrite-url": {}, "x-secret-key": {}, "x-session-token": {}, "x-token": {},
	"method-override": {}, "x-http-method": {}, "x-http-method-override": {},
	"x-method": {}, "x-method-override": {}, "x-override-method": {},
}
