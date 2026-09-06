package engine

import (
	"fmt"
	"net"
	"net/url"
	"path"
	"strings"
)

// TenantOriginSpec selects one connection-level HTTPS origin and an optional
// source-declared API path. It never accepts a command-level route override.
type TenantOriginSpec struct {
	ConfigKey         string `json:"config_key"`
	AppendPath        string `json:"append_path,omitempty"`
	AllowLoopbackHTTP bool   `json:"allow_loopback_http,omitempty"`
}

func resolveTenantOrigin(spec TenantOriginSpec, config map[string]string) (string, error) {
	if strings.TrimSpace(spec.ConfigKey) == "" {
		return "", fmt.Errorf("tenant origin requires config_key")
	}
	raw := strings.TrimSpace(config[spec.ConfigKey])
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Host == "" || parsed.RawPath != "" || parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", fmt.Errorf("tenant origin %q must be an origin without encoded path, query, or fragment", spec.ConfigKey)
	}
	if parsed.Scheme != "https" && (!spec.AllowLoopbackHTTP || parsed.Scheme != "http" || !isLoopbackHost(parsed.Hostname())) {
		return "", fmt.Errorf("tenant origin %q must use HTTPS or declared loopback HTTP", spec.ConfigKey)
	}
	if parsed.User != nil {
		return "", fmt.Errorf("tenant origin %q must not contain user info", spec.ConfigKey)
	}
	appendPath := ""
	if spec.AppendPath != "" {
		appendPath = path.Clean(spec.AppendPath)
		if !strings.HasPrefix(spec.AppendPath, "/") || appendPath != spec.AppendPath || appendPath == "/" {
			return "", fmt.Errorf("tenant origin append_path is unsafe")
		}
	}
	configuredPath := strings.TrimRight(parsed.Path, "/")
	if configuredPath != "" && configuredPath != appendPath {
		return "", fmt.Errorf("tenant origin %q must not contain an undeclared path prefix", spec.ConfigKey)
	}
	parsed.Path = appendPath
	return strings.TrimRight(parsed.String(), "/"), nil
}

func isLoopbackHost(host string) bool {
	return host == "localhost" || (net.ParseIP(host) != nil && net.ParseIP(host).IsLoopback())
}
