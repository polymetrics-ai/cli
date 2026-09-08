package main

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/engine"
	"polymetrics.ai/internal/safety"
)

const vNextFlagAliasRule = "pm-control-provider-alias-v1"

// vNextFlagAlias is authoring provenance only. MapsTo is retained verbatim;
// neither runtime execution nor the alias allocator derives a wire key from Name.
type vNextFlagAlias struct {
	SourceID       string
	OperationIndex int
	CommandIndex   int
	FlagIndex      int
	OriginalName   string
	CanonicalName  string
	MapsTo         string
	Rule           string
}

func vNextSharedControl(command engine.CLICommand, flag engine.CLIFlag) bool {
	return connectors.IsSharedCommandControl(command.Intent, command.Stream, connectors.CommandSurfaceFlag{
		Name: flag.Name, MapsTo: flag.MapsTo, Type: flag.Type,
		Required: flag.Required, Repeatable: flag.Repeatable, EnvOnly: flag.EnvOnly,
	})
}

func vNextValidateFlagNames(command engine.CLICommand) error {
	seen := make(map[string]bool, len(command.Flags))
	for i, flag := range command.Flags {
		if err := safety.ValidateIdentifier(flag.Name, "flag name"); err != nil {
			return fmt.Errorf("flag %d: %w", i, err)
		}
		if seen[flag.Name] {
			return fmt.Errorf("duplicate flag name %q", flag.Name)
		}
		seen[flag.Name] = true
	}
	return nil
}

func projectVNextCommandFlags(raw json.RawMessage, sourceID string, operationIndex, commandIndex int) (json.RawMessage, []vNextFlagAlias, error) {
	var command engine.CLICommand
	if err := decodeStrictJSON(raw, &command); err != nil {
		return nil, nil, err
	}
	if err := vNextValidateFlagNames(command); err != nil {
		return nil, nil, err
	}
	occupied := make(map[string]bool, len(command.Flags))
	var conflicts []int
	for i, flag := range command.Flags {
		occupied[flag.Name] = true
		if connectors.ConnectorControlFlag(flag.Name) != connectors.ProviderCommandFlag && !vNextSharedControl(command, flag) {
			conflicts = append(conflicts, i)
		}
	}
	sort.Slice(conflicts, func(i, j int) bool {
		a, b := command.Flags[conflicts[i]], command.Flags[conflicts[j]]
		if a.Name != b.Name {
			return a.Name < b.Name
		}
		return a.MapsTo < b.MapsTo
	})
	var aliases []vNextFlagAlias
	for _, i := range conflicts {
		flag := command.Flags[i]
		name := "provider-" + flag.Name
		for occupied[name] || connectors.ConnectorControlFlag(name) != connectors.ProviderCommandFlag {
			name = "provider-" + name
		}
		if err := safety.ValidateIdentifier(name, "provider alias"); err != nil {
			return nil, nil, fmt.Errorf("reserved flag %q has no valid provider alias: %w", flag.Name, err)
		}
		occupied[name] = true
		aliases = append(aliases, vNextFlagAlias{SourceID: sourceID, OperationIndex: operationIndex, CommandIndex: commandIndex, FlagIndex: i, OriginalName: flag.Name, CanonicalName: name, MapsTo: flag.MapsTo, Rule: vNextFlagAliasRule})
	}
	if len(aliases) == 0 {
		return cloneRawJSON(raw), nil, nil
	}
	// Edit only the name token; raw field values preserve numeric lexemes and
	// explicitly present optional fields rather than round-tripping typed defaults.
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil {
		return nil, nil, err
	}
	var flags []map[string]json.RawMessage
	if err := json.Unmarshal(object["flags"], &flags); err != nil {
		return nil, nil, err
	}
	for _, alias := range aliases {
		flags[alias.FlagIndex]["name"], _ = json.Marshal(alias.CanonicalName)
	}
	var err error
	object["flags"], err = json.Marshal(flags)
	if err != nil {
		return nil, nil, err
	}
	projected, err := json.Marshal(object)
	return projected, aliases, err
}

// validateVNextProjectedCommands is a final-output invariant, independently
// reached by rendering/admission even if a caller bypasses canonicalization.
func validateVNextProjectedCommands(descriptor vNextCanonicalDescriptor) error {
	for i, operation := range descriptor.Operations {
		for j, command := range operation.Commands {
			fail := func(err error) error {
				return vNextGraphError(vNextOperationPointer(i, "commands", fmt.Sprint(j), "flags"), err)
			}
			var spec engine.CLICommand
			if err := decodeStrictJSON(command.Command, &spec); err != nil {
				return fail(err)
			}
			if err := vNextValidateFlagNames(spec); err != nil {
				return fail(err)
			}
			for _, flag := range spec.Flags {
				if connectors.ConnectorControlFlag(flag.Name) != connectors.ProviderCommandFlag && !vNextSharedControl(spec, flag) {
					return fail(fmt.Errorf("reserved PM flag --%s in command %q must use a distinct canonical provider alias for %q", flag.Name, spec.Path, flag.MapsTo))
				}
			}
			if len(command.AuthoredCommand) == 0 {
				return fail(fmt.Errorf("command %q lacks original flag ownership provenance", spec.Path))
			}
			want, aliases, err := projectVNextCommandFlags(command.AuthoredCommand, operation.ID, i, j)
			if err != nil {
				return fail(err)
			}
			if !vNextJSONEquivalent(json.RawMessage(want), command.Command) || !reflect.DeepEqual(aliases, command.FlagAliases) {
				return fail(fmt.Errorf("command %q changed its reserved-name alias or exact authored provider binding", spec.Path))
			}
		}
	}
	return nil
}
