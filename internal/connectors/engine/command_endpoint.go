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
			return diagnosticAt("/path", "command_target_graphql_shape_invalid", "GRAPHQL command target requires a bounded canonical operation identity", fmt.Errorf("GRAPHQL command target requires a bounded canonical operation identity"))
		}
		if err := safety.ValidateIdentifier(path, "GRAPHQL command target"); err != nil {
			return diagnosticAt("/path", "command_target_graphql_identifier_invalid", "GRAPHQL command target must be a safe operation identifier", err)
		}
		return nil
	}

	switch method {
	case http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodHead:
	default:
		return diagnosticAt("/method", "command_target_method_invalid", "command target method is not supported", fmt.Errorf("command target method %q is not one supported canonical method", method))
	}
	if path == "" || path != strings.TrimSpace(path) || len(path) > maxCommandEndpointPathBytes ||
		!strings.HasPrefix(path, "/") || strings.HasPrefix(path, "//") || strings.Contains(path, "//") ||
		strings.ContainsAny(path, "\\?#") || isAbsoluteHTTPURL(path) {
		return diagnosticAt("/path", "command_target_path_invalid", "command target path must be bounded, canonical, and connector-relative", fmt.Errorf("command target path %q is not one canonical connector-relative path", path))
	}
	if err := safety.RejectDangerousChars(path, "command target path"); err != nil {
		return diagnosticAt("/path", "command_target_path_characters_invalid", "command target path contains forbidden characters", err)
	}
	for _, segment := range strings.Split(path, "/") {
		decoded, err := url.PathUnescape(segment)
		if err != nil {
			return diagnosticAt("/path", "command_target_path_encoding_invalid", "command target path contains invalid percent encoding", fmt.Errorf("command target path %q has invalid percent encoding", path))
		}
		if segment == "." || segment == ".." || decoded == "." || decoded == ".." || strings.ContainsAny(decoded, "/\\") {
			return diagnosticAt("/path", "command_target_path_segment_invalid", "command target path contains a noncanonical segment", fmt.Errorf("command target path %q has a noncanonical path segment", path))
		}
	}
	return nil
}
