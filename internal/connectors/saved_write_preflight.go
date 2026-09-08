package connectors

// SavedWriteActionPreflighter is an optional zero-I/O contract for a selected
// typed saved-write runtime. Individual command execution remains separate.
type SavedWriteActionPreflighter interface {
	PreflightSavedWriteAction(name string) error
}
