package connectors

import (
	"fmt"
	"sort"
	"strings"
)

type GuideSection struct {
	Title string
	Lines []string
}

type GuideExample struct {
	Title   string
	Command string
	Notes   []string
}

type GuideLink struct {
	Label string
	URL   string
}

type ConnectorGuide struct {
	Name        string
	DisplayName string
	Summary     string
	Sections    []GuideSection
	Examples    []GuideExample
	Links       []GuideLink
	AgentNotes  []string
}

type GuideProvider interface {
	Guide() ConnectorGuide
}

// DynamicPollingWatermarkProvider marks a native connector whose effective
// polling declaration is bound from a live catalog. Its static bundle may
// remain planned when it cannot truthfully name a fixed cursor or tie-breaker;
// inspection must not then claim that no polling behavior exists at all.
// Runtime execution still goes through PollingPreflight for every stream.
type DynamicPollingWatermarkProvider interface {
	HasDynamicPollingWatermark() bool
}

func GuideOf(c Connector) ConnectorGuide {
	manifest := ManifestOf(c)
	var guide ConnectorGuide
	if provider, ok := c.(GuideProvider); ok {
		guide = provider.Guide()
	} else {
		guide = guideFromManifest(manifest)
	}
	if provider, ok := c.(CommandSurfaceProvider); ok {
		guide = guideWithCommandSurface(guide, provider.CommandSurface())
	}
	guide = guideWithSyncTransport(guide, c)
	guide = guideWithPollingWatermark(guide, c)
	return guideWithIcon(guide, manifest)
}

func guideWithIcon(guide ConnectorGuide, manifest Manifest) ConnectorGuide {
	for _, section := range guide.Sections {
		if strings.EqualFold(section.Title, "icon") {
			return guide
		}
	}
	guide.Sections = append([]GuideSection{iconSection(manifest)}, guide.Sections...)
	return guide
}

func guideWithCommandSurface(guide ConnectorGuide, surface *CommandSurface) ConnectorGuide {
	if surface == nil || strings.TrimSpace(surface.Usage) == "" {
		return guide
	}
	for _, section := range guide.Sections {
		if strings.EqualFold(section.Title, "command surface") {
			return guide
		}
	}
	guide.Sections = append(guide.Sections, commandSurfaceSection(surface))
	return guide
}

// guideWithSyncTransport projects only an authored closed transport descriptor.
// A missing descriptor stays absent from the human manual; JSON inspection
// still reports both roles as unsupported. This avoids representing ordinary
// read/write capabilities as a transport declaration.
func guideWithSyncTransport(guide ConnectorGuide, connector Connector) ConnectorGuide {
	if _, declared := SyncTransportDescriptorOf(connector); !declared {
		return guide
	}
	for _, section := range guide.Sections {
		if strings.EqualFold(section.Title, "sync transport") {
			return guide
		}
	}
	eligibility := SyncTransportEligibilityOf(connector)
	lines := []string{
		"Source transport: " + eligibility.Source.Status,
		"Destination transport: " + eligibility.Destination.Status,
		"A declared transport still requires runtime preflight and externally verified conformance; it is not a certification claim.",
	}
	if eligibility.Source.Executor != nil {
		lines = append(lines, "Source executor: "+string(eligibility.Source.Executor.Family)+"/"+eligibility.Source.Executor.ID)
	}
	if eligibility.Destination.Executor != nil {
		lines = append(lines, "Destination executor: "+string(eligibility.Destination.Executor.Family)+"/"+eligibility.Destination.Executor.ID)
	}
	guide.Sections = append(guide.Sections, GuideSection{Title: "Sync Transport", Lines: lines})
	return guide
}

// guideWithPollingWatermark exposes only declaration status. Mode execution is
// intentionally not inferred here: the real preflight also needs the selected
// catalog object and destination binding, neither of which inspection reads.
func guideWithPollingWatermark(guide ConnectorGuide, connector Connector) ConnectorGuide {
	definition, ok := DefinitionOf(connector)
	if !ok || definition.PollingWatermark == nil {
		return guide
	}
	for _, section := range guide.Sections {
		if strings.EqualFold(section.Title, "polling watermark") {
			return guide
		}
	}
	declaration := definition.PollingWatermark
	lines := []string{
		"Static declaration status: " + string(declaration.Status),
		"Mechanism: polling_watermark is a bounded polling scan, not CDC or change capture.",
	}
	dynamic, _ := connector.(DynamicPollingWatermarkProvider)
	if dynamic != nil && dynamic.HasDynamicPollingWatermark() {
		lines = append(lines, "Runtime eligibility: this connector constructs an implemented declaration per selected catalog object. Every requested mode still requires runtime preflight for its destination binding, registered native executors, and immutable conformance evidence.")
	} else {
		lines = append(lines, "Runtime eligibility: a static declaration alone does not implement a polling mode. Every requested mode requires runtime preflight for its selected catalog object, destination binding, registered native executors, and immutable conformance evidence.")
	}
	if declaration.Reason != "" {
		lines = append(lines, "Reason: "+declaration.Reason)
	}
	if declaration.Status != PollingWatermarkStatusImplemented && (dynamic == nil || !dynamic.HasDynamicPollingWatermark()) {
		lines = append(lines, "No polling source ordering, checkpoint, snapshot, deletion, or rebootstrap behavior is implemented for this connector while the declaration is non-implemented.")
	}
	guide.Sections = append(guide.Sections, GuideSection{Title: "Polling Watermark", Lines: lines})
	return guide
}

func commandSurfaceSection(surface *CommandSurface) GuideSection {
	lines := []string{}
	if surface.Tagline != "" {
		lines = append(lines, surface.Tagline)
	}
	lines = append(lines, "Usage: "+surface.Usage)
	if surface.SourceCLI != nil && surface.SourceCLI.Name != "" {
		source := "Source CLI: " + surface.SourceCLI.Name
		if surface.SourceCLI.Reference != "" {
			source += " (" + surface.SourceCLI.Reference + ")"
		} else if surface.SourceCLI.Docs != "" {
			source += " (" + surface.SourceCLI.Docs + ")"
		}
		lines = append(lines, source)
	}
	if commandSurfaceHasPMExecutionLimits(surface) {
		lines = append(lines, "PM execution policy "+RequestContractExecutionPolicyVersion+": each max N bytes qualifier is the effective PM request limit, not a provider schema assertion; path/query values are measured after exact wire encoding and rejected rather than truncated.")
	}
	if len(surface.GlobalFlags) > 0 {
		lines = append(lines, "Global flags:")
		for _, flag := range surface.GlobalFlags {
			lines = append(lines, "  "+renderCommandSurfaceFlag(flag))
		}
	}

	commandsByPrefix := map[string][]CommandSurfaceCommand{}
	for _, cmd := range surface.Commands {
		prefix := commandPrefix(cmd.Path)
		commandsByPrefix[prefix] = append(commandsByPrefix[prefix], cmd)
	}
	rendered := map[string]bool{}
	for _, group := range surface.Groups {
		if len(group.Commands) == 0 {
			continue
		}
		title := valueOrDefault(group.Title, titleCase(group.ID))
		lines = append(lines, title)
		for _, prefix := range group.Commands {
			for _, cmd := range commandsByPrefix[prefix] {
				lines = append(lines, "  "+renderCommandSurfaceCommand(cmd))
				rendered[cmd.Path] = true
			}
		}
	}

	var extra []CommandSurfaceCommand
	for _, cmd := range surface.Commands {
		if !rendered[cmd.Path] {
			extra = append(extra, cmd)
		}
	}
	if len(extra) > 0 {
		lines = append(lines, "Other Commands")
		for _, cmd := range extra {
			lines = append(lines, "  "+renderCommandSurfaceCommand(cmd))
		}
	}

	if len(surface.HelpTopics) > 0 {
		lines = append(lines, "Help topics:")
		for _, topic := range surface.HelpTopics {
			if topic.Name == "" {
				continue
			}
			line := topic.Name
			if topic.Summary != "" {
				line += " - " + topic.Summary
			}
			lines = append(lines, "  "+line)
		}
	}
	return GuideSection{Title: "Command Surface", Lines: lines}
}

func renderCommandSurfaceFlag(flag CommandSurfaceFlag) string {
	name := renderCommandSurfaceFlagName(flag)
	parts := []string{name}
	if flag.Type != "" {
		parts[0] += " (" + flag.Type + ")"
	}
	if flag.Summary != "" {
		parts = append(parts, flag.Summary)
	}
	if len(flag.Values) > 0 {
		parts = append(parts, "values="+strings.Join(flag.Values, "|"))
	}
	if flag.MapsTo != "" {
		parts = append(parts, "maps_to="+flag.MapsTo)
	}
	return strings.Join(parts, ": ")
}

// renderCommandSurfaceFlagName is the concise command-list form. It keeps the
// safety contract visible without repeating ordinary type/summary/mapping
// metadata hundreds of times in a generated manual; that complete metadata is
// still retained in JSON/website projection and in the detailed global flags.
func renderCommandSurfaceFlagName(flag CommandSurfaceFlag) string {
	name := "--" + strings.TrimLeft(flag.Name, "-")
	qualifiers := []string{}
	if flag.Required {
		qualifiers = append(qualifiers, "required")
	}
	if flag.Repeatable {
		qualifiers = append(qualifiers, "repeatable")
	}
	if flag.EnvOnly {
		qualifiers = append(qualifiers, "env-only")
	}
	if flag.AllowEmpty != nil && !*flag.AllowEmpty {
		qualifiers = append(qualifiers, "non-empty")
	}
	if flag.MinItems > 0 && flag.MaxItems > 0 {
		qualifiers = append(qualifiers, fmt.Sprintf("%d..%d items", flag.MinItems, flag.MaxItems))
	} else if flag.MinItems > 0 {
		qualifiers = append(qualifiers, fmt.Sprintf("min %d items", flag.MinItems))
	} else if flag.MaxItems > 0 {
		qualifiers = append(qualifiers, fmt.Sprintf("max %d items", flag.MaxItems))
	}
	if flag.MaxBytes > 0 {
		qualifiers = append(qualifiers, fmt.Sprintf("max %d bytes", flag.MaxBytes))
	}
	if flag.Minimum != nil {
		qualifiers = append(qualifiers, "minimum="+flag.Minimum.String())
	}
	if flag.Maximum != nil {
		qualifiers = append(qualifiers, "maximum="+flag.Maximum.String())
	}
	if flag.Format != "" {
		qualifiers = append(qualifiers, "format="+flag.Format)
	}
	if flag.AllowBareString {
		qualifiers = append(qualifiers, "allow-bare-string")
	}
	if len(qualifiers) > 0 {
		name += " (" + strings.Join(qualifiers, ", ") + ")"
	}
	return name
}

func commandSurfaceHasPMExecutionLimits(surface *CommandSurface) bool {
	for _, flag := range surface.GlobalFlags {
		if flag.MaxBytes > 0 && flag.MaxBytesOrigin == "pm_policy" && flag.MaxBytesPolicyVersion != "" {
			return true
		}
	}
	for _, command := range surface.Commands {
		for _, flag := range command.Flags {
			if flag.MaxBytes > 0 && flag.MaxBytesOrigin == "pm_policy" && flag.MaxBytesPolicyVersion != "" {
				return true
			}
		}
	}
	return false
}

// commandSurfaceRenderedFlags returns the flags a command actually accepts.
//
// A binary_download command also accepts the runtime's destination flags, one
// of which is required. They are declared in no bundle, so a renderer showing
// only the bundle's own flags documents a command that cannot be run as
// written. A flag the bundle already declares is never repeated.
func commandSurfaceRenderedFlags(cmd CommandSurfaceCommand) []CommandSurfaceFlag {
	var runtimeFlags []CommandSurfaceFlag
	switch cmd.Intent {
	case "binary_download", "text_export":
		runtimeFlags = BinaryDownloadFlags()
	case "direct_read":
		if strings.TrimSpace(cmd.Stream) != "" {
			return cmd.Flags
		}
		runtimeFlags = DirectReadPageFlags()
	default:
		return cmd.Flags
	}
	declared := make(map[string]bool, len(cmd.Flags))
	for _, flag := range cmd.Flags {
		declared[flag.Name] = true
	}
	rendered := append([]CommandSurfaceFlag(nil), cmd.Flags...)
	for _, flag := range runtimeFlags {
		if !declared[flag.Name] {
			rendered = append(rendered, flag)
		}
	}
	return rendered
}

func renderCommandSurfaceCommand(cmd CommandSurfaceCommand) string {
	line := cmd.Path
	if cmd.Summary != "" {
		line += " - " + cmd.Summary
	}
	meta := []string{}
	if cmd.Intent != "" {
		meta = append(meta, "intent="+cmd.Intent)
	}
	if cmd.Availability != "" {
		meta = append(meta, "availability="+cmd.Availability)
	}
	if cmd.Stream != "" {
		meta = append(meta, "stream="+cmd.Stream)
	}
	if cmd.Write != "" {
		meta = append(meta, "write="+cmd.Write)
	}
	if cmd.Operation != "" {
		meta = append(meta, "operation="+cmd.Operation)
	}
	if cmd.Availability == "unsupported_local" || cmd.Intent == "local_workflow" {
		meta = append(meta, "unsupported local workflow")
	}
	if len(meta) > 0 {
		line += " [" + strings.Join(meta, " ") + "]"
	}
	if cmd.Approval != "" {
		line += "; approval: " + normalizeCommandSurfaceSentence(cmd.Approval)
	}
	if cmd.Risk != "" {
		line += "; risk: " + cmd.Risk
	}
	if cmd.Notes != "" {
		line += "; notes: " + cmd.Notes
	}
	rendered := commandSurfaceRenderedFlags(cmd)
	if len(rendered) > 0 {
		flags := make([]string, 0, len(rendered))
		for _, flag := range rendered {
			flags = append(flags, renderCommandSurfaceFlagName(flag))
		}
		line += "; flags: " + strings.Join(flags, ", ")
	}
	return line
}

func commandPrefix(path string) string {
	fields := strings.Fields(path)
	if len(fields) == 0 {
		return ""
	}
	return fields[0]
}

func normalizeCommandSurfaceSentence(value string) string {
	value = strings.TrimSpace(value)
	if strings.HasPrefix(value, "Reverse ETL ") {
		return "reverse ETL " + strings.TrimPrefix(value, "Reverse ETL ")
	}
	return value
}

func RenderConnectorManual(c Connector) string {
	return RenderGuideManual(GuideOf(c))
}

func RenderConnectorSkill(c Connector) string {
	return RenderGuideSkill(GuideOf(c))
}

func ValidateConnectorGuide(c Connector) error {
	manifest := ManifestOf(c)
	guide := GuideOf(c)
	if strings.TrimSpace(guide.Name) == "" {
		return fmt.Errorf("connector %q guide missing name", c.Name())
	}
	if strings.TrimSpace(guide.Summary) == "" {
		return fmt.Errorf("connector %q guide missing summary", c.Name())
	}
	if manifest.Metadata.Icon == nil {
		return fmt.Errorf("connector %q guide missing icon metadata", c.Name())
	}
	manual := RenderGuideManual(guide)
	for _, required := range []string{"NAME", "SYNOPSIS", "DESCRIPTION", "ICON", "CAPABILITIES", "CONFIGURATION", "SECURITY", "AGENT WORKFLOW"} {
		if !strings.Contains(manual, required) {
			return fmt.Errorf("connector %q guide missing section %s", c.Name(), required)
		}
	}
	if len(manifest.Streams) > 0 && !strings.Contains(manual, "ETL STREAMS") {
		return fmt.Errorf("connector %q guide missing ETL streams", c.Name())
	}
	if len(manifest.WriteActions) > 0 && !strings.Contains(manual, "REVERSE ETL ACTIONS") {
		return fmt.Errorf("connector %q guide missing reverse ETL actions", c.Name())
	}
	if len(manifest.AuthModes) > 0 && !strings.Contains(manual, "AUTHENTICATION") {
		return fmt.Errorf("connector %q guide missing authentication", c.Name())
	}
	if len(guide.Examples) == 0 {
		return fmt.Errorf("connector %q guide missing examples", c.Name())
	}
	skill := RenderGuideSkill(guide)
	if !strings.Contains(skill, "name: pm-"+guide.Name) || !strings.Contains(skill, "## Agent Rules") {
		return fmt.Errorf("connector %q guide skill is incomplete", c.Name())
	}
	return nil
}

func RenderGuideManual(guide ConnectorGuide) string {
	var b strings.Builder
	b.WriteString("NAME\n")
	_, _ = fmt.Fprintf(&b, "  pm connectors inspect %s - %s connector manual\n\n", guide.Name, guide.DisplayName)
	b.WriteString("SYNOPSIS\n")
	_, _ = fmt.Fprintf(&b, "  pm connectors inspect %s\n", guide.Name)
	_, _ = fmt.Fprintf(&b, "  pm connectors inspect %s --json\n", guide.Name)
	_, _ = fmt.Fprintf(&b, "  pm credentials add <name> --connector %s [--config key=value] [--from-env field=ENV] [--value-stdin field]\n\n", guide.Name)
	b.WriteString("DESCRIPTION\n")
	for _, line := range splitParagraphs(guide.Summary) {
		b.WriteString("  " + line + "\n")
	}
	b.WriteString("\n")
	for _, section := range guide.Sections {
		title := strings.TrimSpace(section.Title)
		if title == "" {
			continue
		}
		b.WriteString(strings.ToUpper(title) + "\n")
		for _, line := range section.Lines {
			if strings.TrimSpace(line) == "" {
				b.WriteString("\n")
				continue
			}
			b.WriteString("  " + line + "\n")
		}
		b.WriteString("\n")
	}
	if len(guide.Examples) > 0 {
		b.WriteString("EXAMPLES\n")
		for _, example := range guide.Examples {
			if example.Title != "" {
				b.WriteString("  # " + example.Title + "\n")
			}
			for _, line := range strings.Split(strings.TrimSpace(example.Command), "\n") {
				b.WriteString("  " + line + "\n")
			}
			for _, note := range example.Notes {
				b.WriteString("  " + note + "\n")
			}
			b.WriteString("\n")
		}
	}
	if len(guide.AgentNotes) > 0 {
		b.WriteString("AGENT WORKFLOW\n")
		for _, note := range guide.AgentNotes {
			b.WriteString("  - " + note + "\n")
		}
		b.WriteString("\n")
	}
	if len(guide.Links) > 0 {
		b.WriteString("SEE ALSO\n")
		for _, link := range guide.Links {
			_, _ = fmt.Fprintf(&b, "  %s: %s\n", link.Label, link.URL)
		}
		b.WriteString("\n")
	}
	b.WriteString("EXIT STATUS\n")
	b.WriteString("  0 success\n")
	b.WriteString("  1 runtime error\n")
	b.WriteString("  2 usage error\n")
	return b.String()
}

func RenderGuideSkill(guide ConnectorGuide) string {
	var b strings.Builder
	b.WriteString("---\n")
	b.WriteString("name: pm-" + guide.Name + "\n")
	b.WriteString("description: " + guide.DisplayName + " connector knowledge and safe action guide.\n")
	b.WriteString("---\n\n")
	b.WriteString("# pm-" + guide.Name + "\n\n")
	b.WriteString("## Purpose\n\n")
	b.WriteString(guide.Summary + "\n\n")
	for _, section := range guide.Sections {
		if len(section.Lines) == 0 {
			continue
		}
		b.WriteString("## " + titleCase(section.Title) + "\n\n")
		for _, line := range section.Lines {
			if strings.TrimSpace(line) == "" {
				b.WriteString("\n")
				continue
			}
			if strings.HasPrefix(line, "  ") {
				b.WriteString("  - " + strings.TrimSpace(line) + "\n")
				continue
			}
			b.WriteString("- " + line + "\n")
		}
		b.WriteString("\n")
	}
	if len(guide.Examples) > 0 {
		b.WriteString("## Commands\n\n")
		for _, example := range guide.Examples {
			if example.Title != "" {
				b.WriteString("### " + example.Title + "\n\n")
			}
			b.WriteString("```bash\n" + strings.TrimSpace(example.Command) + "\n```\n\n")
		}
	}
	if len(guide.AgentNotes) > 0 {
		b.WriteString("## Agent Rules\n\n")
		for _, note := range guide.AgentNotes {
			b.WriteString("- " + note + "\n")
		}
		b.WriteString("\n")
	}
	if len(guide.Links) > 0 {
		b.WriteString("## References\n\n")
		for _, link := range guide.Links {
			b.WriteString("- [" + link.Label + "](" + link.URL + ")\n")
		}
	}
	return strings.TrimRight(b.String(), "\n") + "\n"
}

func guideFromManifest(manifest Manifest) ConnectorGuide {
	sections := []GuideSection{
		iconSection(manifest),
		capabilitySection(manifest),
		authSection(manifest),
		configSection(manifest),
		streamSection(manifest),
		syncModeSection(manifest),
		writeActionSection(manifest),
		paginationSection(manifest),
		securitySection(manifest),
	}
	return ConnectorGuide{
		Name:        manifest.Metadata.Name,
		DisplayName: valueOrDefault(manifest.Metadata.DisplayName, manifest.Metadata.Name),
		Summary:     manifest.Metadata.Description,
		Sections:    compactSections(sections),
		Examples:    examplesForManifest(manifest),
		Links:       linksForManifest(manifest),
		AgentNotes:  agentNotesForManifest(manifest),
	}
}

func capabilitySection(manifest Manifest) GuideSection {
	lines := []string{
		fmt.Sprintf("check=%t catalog=%t read=%t write=%t query=%t", manifest.Metadata.Capabilities.Check, manifest.Metadata.Capabilities.Catalog, manifest.Metadata.Capabilities.Read, manifest.Metadata.Capabilities.Write, manifest.Metadata.Capabilities.Query),
		"Integration type: " + manifest.Metadata.IntegrationType,
	}
	return GuideSection{Title: "Capabilities", Lines: lines}
}

func iconSection(manifest Manifest) GuideSection {
	icon := manifest.Metadata.Icon
	if icon == nil {
		return GuideSection{Title: "Icon", Lines: []string{"No icon metadata is registered for this connector."}}
	}
	lines := []string{
		"id: " + icon.ID,
		"asset: " + icon.Path,
	}
	if icon.Title != "" {
		lines = append(lines, "title: "+icon.Title)
	}
	if icon.SimpleIconSlug != "" {
		lines = append(lines, "simple_icon_slug: "+icon.SimpleIconSlug)
	}
	if icon.SimpleIconHex != "" {
		lines = append(lines, "simple_icon_hex: "+icon.SimpleIconHex)
	}
	lines = append(lines, "source: "+icon.Source)
	if icon.License != "" {
		lines = append(lines, "license: "+icon.License)
	}
	if icon.Attribution != "" {
		lines = append(lines, "attribution: "+icon.Attribution)
	}
	lines = append(lines, "review_status: "+icon.ReviewStatus)
	if icon.ReviewURL != "" {
		lines = append(lines, "review_url: "+icon.ReviewURL)
	}
	if icon.Match != "" {
		lines = append(lines, "match: "+icon.Match)
	}
	if icon.MatchedBy != "" {
		lines = append(lines, "matched_by: "+icon.MatchedBy)
	}
	return GuideSection{Title: "Icon", Lines: lines}
}

func authSection(manifest Manifest) GuideSection {
	if len(manifest.AuthModes) == 0 {
		if len(manifest.SecretFields) == 0 {
			return GuideSection{Title: "Authentication", Lines: []string{"No secret authentication is required for this connector."}}
		}
		return GuideSection{Title: "Authentication", Lines: []string{"Use pm credentials add with --from-env or --value-stdin for secret fields."}}
	}
	lines := []string{}
	for _, mode := range manifest.AuthModes {
		lines = append(lines, fmt.Sprintf("%s: %s", mode.Name, mode.Description))
		if len(mode.ConfigFields) > 0 {
			lines = append(lines, "  config: "+strings.Join(mode.ConfigFields, ", "))
		}
		if len(mode.SecretFields) > 0 {
			lines = append(lines, "  secrets: "+strings.Join(mode.SecretFields, ", "))
		}
		lines = append(lines, fmt.Sprintf("  supports: read=%t write=%t", mode.Read, mode.Write))
	}
	return GuideSection{Title: "Authentication", Lines: lines}
}

func configSection(manifest Manifest) GuideSection {
	lines := []string{}
	for _, field := range manifest.ConfigFields {
		line := field.Name
		if field.Required {
			line += " (required)"
		}
		if field.Default != "" {
			line += " default=" + field.Default
		}
		if field.Description != "" {
			line += ": " + field.Description
		}
		lines = append(lines, line)
	}
	for _, field := range manifest.SecretFields {
		line := field.Name + " (secret)"
		if field.Required {
			line += " (required)"
		} else if field.RequiredWhen != "" {
			line += " (required when " + field.RequiredWhen + ")"
		}
		if field.Description != "" {
			line += ": " + field.Description
		}
		lines = append(lines, line)
	}
	if len(lines) == 0 {
		lines = append(lines, "No connector-specific config fields.")
	}
	return GuideSection{Title: "Configuration", Lines: lines}
}

func streamSection(manifest Manifest) GuideSection {
	if len(manifest.Streams) == 0 {
		return GuideSection{}
	}
	lines := []string{}
	for _, stream := range manifest.Streams {
		lines = append(lines, namedDescriptionLine(stream.Name, stream.Description))
		if len(stream.PrimaryKey) > 0 {
			lines = append(lines, "  primary key: "+strings.Join(stream.PrimaryKey, ", "))
		}
		if len(stream.CursorFields) > 0 {
			lines = append(lines, "  cursor: "+strings.Join(stream.CursorFields, ", "))
		}
		if len(stream.Fields) > 0 {
			fields := make([]string, 0, len(stream.Fields))
			for _, field := range stream.Fields {
				fields = append(fields, field.Name+"("+field.Type+")")
			}
			lines = append(lines, "  fields: "+strings.Join(fields, ", "))
		}
	}
	return GuideSection{Title: "ETL Streams", Lines: lines}
}

func namedDescriptionLine(name, description string) string {
	description = strings.TrimSpace(description)
	if description == "" {
		return name + ":"
	}
	return name + ": " + description
}

func syncModeSection(manifest Manifest) GuideSection {
	if len(manifest.SyncModes) == 0 && len(manifest.DestinationSyncModes) == 0 && len(manifest.SourceSyncModes) == 0 {
		return GuideSection{}
	}
	lines := []string{}
	if len(manifest.SyncModes) > 0 {
		lines = append(lines, "ETL sync modes: "+strings.Join(manifest.SyncModes, ", "))
	}
	if len(manifest.SourceSyncModes) > 0 {
		lines = append(lines, "Source modes: "+strings.Join(manifest.SourceSyncModes, ", "))
	}
	if len(manifest.DestinationSyncModes) > 0 {
		lines = append(lines, "Destination modes: "+strings.Join(manifest.DestinationSyncModes, ", "))
	}
	return GuideSection{Title: "Sync Modes", Lines: lines}
}

// writeActionSection renders one line per required/optional field set,
// including the either/or constraint ("A, or B together with C") that
// Manifest.RequiredFields cannot express on its own. That shape is carried by
// RequiredAnyFields — amazon-sqs set_queue_attributes and tag_queue are the
// connectors that declare it — so `pm docs generate --dir docs/cli` reproduces
// their committed MANUAL.md/SKILL.md and stays idempotent.
func writeActionSection(manifest Manifest) GuideSection {
	if len(manifest.WriteActions) == 0 {
		return GuideSection{}
	}
	lines := []string{}
	for _, action := range manifest.WriteActions {
		lines = append(lines, namedDescriptionLine(action.Name, action.Description))
		if action.Method != "" || action.Path != "" {
			lines = append(lines, "  endpoint: "+strings.TrimSpace(action.Method+" "+action.Path))
		}
		if required := requiredFieldsLine(action); required != "" {
			lines = append(lines, "  required fields: "+required)
		}
		if len(action.OptionalFields) > 0 {
			lines = append(lines, "  optional fields: "+strings.Join(action.OptionalFields, ", "))
		}
		if action.Risk != "" {
			lines = append(lines, "  risk: "+action.Risk)
		}
		// Only non-batchable actions render a line, so connectors that never
		// declare the field keep byte-identical help and generated docs.
		if !action.IsBatchable() {
			lines = append(lines, "  bulk reverse ETL: refused (non-batchable; run it as its own pm command, one record at a time)")
		}
	}
	return GuideSection{Title: "Reverse ETL Actions", Lines: lines}
}

// requiredFieldsLine renders the full required-field contract: the always-required
// names first, then each either/or group as "a + b", joined by " or ".
func requiredFieldsLine(action WriteActionSpec) string {
	parts := append([]string(nil), action.RequiredFields...)
	if len(action.RequiredAnyFields) > 0 {
		groups := make([]string, 0, len(action.RequiredAnyFields))
		for _, group := range action.RequiredAnyFields {
			groups = append(groups, strings.Join(group, " + "))
		}
		parts = append(parts, strings.Join(groups, " or "))
	}
	return strings.Join(parts, ", ")
}

func paginationSection(manifest Manifest) GuideSection {
	if manifest.Pagination.Type == "" {
		return GuideSection{}
	}
	lines := []string{"type: " + manifest.Pagination.Type}
	if manifest.Pagination.PageSizeField != "" {
		lines = append(lines, "page size field: "+manifest.Pagination.PageSizeField)
	}
	if manifest.Pagination.PageLimitField != "" {
		lines = append(lines, "page limit field: "+manifest.Pagination.PageLimitField)
	}
	if manifest.Pagination.DefaultLimit != "" {
		lines = append(lines, "default limit: "+manifest.Pagination.DefaultLimit)
	}
	return GuideSection{Title: "Pagination", Lines: lines}
}

func securitySection(manifest Manifest) GuideSection {
	lines := []string{}
	if manifest.Risk.Read != "" {
		lines = append(lines, "read risk: "+manifest.Risk.Read)
	}
	if manifest.Risk.Write != "" {
		lines = append(lines, "write risk: "+manifest.Risk.Write)
	}
	if manifest.Risk.Mutation != "" {
		lines = append(lines, "mutation risk: "+manifest.Risk.Mutation)
	}
	if manifest.Risk.Approval != "" {
		lines = append(lines, "approval: "+manifest.Risk.Approval)
	}
	lines = append(lines, "Never pass secret values in chat, shell arguments, logs, docs, or JSON output.")
	return GuideSection{Title: "Security", Lines: lines}
}

func examplesForManifest(manifest Manifest) []GuideExample {
	name := manifest.Metadata.Name
	examples := []GuideExample{
		{Title: "Inspect as a manual", Command: "pm connectors inspect " + name},
		{Title: "Inspect as structured JSON", Command: "pm connectors inspect " + name + " --json"},
	}
	switch name {
	case "bahmni":
		examples = append(examples,
			GuideExample{Title: "Command discovery", Command: "pm bahmni --help\npm bahmni appointments --help\npm bahmni appointments create --help"},
			GuideExample{Title: "Synthetic appointment read", Command: "pm bahmni appointments list --credential bahmni-local --config appointment_date=2026-01-01T00:00:00.000 --limit 10 --json"},
			GuideExample{Title: "Synthetic patient create plan", Command: "pm bahmni patients create --credential bahmni-local --identifier SYN-CONN-EXAMPLE-001 --identifier-type <identifier-type-uuid> --identifier-location <location-uuid> --given-name Synthetic --family-name Connector --gender O --birthdate 1990-01-01 --preview --json"},
			GuideExample{Title: "Unsupported retained as blocked", Command: "pm bahmni appointments reschedule --help\npm bahmni drug_orders create --help"},
		)
	case "github":
		examples = append(examples,
			GuideExample{Title: "Public repository credential", Command: "pm credentials add github-public --connector github --config owner=octocat --config repo=Hello-World --config public_access=true"},
			GuideExample{Title: "Token credential", Command: "export GITHUB_TOKEN=...\npm credentials add github-token --connector github --config owner=OWNER --config repo=REPO --from-env token=GITHUB_TOKEN"},
			GuideExample{Title: "GitHub App credential", Command: "pm credentials add github-app --connector github --config owner=OWNER --config repo=REPO --config auth_type=github_app --config app_id=12345 --config installation_id=67890 --value-stdin private_key < app-private-key.pem"},
			GuideExample{Title: "Pull request ETL", Command: "pm connections create github_prs_to_warehouse --source github:github-token --destination warehouse:warehouse-local --stream pull_requests --primary-key node_id --cursor updated_at --table github_pull_requests\npm etl run --connection github_prs_to_warehouse --stream pull_requests --batch-size 100 --json"},
			GuideExample{Title: "Approved pull request creation", Command: "pm reverse plan prs_to_github --source-table github_pr_candidates --destination github:github-token --action create_pull_request --map title:title --map body:body --map head:head --map base:base --map reviewers:reviewers\npm reverse preview <plan-id> --json\npm reverse run <plan-id> --approval-token-stdin --json"},
		)
	case "sample":
		examples = append(examples, GuideExample{Title: "Sample ETL", Command: "pm credentials add sample-local --connector sample\npm connections create sample_to_warehouse --source sample:sample-local --destination warehouse:warehouse-local --stream customers --primary-key id --cursor updated_at --table sample_customers\npm etl run --connection sample_to_warehouse --stream customers --json"})
	case "file":
		examples = append(examples, GuideExample{Title: "File ETL", Command: "pm credentials add file-local --connector file --config path=/path/to/records.jsonl\npm connections create file_to_warehouse --source file:file-local --destination warehouse:warehouse-local --stream file --table imported_records\npm etl run --connection file_to_warehouse --stream file --json"})
	case "warehouse":
		examples = append(examples, GuideExample{Title: "Warehouse credential", Command: "pm credentials add warehouse-local --connector warehouse --config path=$ROOT/.polymetrics/warehouse\npm query run --table sample_customers --limit 5 --json"})
	case "outbox":
		examples = append(examples, GuideExample{Title: "Outbox reverse ETL", Command: "pm credentials add outbox-local --connector outbox --config path=$ROOT/.polymetrics/outbox\npm reverse plan customers_to_outbox --source-table sample_customers --destination outbox:outbox-local --map id:external_id --map email:email\npm reverse run <plan-id> --approval-token-stdin --json"})
	}
	return examples
}

func linksForManifest(manifest Manifest) []GuideLink {
	switch manifest.Metadata.Name {
	case "github":
		return []GuideLink{
			{Label: "GitHub REST authentication", URL: "https://docs.github.com/en/rest/authentication/authenticating-to-the-rest-api"},
			{Label: "GitHub App installation auth", URL: "https://docs.github.com/en/apps/creating-github-apps/authenticating-with-a-github-app/authenticating-as-a-github-app-installation"},
			{Label: "GitHub pull requests REST API", URL: "https://docs.github.com/en/rest/pulls/pulls"},
			{Label: "GitHub issues REST API", URL: "https://docs.github.com/en/rest/issues/issues"},
			{Label: "GitHub issue comments REST API", URL: "https://docs.github.com/en/rest/issues/comments"},
			{Label: "GitHub labels REST API", URL: "https://docs.github.com/en/rest/issues/labels"},
			{Label: "GitHub commits REST API", URL: "https://docs.github.com/en/rest/commits/commits"},
			{Label: "GitHub branches REST API", URL: "https://docs.github.com/en/rest/branches/branches"},
			{Label: "GitHub releases REST API", URL: "https://docs.github.com/en/rest/releases/releases"},
			{Label: "GitHub Actions workflows REST API", URL: "https://docs.github.com/en/rest/actions/workflows"},
			{Label: "GitHub Actions workflow runs REST API", URL: "https://docs.github.com/en/rest/actions/workflow-runs"},
			{Label: "GitHub Actions artifacts REST API", URL: "https://docs.github.com/en/rest/actions/artifacts"},
			{Label: "GitHub repository contents REST API", URL: "https://docs.github.com/en/rest/repos/contents"},
		}
	default:
		return nil
	}
}

func agentNotesForManifest(manifest Manifest) []string {
	notes := []string{
		"Run pm connectors inspect " + manifest.Metadata.Name + " before creating credentials or plans.",
		"Use --json only when the caller needs structured output; use the manual for human-readable guidance.",
		"Never ask the user to paste secret values into chat.",
	}
	if len(manifest.WriteActions) > 0 {
		notes = append(notes, "For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.")
	}
	return notes
}

func compactSections(sections []GuideSection) []GuideSection {
	out := make([]GuideSection, 0, len(sections))
	for _, section := range sections {
		if strings.TrimSpace(section.Title) == "" || len(section.Lines) == 0 {
			continue
		}
		out = append(out, section)
	}
	return out
}

func splitParagraphs(value string) []string {
	lines := []string{}
	for _, line := range strings.Split(strings.TrimSpace(value), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			lines = append(lines, line)
		}
	}
	if len(lines) == 0 {
		return []string{"No description provided."}
	}
	return lines
}

func titleCase(value string) string {
	parts := strings.Fields(strings.ToLower(value))
	for i, part := range parts {
		switch part {
		case "etl":
			parts[i] = "ETL"
		default:
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return strings.Join(parts, " ")
}

func valueOrDefault(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func SortGuides(guides []ConnectorGuide) {
	sort.Slice(guides, func(i, j int) bool { return guides[i].Name < guides[j].Name })
}
