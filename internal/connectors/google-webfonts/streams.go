package googlewebfonts

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the sort mode used when querying the
// Google Web Fonts list endpoint, plus the record mapper that flattens each
// font item. The Google Web Fonts Developer API has a single list resource
// (webfonts); the published streams are different sorted views of that list,
// distinguished by the `sort` query parameter (alpha, date, popularity,
// style, trending). The "webfonts" stream applies no sort (API default).
type streamEndpoint struct {
	// sort is the value passed as the `sort` query param. Empty means the API
	// default ordering (no sort param sent).
	sort string
	// mapRecord flattens a raw font family object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// streamEndpoints is the per-stream routing table. Adding a stream means adding
// one entry here plus a Stream definition in streams(); the read path is fully
// data-driven from this table.
var streamEndpoints = map[string]streamEndpoint{
	"webfonts":       {sort: "", mapRecord: fontRecord},
	"popular_fonts":  {sort: "popularity", mapRecord: fontRecord},
	"trending_fonts": {sort: "trending", mapRecord: fontRecord},
	"newest_fonts":   {sort: "date", mapRecord: fontRecord},
	"alpha_fonts":    {sort: "alpha", mapRecord: fontRecord},
}

// streams returns the connector's published stream catalog. Every font item is
// keyed by its unique `family` name (the API has no numeric id), and
// `lastModified` is the natural incremental cursor.
func streams() []connectors.Stream {
	fields := fontFields()
	return []connectors.Stream{
		{
			Name:         "webfonts",
			Description:  "All Google Web Fonts families in the API's default order.",
			PrimaryKey:   []string{"family"},
			CursorFields: []string{"lastModified"},
			Fields:       fields,
		},
		{
			Name:         "popular_fonts",
			Description:  "Google Web Fonts families sorted by popularity.",
			PrimaryKey:   []string{"family"},
			CursorFields: []string{"lastModified"},
			Fields:       fields,
		},
		{
			Name:         "trending_fonts",
			Description:  "Google Web Fonts families sorted by trending usage.",
			PrimaryKey:   []string{"family"},
			CursorFields: []string{"lastModified"},
			Fields:       fields,
		},
		{
			Name:         "newest_fonts",
			Description:  "Google Web Fonts families sorted by date added.",
			PrimaryKey:   []string{"family"},
			CursorFields: []string{"lastModified"},
			Fields:       fields,
		},
		{
			Name:         "alpha_fonts",
			Description:  "Google Web Fonts families sorted alphabetically.",
			PrimaryKey:   []string{"family"},
			CursorFields: []string{"lastModified"},
			Fields:       fields,
		},
	}
}

func fontFields() []connectors.Field {
	return []connectors.Field{
		{Name: "family", Type: "string"},
		{Name: "category", Type: "string"},
		{Name: "version", Type: "string"},
		{Name: "lastModified", Type: "string"},
		{Name: "kind", Type: "string"},
		{Name: "menu", Type: "string"},
		{Name: "variants", Type: "array"},
		{Name: "subsets", Type: "array"},
		{Name: "files", Type: "object"},
		{Name: "axes", Type: "array"},
		{Name: "variant_count", Type: "integer"},
		{Name: "subset_count", Type: "integer"},
	}
}

// fontRecord flattens a raw Google Web Fonts item into a connectors.Record. The
// nested arrays (variants/subsets/files/axes) are passed through unchanged, and
// two convenience counts are derived so downstream consumers can filter without
// re-parsing the arrays.
func fontRecord(item map[string]any) connectors.Record {
	rec := connectors.Record{
		"family":       item["family"],
		"category":     item["category"],
		"version":      item["version"],
		"lastModified": item["lastModified"],
		"kind":         item["kind"],
		"menu":         item["menu"],
		"variants":     item["variants"],
		"subsets":      item["subsets"],
		"files":        item["files"],
		"axes":         item["axes"],
	}
	rec["variant_count"] = lenOf(item["variants"])
	rec["subset_count"] = lenOf(item["subsets"])
	return rec
}

// lenOf returns the length of a JSON array value (decoded as []any), or 0 for
// anything that is not an array.
func lenOf(v any) int {
	if arr, ok := v.([]any); ok {
		return len(arr)
	}
	return 0
}
