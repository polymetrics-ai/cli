package cli

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/safety"
)

type connectorsListFlags struct {
	All []string
}

type connectorsCatalogFlags struct {
	Capabilities []string
	Stages       []string
	Types        []string
}

func newConnectorsCobraCommand(ctx context.Context, root string, stdout io.Writer, jsonOut bool) *cobra.Command {
	return newConnectorsCobraCommandWithRuntime(ctx, root, stdout, jsonOut, defaultCertifyCommandRuntime{})
}

func newConnectorsCobraCommandWithRuntime(ctx context.Context, root string, stdout io.Writer, jsonOut bool, runtime certifyCommandRuntime) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "connectors",
		Args:              cobra.ArbitraryArgs,
		SilenceErrors:     true,
		SilenceUsage:      true,
		ValidArgsFunction: completeNoFile,
		RunE: func(_ *cobra.Command, args []string) error {
			if len(args) != 0 {
				return errUsage
			}
			return markCobraLegacyError(writeManual("connectors", stdout, jsonOut))
		},
	}
	setManualHelp(cmd, "connectors", stdout, jsonOut)
	cmd.AddCommand(newConnectorsListCobraCommand(stdout, jsonOut))
	cmd.AddCommand(newConnectorsCatalogCobraCommand(stdout, jsonOut))
	cmd.AddCommand(newConnectorsInspectCobraCommand(stdout, jsonOut, "inspect"))
	cmd.AddCommand(newConnectorsInspectCobraCommand(stdout, jsonOut, "man"))
	cmd.AddCommand(newConnectorsInspectCobraCommand(stdout, jsonOut, "docs"))
	cmd.AddCommand(newCertifyCobraCommand(ctx, root, stdout, jsonOut, runtime))
	cmd.AddCommand(newConnectorsHelpCobraCommand(stdout, jsonOut))
	return cmd
}

func newConnectorsListCobraCommand(stdout io.Writer, jsonOut bool) *cobra.Command {
	var flags connectorsListFlags
	cmd := newConnectorsActionCobraCommand("list", func(_ *cobra.Command, args []string) error {
		if firstArgIsHelp(args) {
			return markCobraLegacyError(writeManual("connectors", stdout, jsonOut))
		}
		registry := appRegistry()
		if lastString(flags.All) != "" {
			defs, err := connectorCatalogEntries(registry, "", "", "")
			if err != nil {
				return markCobraLegacyError(err)
			}
			if jsonOut {
				return markCobraLegacyError(writeJSON(stdout, envelope{"kind": "ConnectorCatalog", "count": len(defs), "connectors": defs}))
			}
			for _, item := range defs {
				fmt.Fprintf(stdout, "%s\t%s\tread=%t\twrite=%t\tquery=%t\n", item.Name, item.IntegrationType, item.Capabilities.Read, item.Capabilities.Write, item.Capabilities.Query)
			}
			return nil
		}
		list := registry.List()
		if jsonOut {
			return markCobraLegacyError(writeJSON(stdout, envelope{"kind": "ConnectorList", "connectors": list}))
		}
		for _, item := range list {
			fmt.Fprintf(stdout, "%s\t%s\t%+v\n", item.Name, item.IntegrationType, item.Capabilities)
		}
		return nil
	})
	setManualHelp(cmd, "connectors", stdout, jsonOut)
	addConnectorsStringArrayFlag(cmd, &flags.All, "all", "include the complete connector catalog")
	return cmd
}

func newConnectorsCatalogCobraCommand(stdout io.Writer, jsonOut bool) *cobra.Command {
	var flags connectorsCatalogFlags
	cmd := newConnectorsActionCobraCommand("catalog", func(_ *cobra.Command, args []string) error {
		if firstArgIsHelp(args) {
			return markCobraLegacyError(writeManual("connectors", stdout, jsonOut))
		}
		registry := appRegistry()
		defs, err := connectorCatalogEntries(registry, lastString(flags.Capabilities), lastString(flags.Stages), lastString(flags.Types))
		if err != nil {
			return markCobraLegacyError(err)
		}
		if jsonOut {
			return markCobraLegacyError(writeJSON(stdout, envelope{"kind": "ConnectorCatalog", "count": len(defs), "connectors": defs}))
		}
		for _, item := range defs {
			fmt.Fprintf(stdout, "%s\t%s\tread=%t\twrite=%t\tquery=%t\n", item.Name, item.IntegrationType, item.Capabilities.Read, item.Capabilities.Write, item.Capabilities.Query)
		}
		return nil
	})
	setManualHelp(cmd, "connectors", stdout, jsonOut)
	addConnectorsStringArrayFlag(cmd, &flags.Capabilities, "capability", "filter by read, write, cdc, or query")
	addConnectorsStringArrayFlag(cmd, &flags.Stages, "stage", "filter by release stage")
	addConnectorsStringArrayFlag(cmd, &flags.Types, "type", "removed legacy source/destination filter")
	return cmd
}

func newConnectorsInspectCobraCommand(stdout io.Writer, jsonOut bool, action string) *cobra.Command {
	cmd := newConnectorsActionCobraCommand(action+" <name>", func(cmd *cobra.Command, args []string) error {
		if firstArgIsHelp(args) {
			return markCobraLegacyError(writeManual("connectors", stdout, jsonOut))
		}
		state, ok := cmd.Context().Value(connectorsCommandStateKey{}).(connectorsCommandState)
		if !ok || !state.operandSet {
			return errUsage
		}
		name := state.operand
		if err := safety.ValidateIdentifier(name, "connector"); err != nil {
			return markCobraLegacyError(validationErrorf("%v", err))
		}
		if err := connectors.RejectLegacyConnectorName(name); err != nil {
			return markCobraLegacyError(err)
		}
		connector, ok := appRegistry().Get(name)
		if !ok {
			return markCobraLegacyError(fmt.Errorf("connector %q not found", name))
		}
		if jsonOut {
			return markCobraLegacyError(writeJSON(stdout, envelope{"kind": "Connector", "connector": connectors.MetadataWithIcon(connector.Metadata()), "manifest": connectors.ManifestOf(connector)}))
		}
		fmt.Fprint(stdout, connectors.RenderConnectorManual(connector))
		return nil
	})
	setManualHelp(cmd, "connectors", stdout, jsonOut)
	return cmd
}

func newConnectorsHelpCobraCommand(stdout io.Writer, jsonOut bool) *cobra.Command {
	cmd := newConnectorsActionCobraCommand("help", func(_ *cobra.Command, _ []string) error {
		return markCobraLegacyError(writeManual("connectors", stdout, jsonOut))
	})
	cmd.Hidden = true
	setManualHelp(cmd, "connectors", stdout, jsonOut)
	return cmd
}

func newConnectorsActionCobraCommand(use string, run func(*cobra.Command, []string) error) *cobra.Command {
	return &cobra.Command{
		Use:           use,
		Args:          cobra.ArbitraryArgs,
		SilenceErrors: true,
		SilenceUsage:  true,
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			UnknownFlags: true,
		},
		ValidArgsFunction: completeNoFile,
		RunE:              run,
	}
}

func firstArgIsHelp(args []string) bool {
	return len(args) > 0 && isHelpArg(args[0])
}

func addConnectorsStringArrayFlag(cmd *cobra.Command, target *[]string, name, usage string) {
	cmd.Flags().StringArrayVar(target, name, nil, usage)
	if flag := cmd.Flags().Lookup(name); flag != nil {
		flag.NoOptDefVal = "true"
	}
}

func connectorCatalogEntries(registry *connectors.Registry, capability, stage, legacyType string) ([]connectors.Definition, error) {
	if legacyType != "" {
		return nil, validationErrorf("legacy --type source|destination was removed; use --capability read|write|cdc|query")
	}
	capability = strings.TrimSpace(strings.ToLower(capability))
	switch capability {
	case "", "read", "write", "cdc", "query":
	default:
		return nil, validationErrorf("invalid --capability %q, want read|write|cdc|query", capability)
	}
	stage = strings.TrimSpace(stage)
	defs := registry.CatalogEntries()
	out := make([]connectors.Definition, 0, len(defs))
	for _, def := range defs {
		if stage != "" && def.ReleaseStage != stage {
			continue
		}
		if !definitionHasCapability(registry, def, capability) {
			continue
		}
		out = append(out, def)
	}
	return out, nil
}
