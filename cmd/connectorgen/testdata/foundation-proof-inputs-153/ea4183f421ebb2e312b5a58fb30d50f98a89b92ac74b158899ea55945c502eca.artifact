package connectors

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
	"sync"

	_ "embed"
)

const (
	IconSourceOfficial    = "official"
	IconSourceUpstream    = "upstream_registry"
	IconSourcePolymetrics = "polymetrics"
	IconSourceSimpleIcons = "simple-icons"

	IconReviewUpstreamSeeded          = "upstream_seeded"
	IconReviewOfficial                = "official_verified"
	IconReviewManualOverride          = "manual_override"
	IconReviewPolymetrics             = "polymetrics"
	IconReviewSimpleIconsCC0Trademark = "cc0_with_trademark_caveat"
)

// ConnectorIcon identifies a local SVG asset and how it was sourced. The path is
// relative to the connector docs root, e.g. icons/github.svg.
type ConnectorIcon struct {
	ID             string `json:"id"`
	Path           string `json:"path"`
	Title          string `json:"title,omitempty"`
	SimpleIconSlug string `json:"simple_icon_slug,omitempty"`
	SimpleIconHex  string `json:"simple_icon_hex,omitempty"`
	Source         string `json:"source"`
	License        string `json:"license,omitempty"`
	Attribution    string `json:"attribution,omitempty"`
	ReviewStatus   string `json:"review_status"`
	ReviewURL      string `json:"review_url,omitempty"`
	Match          string `json:"match,omitempty"`
	MatchedBy      string `json:"matched_by,omitempty"`
}

type connectorIconEntry struct {
	Connector           string `json:"connector"`
	ID                  string `json:"id"`
	Path                string `json:"path"`
	Source              string `json:"source"`
	ReviewStatus        string `json:"review_status"`
	ReviewURL           string `json:"review_url,omitempty"`
	Implemented         bool   `json:"implemented"`
	FallbackDisposition string `json:"fallback_disposition,omitempty"`
	Title               string `json:"title,omitempty"`
	SimpleIconSlug      string `json:"simple_icon_slug,omitempty"`
	SimpleIconHex       string `json:"simple_icon_hex,omitempty"`
	License             string `json:"license,omitempty"`
	Attribution         string `json:"attribution,omitempty"`
	Match               string `json:"match,omitempty"`
	MatchedBy           string `json:"matched_by,omitempty"`
}

var unsafeSVGPatterns = []struct {
	label string
	re    *regexp.Regexp
}{
	{label: "event handler", re: regexp.MustCompile(`(?i)\son[a-z0-9_-]+\s*=`)},
	{label: "external href", re: regexp.MustCompile(`(?i)\s(?:xlink:)?href\s*=\s*["']?\s*https?://`)},
	{label: "external src", re: regexp.MustCompile(`(?i)\ssrc\s*=\s*["']?\s*https?://`)},
	{label: "external url()", re: regexp.MustCompile(`(?i)url\s*\(\s*https?://`)},
}

//go:embed icon_data.json
var connectorIconData []byte

// ErrConnectorIconPathRuntimeBuiltin reports a connector icon path that is
// declared only by runtime builtin rows carrying implemented: false. Those rows
// are retained dispositions; they never authorize connector definition
// ownership, but the path is declared rather than orphaned.
var ErrConnectorIconPathRuntimeBuiltin = errors.New("connector icon path is a runtime builtin disposition")

var connectorIcons = struct {
	once    sync.Once
	by      map[string]ConnectorIcon
	entries []connectorIconEntry
	err     error
}{}

func ConnectorIconFor(name string) (ConnectorIcon, bool) {
	icons, err := connectorIconRegistry()
	if err != nil {
		return ConnectorIcon{}, false
	}
	icon, ok := icons[strings.TrimSpace(name)]
	return icon, ok
}

func ConnectorIconEntries() []connectorIconEntry {
	entries, err := connectorIconRegistryEntries()
	if err != nil {
		return nil
	}
	return slices.Clone(entries)
}

func MetadataWithIcon(meta Metadata) Metadata {
	meta.Icon = nil
	if icon, ok := ConnectorIconFor(meta.Name); ok {
		meta.Icon = &icon
	}
	return meta
}

func manifestWithIcon(manifest Manifest) Manifest {
	manifest.Metadata = MetadataWithIcon(manifest.Metadata)
	return manifest
}

func loadConnectorIconRegistry() ([]connectorIconEntry, map[string]ConnectorIcon, error) {
	connectorIcons.once.Do(func() {
		entries, err := decodeConnectorIconRegistry()
		if err != nil {
			connectorIcons.err = err
			return
		}
		icons := make(map[string]ConnectorIcon, len(entries))
		for _, entry := range entries {
			icons[entry.Connector] = connectorIconProjection(entry)
		}
		connectorIcons.entries = entries
		connectorIcons.by = icons
	})
	return connectorIcons.entries, connectorIcons.by, connectorIcons.err
}

func connectorIconRegistry() (map[string]ConnectorIcon, error) {
	_, icons, err := loadConnectorIconRegistry()
	return icons, err
}

func connectorIconRegistryEntries() ([]connectorIconEntry, error) {
	entries, _, err := loadConnectorIconRegistry()
	return entries, err
}

func decodeConnectorIconRegistry() ([]connectorIconEntry, error) {
	var entries []connectorIconEntry
	if err := json.Unmarshal(connectorIconData, &entries); err != nil {
		return nil, fmt.Errorf("decode connector icon registry: %w", err)
	}
	seen := make(map[string]bool, len(entries))
	for i := range entries {
		entry := &entries[i]
		entry.Connector = strings.TrimSpace(entry.Connector)
		entry.ID = strings.TrimSpace(entry.ID)
		entry.Path = strings.TrimSpace(entry.Path)
		entry.Source = strings.TrimSpace(entry.Source)
		entry.ReviewStatus = strings.TrimSpace(entry.ReviewStatus)
		entry.ReviewURL = strings.TrimSpace(entry.ReviewURL)
		entry.FallbackDisposition = strings.TrimSpace(entry.FallbackDisposition)
		entry.Title = strings.TrimSpace(entry.Title)
		entry.SimpleIconSlug = strings.TrimSpace(entry.SimpleIconSlug)
		entry.SimpleIconHex = strings.TrimSpace(entry.SimpleIconHex)
		entry.License = strings.TrimSpace(entry.License)
		entry.Attribution = strings.TrimSpace(entry.Attribution)
		entry.Match = strings.TrimSpace(entry.Match)
		entry.MatchedBy = strings.TrimSpace(entry.MatchedBy)
		if entry.Connector == "" {
			return nil, fmt.Errorf("connector icon registry entry missing connector")
		}
		if hasLegacyIconConnectorPrefix(entry.Connector) {
			return nil, fmt.Errorf("connector icon registry entry %q must use a bare connector identifier", entry.Connector)
		}
		if seen[entry.Connector] {
			return nil, fmt.Errorf("duplicate connector icon registry entry %q", entry.Connector)
		}
		seen[entry.Connector] = true
		if entry.ID == "" || entry.Path == "" || entry.Source == "" || entry.ReviewStatus == "" {
			return nil, fmt.Errorf("connector icon registry entry %q has incomplete icon metadata", entry.Connector)
		}
		if !validIconReviewStatus(entry.ReviewStatus) {
			return nil, fmt.Errorf("connector icon registry entry %q has unsupported review_status %q", entry.Connector, entry.ReviewStatus)
		}
		if err := validateConnectorIconPath(entry.Path); err != nil {
			return nil, fmt.Errorf("connector icon registry entry %q: %w", entry.Connector, err)
		}
	}
	sort.SliceStable(entries, func(i, j int) bool { return entries[i].Connector < entries[j].Connector })
	return entries, nil
}

func connectorIconProjection(entry connectorIconEntry) ConnectorIcon {
	return ConnectorIcon{
		ID:             entry.ID,
		Path:           entry.Path,
		Title:          entry.Title,
		SimpleIconSlug: entry.SimpleIconSlug,
		SimpleIconHex:  entry.SimpleIconHex,
		Source:         entry.Source,
		License:        entry.License,
		Attribution:    entry.Attribution,
		ReviewStatus:   entry.ReviewStatus,
		ReviewURL:      entry.ReviewURL,
		Match:          entry.Match,
		MatchedBy:      entry.MatchedBy,
	}
}

func (r *Registry) ValidateIconCoverage() error {
	icons, err := connectorIconRegistry()
	if err != nil {
		return err
	}
	for _, metadata := range r.List() {
		name := metadata.Name
		if name == "" || name != strings.TrimSpace(name) || hasLegacyIconConnectorPrefix(name) {
			return fmt.Errorf("connector registry name %q must be a bare connector identifier", name)
		}
		if _, ok := icons[name]; !ok {
			return fmt.Errorf("missing explicit icon registry entry for connector %q", name)
		}
	}
	return nil
}

// MustValidateIconCoverage enforces canonical icon coverage while constructing a
// process registry. Coverage drift means the embedded registry no longer
// describes the compiled connector set, so it aborts with the remediation
// instead of serving connectors with missing icon identity. Layered
// constructors may each enforce it: a registry that has not been mutated since
// its last successful validation is already covered, so repeat calls are free.
func (r *Registry) MustValidateIconCoverage() {
	r.mu.RLock()
	validated := r.iconCoverageValidated
	r.mu.RUnlock()
	if validated {
		return
	}
	if err := r.ValidateIconCoverage(); err != nil {
		panic("validate connector icon coverage: " + err.Error() + "; regenerate internal/connectors/icon_data.json with `make icons-generate`")
	}
	r.mu.Lock()
	r.iconCoverageValidated = true
	r.mu.Unlock()
}

func ValidateConnectorIcons(connectorsDir string, defs []Definition, metas []Metadata) error {
	for _, def := range defs {
		if def.Icon == nil {
			return fmt.Errorf("connector icon %s: missing icon registry entry", def.Name)
		}
		if err := ValidateConnectorIcon(connectorsDir, def.Name, *def.Icon); err != nil {
			return err
		}
	}
	for _, meta := range metas {
		if meta.Icon == nil {
			return fmt.Errorf("connector icon %s: missing icon registry entry", meta.Name)
		}
		if err := ValidateConnectorIcon(connectorsDir, meta.Name, *meta.Icon); err != nil {
			return err
		}
	}
	return nil
}

func ValidateConnectorIcon(connectorsDir, connector string, icon ConnectorIcon) error {
	if strings.TrimSpace(icon.ID) == "" || strings.TrimSpace(icon.Path) == "" || strings.TrimSpace(icon.Source) == "" || strings.TrimSpace(icon.ReviewStatus) == "" {
		return fmt.Errorf("connector icon %s: incomplete icon metadata", connector)
	}
	if !validIconReviewStatus(icon.ReviewStatus) {
		return fmt.Errorf("connector icon %s: unsupported review_status %q", connector, icon.ReviewStatus)
	}
	if err := validateConnectorIconPath(icon.Path); err != nil {
		return fmt.Errorf("connector icon %s: %w", connector, err)
	}
	clean := path.Clean(icon.Path)
	assetPath := filepath.Join(connectorsDir, filepath.FromSlash(clean))
	data, err := os.ReadFile(assetPath)
	if err != nil {
		return fmt.Errorf("connector icon %s: missing asset %s: %w", connector, icon.Path, err)
	}
	if err := ValidateConnectorIconSVGContent(connector, data); err != nil {
		return err
	}
	return nil
}

func ValidateConnectorIconSVGContent(connector string, data []byte) error {
	trimmed := bytes.TrimSpace(data)
	content := string(trimmed)
	lower := strings.ToLower(content)
	if !strings.HasPrefix(lower, "<svg") && !strings.HasPrefix(lower, "<?xml") {
		return fmt.Errorf("connector icon %s: asset is not an svg document", connector)
	}
	for _, forbidden := range []string{"<script", "<foreignobject", "javascript:"} {
		if strings.Contains(lower, forbidden) {
			return fmt.Errorf("connector icon %s: svg contains forbidden content %q", connector, forbidden)
		}
	}
	for _, pattern := range unsafeSVGPatterns {
		if pattern.re.MatchString(content) {
			return fmt.Errorf("connector icon %s: svg contains forbidden %s", connector, pattern.label)
		}
	}
	if !strings.Contains(lower, "</svg>") {
		return fmt.Errorf("connector icon %s: svg document is missing closing svg tag", connector)
	}
	return nil
}

func validIconReviewStatus(status string) bool {
	switch status {
	case IconReviewUpstreamSeeded, IconReviewOfficial, IconReviewManualOverride, IconReviewPolymetrics, IconReviewSimpleIconsCC0Trademark:
		return true
	default:
		return false
	}
}

func validateConnectorIconPath(iconPath string) error {
	clean := path.Clean(iconPath)
	if clean != iconPath || strings.HasPrefix(clean, "../") || strings.HasPrefix(clean, "/") || !strings.HasPrefix(clean, "icons/") {
		return fmt.Errorf("invalid path %q: must stay under icons/", iconPath)
	}
	if path.Ext(clean) != ".svg" {
		return fmt.Errorf("invalid path %q: icon assets must be .svg", iconPath)
	}
	return nil
}

func hasLegacyIconConnectorPrefix(connector string) bool {
	return strings.HasPrefix(connector, "source-") || strings.HasPrefix(connector, "destination-")
}

func canonicalIconPathForOwnedFile(rel string) (string, bool) {
	clean := strings.Trim(strings.TrimPrefix(filepath.ToSlash(rel), "./"), "/")
	for _, prefix := range []string{"docs/connectors/", "website/public/connectors/"} {
		if strings.HasPrefix(clean, prefix) {
			iconPath := strings.TrimPrefix(clean, prefix)
			if strings.HasPrefix(iconPath, "icons/") && validateConnectorIconPath(iconPath) == nil {
				return iconPath, true
			}
		}
	}
	return "", false
}

func ConnectorIconOwnerForPath(rel string) (string, error) {
	iconPath, ok := canonicalIconPathForOwnedFile(rel)
	if !ok {
		return "", fmt.Errorf("unsupported connector icon path %q: must be under docs/connectors/icons/ or website/public/connectors/icons/", rel)
	}
	entries, err := connectorIconRegistryEntries()
	if err != nil {
		return "", err
	}
	var owners []string
	var builtins []string
	for _, entry := range entries {
		if entry.Path != iconPath {
			continue
		}
		if !entry.Implemented {
			builtins = append(builtins, entry.Connector)
			continue
		}
		owners = append(owners, entry.Connector)
	}
	sort.Strings(owners)
	sort.Strings(builtins)
	switch len(owners) {
	case 0:
		if len(builtins) > 0 {
			return "", fmt.Errorf("%w: %q is declared by %s and does not authorize connector ownership", ErrConnectorIconPathRuntimeBuiltin, rel, strings.Join(builtins, ", "))
		}
		return "", fmt.Errorf("undeclared connector icon path %q", rel)
	case 1:
		return owners[0], nil
	default:
		return "", fmt.Errorf("ambiguous connector icon path %q is declared by %s", rel, strings.Join(owners, ", "))
	}
}

func ValidateConnectorIconOwnershipPaths(paths []string) (map[string]string, error) {
	owners := make(map[string]string, len(paths))
	seen := map[string]bool{}
	for _, rel := range paths {
		clean := strings.Trim(strings.TrimPrefix(filepath.ToSlash(rel), "./"), "/")
		if seen[clean] {
			return nil, fmt.Errorf("duplicate connector icon path %q", rel)
		}
		seen[clean] = true
		owner, err := ConnectorIconOwnerForPath(clean)
		if err != nil {
			return nil, err
		}
		owners[clean] = owner
	}
	return owners, nil
}
