// Command iconregistrygen regenerates connector icon metadata and local SVG
// assets from an upstream connector registry. The generated metadata is embedded
// by package connectors; the SVG assets are released with connector docs under
// docs/connectors/icons.
package main

import (
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
	"sync"
	"time"
	"unicode"

	"polymetrics.ai/internal/connectors"
)

type registryFile struct {
	Sources      []map[string]any `json:"sources"`
	Destinations []map[string]any `json:"destinations"`
}

type iconEntry struct {
	Connector    string `json:"connector"`
	ID           string `json:"id"`
	Path         string `json:"path"`
	Source       string `json:"source"`
	SourceURL    string `json:"-"`
	ReviewStatus string `json:"review_status"`
	ReviewURL    string `json:"review_url,omitempty"`
}

type iconAsset struct {
	Path      string
	SourceURL string
}

func main() {
	source := flag.String("source", os.Getenv("PM_ICON_REGISTRY_SOURCE"), "connector registry JSON URL or local path")
	out := flag.String("out", "internal/connectors/icon_data.json", "embedded connector icon registry output")
	iconsDir := flag.String("icons-dir", "docs/connectors/icons", "local connector SVG asset directory")
	download := flag.Bool("download", true, "download connector icon SVG assets")
	flag.Parse()
	if strings.TrimSpace(*source) == "" {
		fatal(errors.New("icon registry source is required; set --source or PM_ICON_REGISTRY_SOURCE"))
	}

	registry, err := loadRegistry(*source)
	if err != nil {
		fatal(err)
	}
	entries, assets, err := buildIconEntries(registry)
	if err != nil {
		fatal(err)
	}
	if err := writeIconRegistry(*out, entries); err != nil {
		fatal(err)
	}
	if err := writeIconAssets(*iconsDir, assets, *download); err != nil {
		fatal(err)
	}
	fmt.Printf("generated %d connector icon entries and %d SVG assets\n", len(entries), len(assets)+len(localIconIDs()))
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "iconregistrygen:", err)
	os.Exit(1)
}

func loadRegistry(source string) (registryFile, error) {
	var data []byte
	if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		client := http.Client{Timeout: 30 * time.Second}
		resp, err := client.Get(source)
		if err != nil {
			return registryFile{}, fmt.Errorf("fetch registry: %w", err)
		}
		defer func() { _ = resp.Body.Close() }()
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return registryFile{}, fmt.Errorf("fetch registry returned %s", resp.Status)
		}
		data, err = io.ReadAll(resp.Body)
		if err != nil {
			return registryFile{}, fmt.Errorf("read registry response: %w", err)
		}
	} else {
		var err error
		data, err = os.ReadFile(source)
		if err != nil {
			return registryFile{}, err
		}
	}
	var registry registryFile
	if err := json.Unmarshal(data, &registry); err != nil {
		return registryFile{}, fmt.Errorf("decode registry: %w", err)
	}
	return registry, nil
}

func buildIconEntries(registry registryFile) ([]iconEntry, []iconAsset, error) {
	entries := []iconEntry{}
	assetURLs := map[string]string{}
	for _, raw := range registry.Sources {
		entry, ok, err := buildIconEntry(raw)
		if err != nil {
			return nil, nil, err
		}
		if ok {
			entries = append(entries, entry)
			assetURLs[entry.Path] = entry.SourceURL
		}
	}
	for _, raw := range registry.Destinations {
		entry, ok, err := buildIconEntry(raw)
		if err != nil {
			return nil, nil, err
		}
		if ok {
			entries = append(entries, entry)
			assetURLs[entry.Path] = entry.SourceURL
		}
	}
	entries = append(entries, localIconEntries()...)
	sort.SliceStable(entries, func(i, j int) bool { return entries[i].Connector < entries[j].Connector })
	seen := map[string]bool{}
	for _, entry := range entries {
		if seen[entry.Connector] {
			return nil, nil, fmt.Errorf("duplicate connector icon entry %q", entry.Connector)
		}
		seen[entry.Connector] = true
	}
	assets := make([]iconAsset, 0, len(assetURLs))
	for path, sourceURL := range assetURLs {
		assets = append(assets, iconAsset{Path: path, SourceURL: sourceURL})
	}
	sort.SliceStable(assets, func(i, j int) bool { return assets[i].Path < assets[j].Path })
	return entries, assets, nil
}

func buildIconEntry(raw map[string]any) (iconEntry, bool, error) {
	if !boolValue(raw["public"]) || boolValue(raw["tombstone"]) {
		return iconEntry{}, false, nil
	}
	docs := stringValue(raw["documentationUrl"])
	if docs == "" {
		return iconEntry{}, false, nil
	}
	repo := stringValue(raw["dockerRepository"])
	if repo == "" {
		return iconEntry{}, false, errors.New("connector missing docker repository metadata")
	}
	slug := dockerSlug(repo)
	iconName := stringValue(raw["icon"])
	iconURL := stringValue(raw["iconUrl"])
	if iconURL == "" {
		return iconEntry{}, false, nil
	}
	iconID := iconIDFromName(iconName, slug)
	return iconEntry{
		Connector:    slug,
		ID:           iconID,
		Path:         "icons/" + iconID + ".svg",
		Source:       connectors.IconSourceUpstream,
		SourceURL:    iconURL,
		ReviewStatus: connectors.IconReviewUpstreamSeeded,
	}, true, nil
}

func dockerSlug(repo string) string {
	_, slug, ok := strings.Cut(repo, "/")
	if !ok {
		return repo
	}
	return slug
}

func iconIDFromName(iconName, slug string) string {
	base := strings.TrimSpace(iconName)
	if base == "" {
		base = slug + ".svg"
	}
	base = filepath.Base(base)
	base = strings.TrimSuffix(base, filepath.Ext(base))
	if genericIconID(base) {
		base = slug
	}
	return sanitizeIconID(base)
}

func genericIconID(value string) bool {
	switch sanitizeIconID(value) {
	case "icon", "logo", "favicon", "brand", "mark":
		return true
	default:
		return false
	}
}

func sanitizeIconID(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	var b strings.Builder
	lastDash := false
	for _, r := range value {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash {
			b.WriteByte('-')
			lastDash = true
		}
	}
	return strings.Trim(b.String(), "-")
}

func writeIconRegistry(path string, entries []iconEntry) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o644)
}

func writeIconAssets(iconsDir string, assets []iconAsset, download bool) error {
	if err := os.MkdirAll(iconsDir, 0o755); err != nil {
		return err
	}
	for _, entry := range localIconEntries() {
		path := filepath.Join(iconsDir, filepath.Base(entry.Path))
		if err := os.WriteFile(path, []byte(localIconSVG(entry.ID)), 0o644); err != nil {
			return fmt.Errorf("write built-in icon %s: %w", entry.ID, err)
		}
	}
	if !download {
		return nil
	}
	return downloadIconAssets(iconsDir, assets)
}

func downloadIconAssets(iconsDir string, assets []iconAsset) error {
	client := http.Client{Timeout: 30 * time.Second}
	jobs := make(chan iconAsset)
	errs := make(chan error, len(assets))
	var wg sync.WaitGroup
	workers := 12
	if len(assets) < workers {
		workers = len(assets)
	}
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for asset := range jobs {
				if err := downloadIconAsset(client, iconsDir, asset); err != nil {
					errs <- err
				}
			}
		}()
	}
	for _, asset := range assets {
		jobs <- asset
	}
	close(jobs)
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}

func downloadIconAsset(client http.Client, iconsDir string, asset iconAsset) error {
	resp, err := client.Get(asset.SourceURL)
	if err != nil {
		return fmt.Errorf("fetch icon %s: %w", asset.SourceURL, err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("fetch icon %s returned %s", asset.SourceURL, resp.Status)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read icon %s: %w", asset.SourceURL, err)
	}
	if err := connectors.ValidateConnectorIconSVGContent(asset.Path, data); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(iconsDir, filepath.Base(asset.Path)), data, 0o644)
}

func localIconEntries() []iconEntry {
	return []iconEntry{
		localIconEntry("file", "pm-file", connectors.IconSourcePolymetrics, connectors.IconReviewPolymetrics, "https://github.com/polymetrics-ai/cli"),
		localIconEntry("outbox", "pm-outbox", connectors.IconSourcePolymetrics, connectors.IconReviewPolymetrics, "https://github.com/polymetrics-ai/cli"),
		localIconEntry("sample", "pm-sample", connectors.IconSourcePolymetrics, connectors.IconReviewPolymetrics, "https://github.com/polymetrics-ai/cli"),
		localIconEntry("searxng", "searxng", "official_site", connectors.IconReviewManualOverride, "https://docs.searxng.org/"),
		localIconEntry("warehouse", "pm-warehouse", connectors.IconSourcePolymetrics, connectors.IconReviewPolymetrics, "https://github.com/polymetrics-ai/cli"),
	}
}

func localIconIDs() []string {
	return []string{"pm-file", "pm-outbox", "pm-sample", "pm-warehouse", "searxng"}
}

func localIconEntry(connector, id, source, reviewStatus, reviewURL string) iconEntry {
	return iconEntry{
		Connector:    connector,
		ID:           id,
		Path:         "icons/" + id + ".svg",
		Source:       source,
		ReviewStatus: reviewStatus,
		ReviewURL:    reviewURL,
	}
}

func localIconSVG(id string) string {
	label := strings.TrimPrefix(id, "pm-")
	if len(label) > 2 {
		label = label[:2]
	}
	label = strings.ToUpper(label)
	return fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64" role="img" aria-label="Polymetrics %s icon"><rect width="64" height="64" rx="14" fill="#064e3b"/><path d="M16 20h32v6H16zm0 12h24v6H16zm0 12h32v6H16z" fill="#d1fae5"/><text x="32" y="38" text-anchor="middle" font-family="Arial, sans-serif" font-size="16" font-weight="700" fill="#ecfdf5">%s</text></svg>`, id, label)
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
