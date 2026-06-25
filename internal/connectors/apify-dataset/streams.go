package apifydataset

import (
	"fmt"

	"polymetrics.ai/internal/connectors"
)

// streamEndpoint maps a stream name to the Apify API resource it reads from. The
// path is computed per request because the item/dataset streams interpolate the
// configured dataset_id. recordsPath selects where the records live in the JSON
// body (a top-level array for items, "data.items" for management lists, "data"
// for a single dataset object). paginated marks whether offset/limit paging
// applies.
type streamEndpoint struct {
	// name is the stream name (used in error messages).
	name string
	// path returns the request path (relative to base_url) for this stream,
	// resolving the configured dataset_id where needed.
	path func(cfg connectors.RuntimeConfig) (string, error)
	// recordsPath is the dotted JSON path to the records array/object.
	recordsPath string
	// paginated is true when the stream supports offset/limit pagination.
	paginated bool
	// mapRecord flattens a raw Apify object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// apifyStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in apifyStreams.
var apifyStreamEndpoints = map[string]streamEndpoint{
	"item_collection": {
		name: "item_collection",
		path: func(cfg connectors.RuntimeConfig) (string, error) {
			id, err := apifyDatasetID(cfg)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("datasets/%s/items", id), nil
		},
		recordsPath: "", // GET /datasets/{id}/items returns a top-level array.
		paginated:   true,
		mapRecord:   apifyItemRecord,
	},
	"dataset_collection": {
		name:        "dataset_collection",
		path:        func(connectors.RuntimeConfig) (string, error) { return "datasets", nil },
		recordsPath: "data.items", // GET /datasets returns {data:{items:[...]}}.
		paginated:   true,
		mapRecord:   apifyDatasetRecord,
	},
	"dataset": {
		name: "dataset",
		path: func(cfg connectors.RuntimeConfig) (string, error) {
			id, err := apifyDatasetID(cfg)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("datasets/%s", id), nil
		},
		recordsPath: "data", // GET /datasets/{id} returns {data:{...}}.
		paginated:   false,
		mapRecord:   apifyDatasetRecord,
	},
}

// apifyStreams returns the connector's published stream catalog.
func apifyStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "item_collection",
			Description:  "Items stored in the configured Apify dataset (GET /datasets/{datasetId}/items). Each raw item is wrapped under the data field because dataset items have a dynamic, actor-defined schema.",
			PrimaryKey:   nil,
			CursorFields: nil,
			Fields:       apifyItemFields(),
		},
		{
			Name:         "dataset_collection",
			Description:  "All datasets in the Apify account (GET /datasets).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"createdAt"},
			Fields:       apifyDatasetFields(),
		},
		{
			Name:         "dataset",
			Description:  "Metadata for the configured Apify dataset (GET /datasets/{datasetId}).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"modifiedAt"},
			Fields:       apifyDatasetFields(),
		},
	}
}

func apifyItemFields() []connectors.Field {
	// Dataset items have an actor-defined dynamic schema, so the published field
	// is the data envelope wrapping the raw item.
	return []connectors.Field{
		{Name: "data", Type: "object"},
	}
}

func apifyDatasetFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "userId", Type: "string"},
		{Name: "createdAt", Type: "timestamp"},
		{Name: "modifiedAt", Type: "timestamp"},
		{Name: "accessedAt", Type: "timestamp"},
		{Name: "itemCount", Type: "integer"},
		{Name: "cleanItemCount", Type: "integer"},
		{Name: "actId", Type: "string"},
		{Name: "actRunId", Type: "string"},
	}
}

// apifyItemRecord wraps a raw dataset item under the data field. Dataset items
// have an actor-defined dynamic schema with no guaranteed columns, so the whole
// object is preserved verbatim under a stable envelope key.
func apifyItemRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"data": map[string]any(item),
	}
}

func apifyDatasetRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":             item["id"],
		"name":           item["name"],
		"userId":         item["userId"],
		"createdAt":      item["createdAt"],
		"modifiedAt":     item["modifiedAt"],
		"accessedAt":     item["accessedAt"],
		"itemCount":      item["itemCount"],
		"cleanItemCount": item["cleanItemCount"],
		"actId":          item["actId"],
		"actRunId":       item["actRunId"],
	}
}
