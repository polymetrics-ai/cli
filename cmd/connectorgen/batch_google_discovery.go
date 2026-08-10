package main

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

// Google publishes REST Discovery documents for a number of its public APIs.
// They are machine-readable provider contracts, but not OpenAPI documents.
// This narrow extractor retains the documented method/path inventory and its
// source-local provenance; it does not infer request schemas or execution
// capabilities from the Discovery document.
type batchGoogleDiscoveryDocument struct {
	Kind      string                                  `json:"kind"`
	Methods   map[string]batchGoogleDiscoveryMethod   `json:"methods"`
	Resources map[string]batchGoogleDiscoveryResource `json:"resources"`
}

type batchGoogleDiscoveryResource struct {
	Methods   map[string]batchGoogleDiscoveryMethod   `json:"methods"`
	Resources map[string]batchGoogleDiscoveryResource `json:"resources"`
}

type batchGoogleDiscoveryMethod struct {
	ID          string `json:"id"`
	Path        string `json:"path"`
	FlatPath    string `json:"flatPath"`
	HTTPMethod  string `json:"httpMethod"`
	Description string `json:"description"`
}

func parseBatchGoogleDiscoveryArtifact(raw []byte, source batchArtifactSource) (batchArtifactInventory, error) {
	var document batchGoogleDiscoveryDocument
	if err := json.Unmarshal(raw, &document); err != nil {
		return batchArtifactInventory{}, fmt.Errorf("decode Google Discovery artifact: %w", err)
	}
	if document.Kind != "discovery#restDescription" {
		return batchArtifactInventory{}, fmt.Errorf("artifact is not a Google Discovery REST description")
	}
	if source.SHA256 == "" {
		source.SHA256 = fmt.Sprintf("%x", sha256Bytes(raw))
	}
	endpoints := make([]batchArtifactEndpoint, 0)
	if err := appendBatchGoogleDiscoveryMethods(&endpoints, document.Methods, "methods", source); err != nil {
		return batchArtifactInventory{}, err
	}
	if err := appendBatchGoogleDiscoveryResources(&endpoints, document.Resources, "resources", source); err != nil {
		return batchArtifactInventory{}, err
	}
	if len(endpoints) == 0 {
		return batchArtifactInventory{}, fmt.Errorf("google Discovery artifact has no HTTP methods")
	}
	sort.Slice(endpoints, func(i, j int) bool {
		if endpoints[i].Path != endpoints[j].Path {
			return endpoints[i].Path < endpoints[j].Path
		}
		if endpoints[i].Method != endpoints[j].Method {
			return batchArtifactMethodRank(endpoints[i].Method) < batchArtifactMethodRank(endpoints[j].Method)
		}
		return endpoints[i].SourceCoordinate < endpoints[j].SourceCoordinate
	})
	return batchArtifactInventory{Endpoints: endpoints, Sources: []batchArtifactSource{source}}, nil
}

func appendBatchGoogleDiscoveryResources(endpoints *[]batchArtifactEndpoint, resources map[string]batchGoogleDiscoveryResource, coordinate string, source batchArtifactSource) error {
	for _, name := range sortedBatchGoogleDiscoveryResourceNames(resources) {
		resource := resources[name]
		resourceCoordinate := coordinate + "[" + strconv.Quote(name) + "]"
		if err := appendBatchGoogleDiscoveryMethods(endpoints, resource.Methods, resourceCoordinate+".methods", source); err != nil {
			return err
		}
		if err := appendBatchGoogleDiscoveryResources(endpoints, resource.Resources, resourceCoordinate+".resources", source); err != nil {
			return err
		}
	}
	return nil
}

func appendBatchGoogleDiscoveryMethods(endpoints *[]batchArtifactEndpoint, methods map[string]batchGoogleDiscoveryMethod, coordinate string, source batchArtifactSource) error {
	for _, name := range sortedBatchGoogleDiscoveryMethodNames(methods) {
		method := methods[name]
		httpMethod, ok := batchArtifactHTTPMethodForPathItemKey(strings.ToLower(strings.TrimSpace(method.HTTPMethod)))
		if !ok {
			return fmt.Errorf("google Discovery method %q has unsupported HTTP method %q", firstNonEmpty(method.ID, name), method.HTTPMethod)
		}
		path, err := normalizeBatchGoogleDiscoveryPath(firstNonEmpty(method.Path, method.FlatPath))
		if err != nil {
			return fmt.Errorf("google Discovery method %q: %w", firstNonEmpty(method.ID, name), err)
		}
		summary := strings.TrimSpace(method.Description)
		if summary == "" {
			summary = firstNonEmpty(strings.TrimSpace(method.ID), name, fmt.Sprintf("%s %s", httpMethod, path))
		}
		*endpoints = append(*endpoints, batchArtifactEndpoint{
			Method:           httpMethod,
			Path:             path,
			Summary:          summary,
			SourceURL:        source.URL,
			SourceKind:       source.Kind,
			SourceVersion:    source.Version,
			SourceRetrieved:  source.Retrieved,
			SourceSHA256:     source.SHA256,
			SourceCoordinate: coordinate + "[" + strconv.Quote(name) + "]",
		})
	}
	return nil
}

func normalizeBatchGoogleDiscoveryPath(raw string) (string, error) {
	path := strings.TrimSpace(raw)
	if path == "" {
		return "", fmt.Errorf("path is required")
	}
	if strings.ContainsAny(path, "\r\n?#") {
		return "", fmt.Errorf("path %q contains a control or URL component", raw)
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	var normalized strings.Builder
	for offset := 0; offset < len(path); {
		start := strings.Index(path[offset:], "{+")
		if start < 0 {
			normalized.WriteString(path[offset:])
			break
		}
		start += offset
		normalized.WriteString(path[offset:start])
		end := strings.IndexByte(path[start:], '}')
		if end < 0 {
			return "", fmt.Errorf("path %q has an unterminated reserved expansion", raw)
		}
		end += start
		parameter := batchGoogleDiscoverySafePathParameter(path[start+2 : end])
		if parameter == "" {
			return "", fmt.Errorf("path %q has an invalid reserved expansion", raw)
		}
		normalized.WriteByte('{')
		normalized.WriteString(parameter)
		normalized.WriteByte('}')
		offset = end + 1
	}
	return normalized.String(), nil
}

// Reserved Google URI-template variables may contain slash-separated resource
// names. The engine intentionally exposes its safe path placeholder spelling;
// the source document remains identifiable through endpoint provenance.
func batchGoogleDiscoverySafePathParameter(raw string) string {
	runes := []rune(strings.TrimSpace(raw))
	var normalized strings.Builder
	for index, current := range runes {
		if unicode.IsUpper(current) && index > 0 && (unicode.IsLower(runes[index-1]) || unicode.IsDigit(runes[index-1])) {
			normalized.WriteByte('_')
		}
		switch {
		case unicode.IsLetter(current), unicode.IsDigit(current):
			normalized.WriteRune(unicode.ToLower(current))
		case current == '_' || current == '-':
			if normalized.Len() > 0 {
				normalized.WriteByte('_')
			}
		default:
			return ""
		}
	}
	return strings.Trim(normalized.String(), "_")
}

func sortedBatchGoogleDiscoveryMethodNames(methods map[string]batchGoogleDiscoveryMethod) []string {
	names := make([]string, 0, len(methods))
	for name := range methods {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func sortedBatchGoogleDiscoveryResourceNames(resources map[string]batchGoogleDiscoveryResource) []string {
	names := make([]string, 0, len(resources))
	for name := range resources {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
