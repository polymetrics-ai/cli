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
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		machineSource := current.Source
		machineSource.Kind = "official-reference"
		machine, isMachine, machineErr := parseBatchReferenceMachineArtifact(current.Raw, machineSource, fetch)
		if machineErr != nil {
			return batchArtifactInventory{}, machineErr
		}
		if isMachine {
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

		currentURL, currentURLErr := parseBatchReferenceURL(current.URL)
		if currentURLErr != nil {
			return batchArtifactInventory{}, batchArtifactInventoryUnknown("official reference page URL %q is unsafe: %v", current.URL, currentURLErr)
		}
		links, linksErr := batchHTMLReferenceLinks(current.Raw, currentURL)
		if linksErr != nil {
			return batchArtifactInventory{}, linksErr
		}
		for _, link := range links {
			if seen[link] {
				continue
			}
			if fetch == nil {
				return batchArtifactInventory{}, batchArtifactInventoryUnknown("official reference link %q cannot be fetched during traversal", link)
			}
			if len(seen) >= maxBatchArtifactReferenceDocuments {
				return batchArtifactInventory{}, batchArtifactInventoryUnknown("official reference traversal exceeds the %d-document limit", maxBatchArtifactReferenceDocuments)
			}
			linkRaw, linkErr := fetch(link)
			if linkErr != nil {
				return batchArtifactInventory{}, batchArtifactInventoryUnknown("official reference link %q could not be fetched: %v", link, linkErr)
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

func parseBatchReferenceArtifact(raw []byte, source batchArtifactSource, fetch batchArtifactFetchFunc) (batchArtifactInventory, error) {
	inventory, isMachine, err := parseBatchReferenceMachineArtifact(raw, source, fetch)
	if err != nil {
		return batchArtifactInventory{}, err
	}
	if isMachine {
		return inventory, nil
	}
	return parseBatchHTMLReference(raw, source, fetch)
}

func parseBatchReferenceMachineArtifact(raw []byte, source batchArtifactSource, fetch batchArtifactFetchFunc) (batchArtifactInventory, bool, error) {
	inventory, openAPIErr := parseBatchOpenAPIArtifactSource(raw, source, fetch)
	if openAPIErr == nil {
		return inventory, true, nil
	}
	inventory, postmanErr := parseBatchPostmanArtifact(raw, source)
	if postmanErr == nil {
		return inventory, true, nil
	}
	if batchReferenceLooksLikeMachineArtifact(raw, source.URL) {
		return batchArtifactInventory{}, true, batchArtifactInventoryUnknown("official reference %q is a machine-readable artifact that could not be parsed: OpenAPI/Swagger: %v; Postman: %v", source.URL, openAPIErr, postmanErr)
	}
	if !batchReferenceLooksLikeText(raw, source.URL) {
		return batchArtifactInventory{}, true, batchArtifactInventoryUnknown("official reference %q is neither a parseable OpenAPI/Swagger or Postman artifact nor a recognized reference document", source.URL)
	}
	return batchArtifactInventory{}, false, nil
}

func batchReferenceLooksLikeMachineArtifact(raw []byte, sourceURL string) bool {
	trimmed := bytes.TrimSpace(raw)
	document, err := parseBatchArtifactDocument(raw, batchArtifactSource{})
	if err == nil {
		fields, fieldsErr := batchYAMLFields(document.Root)
		if fieldsErr == nil {
			if _, ok := fields["openapi"]; ok {
				return true
			}
			if _, ok := fields["swagger"]; ok {
				return true
			}
			if info, ok := fields["info"]; ok {
				if infoFields, infoErr := batchYAMLFields(info); infoErr == nil {
					if schema, schemaErr := batchYAMLFieldString(infoFields, "schema"); schemaErr == nil && strings.Contains(strings.ToLower(schema), "postman") {
						return true
					}
				}
			}
			if len(trimmed) > 0 && trimmed[0] == '{' {
				if _, ok := fields["item"]; ok {
					return true
				}
			}
		}
	}
	return batchReferenceURLHasSuffix(sourceURL, ".json", ".yaml", ".yml")
}

func batchReferenceLooksLikeText(raw []byte, sourceURL string) bool {
	lower := strings.ToLower(string(raw))
	if batchHTMLExplicitOperation.Match(raw) || batchMarkdownReferenceLink.Match(raw) {
		return true
	}
	for _, marker := range []string{"<html", "<!doctype html", "<a ", "<a>", "<body", "<article", "<main"} {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return batchReferenceURLHasSuffix(sourceURL, ".html", ".htm", ".md", ".txt")
}

func batchReferenceURLHasSuffix(rawURL string, suffixes ...string) bool {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	path := strings.ToLower(parsed.Path)
	for _, suffix := range suffixes {
		if strings.HasSuffix(path, suffix) {
			return true
		}
	}
	return false
}

func batchHTMLReferenceLinks(raw []byte, base *url.URL) ([]string, error) {
	if base == nil {
		return nil, nil
	}
	links := map[string]bool{}
	addLink := func(candidate string) error {
		candidate = strings.TrimSpace(nethtml.UnescapeString(candidate))
		if !isLikelyBatchReferenceLink(candidate) {
			return nil
		}
		parsed, err := url.Parse(candidate)
		if err != nil {
			return batchArtifactInventoryUnknown("official reference contains a malformed selected link")
		}
		resolved := base.ResolveReference(parsed)
		resolved.Fragment = ""
		resolved.RawFragment = ""
		if resolved.Scheme != "https" || resolved.Host == "" || resolved.User != nil {
			return nil
		}
		if !strings.EqualFold(strings.TrimSuffix(resolved.Hostname(), "."), strings.TrimSuffix(base.Hostname(), ".")) {
			return nil
		}
		if validateBatchReferenceURLObject(resolved) == nil {
			links[resolved.String()] = true
		}
		return nil
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
			if err := addLink(attribute.Val); err != nil {
				return nil, err
			}
		}
	}
	for _, match := range batchMarkdownReferenceLink.FindAllSubmatch(raw, -1) {
		if err := addLink(string(match[1])); err != nil {
			return nil, err
		}
	}
	out := make([]string, 0, len(links))
	for link := range links {
		out = append(out, link)
	}
	sort.Strings(out)
	return out, nil
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
