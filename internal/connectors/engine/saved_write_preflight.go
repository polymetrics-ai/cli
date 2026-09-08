package engine

import (
	"errors"
	"fmt"
)

// ErrSavedWriteNonBatchable identifies a valid individual-only action. It is
// distinct from malformed typed execution, which admission must not ignore.
var ErrSavedWriteNonBatchable = errors.New("write action is not batchable")

// ErrSavedWriteNoInputBatchability identifies a valid no-input action whose
// declaration does not explicitly support bounded saved batch delivery.
var ErrSavedWriteNoInputBatchability = errors.New("no-input action lacks explicit saved batchability")

// PreflightSavedWriteAction checks the loaded declarative action without
// acquiring credentials, a warehouse, approvals or a provider connection.
func (c *Connector) PreflightSavedWriteAction(name string) error {
	action, err := findWriteAction(c.bundle, name)
	if err != nil {
		return err
	}
	if !c.bundle.Metadata.Capabilities.Write {
		return fmt.Errorf("saved action %q requires selected write capability", name)
	}
	if !action.IsBatchable() {
		return fmt.Errorf("saved action %q: %w", name, ErrSavedWriteNonBatchable)
	}
	if len(action.RecordSchema) == 0 {
		return fmt.Errorf("saved action %q lacks a concrete record schema", name)
	}
	if legacyWriteHookClaimsAction(c.hooks, action) {
		return fmt.Errorf("saved action %q requires a closed selected hook preflight contract", name)
	}
	if action.Hook != "" {
		_, prepared := c.hooks.(PreparedWriteHook)
		classifier, classified := c.hooks.(WriteHookClassifier)
		// Record mappers retain the ordinary declaration-owned encoder and
		// frozen prepared request; they cannot select physical request fields.
		_, recordMapper := c.hooks.(WriteRecordHook)
		if !recordMapper && (!prepared || !classified || !classifier.HandlesWriteAction(action)) {
			return fmt.Errorf("saved action %q requires a closed selected hook preflight contract", name)
		}
	}
	if _, err := compiledRecordSchema(action); err != nil {
		return err
	}
	if err := validateWriteBodies(c.bundle.Writes); err != nil {
		return err
	}
	shape, err := InspectRecordSchema(action.RecordSchema)
	if err != nil {
		return err
	}
	if shape.AdmitsOnlyEmptyObject && writeActionConsumesNoRecord(action) && (action.Batchable == nil || !*action.Batchable) {
		return fmt.Errorf("saved no-input action %q: %w", name, ErrSavedWriteNoInputBatchability)
	}
	return c.PreflightWriteAction(name)
}
