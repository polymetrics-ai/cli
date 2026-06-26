package jinaaireader

import "polymetrics.ai/internal/connectors"

func streams() []connectors.Stream {
	return []connectors.Stream{{Name: "pages", Description: "Pages fetched and converted by Jina AI Reader.", PrimaryKey: []string{"url"}, Fields: []connectors.Field{{Name: "url", Type: "string"}, {Name: "title", Type: "string"}, {Name: "description", Type: "string"}, {Name: "content", Type: "string"}}}}
}
