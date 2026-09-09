package engine

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"polymetrics.ai/internal/safety"
)

const maxCommandEndpointPathBytes = 8192

// ValidateCommandEndpoint validates the fixed provider-operation identity used
// by CLI command mappings and deferred runtime targets. HTTP transports use a
// canonical connector-relative path; GRAPHQL uses a fixed document/field
// identifier rather than a caller-supplied URL.
func ValidateCommandEndpoint(method, path string) error {
	if method == "GRAPHQL" {
		if len(path) > maxCommandEndpointPathBytes || path != strings.TrimSpace(path) {
			return fmt.Errorf("GRAPHQL command target requires a bounded canonical operation identity")
		}
		if err := safety.ValidateIdentifier(path, "GRAPHQL command target"); err != nil {
			return err
		}
		return nil
	}

	switch method {
	case http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodHead:
	default:
		return fmt.Errorf("command target method %q is not one supported canonical method", method)
	}
	if path == "" || path != strings.TrimSpace(path) || len(path) > maxCommandEndpointPathBytes ||
		!strings.HasPrefix(path, "/") || strings.HasPrefix(path, "//") || strings.Contains(path, "//") ||
		strings.ContainsAny(path, "\\?#") || isAbsoluteHTTPURL(path) {
		return fmt.Errorf("command target path %q is not one canonical connector-relative path", path)
	}
	if err := safety.RejectDangerousChars(path, "command target path"); err != nil {
		return err
	}
	for _, segment := range strings.Split(path, "/") {
		decoded, err := url.PathUnescape(segment)
		if err != nil {
			return fmt.Errorf("command target path %q has invalid percent encoding", path)
		}
		if segment == "." || segment == ".." || decoded == "." || decoded == ".." || strings.ContainsAny(decoded, "/\\") {
			return fmt.Errorf("command target path %q has a noncanonical path segment", path)
		}
	}
	return nil
}
