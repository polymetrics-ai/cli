package connectors

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"

	_ "embed"
)

const (
	IconSourceOfficial    = "official"
	IconSourceUpstream    = "upstream_registry"
	IconSourcePolymetrics = "polymetrics"

	IconReviewUpstreamSeeded = "upstream_seeded"
	IconReviewOfficial       = "official_verified"
	IconReviewManualOverride = "manual_override"
	IconReviewPolymetrics    = "polymetrics"
)

// ConnectorIcon identifies a local SVG asset and how it was sourced. The path is
// relative to the connector docs root, e.g. icons/github.svg.
type ConnectorIcon struct {
	ID           string `json:"id"`
	Path         string `json:"path"`
	Source       string `json:"source"`
	ReviewStatus string `json:"review_status"`
	ReviewURL    string `json:"review_url,omitempty"`
}

type connectorIconEntry struct {
	Connector    string `json:"connector"`
	ID           string `json:"id"`
	Path         string `json:"path"`
	Source       string `json:"source"`
	ReviewStatus string `json:"review_status"`
	ReviewURL    string `json:"review_url,omitempty"`
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

var connectorIcons = struct {
	once sync.Once
	by   map[string]ConnectorIcon
	err  error
}{}

func ConnectorIconFor(name string) (ConnectorIcon, bool) {
	icons, err := connectorIconRegistry()
	if err != nil {
		return ConnectorIcon{}, false
	}
	for _, candidate := range []string{name, "source-" + name, "destination-" + name} {
		icon, ok := icons[candidate]
		if ok {
			return icon, true
		}
	}
	return ConnectorIcon{}, false
}

func ConnectorIconEntries() []connectorIconEntry {
	var entries []connectorIconEntry
	if err := json.Unmarshal(connectorIconData, &entries); err != nil {
		return nil
	}
	sort.SliceStable(entries, func(i, j int) bool { return entries[i].Connector < entries[j].Connector })
	return entries
}

func MetadataWithIcon(meta Metadata) Metadata {
	if meta.Icon != nil {
		icon := *meta.Icon
		meta.Icon = &icon
		return meta
	}
	if icon, ok := ConnectorIconFor(meta.Name); ok {
		meta.Icon = &icon
	} else {
		icon := fallbackConnectorIcon(meta.Name)
		meta.Icon = &icon
	}
	return meta
}

func fallbackConnectorIcon(name string) ConnectorIcon {
	return ConnectorIcon{
		ID:           "pm-" + name,
		Path:         "icons/pm-sample.svg",
		Source:       IconSourcePolymetrics,
		ReviewStatus: IconReviewPolymetrics,
		ReviewURL:    "https://github.com/polymetrics-ai/cli",
	}
}

func manifestWithIcon(manifest Manifest) Manifest {
	manifest.Metadata = MetadataWithIcon(manifest.Metadata)
	return manifest
}

func connectorIconRegistry() (map[string]ConnectorIcon, error) {
	connectorIcons.once.Do(func() {
		var entries []connectorIconEntry
		if err := json.Unmarshal(connectorIconData, &entries); err != nil {
			connectorIcons.err = fmt.Errorf("decode connector icon registry: %w", err)
			return
		}
		icons := make(map[string]ConnectorIcon, len(entries))
		for _, entry := range entries {
			connector := strings.TrimSpace(entry.Connector)
			if connector == "" {
				connectorIcons.err = fmt.Errorf("connector icon registry entry missing connector")
				return
			}
			icons[connector] = ConnectorIcon{ID: entry.ID, Path: entry.Path, Source: entry.Source, ReviewStatus: entry.ReviewStatus, ReviewURL: entry.ReviewURL}
		}
		connectorIcons.by = icons
	})
	return connectorIcons.by, connectorIcons.err
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
	clean := path.Clean(icon.Path)
	if clean != icon.Path || strings.HasPrefix(clean, "../") || strings.HasPrefix(clean, "/") || !strings.HasPrefix(clean, "icons/") {
		return fmt.Errorf("connector icon %s: invalid path %q: must stay under icons/", connector, icon.Path)
	}
	if path.Ext(clean) != ".svg" {
		return fmt.Errorf("connector icon %s: invalid path %q: icon assets must be .svg", connector, icon.Path)
	}
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
	case IconReviewUpstreamSeeded, IconReviewOfficial, IconReviewManualOverride, IconReviewPolymetrics:
		return true
	default:
		return false
	}
}
