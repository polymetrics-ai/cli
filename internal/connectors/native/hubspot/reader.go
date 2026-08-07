package hubspot

import (
	"context"
	"errors"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/engine"
)

const (
	hubspotReadPageSize = 100
	hubspotReadMaxPages = 1000
)

type objectPage struct {
	Results []struct {
		Properties map[string]any `json:"properties"`
	} `json:"results"`
	Paging struct {
		Next struct {
			After string `json:"after"`
		} `json:"next"`
	} `json:"paging"`
}

// Read reaches only the fixed collection route for a stream that was just
// discovered. It obtains the catalog first so an arbitrary tenant object type
// never becomes a generic path escape or a read for fields we did not receive
// from the provider's property description.
func (c *Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	var catalog connectors.Catalog
	if req.Config.ResolvedCatalog != nil {
		catalog = *req.Config.ResolvedCatalog
	} else {
		var err error
		catalog, err = c.Catalog(ctx, req.Config)
		if err != nil {
			return err
		}
	}
	stream, ok := discoveredStream(catalog, req.Stream)
	if !ok {
		return errUnknownStream
	}
	if !safeObjectType(stream.Name) {
		return errUnknownStream
	}
	runtime, err := engine.NewRuntime(ctx, c.bundle, req.Config, nil)
	if err != nil {
		return err
	}
	requester, err := runtime.RequesterFor(httpMethodGet, objectsPath)
	if err != nil {
		return err
	}
	properties := discoveredPropertyNames(stream)
	if len(properties) == 0 {
		return errors.New("HubSpot discovered stream has no properties")
	}

	remaining := req.Limit
	after := ""
	for page := 0; page < hubspotReadMaxPages && (remaining != 0); page++ {
		query := url.Values{}
		pageSize := hubspotReadPageSize
		if remaining > 0 && remaining < pageSize {
			pageSize = remaining
		}
		query.Set("limit", strconv.Itoa(pageSize))
		query.Set("properties", strings.Join(properties, ","))
		if after != "" {
			query.Set("after", after)
		}
		var response objectPage
		if err := requester.DoJSON(ctx, httpMethodGet, "/crm/v3/objects/"+stream.Name, query, nil, &response); err != nil {
			return err
		}
		for _, item := range response.Results {
			if err := ctx.Err(); err != nil {
				return err
			}
			record := make(connectors.Record, len(properties))
			for _, name := range properties {
				if value, exists := item.Properties[name]; exists {
					record[name] = value
				}
			}
			if err := emit(record); err != nil {
				return err
			}
			if remaining > 0 {
				remaining--
				if remaining == 0 {
					return nil
				}
			}
		}
		after = strings.TrimSpace(response.Paging.Next.After)
		if after == "" || len(response.Results) == 0 {
			return nil
		}
	}
	return nil
}

func discoveredStream(catalog connectors.Catalog, name string) (connectors.Stream, bool) {
	for _, stream := range catalog.Streams {
		if stream.Name == name {
			return stream, true
		}
	}
	return connectors.Stream{}, false
}

func discoveredPropertyNames(stream connectors.Stream) []string {
	properties := make([]string, 0, len(stream.Fields))
	for _, field := range stream.Fields {
		properties = append(properties, field.Name)
	}
	sort.Strings(properties)
	return properties
}
