package dockerhub

import "polymetrics.ai/internal/connectors"

// streamKind classifies how a stream is read off the Docker Hub registry API.
type streamKind int

const (
	// kindPaginated reads a {count,next,previous,results} envelope, following the
	// absolute `next` URL until it is null.
	kindPaginated streamKind = iota
	// kindSingle reads one object (the namespace/user profile) into a single
	// record.
	kindSingle
)

// streamDef describes one published stream: its kind, the relative path template
// (filled with username/repository at read time), and the mapper that flattens
// raw API objects into connectors.Record values.
type streamDef struct {
	kind      streamKind
	path      func(username, repository string) string
	mapRecord func(map[string]any) connectors.Record
	// requiresRepository is true for streams scoped to a single repository (tags),
	// which need the `repository` config field.
	requiresRepository bool
}

// dockerhubStreamDefs is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in dockerhubStreams; the read
// path is data-driven from this table.
var dockerhubStreamDefs = map[string]streamDef{
	"repositories": {
		kind:      kindPaginated,
		path:      func(username, _ string) string { return "repositories/" + username + "/" },
		mapRecord: repositoryRecord,
	},
	"tags": {
		kind: kindPaginated,
		path: func(username, repository string) string {
			return "repositories/" + username + "/" + repository + "/tags"
		},
		mapRecord:          tagRecord,
		requiresRepository: true,
	},
	"namespace": {
		kind:      kindSingle,
		path:      func(username, _ string) string { return "users/" + username + "/" },
		mapRecord: namespaceRecord,
	},
}

// dockerhubStreams returns the connector's published stream catalog.
func dockerhubStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "repositories",
			Description:  "Public Docker Hub repositories owned by the configured username/organization.",
			PrimaryKey:   []string{"name"},
			CursorFields: []string{"last_updated"},
			Fields:       repositoryFields(),
		},
		{
			Name:         "tags",
			Description:  "Image tags for the repository named by the `repository` config field.",
			PrimaryKey:   []string{"name"},
			CursorFields: []string{"last_updated"},
			Fields:       tagFields(),
		},
		{
			Name:        "namespace",
			Description: "Profile for the configured Docker Hub user or organization.",
			PrimaryKey:  []string{"id"},
			Fields:      namespaceFields(),
		},
	}
}

func repositoryFields() []connectors.Field {
	return []connectors.Field{
		{Name: "name", Type: "string"},
		{Name: "namespace", Type: "string"},
		{Name: "repository_type", Type: "string"},
		{Name: "status", Type: "integer"},
		{Name: "status_description", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "is_private", Type: "boolean"},
		{Name: "star_count", Type: "integer"},
		{Name: "pull_count", Type: "integer"},
		{Name: "storage_size", Type: "integer"},
		{Name: "last_updated", Type: "timestamp"},
		{Name: "last_modified", Type: "timestamp"},
		{Name: "date_registered", Type: "timestamp"},
	}
}

func tagFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "repository", Type: "integer"},
		{Name: "full_size", Type: "integer"},
		{Name: "digest", Type: "string"},
		{Name: "media_type", Type: "string"},
		{Name: "content_type", Type: "string"},
		{Name: "tag_status", Type: "string"},
		{Name: "last_updated", Type: "timestamp"},
		{Name: "last_pushed", Type: "timestamp"},
		{Name: "last_updater_username", Type: "string"},
	}
}

func namespaceFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "uuid", Type: "string"},
		{Name: "orgname", Type: "string"},
		{Name: "full_name", Type: "string"},
		{Name: "company", Type: "string"},
		{Name: "location", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "badge", Type: "string"},
		{Name: "is_active", Type: "boolean"},
		{Name: "date_joined", Type: "timestamp"},
	}
}

func repositoryRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"name":               item["name"],
		"namespace":          item["namespace"],
		"repository_type":    item["repository_type"],
		"status":             item["status"],
		"status_description": item["status_description"],
		"description":        item["description"],
		"is_private":         item["is_private"],
		"star_count":         item["star_count"],
		"pull_count":         item["pull_count"],
		"storage_size":       item["storage_size"],
		"last_updated":       item["last_updated"],
		"last_modified":      item["last_modified"],
		"date_registered":    item["date_registered"],
	}
}

func tagRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                    item["id"],
		"name":                  item["name"],
		"repository":            item["repository"],
		"full_size":             item["full_size"],
		"digest":                item["digest"],
		"media_type":            item["media_type"],
		"content_type":          item["content_type"],
		"tag_status":            item["tag_status"],
		"last_updated":          item["last_updated"],
		"last_pushed":           item["last_pushed"],
		"last_updater_username": item["last_updater_username"],
	}
}

func namespaceRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"uuid":        item["uuid"],
		"orgname":     item["orgname"],
		"full_name":   item["full_name"],
		"company":     item["company"],
		"location":    item["location"],
		"type":        item["type"],
		"badge":       item["badge"],
		"is_active":   item["is_active"],
		"date_joined": item["date_joined"],
	}
}
