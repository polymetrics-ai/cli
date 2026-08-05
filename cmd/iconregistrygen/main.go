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
	"path"
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
	Connector           string `json:"connector"`
	ID                  string `json:"id"`
	Title               string `json:"title,omitempty"`
	SimpleIconSlug      string `json:"simple_icon_slug,omitempty"`
	SimpleIconHex       string `json:"simple_icon_hex,omitempty"`
	Path                string `json:"path"`
	Source              string `json:"source"`
	SourceURL           string `json:"-"`
	UpstreamRecord      string `json:"-"`
	ReviewStatus        string `json:"review_status"`
	ReviewURL           string `json:"review_url,omitempty"`
	License             string `json:"license,omitempty"`
	Attribution         string `json:"attribution,omitempty"`
	Match               string `json:"match,omitempty"`
	MatchedBy           string `json:"matched_by,omitempty"`
	FallbackDisposition string `json:"fallback_disposition,omitempty"`
	Implemented         bool   `json:"implemented"`
}

type iconAsset struct {
	Path      string
	SourceURL string
}

type buildOptions struct {
	ImplementedConnectors map[string]bool
	CuratedEntries        []iconEntry
	IncludeLocalBuiltins  bool
}

func main() {
	source := flag.String("source", os.Getenv("PM_ICON_REGISTRY_SOURCE"), "connector registry JSON URL or local path")
	out := flag.String("out", "internal/connectors/icon_data.json", "embedded connector icon registry output")
	iconsDir := flag.String("icons-dir", "docs/connectors/icons", "local connector SVG asset directory")
	defsDir := flag.String("defs-dir", "internal/connectors/defs", "connector definition directory used to scope implemented icon entries")
	curated := flag.String("curated", "", "existing canonical icon registry whose authored rows override upstream records; defaults to --out when present")
	download := flag.Bool("download", true, "download connector icon SVG assets")
	flag.Parse()
	if strings.TrimSpace(*source) == "" {
		fatal(errors.New("icon registry source is required; set --source or PM_ICON_REGISTRY_SOURCE"))
	}

	registry, err := loadRegistry(*source)
	if err != nil {
		fatal(err)
	}
	implemented, err := loadImplementedConnectors(*defsDir)
	if err != nil {
		fatal(err)
	}
	curatedPath := strings.TrimSpace(*curated)
	if curatedPath == "" {
		curatedPath = *out
	}
	curatedEntries, err := loadCuratedIconEntries(curatedPath)
	if err != nil {
		fatal(err)
	}
	entries, assets, err := buildIconEntries(registry, buildOptions{
		ImplementedConnectors: implemented,
		CuratedEntries:        curatedEntries,
		IncludeLocalBuiltins:  true,
	})
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
	var err error
	if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		client := http.Client{Timeout: 30 * time.Second}
		var resp *http.Response
		resp, err = client.Get(source)
		if err != nil {
			return registryFile{}, fmt.Errorf("fetch registry: %w", err)
		}
		defer func() { _ = resp.Body.Close() }()
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

func buildIconEntries(registry registryFile, opts buildOptions) ([]iconEntry, []iconAsset, error) {
	builtinConnectors := map[string]bool{}
	if opts.IncludeLocalBuiltins {
		for _, entry := range localIconEntries() {
			builtinConnectors[entry.Connector] = true
		}
	}

	// The canonical registry is authored state. Its source/review fields record
	// provenance, not whether an existing bare key may settle an upstream
	// disagreement before raw source/destination rows are collapsed.
	curatedByConnector := make(map[string]iconEntry, len(opts.CuratedEntries))
	for _, entry := range opts.CuratedEntries {
		if _, exists := curatedByConnector[entry.Connector]; exists {
			return nil, nil, fmt.Errorf("duplicate curated icon entry for %q", entry.Connector)
		}
		if !includeConnector(entry.Connector, opts.ImplementedConnectors) && !builtinConnectors[entry.Connector] {
			return nil, nil, fmt.Errorf("curated icon entry %q has no connector definition or runtime builtin owner", entry.Connector)
		}
		entry.Implemented = opts.ImplementedConnectors[entry.Connector]
		curatedByConnector[entry.Connector] = entry
	}

	byConnector := map[string]iconEntry{}
	for _, rawGroup := range [][]map[string]any{registry.Sources, registry.Destinations} {
		for _, raw := range rawGroup {
			entry, ok, err := buildIconEntry(raw)
			if err != nil {
				return nil, nil, err
			}
			if !ok || !includeConnector(entry.Connector, opts.ImplementedConnectors) {
				continue
			}
			if _, curated := curatedByConnector[entry.Connector]; curated {
				continue
			}
			if existing, exists := byConnector[entry.Connector]; exists {
				merged, err := mergeCollapsedIconEntry(existing, entry)
				if err != nil {
					return nil, nil, err
				}
				byConnector[entry.Connector] = merged
				continue
			}
			entry.Implemented = opts.ImplementedConnectors[entry.Connector]
			byConnector[entry.Connector] = entry
		}
	}

	for connector, entry := range curatedByConnector {
		entry.Implemented = opts.ImplementedConnectors[connector]
		byConnector[entry.Connector] = entry
	}

	for connector := range opts.ImplementedConnectors {
		if _, ok := byConnector[connector]; !ok {
			byConnector[connector] = fallbackIconEntry(connector, true)
		}
	}

	if opts.IncludeLocalBuiltins {
		for _, entry := range localIconEntries() {
			if _, exists := byConnector[entry.Connector]; exists {
				continue
			}
			entry.Implemented = opts.ImplementedConnectors[entry.Connector]
			byConnector[entry.Connector] = entry
		}
	}

	entries := make([]iconEntry, 0, len(byConnector))
	for _, entry := range byConnector {
		if err := validateBuiltIconEntry(entry); err != nil {
			return nil, nil, err
		}
		entries = append(entries, entry)
	}
	sort.SliceStable(entries, func(i, j int) bool { return entries[i].Connector < entries[j].Connector })

	type assetOwner struct {
		connector string
		sourceURL string
	}
	assetOwners := map[string]assetOwner{}
	assetURLs := map[string]string{}
	for _, entry := range entries {
		existing, exists := assetOwners[entry.Path]
		switch {
		case !exists:
			assetOwners[entry.Path] = assetOwner{connector: entry.Connector, sourceURL: entry.SourceURL}
		case existing.sourceURL == "" && entry.SourceURL != "":
			assetOwners[entry.Path] = assetOwner{connector: entry.Connector, sourceURL: entry.SourceURL}
		case entry.SourceURL != "" && existing.sourceURL != entry.SourceURL:
			return nil, nil, fmt.Errorf(
				"conflicting source URLs for shared icon path %q: %q (%s) vs %q (%s)",
				entry.Path,
				existing.sourceURL,
				existing.connector,
				entry.SourceURL,
				entry.Connector,
			)
		}
		if entry.SourceURL != "" {
			assetURLs[entry.Path] = entry.SourceURL
		}
	}
	assets := make([]iconAsset, 0, len(assetURLs))
	for path, sourceURL := range assetURLs {
		assets = append(assets, iconAsset{Path: path, SourceURL: sourceURL})
	}
	sort.SliceStable(assets, func(i, j int) bool { return assets[i].Path < assets[j].Path })
	return entries, assets, nil
}

func includeConnector(connector string, implemented map[string]bool) bool {
	return len(implemented) == 0 || implemented[connector]
}

func mergeCollapsedIconEntry(a, b iconEntry) (iconEntry, error) {
	if a.ID != b.ID || a.Path != b.Path {
		return iconEntry{}, collapsedIconConflictError(
			a,
			b,
			fmt.Sprintf("conflicting icon identities %s/%s vs %s/%s", a.ID, a.Path, b.ID, b.Path),
		)
	}
	if a.SourceURL != b.SourceURL {
		return iconEntry{}, collapsedIconConflictError(a, b, "conflicting source URLs")
	}
	if reviewRank(b.ReviewStatus) > reviewRank(a.ReviewStatus) || (b.ReviewURL != "" && a.ReviewURL == "") {
		b.Implemented = a.Implemented || b.Implemented
		return b, nil
	}
	if a.SourceURL == "" {
		a.SourceURL = b.SourceURL
	}
	a.Implemented = a.Implemented || b.Implemented
	return a, nil
}

func collapsedIconConflictError(a, b iconEntry, reason string) error {
	return fmt.Errorf(
		"ambiguous source/destination icon collapse for %q: %s between upstream records %q (%q) and %q (%q); add a curated icon entry for %q to resolve",
		a.Connector,
		reason,
		a.UpstreamRecord,
		a.SourceURL,
		b.UpstreamRecord,
		b.SourceURL,
		a.Connector,
	)
}

func reviewRank(status string) int {
	switch status {
	case connectors.IconReviewOfficial:
		return 4
	case connectors.IconReviewManualOverride, connectors.IconReviewSimpleIconsCC0Trademark:
		return 3
	case connectors.IconReviewUpstreamSeeded:
		return 2
	case connectors.IconReviewPolymetrics:
		return 1
	default:
		return 0
	}
}

func validateBuiltIconEntry(entry iconEntry) error {
	if entry.Connector == "" {
		return errors.New("connector icon entry missing connector")
	}
	if strings.HasPrefix(entry.Connector, "source-") || strings.HasPrefix(entry.Connector, "destination-") {
		return fmt.Errorf("connector icon entry %q must use a bare connector identifier", entry.Connector)
	}
	if entry.ID == "" || entry.Path == "" || entry.Source == "" || entry.ReviewStatus == "" {
		return fmt.Errorf("connector icon entry %q has incomplete metadata", entry.Connector)
	}
	clean := path.Clean(entry.Path)
	if strings.Contains(entry.Path, `\`) || !strings.HasPrefix(entry.Path, "icons/") || clean != entry.Path || path.Ext(clean) != ".svg" {
		return fmt.Errorf("connector icon entry %q has invalid path %q", entry.Connector, entry.Path)
	}
	return nil
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
	slug := canonicalConnectorKey(dockerSlug(repo))
	if slug == "" {
		return iconEntry{}, false, errors.New("connector has empty canonical name")
	}
	iconName := stringValue(raw["icon"])
	iconURL := stringValue(raw["iconUrl"])
	if iconURL == "" {
		return iconEntry{}, false, nil
	}
	iconID := iconIDFromName(iconName, slug)
	return iconEntry{
		Connector:      slug,
		ID:             iconID,
		Path:           "icons/" + iconID + ".svg",
		Source:         connectors.IconSourceUpstream,
		SourceURL:      iconURL,
		UpstreamRecord: dockerSlug(repo),
		ReviewStatus:   connectors.IconReviewUpstreamSeeded,
	}, true, nil
}

func dockerSlug(repo string) string {
	_, slug, ok := strings.Cut(repo, "/")
	if !ok {
		return strings.TrimSpace(repo)
	}
	return strings.TrimSpace(slug)
}

func canonicalConnectorKey(slug string) string {
	slug = strings.TrimSpace(slug)
	for _, prefix := range []string{"source-", "destination-"} {
		slug = strings.TrimPrefix(slug, prefix)
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

func loadImplementedConnectors(defsDir string) (map[string]bool, error) {
	implemented := map[string]bool{}
	entries, err := os.ReadDir(defsDir)
	if err != nil {
		return nil, fmt.Errorf("read connector definitions: %w", err)
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		metadataPath := filepath.Join(defsDir, entry.Name(), "metadata.json")
		data, err := os.ReadFile(metadataPath)
		if err != nil {
			return nil, fmt.Errorf("read connector metadata %s: %w", entry.Name(), err)
		}
		var metadata struct {
			Name string `json:"name"`
		}
		if err := json.Unmarshal(data, &metadata); err != nil {
			return nil, fmt.Errorf("decode connector metadata %s: %w", entry.Name(), err)
		}
		name := strings.TrimSpace(metadata.Name)
		if name == "" {
			return nil, fmt.Errorf("connector metadata %s: name is required", entry.Name())
		}
		if name != entry.Name() {
			return nil, fmt.Errorf("connector metadata %s: name %q must match directory", entry.Name(), name)
		}
		if name != canonicalConnectorKey(name) {
			return nil, fmt.Errorf("connector metadata %s: name must be a bare identifier", entry.Name())
		}
		implemented[name] = true
	}
	return implemented, nil
}

func loadCuratedIconEntries(path string) ([]iconEntry, error) {
	if strings.TrimSpace(path) == "" {
		return nil, nil
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read curated icon registry: %w", err)
	}
	var entries []iconEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("decode curated icon registry: %w", err)
	}
	for i := range entries {
		entry := &entries[i]
		entry.Connector = strings.TrimSpace(entry.Connector)
		entry.ID = strings.TrimSpace(entry.ID)
		entry.Path = strings.TrimSpace(entry.Path)
		entry.Source = strings.TrimSpace(entry.Source)
		entry.ReviewStatus = strings.TrimSpace(entry.ReviewStatus)
		entry.ReviewURL = strings.TrimSpace(entry.ReviewURL)
		// Curated entries are authored state, not external input: an empty or
		// legacy-prefixed connector key must fail loudly rather than being
		// silently dropped and backfilled from upstream/fallback data, which
		// would destroy hand-curated review status, review URL, and source
		// attribution while reporting success.
		if entry.Connector == "" {
			return nil, fmt.Errorf("curated icon registry %s: entry missing connector", path)
		}
		if entry.Connector != canonicalConnectorKey(entry.Connector) {
			return nil, fmt.Errorf("curated icon registry %s: connector %q must be a bare connector identifier", path, entry.Connector)
		}
	}
	return entries, nil
}

func fallbackIconEntry(connector string, implemented bool) iconEntry {
	entry := localIconEntry(connector, "pm-sample", connectors.IconSourcePolymetrics, connectors.IconReviewPolymetrics, "https://github.com/polymetrics-ai/cli")
	entry.Implemented = implemented
	entry.FallbackDisposition = "reviewed-polymetrics-sample-fallback"
	return entry
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
