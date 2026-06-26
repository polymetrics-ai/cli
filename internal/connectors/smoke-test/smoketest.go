// Package smoketest implements a deterministic in-process read-only connector.
package smoketest

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"polymetrics.ai/internal/connectors"
)

func init() { connectors.RegisterFactory("smoke-test", New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{}

func (Connector) Name() string { return "smoke-test" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: "smoke-test", DisplayName: "Smoke Test", IntegrationType: "api", Description: "Deterministic read-only in-process source for smoke-testing connector execution.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
}

func (Connector) Check(ctx context.Context, _ connectors.RuntimeConfig) error { return ctx.Err() }

func (Connector) Catalog(ctx context.Context, _ connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: "smoke-test", Streams: []connectors.Stream{
		{Name: "users", Description: "Deterministic user records.", PrimaryKey: []string{"id"}, Fields: []connectors.Field{{Name: "id", Type: "integer"}, {Name: "name", Type: "string"}, {Name: "active", Type: "boolean"}}},
		{Name: "events", Description: "Deterministic event records.", PrimaryKey: []string{"id"}, CursorFields: []string{"occurred_at"}, Fields: []connectors.Field{{Name: "id", Type: "integer"}, {Name: "user_id", Type: "integer"}, {Name: "occurred_at", Type: "string"}}},
		{Name: "large_batch_stream", Description: "Bounded deterministic batch records.", PrimaryKey: []string{"id"}, Fields: []connectors.Field{{Name: "id", Type: "integer"}, {Name: "payload", Type: "string"}}},
	}}, nil
}

func (Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	stream := req.Stream
	if stream == "" {
		stream = "users"
	}
	switch stream {
	case "users":
		return emitAll(ctx, emit, []connectors.Record{{"id": 1, "name": "Ada", "active": true}, {"id": 2, "name": "Grace", "active": true}})
	case "events":
		return emitAll(ctx, emit, []connectors.Record{{"id": 1, "user_id": 1, "occurred_at": "2026-01-01T00:00:00Z"}, {"id": 2, "user_id": 2, "occurred_at": "2026-01-02T00:00:00Z"}})
	case "large_batch_stream":
		count := configInt(req.Config, "large_batch_record_count", 1000)
		if count < 0 {
			count = 0
		}
		for i := 1; i <= count; i++ {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(connectors.Record{"id": i, "payload": fmt.Sprintf("record-%d", i)}); err != nil {
				return err
			}
		}
		return nil
	default:
		return fmt.Errorf("smoke-test stream %q not found", stream)
	}
}

func (Connector) Write(context.Context, connectors.WriteRequest, []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func emitAll(ctx context.Context, emit func(connectors.Record) error, records []connectors.Record) error {
	for _, rec := range records {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(rec); err != nil {
			return err
		}
	}
	return nil
}

func configInt(cfg connectors.RuntimeConfig, key string, def int) int {
	if cfg.Config == nil {
		return def
	}
	v, err := strconv.Atoi(strings.TrimSpace(cfg.Config[key]))
	if err != nil {
		return def
	}
	return v
}
