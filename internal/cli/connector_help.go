package cli

import (
	"fmt"
	"io"
	"strings"

	"polymetrics.ai/internal/connectors"
)

type connectorHelpResolution struct {
	handled bool
	command string
	manual  string
}

// tryWriteConnectorHelp resolves connector discovery before application state,
// credentials, or connector execution are opened. A false return means the
// arguments name an exact command and must continue through normal execution,
// or do not name a connector command surface at all.
func tryWriteConnectorHelp(connectorName string, args []string, stdout io.Writer, jsonOut bool) (bool, error) {
	connector, ok := appRegistry().Get(connectorName)
	if !ok {
		return false, nil
	}
	provider, ok := connector.(connectors.CommandSurfaceProvider)
	if !ok || provider.CommandSurface() == nil {
		return false, nil
	}
	return tryWriteConnectorSurfaceHelp(connectorName, provider.CommandSurface(), args, stdout, jsonOut)
}

func tryWriteConnectorSurfaceHelp(connectorName string, surface *connectors.CommandSurface, args []string, stdout io.Writer, jsonOut bool) (bool, error) {
	resolution, err := resolveConnectorHelp(connectorName, surface, args)
	if err != nil {
		return true, err
	}
	if !resolution.handled {
		return false, nil
	}
	if jsonOut {
		return true, writeJSON(stdout, envelope{
			"kind":    "CommandManual",
			"command": resolution.command,
			"manual":  resolution.manual,
		})
	}
	_, err = fmt.Fprint(stdout, resolution.manual)
	return true, err
}

func resolveConnectorHelp(connectorName string, surface *connectors.CommandSurface, args []string) (connectorHelpResolution, error) {
	if surface == nil {
		return connectorHelpResolution{}, usageErrorf("connector %q has no command help", connectorName)
	}

	explicit := false
	remaining := append([]string(nil), args...)
	if len(remaining) > 0 && remaining[0] == "help" {
		explicit = true
		remaining = remaining[1:]
	}
	if len(remaining) > 0 && isConnectorHelpFlag(remaining[len(remaining)-1]) {
		explicit = true
		remaining = remaining[:len(remaining)-1]
	}

	flags := parseFlags(remaining)
	path := flags.values["_"]
	hasFlags := connectorHelpHasFlags(remaining)
	if len(path) == 0 {
		if len(remaining) > 0 && !explicit {
			return connectorHelpResolution{}, usageErrorf("missing connector command path")
		}
		manual, err := renderConnectorHelpManual(connectorName, surface, nil)
		return connectorHelpResolution{handled: true, command: connectorName, manual: manual}, err
	}

	exact, prefix := connectorHelpPathKind(surface, path)
	if !exact && !prefix {
		return connectorHelpResolution{}, usageErrorf("unknown connector help path %q for %s", strings.Join(path, " "), connectorName)
	}
	if exact && !explicit {
		return connectorHelpResolution{}, nil
	}
	if prefix && hasFlags && !explicit {
		return connectorHelpResolution{}, usageErrorf("connector help prefix %q does not accept command flags", strings.Join(path, " "))
	}

	manual, err := renderConnectorHelpManual(connectorName, surface, path)
	return connectorHelpResolution{
		handled: true,
		command: strings.TrimSpace(connectorName + " " + strings.Join(path, " ")),
		manual:  manual,
	}, err
}

func isConnectorHelpFlag(value string) bool {
	return value == "--help" || value == "-h"
}

func connectorHelpHasFlags(args []string) bool {
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			return true
		}
	}
	return false
}

func connectorHelpPathKind(surface *connectors.CommandSurface, path []string) (exact, prefix bool) {
	for _, command := range surface.Commands {
		fields := strings.Fields(command.Path)
		if !connectorHelpPathPrefix(fields, path) {
			continue
		}
		if len(fields) == len(path) {
			exact = true
		} else {
			prefix = true
		}
	}
	return exact, prefix
}

func connectorHelpPathPrefix(command, prefix []string) bool {
	if len(prefix) > len(command) {
		return false
	}
	for i := range prefix {
		if command[i] != prefix[i] {
			return false
		}
	}
	return true
}

func renderConnectorHelpManual(connectorName string, surface *connectors.CommandSurface, path []string) (string, error) {
	if surface == nil {
		return "", usageErrorf("connector %q has no command help", connectorName)
	}
	exact, prefix := connectorHelpPathKind(surface, path)
	if len(path) > 0 && !exact && !prefix {
		return "", usageErrorf("unknown connector help path %q for %s", strings.Join(path, " "), connectorName)
	}
	if exact {
		for _, command := range surface.Commands {
			if command.Path == strings.Join(path, " ") {
				return renderConnectorLeafManual(connectorName, surface, command), nil
			}
		}
	}
	return renderConnectorPrefixManual(connectorName, surface, path), nil
}

func renderConnectorPrefixManual(connectorName string, surface *connectors.CommandSurface, path []string) string {
	commandName := strings.TrimSpace("pm " + connectorName + " " + strings.Join(path, " "))
	var b strings.Builder
	writeConnectorHelpSection(&b, "NAME", commandName+" - "+surface.Tagline)
	if len(path) == 0 {
		writeConnectorHelpSection(&b, "SYNOPSIS", surface.Usage)
	} else {
		writeConnectorHelpSection(&b, "SYNOPSIS", commandName+" <command> [flags]")
	}
	writeConnectorHelpSection(&b, "DESCRIPTION", surface.Tagline)

	b.WriteString("COMMANDS\n")
	if len(path) == 0 {
		renderConnectorNamespaceCommands(&b, surface)
	} else {
		for _, command := range surface.Commands {
			if connectorHelpPathPrefix(strings.Fields(command.Path), path) {
				fmt.Fprintf(&b, "  %s\n", renderConnectorHelpCommand(command))
			}
		}
	}
	b.WriteByte('\n')
	renderConnectorGlobalFlags(&b, surface.GlobalFlags)
	writeConnectorSafety(&b)
	writeConnectorExitStatus(&b)
	return b.String()
}

func renderConnectorNamespaceCommands(b *strings.Builder, surface *connectors.CommandSurface) {
	rendered := make(map[string]bool, len(surface.Commands))
	for _, group := range surface.Groups {
		if len(group.Commands) == 0 {
			continue
		}
		title := strings.TrimSpace(group.Title)
		if title == "" {
			title = connectorHelpTitle(group.ID)
		}
		fmt.Fprintf(b, "  %s\n", title)
		for _, topic := range group.Commands {
			for _, command := range surface.Commands {
				fields := strings.Fields(command.Path)
				if len(fields) == 0 || fields[0] != topic || rendered[command.Path] {
					continue
				}
				fmt.Fprintf(b, "    %s\n", renderConnectorHelpCommand(command))
				rendered[command.Path] = true
			}
		}
	}
	for _, command := range surface.Commands {
		if rendered[command.Path] {
			continue
		}
		fmt.Fprintf(b, "  %s\n", renderConnectorHelpCommand(command))
	}
}

func renderConnectorLeafManual(connectorName string, surface *connectors.CommandSurface, command connectors.CommandSurfaceCommand) string {
	commandName := "pm " + connectorName + " " + command.Path
	var b strings.Builder
	name := commandName
	if command.Summary != "" {
		name += " - " + command.Summary
	}
	writeConnectorHelpSection(&b, "NAME", name)
	writeConnectorHelpSection(&b, "SYNOPSIS", commandName+" [flags]")
	if command.Summary != "" || command.Notes != "" {
		lines := make([]string, 0, 2)
		if command.Summary != "" {
			lines = append(lines, command.Summary)
		}
		if command.Notes != "" {
			lines = append(lines, command.Notes)
		}
		writeConnectorHelpSection(&b, "DESCRIPTION", strings.Join(lines, "\n"))
	}
	mapping := connectorHelpMapping(command)
	if len(mapping) > 0 {
		writeConnectorHelpSection(&b, "MAPPING", strings.Join(mapping, "\n"))
	}
	renderConnectorFlags(&b, "FLAGS", command.Flags)
	renderConnectorGlobalFlags(&b, surface.GlobalFlags)
	if len(command.Examples) > 0 {
		writeConnectorHelpSection(&b, "EXAMPLES", strings.Join(command.Examples, "\n"))
	}
	if command.OutputPolicy != "" {
		writeConnectorHelpSection(&b, "OUTPUT POLICY", command.OutputPolicy)
	}
	if command.Risk != "" {
		writeConnectorHelpSection(&b, "RISK", command.Risk)
	}
	if command.Approval != "" {
		writeConnectorHelpSection(&b, "APPROVAL", command.Approval)
	}
	writeConnectorSafety(&b)
	writeConnectorExitStatus(&b)
	return b.String()
}

func renderConnectorHelpCommand(command connectors.CommandSurfaceCommand) string {
	line := command.Path
	if command.Summary != "" {
		line += " - " + command.Summary
	}
	if mapping := connectorHelpMapping(command); len(mapping) > 0 {
		line += " [" + strings.Join(mapping, " ") + "]"
	}
	if command.OutputPolicy != "" {
		line += "; output: " + command.OutputPolicy
	}
	if command.Risk != "" {
		line += "; risk: " + command.Risk
	}
	if command.Approval != "" {
		line += "; approval: " + command.Approval
	}
	return line
}

func connectorHelpMapping(command connectors.CommandSurfaceCommand) []string {
	mapping := make([]string, 0, 5)
	if command.Intent != "" {
		mapping = append(mapping, "intent="+command.Intent)
	}
	if command.Availability != "" {
		mapping = append(mapping, "availability="+command.Availability)
	}
	if command.Stream != "" {
		mapping = append(mapping, "stream="+command.Stream)
	}
	if command.Write != "" {
		mapping = append(mapping, "write="+command.Write)
	}
	if command.Operation != "" {
		mapping = append(mapping, "operation="+command.Operation)
	}
	return mapping
}

func renderConnectorGlobalFlags(b *strings.Builder, flags []connectors.CommandSurfaceFlag) {
	renderConnectorFlags(b, "GLOBAL FLAGS", flags)
}

func renderConnectorFlags(b *strings.Builder, title string, flags []connectors.CommandSurfaceFlag) {
	if len(flags) == 0 {
		return
	}
	b.WriteString(title)
	b.WriteByte('\n')
	for _, flag := range flags {
		fmt.Fprintf(b, "  %s\n", renderConnectorHelpFlag(flag))
	}
	b.WriteByte('\n')
}

func renderConnectorHelpFlag(flag connectors.CommandSurfaceFlag) string {
	name := "--" + strings.TrimLeft(flag.Name, "-")
	if flag.Type != "" && flag.Type != "boolean" {
		name += " <" + flag.Type + ">"
	}
	parts := []string{name}
	if flag.Summary != "" {
		parts = append(parts, flag.Summary)
	}
	if len(flag.Values) > 0 {
		parts = append(parts, "values="+strings.Join(flag.Values, "|"))
	}
	if flag.MapsTo != "" {
		parts = append(parts, "maps_to="+flag.MapsTo)
	}
	return strings.Join(parts, " - ")
}

func writeConnectorHelpSection(b *strings.Builder, title, content string) {
	if strings.TrimSpace(content) == "" {
		return
	}
	b.WriteString(title)
	b.WriteByte('\n')
	for _, line := range strings.Split(content, "\n") {
		b.WriteString("  ")
		b.WriteString(line)
		b.WriteByte('\n')
	}
	b.WriteByte('\n')
}

func writeConnectorSafety(b *strings.Builder) {
	writeConnectorHelpSection(b, "SAFETY", "Help is generated from validated connector metadata. It does not read credentials, open project state, contact connector APIs, or execute commands.\nMutation commands retain their declared plan, preview, approval, and execution policy.")
}

func writeConnectorExitStatus(b *strings.Builder) {
	writeConnectorHelpSection(b, "EXIT STATUS", "0  Help rendered successfully.\n2  The connector help path is invalid.")
}

func connectorHelpTitle(value string) string {
	parts := strings.FieldsFunc(value, func(r rune) bool { return r == '-' || r == '_' })
	for i, part := range parts {
		if part == "" {
			continue
		}
		parts[i] = strings.ToUpper(part[:1]) + part[1:]
	}
	return strings.Join(parts, " ")
}
