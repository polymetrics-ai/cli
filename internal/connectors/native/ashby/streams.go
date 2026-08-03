package ashby

import (
	"sort"

	"polymetrics.ai/internal/connectors"
)

// streamEndpoint maps a stream name to a fixed Ashby API path plus the
// documented request-body fields and catalog projection generated from the
// public Ashby OpenAPI. Paths remain fixed constants; caller input can only
// populate declared request fields.
type streamEndpoint struct {
	path               string
	requestFields      map[string]string
	fixedRequestFields map[string]string
	// fixedRequestFieldGaps names the foundation gap that blocks each pinned
	// field's non-default values, keyed by the same field name as
	// fixedRequestFields.
	fixedRequestFieldGaps map[string]string
	requiredFields        []string
	requiredAnyFields     []string
	cursorField           string
	syntheticFields       []string
	primaryKey            []string
	fields                []connectors.Field
}

// ashbyStreams returns the connector's published stream catalog. The field
// lists, primary keys, and cursor fields are generated from Ashby's documented
// response schemas and kept alongside the native POST cursor reader.
func ashbyStreams() []connectors.Stream {
	out := make([]connectors.Stream, 0, len(ashbyStreamEndpoints))
	for _, stream := range ashbyStreamOrder() {
		endpoint := ashbyStreamEndpoints[stream]
		out = append(out, connectors.Stream{
			Name:         stream,
			Description:  "Ashby " + stream + ".",
			PrimaryKey:   append([]string(nil), endpoint.primaryKey...),
			CursorFields: cursorFields(endpoint),
			Fields:       append([]connectors.Field(nil), endpoint.fields...),
		})
	}
	return out
}

func ashbyStreamOrder() []string {
	ordered := []string{
		"candidates",
		"jobs",
		"applications",
		"users",
	}
	seen := make(map[string]bool, len(ashbyStreamEndpoints))
	for _, name := range ordered {
		if _, ok := ashbyStreamEndpoints[name]; ok {
			seen[name] = true
		}
	}
	var rest []string
	for name := range ashbyStreamEndpoints {
		if !seen[name] {
			rest = append(rest, name)
		}
	}
	sort.Strings(rest)
	ordered = append(ordered, rest...)
	return ordered
}

func cursorFields(endpoint streamEndpoint) []string {
	if endpoint.cursorField == "" {
		return nil
	}
	return []string{endpoint.cursorField}
}
