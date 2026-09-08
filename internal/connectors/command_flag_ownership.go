package connectors

// CommandFlagOwner is the closed PM keyword classification. Navigation and
// bootstrap controls have different consumers from connector lifecycle controls.
type CommandFlagOwner uint8

const (
	ProviderCommandFlag CommandFlagOwner = iota
	PMConnectorControl
	PMGlobalControl
	PMNavigationControl
)

// ConnectorControlFlag is the single reserved-word inventory shared by canonical
// authoring and CLI preparation. It does not decide provider request mappings.
func ConnectorControlFlag(name string) CommandFlagOwner {
	switch name {
	case "_", "credential", "connection", "config", "limit", "max-bytes", "plan", "preview", "approve", "approval-token-stdin", "confirm", "plan-name", "dest-root", "file-name", "from-env":
		return PMConnectorControl
	case "root", "json", "help":
		return PMGlobalControl
	case "page", "page-cursor":
		return PMNavigationControl
	default:
		return ProviderCommandFlag
	}
}

// IsSharedCommandControl recognizes the existing stream row-limit declaration.
// A numeric provider query/body limit is a different contract and must be aliased.
func IsSharedCommandControl(intent, stream string, flag CommandSurfaceFlag) bool {
	return stream != "" && (intent == "etl" || intent == "direct_read") &&
		flag.Name == "limit" && flag.MapsTo == "limit" && flag.Type == "integer" &&
		!flag.Required && !flag.Repeatable && !flag.EnvOnly
}
