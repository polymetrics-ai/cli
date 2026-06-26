// Package hardcodedrecords implements a deterministic, in-process source of
// hardcoded records. It is marked as an API integration to match the catalog.
package hardcodedrecords

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"polymetrics.ai/internal/connectors"
)

const connectorName = "hardcoded-records"

func init() { connectors.RegisterFactory(connectorName, New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{}

func (Connector) Name() string { return connectorName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            connectorName,
		DisplayName:     "Hardcoded Records",
		IntegrationType: "api",
		Description:     "Emits deterministic hardcoded records without network access.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

func (Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error { return ctx.Err() }

func (Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: connectorName, Streams: []connectors.Stream{{
		Name: "records", Description: "Deterministic hardcoded records.", PrimaryKey: []string{"id"}, CursorFields: []string{"updated_at"},
		Fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "name", Type: "string"}, {Name: "status", Type: "string"}, {Name: "value", Type: "integer"}, {Name: "updated_at", Type: "timestamp"}},
	}}}, nil
}

func (Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	stream := req.Stream
	if stream == "" {
		stream = "records"
	}
	if stream != "records" {
		return fmt.Errorf("hardcoded-records stream %q not found", stream)
	}
	count, err := count(req.Config.Config["count"], 1000)
	if err != nil {
		return err
	}
	for i := 1; i <= count; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		status := "active"
		if i%2 == 0 {
			status = "inactive"
		}
		if err := emit(connectors.Record{"id": fmt.Sprintf("record_%03d", i), "name": fmt.Sprintf("Record %03d", i), "status": status, "value": i, "updated_at": fmt.Sprintf("2026-01-%02dT00:00:00Z", ((i-1)%28)+1)}); err != nil {
			return err
		}
	}
	return nil
}

func (Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func count(raw string, def int) (int, error) {
	if strings.TrimSpace(raw) == "" {
		return def, nil
	}
	n, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || n < 1 {
		return 0, fmt.Errorf("count must be a positive integer")
	}
	return n, nil
}
