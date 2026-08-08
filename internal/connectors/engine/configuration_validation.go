package engine

import (
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"
)

// HasConfigurationConstraints reports whether the root schema declares a
// configuration constraint that can be evaluated against the credential
// boundary's flat map[string]string representation.
func (s *Schema) HasConfigurationConstraints() bool {
	if s == nil || s.node == nil {
		return false
	}
	for name, property := range s.node.properties {
		if property.hasConfigurationConstraints() || s.admitsRequiredConfigurationKey(name) {
			return true
		}
	}
	return false
}

func (n *schemaNode) hasConfigurationConstraints() bool {
	return len(n.enum) > 0 || n.pattern != nil || n.format != ""
}

// admitsRequiredConfigurationKey reports whether name is a root property whose
// absence from a credential's configuration is checkable at the credential
// boundary: declared required, non-secret (secrets are supplied through a
// separate map and never appear in config), and carrying no "default" (which
// materializeConfigDefaults supplies for the caller, so an omitted value is
// complete, not missing).
func (s *Schema) admitsRequiredConfigurationKey(name string) bool {
	property, declared := s.node.properties[name]
	if !declared || property.secret || property.hasDefault {
		return false
	}
	for _, req := range s.node.required {
		if req == name {
			return true
		}
	}
	return false
}

// ValidateConfiguration evaluates declared configuration constraints, and the
// presence of declared-required keys, on supplied top-level credential fields.
// It deliberately does not apply the full schema's type/additional-properties
// rules: credentials are accepted as a flat string map, and changing those
// existing semantics is not part of configuration-constraint validation.
//
// Required-key presence IS applied, because the alternative is admitting a
// credential the connector can never use: dockerhub declares `namespace`
// required and interpolates it into every stream path and the connection
// check, so a credential saved without it fails at read/check time with a
// connector-internal template error, long after the point where the operator
// could have been told what was missing. Only the keys this boundary can
// actually see are checked (admitsRequiredConfigurationKey): a required secret
// lives in the separate secrets map, and a required property with a declared
// default is filled in by the engine rather than the caller.
//
// Supplied values are checked before omitted ones: a caller who typed a value
// this schema rejects is told what is wrong with what they typed, not handed a
// different key's absence first.
func (s *Schema) ValidateConfiguration(config map[string]string) error {
	if s == nil || s.node == nil || len(s.node.properties) == 0 {
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
	return s.validateRequiredConfiguration(config)
}

func (s *Schema) validateRequiredConfiguration(config map[string]string) error {
	for _, name := range s.node.required {
		if !s.admitsRequiredConfigurationKey(name) {
			continue
		}
		if strings.TrimSpace(config[name]) == "" {
			return fmt.Errorf("%s: required property missing", displayPath("/"+name))
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
