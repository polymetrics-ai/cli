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
	name := "--" + strings.TrimLeft(flag.Name, "-")
	if flag.Type != "" {
		name += " (" + flag.Type + ")"
	}
	meta := []string{}
	if len(flag.Values) > 0 {
		meta = append(meta, "values="+strings.Join(flag.Values, "|"))
	}
	if flag.MapsTo != "" {
		meta = append(meta, "maps_to="+flag.MapsTo)
	}
	if flag.MaxItems > 0 {
		meta = append(meta, fmt.Sprintf("max_items=%d", flag.MaxItems))
	}
	parts := []string{}
	if flag.Summary != "" {
		summary := flag.Summary
		if len(meta) > 0 {
			summary = strings.TrimRight(summary, ".")
		}
		parts = append(parts, summary)
	}
	parts = append(parts, meta...)
	if len(parts) == 0 {
		return name
	}
	return name + ": " + strings.Join(parts, "; ")
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
	if cmd.Intent == "local_workflow" {
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
	if len(cmd.Flags) > 0 {
		flags := make([]string, 0, len(cmd.Flags))
		for _, flag := range cmd.Flags {
			flags = append(flags, "--"+strings.TrimLeft(flag.Name, "-"))
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
	b.WriteString(fmt.Sprintf("  pm connectors inspect %s - %s connector manual\n\n", guide.Name, guide.DisplayName))
	b.WriteString("SYNOPSIS\n")
	b.WriteString(fmt.Sprintf("  pm connectors inspect %s\n", guide.Name))
	b.WriteString(fmt.Sprintf("  pm connectors inspect %s --json\n", guide.Name))
	b.WriteString(fmt.Sprintf("  pm credentials add <name> --connector %s [--config key=value] [--from-env field=ENV] [--value-stdin field]\n\n", guide.Name))
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
			b.WriteString(fmt.Sprintf("  %s: %s\n", link.Label, link.URL))
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
		"asset: " + icon.Path,
		"source: " + icon.Source,
		"review_status: " + icon.ReviewStatus,
	}
	if icon.ReviewURL != "" {
		lines = append(lines, "review_url: "+icon.ReviewURL)
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
		lines = append(lines, stream.Name+": "+stream.Description)
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

func writeActionSection(manifest Manifest) GuideSection {
	if len(manifest.WriteActions) == 0 {
		return GuideSection{}
	}
	lines := []string{}
	for _, action := range manifest.WriteActions {
		lines = append(lines, action.Name+": "+action.Description)
		if action.Method != "" || action.Path != "" {
			lines = append(lines, "  endpoint: "+strings.TrimSpace(action.Method+" "+action.Path))
		}
		if len(action.RequiredFields) > 0 {
			lines = append(lines, "  required fields: "+strings.Join(action.RequiredFields, ", "))
		}
		if len(action.OptionalFields) > 0 {
			lines = append(lines, "  optional fields: "+strings.Join(action.OptionalFields, ", "))
		}
		if action.Risk != "" {
			lines = append(lines, "  risk: "+action.Risk)
		}
	}
	return GuideSection{Title: "Reverse ETL Actions", Lines: lines}
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
	case "github":
		examples = append(examples,
			GuideExample{Title: "Public repository credential", Command: "pm credentials add github-public --connector github --config repository=octocat/Hello-World"},
			GuideExample{Title: "Token credential", Command: "export GITHUB_TOKEN=...\npm credentials add github-token --connector github --config repository=OWNER/REPO --from-env token=GITHUB_TOKEN"},
			GuideExample{Title: "GitHub App credential", Command: "pm credentials add github-app --connector github --config repository=OWNER/REPO --config auth_type=github_app --config app_id=12345 --config installation_id=67890 --value-stdin private_key < app-private-key.pem"},
			GuideExample{Title: "Pull request ETL", Command: "pm connections create github_prs_to_warehouse --source github:github-token --destination warehouse:warehouse-local --stream pull_requests --primary-key node_id --cursor updated_at --table github_pull_requests\npm etl run --connection github_prs_to_warehouse --stream pull_requests --batch-size 100 --json"},
			GuideExample{Title: "Approved pull request creation", Command: "pm reverse plan prs_to_github --source-table github_pr_candidates --destination github:github-token --action create_pull_request --map title:title --map body:body --map head:head --map base:base --map reviewers:reviewers\npm reverse preview <plan-id> --json\npm reverse run <plan-id> --approve <approval-token> --json"},
		)
	case "sample":
		examples = append(examples, GuideExample{Title: "Sample ETL", Command: "pm credentials add sample-local --connector sample\npm connections create sample_to_warehouse --source sample:sample-local --destination warehouse:warehouse-local --stream customers --primary-key id --cursor updated_at --table sample_customers\npm etl run --connection sample_to_warehouse --stream customers --json"})
	case "file":
		examples = append(examples, GuideExample{Title: "File ETL", Command: "pm credentials add file-local --connector file --config path=/path/to/records.jsonl\npm connections create file_to_warehouse --source file:file-local --destination warehouse:warehouse-local --stream file --table imported_records\npm etl run --connection file_to_warehouse --stream file --json"})
	case "warehouse":
		examples = append(examples, GuideExample{Title: "Warehouse credential", Command: "pm credentials add warehouse-local --connector warehouse --config path=$ROOT/.polymetrics/warehouse\npm query run --table sample_customers --limit 5 --json"})
	case "outbox":
		examples = append(examples, GuideExample{Title: "Outbox reverse ETL", Command: "pm credentials add outbox-local --connector outbox --config path=$ROOT/.polymetrics/outbox\npm reverse plan customers_to_outbox --source-table sample_customers --destination outbox:outbox-local --map id:external_id --map email:email\npm reverse run <plan-id> --approve <approval-token> --json"})
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
