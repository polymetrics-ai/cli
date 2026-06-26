package gutendex

import (
	"fmt"
	"net/url"
	"strings"

	"polymetrics.ai/internal/connectors"
)

// streamEndpoint maps a stream name to the Gutendex query it reads. The Gutendex
// API exposes a single resource (/books); each stream is a fixed, useful view of
// that resource (a different sort/filter). resource is always "books"; baseQuery
// holds the query params that define the view; mapRecord flattens a raw book
// object into a connectors.Record.
type streamEndpoint struct {
	resource  string
	baseQuery url.Values
	mapRecord func(map[string]any) connectors.Record
}

// gutendexStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in gutendexStreams; the read
// path is fully data-driven from this table.
var gutendexStreamEndpoints = map[string]streamEndpoint{
	// books: the full catalog in the API default order (popular first).
	"books": {resource: "books", baseQuery: url.Values{}, mapRecord: gutendexBookRecord},
	// popular_books: explicitly most-downloaded first.
	"popular_books": {resource: "books", baseQuery: url.Values{"sort": []string{"popular"}}, mapRecord: gutendexBookRecord},
	// latest_books: highest Project Gutenberg IDs first (newest additions).
	"latest_books": {resource: "books", baseQuery: url.Values{"sort": []string{"descending"}}, mapRecord: gutendexBookRecord},
	// english_books: books available in English.
	"english_books": {resource: "books", baseQuery: url.Values{"languages": []string{"en"}}, mapRecord: gutendexBookRecord},
}

// gutendexStreams returns the connector's published stream catalog. Every book
// has a stable integer id, so the primary key is ["id"]. The API has no
// updated-at field, so there is no incremental cursor (full-refresh only), which
// matches the catalog's supported_sync_modes.
func gutendexStreams() []connectors.Stream {
	fields := gutendexBookFields()
	return []connectors.Stream{
		{
			Name:        "books",
			Description: "All Project Gutenberg books in the API default (popular) order.",
			PrimaryKey:  []string{"id"},
			Fields:      fields,
		},
		{
			Name:        "popular_books",
			Description: "Project Gutenberg books sorted by download count, most popular first.",
			PrimaryKey:  []string{"id"},
			Fields:      fields,
		},
		{
			Name:        "latest_books",
			Description: "Project Gutenberg books sorted by descending Gutenberg ID (newest additions first).",
			PrimaryKey:  []string{"id"},
			Fields:      fields,
		},
		{
			Name:        "english_books",
			Description: "Project Gutenberg books available in English.",
			PrimaryKey:  []string{"id"},
			Fields:      fields,
		},
	}
}

func gutendexBookFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "title", Type: "string"},
		{Name: "author_name", Type: "string"},
		{Name: "author_birth_year", Type: "integer"},
		{Name: "author_death_year", Type: "integer"},
		{Name: "authors", Type: "string"},
		{Name: "translators", Type: "string"},
		{Name: "subjects", Type: "string"},
		{Name: "bookshelves", Type: "string"},
		{Name: "languages", Type: "string"},
		{Name: "copyright", Type: "boolean"},
		{Name: "media_type", Type: "string"},
		{Name: "download_count", Type: "integer"},
	}
}

// gutendexBookRecord flattens a raw Gutendex book object into a connectors.Record.
// List-valued fields (authors, subjects, etc.) are joined into a comma-separated
// string so the record is a flat scalar map; the first author is also promoted to
// dedicated author_* columns for convenience.
func gutendexBookRecord(item map[string]any) connectors.Record {
	rec := connectors.Record{
		"id":             item["id"],
		"title":          item["title"],
		"copyright":      item["copyright"],
		"media_type":     item["media_type"],
		"download_count": item["download_count"],
		"languages":      joinStrings(item["languages"]),
		"subjects":       joinStrings(item["subjects"]),
		"bookshelves":    joinStrings(item["bookshelves"]),
		"translators":    joinAuthorNames(item["translators"]),
		"authors":        joinAuthorNames(item["authors"]),
	}
	if name, birth, death, ok := firstAuthor(item["authors"]); ok {
		rec["author_name"] = name
		rec["author_birth_year"] = birth
		rec["author_death_year"] = death
	} else {
		rec["author_name"] = nil
		rec["author_birth_year"] = nil
		rec["author_death_year"] = nil
	}
	return rec
}

// firstAuthor returns the name, birth_year and death_year of the first author in
// a raw authors list (which is []any of map[string]any). ok is false when the
// list is empty or malformed.
func firstAuthor(v any) (name string, birth any, death any, ok bool) {
	list, isList := v.([]any)
	if !isList || len(list) == 0 {
		return "", nil, nil, false
	}
	author, isMap := list[0].(map[string]any)
	if !isMap {
		return "", nil, nil, false
	}
	name, _ = author["name"].(string)
	return name, author["birth_year"], author["death_year"], true
}

// joinAuthorNames collapses an authors/translators list into a comma-separated
// list of names.
func joinAuthorNames(v any) string {
	list, ok := v.([]any)
	if !ok {
		return ""
	}
	names := make([]string, 0, len(list))
	for _, item := range list {
		if author, ok := item.(map[string]any); ok {
			if name, ok := author["name"].(string); ok {
				names = append(names, name)
			}
		}
	}
	return strings.Join(names, ", ")
}

// joinStrings collapses a JSON string array into a comma-separated string.
func joinStrings(v any) string {
	list, ok := v.([]any)
	if !ok {
		return ""
	}
	parts := make([]string, 0, len(list))
	for _, item := range list {
		switch s := item.(type) {
		case string:
			parts = append(parts, s)
		default:
			parts = append(parts, fmt.Sprintf("%v", s))
		}
	}
	return strings.Join(parts, ", ")
}
