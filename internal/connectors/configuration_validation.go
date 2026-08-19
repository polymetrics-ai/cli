package connectors

// ConfigurationConstraintValidator is an optional connector capability for
// validating the configuration constraints its definition actually declares.
// HasConfigurationConstraints distinguishes a real declared constraint from a
// connector that has no configuration-time constraints at all; callers must
// not treat the mere presence of this interface as validation coverage.
//
// Configuration is deliberately a map of non-secret strings, matching the
// credential configuration boundary. Secret validation and storage remain
// outside this contract.
type ConfigurationConstraintValidator interface {
	HasConfigurationConstraints() bool
	ValidateConfiguration(config map[string]string) error
}

// ValidateConfiguration applies declared configuration constraints only when
// c advertises them. Connectors without a declarative validator, and validators
// whose definition declares no constraints, retain their existing unrestricted
// configuration behavior.
func ValidateConfiguration(c Connector, config map[string]string) error {
	validator, ok := c.(ConfigurationConstraintValidator)
	if !ok || !validator.HasConfigurationConstraints() {
		return nil
	}
	return validator.ValidateConfiguration(config)
}
