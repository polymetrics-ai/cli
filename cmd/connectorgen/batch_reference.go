package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strings"

	nethtml "golang.org/x/net/html"
)

var batchHTMLExplicitOperation = regexp.MustCompile(`(?i)\b(GET|POST|PUT|PATCH|DELETE|HEAD|OPTIONS|TRACE)\s+(/[A-Za-z0-9._~%!$&'()*+,;=:@/{}\[\]-]+)`)
var batchMarkdownReferenceLink = regexp.MustCompile(`\[[^\]]+\]\(([^)\s]+)\)`)

// parseBatchHTMLReference is deliberately conservative: it only normalizes
// method/path pairs printed by the provider and machine-readable sources
// linked by provider-owned pages. It never turns prose such as "get started"
// into an operation, and traversal is capped and cacheable through fetch.
func parseBatchHTMLReference(raw []byte, source batchArtifactSource, fetch batchArtifactFetchFunc) (batchArtifactInventory, error) {
	if source.URL == "" {
		return batchArtifactInventory{}, batchArtifactInventoryUnknown("official HTML reference has no source URL")
	}
	if source.Kind == "" {
		source.Kind = "html_reference"
	}
	if source.SHA256 == "" {
		source.SHA256 = fmt.Sprintf("%x", sha256Bytes(raw))
	}
	rootURL, err := parseBatchReferenceURL(source.URL)
	if err != nil {
		return batchArtifactInventory{}, batchArtifactInventoryUnknown("official HTML reference URL is unsafe: %v", err)
	}
	type page struct {
		URL    string
		Raw    []byte
		Source batchArtifactSource
	}
	queue := []page{{URL: rootURL.String(), Raw: raw, Source: source}}
	seen := map[string]bool{rootURL.String(): true}
	sources := []batchArtifactSource{source}
	inventory := batchArtifactInventory{}
	totalBytes := len(raw)
	for len(queue) > 0 && len(seen) <= maxBatchArtifactReferenceDocuments {
		current := queue[0]
		queue = queue[1:]

		machineSource := current.Source
		machineSource.Kind = "official-reference"
		if machine, machineErr := parseBatchOpenAPIArtifactSource(current.Raw, machineSource, fetch); machineErr == nil {
			inventory = mergeBatchArtifactInventories(inventory, machine)
		} else if machine, machineErr := parseBatchPostmanArtifact(current.Raw, machineSource); machineErr == nil {
			inventory = mergeBatchArtifactInventories(inventory, machine)
		}

		for _, match := range batchHTMLExplicitOperation.FindAllSubmatch(current.Raw, -1) {
			method := strings.ToUpper(string(match[1]))
			path := normalizeBatchHTMLOperationPath(string(match[2]))
			if path == "" {
				continue
			}
			endpoint := batchArtifactEndpoint{
				Method:           method,
				Path:             path,
				Summary:          fmt.Sprintf("%s %s", method, path),
				SourceURL:        current.Source.URL,
				SourceKind:       current.Source.Kind,
				SourceVersion:    current.Source.Version,
				SourceRetrieved:  current.Source.Retrieved,
				SourceSHA256:     current.Source.SHA256,
				SourceCoordinate: fmt.Sprintf("%s#%s %s", current.URL, method, path),
			}
			inventory = mergeBatchArtifactInventories(inventory, batchArtifactInventory{Endpoints: []batchArtifactEndpoint{endpoint}})
		}

		for _, link := range batchHTMLReferenceLinks(current.Raw, rootURL) {
			if seen[link] || len(seen) >= maxBatchArtifactReferenceDocuments || fetch == nil {
				continue
			}
			linkRaw, linkErr := fetch(link)
			if linkErr != nil {
				continue
			}
			if len(linkRaw) > maxMaterializeArtifactBytes || totalBytes > maxBatchArtifactReferenceBytes-len(linkRaw) {
				return batchArtifactInventory{}, batchArtifactInventoryUnknown("official reference traversal exceeds the bounded %d-byte source budget", maxBatchArtifactReferenceBytes)
			}
			totalBytes += len(linkRaw)
			seen[link] = true
			linkSource := batchArtifactSource{
				URL:       link,
				Kind:      "html_reference",
				Version:   source.Version,
				Retrieved: source.Retrieved,
				SHA256:    fmt.Sprintf("%x", sha256Bytes(linkRaw)),
			}
			sources = append(sources, linkSource)
			queue = append(queue, page{URL: link, Raw: linkRaw, Source: linkSource})
		}
	}
	inventory.Sources = append(inventory.Sources, sources...)
	inventory = deduplicateBatchArtifactSources(inventory)
	if len(inventory.Endpoints) == 0 {
		return batchArtifactInventory{}, batchArtifactInventoryUnknown("official HTML reference traversal found no explicit operations or linked machine-readable contract")
	}
	return inventory, nil
}

func batchHTMLReferenceLinks(raw []byte, base *url.URL) []string {
	if base == nil {
		return nil
	}
	links := map[string]bool{}
	addLink := func(candidate string) {
		candidate = strings.TrimSpace(nethtml.UnescapeString(candidate))
		if !isLikelyBatchReferenceLink(candidate) {
			return
		}
		parsed, err := url.Parse(candidate)
		if err != nil {
			return
		}
		resolved := base.ResolveReference(parsed)
		if resolved.Scheme != "https" || resolved.Host == "" || resolved.User != nil || resolved.Fragment != "" {
			return
		}
		if !strings.EqualFold(strings.TrimSuffix(resolved.Hostname(), "."), strings.TrimSuffix(base.Hostname(), ".")) {
			return
		}
		resolved.Fragment = ""
		if validateBatchReferenceURLObject(resolved) == nil {
			links[resolved.String()] = true
		}
	}
	tokenizer := nethtml.NewTokenizer(bytes.NewReader(raw))
	for {
		tokenType := tokenizer.Next()
		if tokenType == nethtml.ErrorToken {
			break
		}
		if tokenType != nethtml.StartTagToken && tokenType != nethtml.SelfClosingTagToken {
			continue
		}
		token := tokenizer.Token()
		for _, attribute := range token.Attr {
			if attribute.Key != "href" && attribute.Key != "src" {
				continue
			}
			addLink(attribute.Val)
		}
	}
	for _, match := range batchMarkdownReferenceLink.FindAllSubmatch(raw, -1) {
		addLink(string(match[1]))
	}
	out := make([]string, 0, len(links))
	for link := range links {
		out = append(out, link)
	}
	sort.Strings(out)
	return out
}

func validateBatchReferenceURLObject(parsed *url.URL) error {
	if parsed == nil {
		return fmt.Errorf("reference URL is nil")
	}
	return validateBatchArtifactURLObjectWithQuery(parsed, true)
}

func isLikelyBatchReferenceLink(raw string) bool {
	lower := strings.ToLower(strings.TrimSpace(raw))
	if lower == "" || strings.HasPrefix(lower, "#") || strings.HasPrefix(lower, "mailto:") || strings.HasPrefix(lower, "javascript:") {
		return false
	}
	for _, marker := range []string{".json", ".yaml", ".yml", "openapi", "swagger", "postman", "reference", "endpoint", "operation", "api-doc", "api/"} {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return false
}

func normalizeBatchHTMLOperationPath(raw string) string {
	path := nethtml.UnescapeString(strings.TrimSpace(raw))
	if cut := strings.IndexAny(path, "?#\"'<>"); cut >= 0 {
		path = path[:cut]
	}
	if path == "" || !strings.HasPrefix(path, "/") || strings.HasPrefix(path, "//") {
		return ""
	}
	return path
}

func deduplicateBatchArtifactSources(inventory batchArtifactInventory) batchArtifactInventory {
	seen := map[string]bool{}
	sources := make([]batchArtifactSource, 0, len(inventory.Sources))
	for _, source := range inventory.Sources {
		if source.URL == "" || seen[source.URL] {
			continue
		}
		seen[source.URL] = true
		sources = append(sources, source)
	}
	inventory.Sources = sources
	return inventory
}

func sha256Bytes(raw []byte) [32]byte {
	return sha256.Sum256(raw)
}
