package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"polymetrics.ai/internal/connectors"
)

func writeConnectorDocs(dir string, registry *connectors.Registry) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create connector docs dir: %w", err)
	}
	if err := copyConnectorIconAssets(dir); err != nil {
		return err
	}
	if err := writeConnectorCatalogDocs(filepath.Join(dir, "catalog"), registry.CatalogEntries()); err != nil {
		return err
	}
	index, err := renderConnectorDocsIndex(registry)
	if err != nil {
		return err
	}
	for _, meta := range registry.List() {
		connector, ok := registry.Get(meta.Name)
		if !ok {
			continue
		}
		if err := writeOneConnectorDocs(dir, meta.Name, connector); err != nil {
			return err
		}
	}
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte(index), 0o644); err != nil {
		return fmt.Errorf("write connector docs index: %w", err)
	}
	return nil
}

// writeSelectedConnectorDocs writes only the named connector's manual, skill,
// and icon asset. It deliberately leaves the all-connector catalog and every
// other connector directory alone; use writeConnectorDocs for the full docs
// publication pass.
func writeSelectedConnectorDocs(dir, name string, registry *connectors.Registry) error {
	connector, ok := registry.Get(name)
	if !ok {
		return fmt.Errorf("unknown connector %q", name)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create connector docs dir: %w", err)
	}
	if err := copySelectedConnectorIconAsset(dir, name, connector); err != nil {
		return err
	}
	return writeOneConnectorDocs(dir, name, connector)
}

func writeOneConnectorDocs(dir, name string, connector connectors.Connector) error {
	connectorDir := filepath.Join(dir, name)
	if err := os.MkdirAll(connectorDir, 0o755); err != nil {
		return fmt.Errorf("create connector docs %s: %w", name, err)
	}
	manual := renderConnectorManual(name, connector)
	if err := os.WriteFile(filepath.Join(connectorDir, "MANUAL.md"), []byte(manual), 0o644); err != nil {
		return fmt.Errorf("write connector manual %s: %w", name, err)
	}
	skill := connectors.RenderConnectorSkill(connector)
	if err := os.WriteFile(filepath.Join(connectorDir, "SKILL.md"), []byte(skill), 0o644); err != nil {
		return fmt.Errorf("write connector skill %s: %w", name, err)
	}
	return nil
}

func validateConnectorDocs(dir string, registry *connectors.Registry) error {
	if err := validateConnectorCatalogDocs(filepath.Join(dir, "catalog"), registry.CatalogEntries()); err != nil {
		return err
	}
	if err := connectors.ValidateConnectorIcons(dir, registry.CatalogEntries(), registry.List()); err != nil {
		return err
	}
	expectedIndex, err := renderConnectorDocsIndex(registry)
	if err != nil {
		return err
	}
	indexPath := filepath.Join(dir, "README.md")
	index, err := os.ReadFile(indexPath)
	if err != nil {
		return fmt.Errorf("connector docs missing index at %s: %w", indexPath, err)
	}
	if string(index) != expectedIndex {
		return fmt.Errorf("connector docs index is stale; run pm docs generate")
	}
	for _, meta := range registry.List() {
		connector, ok := registry.Get(meta.Name)
		if !ok {
			continue
		}
		if err := validateOneConnectorDocs(dir, meta.Name, connector, "pm docs generate"); err != nil {
			return err
		}
	}
	return nil
}

// validateSelectedConnectorDocs validates only the named connector's generated
// files. Full catalog and all-connector validation remain owned by
// validateConnectorDocs.
func validateSelectedConnectorDocs(dir, name string, registry *connectors.Registry) error {
	connector, ok := registry.Get(name)
	if !ok {
		return fmt.Errorf("unknown connector %q", name)
	}
	metadata := connectors.MetadataOf(connector)
	if metadata.Icon == nil {
		return fmt.Errorf("connector %s has no icon metadata", name)
	}
	if err := connectors.ValidateConnectorIcon(dir, name, *metadata.Icon); err != nil {
		return err
	}
	return validateOneConnectorDocs(dir, name, connector, "pm docs connector generate --connector "+name)
}

func validateOneConnectorDocs(dir, name string, connector connectors.Connector, regenerationCommand string) error {
	if err := connectors.ValidateConnectorGuide(connector); err != nil {
		return err
	}
	manualPath := filepath.Join(dir, name, "MANUAL.md")
	manual, err := os.ReadFile(manualPath)
	if err != nil {
		return fmt.Errorf("connector %s missing manual at %s: %w", name, manualPath, err)
	}
	for _, required := range []string{"NAME", "SYNOPSIS", "DESCRIPTION", "ICON", "SECURITY", "AGENT WORKFLOW"} {
		if !strings.Contains(string(manual), required) {
			return fmt.Errorf("connector %s manual missing %s", name, required)
		}
	}
	expectedManual := connectors.RenderConnectorManual(connector)
	if err := validateGeneratedConnectorIconMetadata(string(manual), expectedManual, "ICON\n", name, "manual"); err != nil {
		return err
	}
	if string(manual) != renderConnectorManual(name, connector) {
		return fmt.Errorf("connector %s manual is stale; run %s", name, regenerationCommand)
	}
	if err := validateConnectorSkillFile(dir, name, connector, regenerationCommand); err != nil {
		return err
	}
	return nil
}

func renderConnectorDocsIndex(registry *connectors.Registry) (string, error) {
	index := strings.Builder{}
	index.WriteString("# Connector Manuals\n\n")
	index.WriteString("> Auto-generated by `pm docs generate`.\n\n")
	index.WriteString("## Connector Catalog\n\n")
	index.WriteString("- [All connectors](catalog/all-connectors.md): generated catalog from declarative bundles and Tier-3 natives.\n\n")
	index.WriteString("## Runtime Connectors\n\n")
	for _, meta := range registry.List() {
		connector, ok := registry.Get(meta.Name)
		if !ok {
			continue
		}
		if err := connectors.ValidateConnectorGuide(connector); err != nil {
			return "", err
		}
		index.WriteString("- [" + meta.Name + "](" + meta.Name + "/MANUAL.md): " + meta.Description + "\n")
	}
	return index.String(), nil
}

func renderConnectorManual(name string, connector connectors.Connector) string {
	return "# pm connectors inspect " + name + "\n\n```text\n" + connectors.RenderConnectorManual(connector) + "\n```\n"
}

func validateConnectorSkillFile(dir, name string, connector connectors.Connector, regenerationCommand string) error {
	skillPath := filepath.Join(dir, name, "SKILL.md")
	skill, err := os.ReadFile(skillPath)
	if err != nil {
		return fmt.Errorf("connector %s missing skill at %s: %w", name, skillPath, err)
	}
	for _, required := range []string{"name: pm-" + name, "## Agent Rules"} {
		if !strings.Contains(string(skill), required) {
			return fmt.Errorf("connector %s skill missing %q", name, required)
		}
	}
	if err := validateGeneratedConnectorIconMetadata(string(skill), connectors.RenderConnectorSkill(connector), "## Icon\n\n", name, "skill"); err != nil {
		return err
	}
	if string(skill) != connectors.RenderConnectorSkill(connector) {
		return fmt.Errorf("connector %s skill is stale; run %s", name, regenerationCommand)
	}
	return nil
}

func validateGeneratedConnectorIconMetadata(document, expected, heading, name, surface string) error {
	got, err := generatedConnectorIconBlock(document, heading)
	if err != nil {
		return fmt.Errorf("connector %s %s icon metadata: %w", name, surface, err)
	}
	want, err := generatedConnectorIconBlock(expected, heading)
	if err != nil {
		return fmt.Errorf("connector %s canonical %s icon metadata: %w", name, surface, err)
	}
	if got != want {
		return fmt.Errorf("connector %s %s icon metadata does not match canonical registry", name, surface)
	}
	return nil
}

func generatedConnectorIconBlock(document, heading string) (string, error) {
	headingLine := strings.TrimRight(heading, "\n")
	if headingLine == "" || strings.Contains(headingLine, "\n") {
		return "", fmt.Errorf("invalid section heading")
	}
	separator := heading[len(headingLine):]
	starts := make([]int, 0, 1)
	for lineStart := 0; lineStart <= len(document); {
		relativeEnd := strings.IndexByte(document[lineStart:], '\n')
		lineEnd := len(document)
		if relativeEnd >= 0 {
			lineEnd = lineStart + relativeEnd
		}
		if document[lineStart:lineEnd] == headingLine {
			starts = append(starts, lineStart)
		}
		if lineEnd == len(document) {
			break
		}
		lineStart = lineEnd + 1
	}
	if len(starts) == 0 {
		return "", fmt.Errorf("missing section")
	}
	if len(starts) > 1 {
		return "", fmt.Errorf("duplicate sections")
	}
	start := starts[0]
	remainder := document[start+len(headingLine):]
	if !strings.HasPrefix(remainder, separator) {
		return "", fmt.Errorf("invalid section separator")
	}
	remainder = remainder[len(separator):]
	end := strings.Index(remainder, "\n\n")
	if end < 0 {
		return "", fmt.Errorf("unterminated section")
	}
	return remainder[:end], nil
}

func writeConnectorCatalogDocs(dir string, defs []connectors.Definition) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create connector catalog docs dir: %w", err)
	}
	data, err := renderConnectorCatalogJSON(defs)
	if err != nil {
		return fmt.Errorf("encode connector catalog: %w", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "all-connectors.json"), data, 0o644); err != nil {
		return fmt.Errorf("write connector catalog json: %w", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "all-connectors.md"), []byte(renderConnectorCatalogMarkdown(defs)), 0o644); err != nil {
		return fmt.Errorf("write connector catalog markdown: %w", err)
	}
	return nil
}

func validateConnectorCatalogDocs(dir string, wantDefs []connectors.Definition) error {
	jsonPath := filepath.Join(dir, "all-connectors.json")
	mdPath := filepath.Join(dir, "all-connectors.md")
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return fmt.Errorf("connector catalog missing json at %s: %w", jsonPath, err)
	}
	var defs []connectors.Definition
	if err := json.Unmarshal(data, &defs); err != nil {
		return fmt.Errorf("decode connector catalog json: %w", err)
	}
	var rawDefs []struct {
		Name string          `json:"name"`
		Icon json.RawMessage `json:"icon"`
	}
	if err := json.Unmarshal(data, &rawDefs); err != nil {
		return fmt.Errorf("decode raw connector catalog json: %w", err)
	}
	if len(defs) != len(wantDefs) {
		return fmt.Errorf("connector catalog json has %d entries, want %d", len(defs), len(wantDefs))
	}
	wantByName := make(map[string]connectors.Definition, len(wantDefs))
	for _, def := range wantDefs {
		wantByName[def.Name] = def
	}
	seen := make(map[string]struct{}, len(defs))
	for i, def := range defs {
		want, ok := wantByName[def.Name]
		if !ok {
			return fmt.Errorf("connector catalog json has unexpected connector %q", def.Name)
		}
		if _, duplicate := seen[def.Name]; duplicate {
			return fmt.Errorf("connector catalog json has duplicate connector %q", def.Name)
		}
		seen[def.Name] = struct{}{}
		if err := validateCatalogIconMetadata("json", def.Name, rawDefs[i].Icon, want.Icon); err != nil {
			return err
		}
	}
	expectedJSON, err := renderConnectorCatalogJSON(wantDefs)
	if err != nil {
		return fmt.Errorf("encode canonical connector catalog: %w", err)
	}
	if !bytes.Equal(data, expectedJSON) {
		return fmt.Errorf("connector catalog json is stale; run pm docs generate")
	}
	markdown, err := os.ReadFile(mdPath)
	if err != nil {
		return fmt.Errorf("connector catalog missing markdown at %s: %w", mdPath, err)
	}
	for _, required := range []string{"# Connector Catalog", "Icon", "icons/github.svg", "Documentation", "Connector Metadata", "`github`", "`postgres`"} {
		if !strings.Contains(string(markdown), required) {
			return fmt.Errorf("connector catalog markdown missing %q", required)
		}
	}
	for _, def := range wantDefs {
		rowPrefix := "| `" + def.Name + "` | " + catalogIconMarkdown(def.Icon) + " |"
		if !strings.Contains(string(markdown), rowPrefix) {
			return fmt.Errorf("connector %s catalog markdown icon path does not match canonical registry", def.Name)
		}
	}
	if string(markdown) != renderConnectorCatalogMarkdown(wantDefs) {
		return fmt.Errorf("connector catalog markdown is stale; run pm docs generate")
	}
	return nil
}

func renderConnectorCatalogJSON(defs []connectors.Definition) ([]byte, error) {
	data, err := json.MarshalIndent(defs, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

func renderConnectorCatalogMarkdown(defs []connectors.Definition) string {
	var b strings.Builder
	b.WriteString("# Connector Catalog\n\n")
	b.WriteString("> Auto-generated by `pm docs generate`. Image references are metadata only; runtime capabilities come from connector bundles and Tier-3 natives.\n\n")
	b.WriteString("| Name | Icon | Display Name | Type | Release | Capabilities | Streams | Writes | Documentation | Connector Metadata |\n")
	b.WriteString("| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |\n")
	for _, def := range defs {
		b.WriteString("| `" + def.Name + "` | " + catalogIconMarkdown(def.Icon) + " | " + markdownCell(def.DisplayName) + " | `" + markdownCell(def.IntegrationType) + "` | " + markdownCell(def.ReleaseStage) + " | " + catalogCapabilitySummary(def.Capabilities) + " | " + fmt.Sprintf("%d", len(def.Streams)) + " | " + fmt.Sprintf("%d", len(def.WriteActions)) + " | " + definitionDocsMarkdown(def) + " | bundle definition |\n")
	}
	return b.String()
}

func validateCatalogIconMetadata(surface, name string, got json.RawMessage, want *connectors.ConnectorIcon) error {
	if want == nil {
		return fmt.Errorf("connector %s canonical icon metadata is missing", name)
	}
	var gotData bytes.Buffer
	if err := json.Compact(&gotData, got); err != nil {
		return fmt.Errorf("decode connector %s catalog %s icon metadata: %w", name, surface, err)
	}
	wantData, err := json.Marshal(want)
	if err != nil {
		return fmt.Errorf("encode connector %s canonical icon metadata: %w", name, err)
	}
	if !bytes.Equal(gotData.Bytes(), wantData) {
		return fmt.Errorf("connector %s catalog %s icon metadata does not match canonical registry", name, surface)
	}
	return nil
}

func catalogIconMarkdown(icon *connectors.ConnectorIcon) string {
	if icon == nil || icon.Path == "" {
		return "missing"
	}
	return "[`" + markdownCell(icon.Path) + "`](../" + markdownCell(icon.Path) + ")"
}

func catalogCapabilitySummary(caps connectors.Capabilities) string {
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
	if len(enabled) == 0 {
		return "metadata"
	}
	return strings.Join(enabled, ", ")
}

func definitionDocsMarkdown(def connectors.Definition) string {
	if strings.TrimSpace(def.DocsURL) == "" {
		return "manual intervention needed"
	}
	return "[Documentation](" + markdownCell(def.DocsURL) + ")"
}

func markdownCell(value string) string {
	value = strings.ReplaceAll(value, "|", "\\|")
	value = strings.ReplaceAll(value, "\n", " ")
	return value
}

func copyConnectorIconAssets(connectorsDir string) error {
	src, err := connectorIconSourceDir(connectorsDir)
	if err != nil {
		return err
	}
	dst := filepath.Join(connectorsDir, "icons")
	if samePath(src, dst) {
		return nil
	}
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return fmt.Errorf("create connector icons dir: %w", err)
	}
	return filepath.WalkDir(src, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".svg") {
			return nil
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return fmt.Errorf("resolve connector icon %s: %w", path, err)
		}
		if strings.HasPrefix(filepath.ToSlash(rel), "../") {
			return fmt.Errorf("connector icon %s escapes icon source root", path)
		}
		target := filepath.Join(dst, rel)
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return fmt.Errorf("create connector icon dir %s: %w", filepath.Dir(target), err)
		}
		return copyFile(path, target)
	})
}

func copySelectedConnectorIconAsset(connectorsDir, connectorName string, connector connectors.Connector) error {
	metadata := connectors.MetadataOf(connector)
	if metadata.Icon == nil {
		return fmt.Errorf("connector icon %s: missing icon registry entry", connectorName)
	}
	iconPath := path.Clean(metadata.Icon.Path)
	if iconPath != metadata.Icon.Path || strings.HasPrefix(iconPath, "../") || strings.HasPrefix(iconPath, "/") || !strings.HasPrefix(iconPath, "icons/") || path.Ext(iconPath) != ".svg" {
		return fmt.Errorf("connector icon %s: invalid path %q", connectorName, metadata.Icon.Path)
	}
	srcRoot, err := connectorIconSourceDir(connectorsDir)
	if err != nil {
		return err
	}
	src := filepath.Join(srcRoot, filepath.FromSlash(strings.TrimPrefix(iconPath, "icons/")))
	dst := filepath.Join(connectorsDir, filepath.FromSlash(iconPath))
	if !pathWithin(srcRoot, src) || !pathWithin(connectorsDir, dst) {
		return fmt.Errorf("connector icon %s path escapes its root", connectorName)
	}
	if samePath(src, dst) {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return fmt.Errorf("create connector icon dir %s: %w", filepath.Dir(dst), err)
	}
	return copyFile(src, dst)
}

func pathWithin(root, target string) bool {
	rel, err := filepath.Rel(root, target)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel)
}

func connectorIconSourceDir(connectorsDir string) (string, error) {
	root, err := repoRootFromWorkingDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "docs", "connectors", "icons"), nil
}

func repoRootFromWorkingDir() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found from working directory")
		}
		dir = parent
	}
}

func samePath(a, b string) bool {
	absA, errA := filepath.Abs(a)
	absB, errB := filepath.Abs(b)
	if errA != nil || errB != nil {
		return filepath.Clean(a) == filepath.Clean(b)
	}
	return filepath.Clean(absA) == filepath.Clean(absB)
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open connector icon %s: %w", src, err)
	}
	defer func() { _ = in.Close() }()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("create connector icon %s: %w", dst, err)
	}
	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		return fmt.Errorf("copy connector icon %s: %w", dst, err)
	}
	if err := out.Close(); err != nil {
		return fmt.Errorf("close connector icon %s: %w", dst, err)
	}
	return nil
}
