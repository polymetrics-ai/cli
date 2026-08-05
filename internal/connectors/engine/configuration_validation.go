package engine

import (
	"fmt"
	"net/url"
	"sort"
	"time"
)

// HasConfigurationConstraints reports whether the root schema declares a
// configuration constraint that can be evaluated against the credential
// boundary's flat map[string]string representation.
func (s *Schema) HasConfigurationConstraints() bool {
	if s == nil || s.node == nil {
		return false
	}
	for _, property := range s.node.properties {
		if property.hasConfigurationConstraints() {
			return true
		}
	}
	return false
}

func (n *schemaNode) hasConfigurationConstraints() bool {
	return len(n.enum) > 0 || n.pattern != nil || n.format != ""
}

// ValidateConfiguration evaluates only declared configuration constraints on
// supplied top-level credential fields. It deliberately does not apply the
// full schema's required/type/additional-properties rules: credentials are
// accepted as a flat string map, and changing those existing semantics is not
// part of configuration-constraint validation.
func (s *Schema) ValidateConfiguration(config map[string]string) error {
	if s == nil || s.node == nil || len(config) == 0 || len(s.node.properties) == 0 {
		return nil
	}

	keys := make([]string, 0, len(config))
	for key := range config {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		property, declared := s.node.properties[key]
		if !declared || !property.hasConfigurationConstraints() {
			continue
		}
		if err := property.validateConfigurationString(config[key], "/"+key); err != nil {
			return err
		}
	}
	return nil
}

func (n *schemaNode) validateConfigurationString(value, path string) error {
	if len(n.enum) > 0 {
		matched := false
		for _, want := range n.enum {
			if enumEquals(value, want) {
				matched = true
				break
			}
		}
		if !matched {
			return fmt.Errorf("%s: value not in enum %v", displayPath(path), n.enum)
		}
	}

	if n.pattern != nil && !n.pattern.MatchString(value) {
		return fmt.Errorf("%s: value does not match pattern %q", displayPath(path), n.pattern.String())
	}

	if n.format == "" {
		return nil
	}
	if !matchesConfigurationFormat(value, n.format) {
		if isSupportedConfigurationFormat(n.format) {
			return fmt.Errorf("%s: value does not match format %q", displayPath(path), n.format)
		}
		return fmt.Errorf("%s: declared format %q is not supported for configuration validation", displayPath(path), n.format)
	}
	return nil
}

func matchesConfigurationFormat(value, format string) bool {
	switch format {
	case "uri":
		parsed, err := url.Parse(value)
		return err == nil && parsed.IsAbs()
	case "date":
		_, err := time.Parse(time.DateOnly, value)
		return err == nil
	case "date-time":
		_, err := time.Parse(time.RFC3339, value)
		return err == nil
	default:
		return false
	}
}

func isSupportedConfigurationFormat(format string) bool {
	switch format {
	case "uri", "date", "date-time":
		return true
	default:
		return false
	}
}
