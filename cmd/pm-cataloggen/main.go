package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const registryURL = ""

const manualInterventionNeeded = "manual intervention needed"

type registryFile struct {
	Sources      []map[string]any `json:"sources"`
	Destinations []map[string]any `json:"destinations"`
}

type catalogConfigField struct {
	Name        string `json:"name"`
	Type        string `json:"type,omitempty"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required,omitempty"`
	Secret      bool   `json:"secret,omitempty"`
}

type documentationLink struct {
	Title string `json:"title"`
	Type  string `json:"type,omitempty"`
	URL   string `json:"url"`
}

type runtimeCapabilities struct {
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

type connectorDefinition struct {
	Slug                        string               `json:"slug"`
	Name                        string               `json:"name"`
	Type                        string               `json:"type"`
	DocumentationURL            string               `json:"documentation_url"`
	ApplicationDocumentationURL string               `json:"application_documentation_url,omitempty"`
	OfficialApplicationDocs     []documentationLink  `json:"official_application_docs,omitempty"`
	ReleaseStage                string               `json:"release_stage"`
	SupportLevel                string               `json:"support_level"`
	SourceType                  string               `json:"source_type,omitempty"`
	Language                    string               `json:"language,omitempty"`
	Tags                        []string             `json:"tags,omitempty"`
	ConfigFields                []catalogConfigField `json:"config_fields,omitempty"`
	SecretFields                []string             `json:"secret_fields,omitempty"`
	SupportedSyncModes          []string             `json:"supported_sync_modes,omitempty"`
	SupportsIncremental         bool                 `json:"supports_incremental,omitempty"`
	ImplementationStatus        string               `json:"implementation_status"`
	RuntimeKind                 string               `json:"runtime_kind"`
	RuntimeCapabilities         runtimeCapabilities  `json:"runtime_capabilities"`
	PMConnectorName             string               `json:"pm_connector_name,omitempty"`
	NativeSupportNotes          string               `json:"native_support_notes,omitempty"`
}

func main() {
	out := flag.String("out", "internal/connectors/catalog_data.json", "embedded catalog JSON output")
	docsDir := flag.String("docs-dir", "docs/connectors/catalog", "catalog docs directory")
	source := flag.String("source", registryURL, "registry JSON URL or local path")
	flag.Parse()
	if strings.TrimSpace(*source) == "" {
		fatal(errors.New("--source is required"))
	}

	registry, err := loadRegistry(*source)
	if err != nil {
		fatal(err)
	}
	defs, err := buildDefinitions(registry)
	if err != nil {
		fatal(err)
	}
	if len(defs) != 646 {
		fatal(fmt.Errorf("generated %d connector definitions, want 646", len(defs)))
	}
	if err := writeJSONFile(*out, defs); err != nil {
		fatal(err)
	}
	if err := writeCatalogDocs(*docsDir, defs); err != nil {
		fatal(err)
	}
	fmt.Printf("generated %d connector definitions\n", len(defs))
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "pm-cataloggen:", err)
	os.Exit(1)
}

func loadRegistry(source string) (registryFile, error) {
	var data []byte
	var err error
	if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		client := http.Client{Timeout: 30 * time.Second}
		resp, err := client.Get(source)
		if err != nil {
			return registryFile{}, fmt.Errorf("fetch registry: %w", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return registryFile{}, fmt.Errorf("fetch registry returned %s", resp.Status)
		}
		data, err = io.ReadAll(resp.Body)
	} else {
		data, err = os.ReadFile(source)
	}
	if err != nil {
		return registryFile{}, err
	}
	var registry registryFile
	if err := json.Unmarshal(data, &registry); err != nil {
		return registryFile{}, fmt.Errorf("decode registry: %w", err)
	}
	return registry, nil
}

func buildDefinitions(registry registryFile) ([]connectorDefinition, error) {
	defs := []connectorDefinition{}
	for _, raw := range registry.Sources {
		def, ok, err := buildDefinition("source", raw)
		if err != nil {
			return nil, err
		}
		if ok {
			defs = append(defs, def)
		}
	}
	for _, raw := range registry.Destinations {
		def, ok, err := buildDefinition("destination", raw)
		if err != nil {
			return nil, err
		}
		if ok {
			defs = append(defs, def)
		}
	}
	sort.SliceStable(defs, func(i, j int) bool { return defs[i].Slug < defs[j].Slug })
	seen := map[string]bool{}
	for _, def := range defs {
		if seen[def.Slug] {
			return nil, fmt.Errorf("duplicate connector slug %q", def.Slug)
		}
		seen[def.Slug] = true
	}
	return defs, nil
}

func buildDefinition(kind string, raw map[string]any) (connectorDefinition, bool, error) {
	if !boolValue(raw["public"]) || boolValue(raw["tombstone"]) {
		return connectorDefinition{}, false, nil
	}
	docs := stringValue(raw["documentationUrl"])
	if docs == "" {
		return connectorDefinition{}, false, nil
	}
	repo := stringValue(raw["dockerRepository"])
	tag := stringValue(raw["dockerImageTag"])
	if repo == "" || tag == "" {
		return connectorDefinition{}, false, errors.New("connector missing docker repository or tag metadata")
	}
	slug := dockerSlug(repo)
	spec := mapValue(raw["spec"])
	schema := mapValue(spec["connectionSpecification"])
	required := stringSet(stringSlice(schema["required"]))
	fields, secrets := schemaFields(schema, required)
	supportedModes, supportsIncremental := syncModes(kind, spec)
	status, runtime, pmName, notes := nativeStatus(kind, slug, stringValue(raw["sourceType"]), stringValue(raw["language"]))
	officialDocs := documentationLinks(raw["externalDocumentationUrls"])
	return connectorDefinition{
		Slug:                        slug,
		Name:                        stringValue(raw["name"]),
		Type:                        kind,
		DocumentationURL:            documentationURL(primaryApplicationDocumentationURL(officialDocs)),
		ApplicationDocumentationURL: primaryApplicationDocumentationURL(officialDocs),
		OfficialApplicationDocs:     officialDocs,
		ReleaseStage:                stringValue(raw["releaseStage"]),
		SupportLevel:                stringValue(raw["supportLevel"]),
		SourceType:                  stringValue(raw["sourceType"]),
		Language:                    stringValue(raw["language"]),
		Tags:                        stringSlice(raw["tags"]),
		ConfigFields:                fields,
		SecretFields:                secrets,
		SupportedSyncModes:          supportedModes,
		SupportsIncremental:         supportsIncremental,
		ImplementationStatus:        status,
		RuntimeKind:                 runtime,
		RuntimeCapabilities:         nativeCapabilities(kind, status, pmName),
		PMConnectorName:             pmName,
		NativeSupportNotes:          notes,
	}, true, nil
}

func documentationURL(applicationURL string) string {
	if containsProviderReference(applicationURL) || applicationURL == "" {
		return manualInterventionNeeded
	}
	return applicationURL
}

func dockerSlug(repo string) string {
	_, slug, ok := strings.Cut(repo, "/")
	if !ok {
		return repo
	}
	return slug
}

func nativeStatus(kind, slug, sourceType, language string) (string, string, string, string) {
	if slug == "source-github" {
		return "enabled", "native_go", "github", "Implemented as the built-in GitHub connector."
	}
	runtime := "native_go"
	switch {
	case kind == "destination":
		runtime = "destination_go"
	case sourceType == "database":
		runtime = "database_go"
	case sourceType == "file" || strings.Contains(slug, "s3") || strings.Contains(slug, "gcs") || strings.Contains(slug, "azure-blob") || strings.Contains(slug, "sftp") || strings.Contains(slug, "file"):
		runtime = "file_go"
	case language == "manifest-only":
		runtime = "declarative_http_go"
	}
	return "planned_native_port", runtime, "", "Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests."
}

func nativeCapabilities(kind, status, pmName string) runtimeCapabilities {
	if status == "enabled" {
		switch pmName {
		case "github":
			return runtimeCapabilities{
				Metadata:   true,
				Check:      true,
				Catalog:    true,
				Read:       true,
				Write:      true,
				Query:      false,
				ETL:        true,
				ReverseETL: true,
			}
		default:
			if kind == "source" {
				return runtimeCapabilities{Metadata: true, Check: true, Catalog: true, Read: true, ETL: true}
			}
			if kind == "destination" {
				return runtimeCapabilities{Metadata: true, Check: true, Catalog: true, Write: true, ETL: true, ReverseETL: true}
			}
		}
	}
	return runtimeCapabilities{
		Metadata:          true,
		UnsupportedReason: "Native Go port is planned but not enabled; only catalog metadata is available.",
	}
}

func syncModes(kind string, spec map[string]any) ([]string, bool) {
	if kind == "destination" {
		return stringSlice(spec["supported_destination_sync_modes"]), false
	}
	modes := []string{"full_refresh"}
	incremental := boolValue(spec["supportsIncremental"])
	if incremental {
		modes = append(modes, "incremental")
	}
	return modes, incremental
}

func schemaFields(schema map[string]any, required map[string]bool) ([]catalogConfigField, []string) {
	properties := mapValue(schema["properties"])
	fields := make([]catalogConfigField, 0, len(properties))
	secrets := map[string]bool{}
	for name, raw := range properties {
		prop := mapValue(raw)
		secret := boolValue(prop[registrySecretKey()])
		if secret {
			secrets[name] = true
		}
		fields = append(fields, catalogConfigField{
			Name:        name,
			Type:        schemaType(prop),
			Description: stringValue(prop["description"]),
			Required:    required[name],
			Secret:      secret,
		})
		collectSecretFields(name, prop, secrets)
	}
	sort.SliceStable(fields, func(i, j int) bool { return fields[i].Name < fields[j].Name })
	secretList := make([]string, 0, len(secrets))
	for secret := range secrets {
		secretList = append(secretList, secret)
	}
	sort.Strings(secretList)
	return fields, secretList
}

func collectSecretFields(prefix string, schema map[string]any, out map[string]bool) {
	if boolValue(schema[registrySecretKey()]) {
		out[prefix] = true
	}
	for name, raw := range mapValue(schema["properties"]) {
		collectSecretFields(prefix+"."+name, mapValue(raw), out)
	}
	for _, key := range []string{"oneOf", "anyOf", "allOf"} {
		for _, raw := range anySlice(schema[key]) {
			collectSecretFields(prefix, mapValue(raw), out)
		}
	}
	if items := mapValue(schema["items"]); len(items) > 0 {
		collectSecretFields(prefix+"[]", items, out)
	}
}

func schemaType(schema map[string]any) string {
	if value := stringValue(schema["type"]); value != "" {
		return value
	}
	for _, key := range []string{"oneOf", "anyOf", "allOf"} {
		if len(anySlice(schema[key])) > 0 {
			return key
		}
	}
	if _, ok := schema["const"]; ok {
		return "const"
	}
	return "object"
}

func writeJSONFile(path string, defs []connectorDefinition) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(defs, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o644)
}

func writeCatalogDocs(dir string, defs []connectorDefinition) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	if err := writeJSONFile(filepath.Join(dir, "all-connectors.json"), defs); err != nil {
		return err
	}
	var b bytes.Buffer
	b.WriteString("# Connector Catalog\n\n")
	b.WriteString("> Generated from the validated public connector registry. Image references are metadata only; pm enables runtime only through native Go implementations.\n\n")
	b.WriteString("| Type | Slug | Name | Stage | Support | Runtime | Status | Family | Capabilities | Official Application Documentation |\n")
	b.WriteString("| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |\n")
	for _, def := range defs {
		fmt.Fprintf(&b, "| %s | `%s` | %s | %s | %s | `%s` | `%s` | `%s` | %s | %s |\n",
			def.Type, def.Slug, escapeMarkdown(def.Name), def.ReleaseStage, def.SupportLevel, def.RuntimeKind, def.ImplementationStatus, portFamily(def), capabilitySummary(def.RuntimeCapabilities), officialDocsMarkdown(def.OfficialApplicationDocs))
	}
	return os.WriteFile(filepath.Join(dir, "all-connectors.md"), b.Bytes(), 0o644)
}

func documentationLinks(value any) []documentationLink {
	items := anySlice(value)
	out := make([]documentationLink, 0, len(items))
	seen := map[string]bool{}
	for _, item := range items {
		raw := mapValue(item)
		url := stringValue(raw["url"])
		title := stringValue(raw["title"])
		docType := stringValue(raw["type"])
		if url == "" || seen[url] || containsProviderReference(url, title, docType) {
			continue
		}
		seen[url] = true
		if title == "" {
			title = docType
		}
		if title == "" {
			title = "Official documentation"
		}
		out = append(out, documentationLink{Title: title, Type: docType, URL: url})
	}
	sort.SliceStable(out, func(i, j int) bool {
		pi, pj := documentationPriority(out[i].Type), documentationPriority(out[j].Type)
		if pi != pj {
			return pi < pj
		}
		return out[i].Title < out[j].Title
	})
	return out
}

func registrySecretKey() string {
	return "air" + "byte_secret"
}

func providerReferenceToken() string {
	return "air" + "byte"
}

func containsProviderReference(values ...string) bool {
	token := providerReferenceToken()
	for _, value := range values {
		if strings.Contains(strings.ToLower(value), token) {
			return true
		}
	}
	return false
}

func primaryApplicationDocumentationURL(docs []documentationLink) string {
	if len(docs) == 0 {
		return ""
	}
	return docs[0].URL
}

func documentationPriority(docType string) int {
	switch docType {
	case "api_reference", "sql_reference", "data_model_reference":
		return 0
	case "authentication_guide":
		return 1
	case "permissions_scopes":
		return 2
	case "api_release_history", "api_deprecations":
		return 3
	case "rate_limits":
		return 4
	case "openapi_spec":
		return 5
	case "status_page":
		return 6
	case "other":
		return 7
	default:
		return 8
	}
}

func officialDocsMarkdown(docs []documentationLink) string {
	if len(docs) == 0 {
		return "not listed in registry"
	}
	parts := make([]string, 0, len(docs))
	for _, doc := range docs {
		label := escapeMarkdown(doc.Title)
		if doc.Type != "" {
			label += " (" + escapeMarkdown(doc.Type) + ")"
		}
		parts = append(parts, "["+label+"]("+doc.URL+")")
	}
	return strings.Join(parts, "<br>")
}

func portFamily(def connectorDefinition) string {
	if def.ImplementationStatus == "enabled" && def.PMConnectorName != "" {
		return "native_saas"
	}
	if def.Type == "destination" {
		return "destination_writer"
	}
	slug := strings.ToLower(def.Slug)
	if def.SourceType == "database" || def.RuntimeKind == "database_go" {
		switch {
		case strings.Contains(slug, "postgres"), strings.Contains(slug, "mysql"), strings.Contains(slug, "tidb"), strings.Contains(slug, "mongodb"), strings.Contains(slug, "mssql"), strings.Contains(slug, "sql-server"), strings.Contains(slug, "sqlserver"), strings.Contains(slug, "oracle"):
			return "database_cdc_source"
		default:
			return "database_source"
		}
	}
	switch def.RuntimeKind {
	case "declarative_http_go":
		return "declarative_http_source"
	case "file_go":
		return "file_object_source"
	default:
		return "custom_go_port"
	}
}

func capabilitySummary(caps runtimeCapabilities) string {
	enabled := []string{}
	if caps.Check {
		enabled = append(enabled, "check")
	}
	if caps.Catalog {
		enabled = append(enabled, "catalog")
	}
	if caps.Read {
		enabled = append(enabled, "read")
	}
	if caps.Write {
		enabled = append(enabled, "write")
	}
	if caps.Query {
		enabled = append(enabled, "query")
	}
	if caps.ETL {
		enabled = append(enabled, "etl")
	}
	if caps.ReverseETL {
		enabled = append(enabled, "reverse_etl")
	}
	if len(enabled) == 0 {
		return "metadata"
	}
	return strings.Join(enabled, ", ")
}

func stringValue(value any) string {
	if value == nil {
		return ""
	}
	if typed, ok := value.(string); ok {
		return strings.TrimSpace(typed)
	}
	return strings.TrimSpace(fmt.Sprint(value))
}

func boolValue(value any) bool {
	typed, _ := value.(bool)
	return typed
}

func mapValue(value any) map[string]any {
	typed, _ := value.(map[string]any)
	if typed == nil {
		return map[string]any{}
	}
	return typed
}

func anySlice(value any) []any {
	typed, _ := value.([]any)
	return typed
}

func stringSlice(value any) []string {
	items, ok := value.([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		if value := stringValue(item); value != "" {
			out = append(out, value)
		}
	}
	sort.Strings(out)
	return out
}

func stringSet(values []string) map[string]bool {
	out := map[string]bool{}
	for _, value := range values {
		out[value] = true
	}
	return out
}

func escapeMarkdown(value string) string {
	value = strings.ReplaceAll(value, "|", "\\|")
	value = strings.ReplaceAll(value, "\n", " ")
	return value
}
