package synccontract

import "strings"

// PublicMode records one accepted public sync-mode spelling and its closed
// contract counterpart.
type PublicMode struct {
	Name                        string
	ContractMode                Mode
	RequiresCursor              bool
	RequiresPrimaryKey          bool
	RequiresIncrementalExecutor bool
	TypedOnly                   bool
	aliases                     []string
}

// PublicModeCapabilities records the stream facts used to project public modes.
type PublicModeCapabilities struct {
	HasPrimaryKey          bool
	HasCursor              bool
	HasIncrementalExecutor bool
}

var publicModes = []PublicMode{
	{
		Name:         "full_refresh_append",
		ContractMode: ModeFullAppend,
	},
	{
		Name:         "full_refresh_overwrite",
		ContractMode: ModeFullOverwrite,
	},
	{
		Name:               "full_refresh_overwrite_deduped",
		ContractMode:       ModeFullOverwrite,
		RequiresCursor:     true,
		RequiresPrimaryKey: true,
		TypedOnly:          true,
		aliases:            []string{"full_refresh_overwrite_dedup", "full_refresh_deduped"},
	},
	{
		Name:                        "incremental_append",
		ContractMode:                ModeIncrementalAppend,
		RequiresCursor:              true,
		RequiresIncrementalExecutor: true,
	},
	{
		Name:                        "incremental_append_deduped",
		ContractMode:                ModeIncrementalDedupe,
		RequiresCursor:              true,
		RequiresPrimaryKey:          true,
		RequiresIncrementalExecutor: true,
		TypedOnly:                   true,
		aliases:                     []string{"incremental_append_dedup"},
	},
}

// PublicModes returns the accepted public modes in stable display order.
func PublicModes() []PublicMode {
	return append([]PublicMode(nil), publicModes...)
}

// PublicModeNames returns accepted public mode spellings in stable display order.
func PublicModeNames() []string {
	names := make([]string, 0, len(publicModes))
	for _, mode := range publicModes {
		names = append(names, mode.Name)
	}
	return names
}

// LookupPublicMode resolves a public spelling, including retained aliases.
func LookupPublicMode(raw string) (PublicMode, bool) {
	name := strings.ToLower(strings.TrimSpace(raw))
	for _, mode := range publicModes {
		if name == mode.Name {
			return mode, true
		}
		for _, alias := range mode.aliases {
			if name == alias {
				return mode, true
			}
		}
	}
	return PublicMode{}, false
}

// SupportedPublicModeNames derives accepted public modes from stream capabilities.
func SupportedPublicModeNames(capabilities PublicModeCapabilities) []string {
	names := make([]string, 0, len(publicModes))
	for _, mode := range publicModes {
		if mode.RequiresPrimaryKey && !capabilities.HasPrimaryKey {
			continue
		}
		if mode.RequiresCursor && !capabilities.HasCursor {
			continue
		}
		if mode.RequiresIncrementalExecutor && !capabilities.HasIncrementalExecutor {
			continue
		}
		names = append(names, mode.Name)
	}
	return names
}

// MaterializingPublicModeNames returns public modes that can use the ordinary
// local materialization path without an admitted closed transport.
func MaterializingPublicModeNames() []string {
	names := make([]string, 0, len(publicModes))
	for _, mode := range publicModes {
		if !mode.TypedOnly {
			names = append(names, mode.Name)
		}
	}
	return names
}

// DefaultPublicModeName chooses a materializing default for the discovered stream capabilities.
func DefaultPublicModeName(capabilities PublicModeCapabilities) string {
	switch {
	case capabilities.HasCursor && capabilities.HasIncrementalExecutor:
		return "incremental_append"
	case capabilities.HasPrimaryKey:
		return "full_refresh_overwrite"
	default:
		return "full_refresh_append"
	}
}
