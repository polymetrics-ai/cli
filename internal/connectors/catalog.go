package connectors

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	_ "embed"
)

type ConnectorType string

const (
	ConnectorTypeSource      ConnectorType = "source"
	ConnectorTypeDestination ConnectorType = "destination"
)

type ImplementationStatus string

const (
	ImplementationEnabled           ImplementationStatus = "enabled"
	ImplementationPlannedNativePort ImplementationStatus = "planned_native_port"
	ImplementationUnsupported       ImplementationStatus = "unsupported_deprecated"
)

type RuntimeKind string

const (
	RuntimeNativeGo          RuntimeKind = "native_go"
	RuntimeDeclarativeHTTPGo RuntimeKind = "declarative_http_go"
	RuntimeDatabaseGo        RuntimeKind = "database_go"
	RuntimeFileGo            RuntimeKind = "file_go"
	RuntimeDestinationGo     RuntimeKind = "destination_go"
)

type RuntimeCapabilities struct {
	Metadata          bool   `json:"metadata"`
	Check             bool   `json:"check"`
	Catalog           bool   `json:"catalog"`
	Read              bool   `json:"read"`
	Write             bool   `json:"write"`
	Query             bool   `json:"query"`
	ETL               bool   `json:"etl"`
	ReverseETL        bool   `json:"reverse_etl"`
	UnsupportedReason string `json:"unsupported_reason,omitempty"`
}

type CatalogConfigField struct {
	Name        string `json:"name"`
	Type        string `json:"type,omitempty"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required,omitempty"`
	Secret      bool   `json:"secret,omitempty"`
}

type DocumentationLink struct {
	Title string `json:"title"`
	Type  string `json:"type,omitempty"`
	URL   string `json:"url"`
}

type ConnectorDefinition struct {
	Slug                             string               `json:"slug"`
	Name                             string               `json:"name"`
	Type                             ConnectorType        `json:"type"`
	DocumentationURL                 string               `json:"documentation_url"`
	AirbyteConnectorDocumentationURL string               `json:"airbyte_connector_documentation_url,omitempty"`
	ApplicationDocumentationURL      string               `json:"application_documentation_url,omitempty"`
	OfficialApplicationDocs          []DocumentationLink  `json:"official_application_docs,omitempty"`
	ReleaseStage                     string               `json:"release_stage"`
	SupportLevel                     string               `json:"support_level"`
	SourceType                       string               `json:"source_type,omitempty"`
	Language                         string               `json:"language,omitempty"`
	Tags                             []string             `json:"tags,omitempty"`
	ConfigFields                     []CatalogConfigField `json:"config_fields,omitempty"`
	ConfigSchema                     json.RawMessage      `json:"config_schema,omitempty"`
	SecretFields                     []string             `json:"secret_fields,omitempty"`
	SupportedSyncModes               []string             `json:"supported_sync_modes,omitempty"`
	SupportsIncremental              bool                 `json:"supports_incremental,omitempty"`
	ImplementationStatus             ImplementationStatus `json:"implementation_status"`
	RuntimeKind                      RuntimeKind          `json:"runtime_kind"`
	RuntimeCapabilities              RuntimeCapabilities  `json:"runtime_capabilities"`
	PMConnectorName                  string               `json:"pm_connector_name,omitempty"`
	UpstreamImageReference           string               `json:"upstream_image_reference,omitempty"`
	NativeSupportNotes               string               `json:"native_support_notes,omitempty"`
}

type ConnectorCatalogFilter struct {
	Type  ConnectorType
	Stage string
}

type ConnectorCatalogSummary struct {
	Total                 int `json:"total"`
	Sources               int `json:"sources"`
	Destinations          int `json:"destinations"`
	DocsPresent           int `json:"docs_present"`
	Enabled               int `json:"enabled"`
	PlannedNativePort     int `json:"planned_native_port"`
	UnsupportedDeprecated int `json:"unsupported_deprecated"`
}

//go:embed catalog_data.json
var connectorCatalogData []byte

func ConnectorCatalog() []ConnectorDefinition {
	var defs []ConnectorDefinition
	if err := json.Unmarshal(connectorCatalogData, &defs); err != nil {
		return nil
	}
	for i := range defs {
		defs[i] = normalizeConnectorDefinition(defs[i])
	}
	return cloneConnectorDefinitions(defs)
}

func ConnectorDefinitionBySlug(slug string) (ConnectorDefinition, bool) {
	slug = strings.TrimSpace(slug)
	catalog := ConnectorCatalog()
	// Exact legacy slug match wins (back-compat with source-/destination- slugs).
	for _, entry := range catalog {
		if entry.Slug == slug {
			return entry, true
		}
	}
	// Bare-name resolution: accept the unified per-system name (e.g. "github").
	// For unify pairs, prefer the source entry, then the destination, so lookups
	// stay deterministic until catalog_data.json is migrated to merged entries.
	bare := BareName(slug)
	srcIdx, dstIdx := -1, -1
	for i := range catalog {
		if BareName(catalog[i].Slug) != bare {
			continue
		}
		switch catalog[i].Type {
		case ConnectorTypeSource:
			if srcIdx < 0 {
				srcIdx = i
			}
		case ConnectorTypeDestination:
			if dstIdx < 0 {
				dstIdx = i
			}
		}
	}
	if srcIdx >= 0 {
		return catalog[srcIdx], true
	}
	if dstIdx >= 0 {
		return catalog[dstIdx], true
	}
	return ConnectorDefinition{}, false
}

func FilterConnectorCatalog(defs []ConnectorDefinition, filter ConnectorCatalogFilter) []ConnectorDefinition {
	out := make([]ConnectorDefinition, 0, len(defs))
	stage := strings.TrimSpace(filter.Stage)
	for _, entry := range defs {
		if filter.Type != "" && entry.Type != filter.Type {
			continue
		}
		if stage != "" && entry.ReleaseStage != stage {
			continue
		}
		out = append(out, entry)
	}
	return cloneConnectorDefinitions(out)
}

func ConnectorCatalogCounts(defs []ConnectorDefinition) ConnectorCatalogSummary {
	summary := ConnectorCatalogSummary{Total: len(defs)}
	for _, entry := range defs {
		if entry.DocumentationURL != "" {
			summary.DocsPresent++
		}
		switch entry.Type {
		case ConnectorTypeSource:
			summary.Sources++
		case ConnectorTypeDestination:
			summary.Destinations++
		}
		switch entry.ImplementationStatus {
		case ImplementationEnabled:
			summary.Enabled++
		case ImplementationPlannedNativePort:
			summary.PlannedNativePort++
		case ImplementationUnsupported:
			summary.UnsupportedDeprecated++
		}
	}
	return summary
}

func ConnectorDefinitionGuide(def ConnectorDefinition) ConnectorGuide {
	sections := []GuideSection{
		{
			Title: "Capabilities",
			Lines: []string{
				"catalog_metadata=true",
				"connector type: " + string(def.Type),
				"release stage: " + valueOrDefault(def.ReleaseStage, "unknown"),
				"support level: " + valueOrDefault(def.SupportLevel, "unknown"),
			},
		},
		implementationSection(def),
		runtimeCapabilitiesSection(def),
		nativePortPlanSection(def),
		officialApplicationDocsSection(def),
		catalogConfigSection(def),
		catalogSyncSection(def),
		{
			Title: "Security",
			Lines: []string{
				"Secret values are never rendered; only secret field names are shown.",
				"Upstream image references are metadata only and are not executed by pm.",
				"Catalog-only connectors cannot run ETL until a native Go implementation is enabled.",
			},
		},
	}
	if def.DocumentationURL != "" {
		sections = append(sections, GuideSection{Title: "Documentation", Lines: []string{def.DocumentationURL}})
	}
	return ConnectorGuide{
		Name:        def.Slug,
		DisplayName: def.Name,
		Summary:     fmt.Sprintf("%s catalog connector for %s. Native implementation status: %s.", def.Name, def.DocumentationURL, def.ImplementationStatus),
		Sections:    compactSections(sections),
		Examples: []GuideExample{
			{Title: "Inspect catalog entry", Command: "pm connectors inspect " + def.Slug},
			{Title: "Inspect as JSON", Command: "pm connectors inspect " + def.Slug + " --json"},
		},
		Links: []GuideLink{{Label: def.Name + " documentation", URL: def.DocumentationURL}},
		AgentNotes: []string{
			"Read implementation_status before planning ETL or reverse ETL.",
			"If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.",
			"Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.",
		},
	}
}

func RenderConnectorDefinitionManual(def ConnectorDefinition) string {
	return RenderGuideManual(ConnectorDefinitionGuide(def))
}

func RenderConnectorDefinitionSkill(def ConnectorDefinition) string {
	return RenderGuideSkill(ConnectorDefinitionGuide(def))
}

func ValidateConnectorDefinitionGuide(def ConnectorDefinition) error {
	if strings.TrimSpace(def.Slug) == "" || strings.TrimSpace(def.Name) == "" {
		return fmt.Errorf("catalog connector missing slug or name: %+v", def)
	}
	manual := RenderConnectorDefinitionManual(def)
	for _, required := range []string{"NAME", "SYNOPSIS", "DESCRIPTION", "CAPABILITIES", "IMPLEMENTATION STATUS", "RUNTIME CAPABILITIES", "NATIVE PORT PLAN", "OFFICIAL APPLICATION DOCUMENTATION", "CONFIGURATION", "SECURITY", "AGENT WORKFLOW"} {
		if !strings.Contains(manual, required) {
			return fmt.Errorf("catalog connector %q manual missing %s", def.Slug, required)
		}
	}
	skill := RenderConnectorDefinitionSkill(def)
	if !strings.Contains(skill, "name: pm-"+def.Slug) || !strings.Contains(skill, "## Agent Rules") {
		return fmt.Errorf("catalog connector %q skill is incomplete", def.Slug)
	}
	return nil
}

func implementationSection(def ConnectorDefinition) GuideSection {
	lines := []string{
		"implementation_status: " + string(def.ImplementationStatus),
		"runtime_kind: " + string(def.RuntimeKind),
	}
	if def.PMConnectorName != "" {
		lines = append(lines, "pm connector: "+def.PMConnectorName)
	}
	if def.NativeSupportNotes != "" {
		lines = append(lines, "notes: "+def.NativeSupportNotes)
	}
	if def.UpstreamImageReference != "" {
		lines = append(lines, "upstream image reference: "+def.UpstreamImageReference+" (metadata only; not executed)")
	}
	return GuideSection{Title: "Implementation Status", Lines: lines}
}

func runtimeCapabilitiesSection(def ConnectorDefinition) GuideSection {
	caps := def.RuntimeCapabilities
	if caps == (RuntimeCapabilities{}) {
		caps = RuntimeCapabilitiesForDefinition(def)
	}
	lines := []string{
		fmt.Sprintf("metadata=%t", caps.Metadata),
		fmt.Sprintf("check=%t", caps.Check),
		fmt.Sprintf("catalog=%t", caps.Catalog),
		fmt.Sprintf("read=%t", caps.Read),
		fmt.Sprintf("write=%t", caps.Write),
		fmt.Sprintf("query=%t", caps.Query),
		fmt.Sprintf("etl=%t", caps.ETL),
		fmt.Sprintf("reverse_etl=%t", caps.ReverseETL),
	}
	if caps.UnsupportedReason != "" {
		lines = append(lines, "unsupported_reason: "+caps.UnsupportedReason)
	}
	return GuideSection{Title: "Runtime Capabilities", Lines: lines}
}

func RuntimeCapabilitiesForDefinition(def ConnectorDefinition) RuntimeCapabilities {
	if def.RuntimeCapabilities != (RuntimeCapabilities{}) {
		return def.RuntimeCapabilities
	}
	if def.ImplementationStatus != ImplementationEnabled {
		return RuntimeCapabilities{
			Metadata:          true,
			UnsupportedReason: "Native Go port is planned but not enabled; only catalog metadata is available.",
		}
	}
	return enabledRuntimeCapabilitiesForDefinition(def)
}

func enabledRuntimeCapabilitiesForDefinition(def ConnectorDefinition) RuntimeCapabilities {
	caps := RuntimeCapabilities{Metadata: true, Check: true, Catalog: true, ETL: true}
	if def.Type == ConnectorTypeSource {
		caps.Read = true
		caps.Query = nativeQuerySupported(def)
	}
	if def.Type == ConnectorTypeDestination {
		caps.Write = true
		caps.Query = nativeQuerySupported(def)
		caps.ReverseETL = true
	}
	if def.PMConnectorName == "github" {
		caps.Write = true
		caps.Query = false
		caps.ReverseETL = true
	}
	return caps
}

func normalizeConnectorDefinition(def ConnectorDefinition) ConnectorDefinition {
	if def.RuntimeKind == "" {
		def.RuntimeKind = runtimeKindForNativeDefinition(def)
	}
	def.RuntimeCapabilities = RuntimeCapabilitiesForDefinition(def)
	if def.NativeSupportNotes == "" {
		if def.ImplementationStatus == ImplementationEnabled {
			def.NativeSupportNotes = "Native Go implementation is enabled in the current binary."
		} else {
			def.NativeSupportNotes = "Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests."
		}
	}
	return def
}

func runtimeKindForNativeDefinition(def ConnectorDefinition) RuntimeKind {
	if def.Type == ConnectorTypeDestination {
		return RuntimeDestinationGo
	}
	if def.SourceType == "database" {
		return RuntimeDatabaseGo
	}
	if def.SourceType == "file" {
		return RuntimeFileGo
	}
	return RuntimeDeclarativeHTTPGo
}

func nativeQuerySupported(def ConnectorDefinition) bool {
	if def.RuntimeKind == RuntimeDatabaseGo || def.RuntimeKind == RuntimeDestinationGo {
		return true
	}
	if strings.Contains(strings.ToLower(def.Slug), "warehouse") {
		return true
	}
	return false
}

func (caps RuntimeCapabilities) toCapabilities() Capabilities {
	return Capabilities{
		Check:   caps.Check,
		Catalog: caps.Catalog,
		Read:    caps.Read,
		Write:   caps.Write,
		Query:   caps.Query,
	}
}

func officialApplicationDocsSection(def ConnectorDefinition) GuideSection {
	lines := []string{}
	if len(def.OfficialApplicationDocs) == 0 {
		lines = append(lines, "No upstream application documentation URL was listed in the imported connector registry.")
	} else {
		for _, doc := range def.OfficialApplicationDocs {
			label := doc.Title
			if label == "" {
				label = doc.Type
			}
			if label == "" {
				label = "documentation"
			}
			lines = append(lines, label+": "+doc.URL)
		}
	}
	if def.DocumentationURL != "" {
		lines = append(lines, "Airbyte connector documentation: "+def.DocumentationURL)
	}
	return GuideSection{Title: "Official Application Documentation", Lines: lines}
}

func catalogConfigSection(def ConnectorDefinition) GuideSection {
	lines := []string{}
	if len(def.ConfigFields) == 0 {
		lines = append(lines, "No config schema fields were advertised in the catalog.")
	} else {
		for _, field := range def.ConfigFields {
			line := field.Name
			if field.Type != "" {
				line += " (" + field.Type + ")"
			}
			if field.Required {
				line += " required"
			}
			if field.Secret {
				line += " secret"
			}
			if field.Description != "" {
				line += ": " + compactDescription(field.Description)
			}
			lines = append(lines, line)
		}
	}
	if len(def.SecretFields) > 0 {
		lines = append(lines, "secret fields: "+strings.Join(def.SecretFields, ", "))
	}
	return GuideSection{Title: "Configuration", Lines: lines}
}

func catalogSyncSection(def ConnectorDefinition) GuideSection {
	lines := []string{}
	if len(def.SupportedSyncModes) > 0 {
		lines = append(lines, "supported sync modes: "+strings.Join(def.SupportedSyncModes, ", "))
	}
	lines = append(lines, fmt.Sprintf("supports incremental: %t", def.SupportsIncremental))
	return GuideSection{Title: "Sync Modes", Lines: lines}
}

func cloneConnectorDefinitions(in []ConnectorDefinition) []ConnectorDefinition {
	out := make([]ConnectorDefinition, len(in))
	copy(out, in)
	for i := range out {
		out[i].Tags = append([]string(nil), out[i].Tags...)
		out[i].SecretFields = append([]string(nil), out[i].SecretFields...)
		out[i].SupportedSyncModes = append([]string(nil), out[i].SupportedSyncModes...)
		out[i].ConfigFields = append([]CatalogConfigField(nil), out[i].ConfigFields...)
		out[i].OfficialApplicationDocs = append([]DocumentationLink(nil), out[i].OfficialApplicationDocs...)
		out[i].ConfigSchema = append(json.RawMessage(nil), out[i].ConfigSchema...)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Slug < out[j].Slug })
	return out
}

func compactDescription(value string) string {
	value = strings.Join(strings.Fields(value), " ")
	if len(value) <= 180 {
		return value
	}
	return value[:177] + "..."
}
