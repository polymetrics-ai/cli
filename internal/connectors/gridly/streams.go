package gridly

import "polymetrics.ai/internal/connectors"

type streamEndpoint struct {
	resource    string
	recordsPath string
	perView     bool
	mapRecord   func(map[string]any) connectors.Record
}

var streamEndpoints = map[string]streamEndpoint{
	"views":    {resource: "views", recordsPath: "", mapRecord: viewRecord},
	"records":  {resource: "views/{view}/records", recordsPath: "", perView: true, mapRecord: gridRecord},
	"branches": {resource: "views/{view}/branches", recordsPath: "", perView: true, mapRecord: branchRecord},
}

func streams() []connectors.Stream {
	return []connectors.Stream{
		{Name: "views", Description: "Gridly views.", PrimaryKey: []string{"id"}, Fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "name", Type: "string"}}},
		{Name: "records", Description: "Gridly records for configured views.", PrimaryKey: []string{"view_id", "id"}, Fields: []connectors.Field{{Name: "view_id", Type: "string"}, {Name: "id", Type: "string"}, {Name: "path", Type: "string"}, {Name: "cells", Type: "array"}}},
		{Name: "branches", Description: "Gridly branches for configured views.", PrimaryKey: []string{"view_id", "id"}, Fields: []connectors.Field{{Name: "view_id", Type: "string"}, {Name: "id", Type: "string"}, {Name: "name", Type: "string"}}},
	}
}

func viewRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["id"], "name": item["name"]}
}

func branchRecord(item map[string]any) connectors.Record { return viewRecord(item) }

func gridRecord(item map[string]any) connectors.Record {
	rec := connectors.Record{"id": item["id"], "path": item["path"], "cells": item["cells"]}
	for _, raw := range asSlice(item["cells"]) {
		cell, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		column, _ := cell["columnId"].(string)
		if column == "" {
			continue
		}
		rec[cellKey(column)] = cell["value"]
	}
	return rec
}

func asSlice(v any) []any {
	if s, ok := v.([]any); ok {
		return s
	}
	return nil
}
