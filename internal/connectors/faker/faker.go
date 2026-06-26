// Package faker implements a deterministic, in-process sample-data connector.
// It mirrors the catalog's API integration type without making network calls.
package faker

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"polymetrics.ai/internal/connectors"
)

const connectorName = "faker"

func init() { connectors.RegisterFactory(connectorName, New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{}

func (Connector) Name() string { return connectorName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            connectorName,
		DisplayName:     "Sample Data",
		IntegrationType: "api",
		Description:     "Generates deterministic sample users, purchases, and products without network access.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

func (Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error { return ctx.Err() }

func (Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: connectorName, Streams: []connectors.Stream{
		{Name: "users", Description: "Deterministic fake users.", PrimaryKey: []string{"id"}, CursorFields: []string{"updated_at"}, Fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "name", Type: "string"}, {Name: "email", Type: "string"}, {Name: "updated_at", Type: "timestamp"}}},
		{Name: "purchases", Description: "Deterministic fake purchases tied to generated users and products.", PrimaryKey: []string{"id"}, CursorFields: []string{"updated_at"}, Fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "user_id", Type: "string"}, {Name: "product_id", Type: "string"}, {Name: "amount", Type: "number"}, {Name: "updated_at", Type: "timestamp"}}},
		{Name: "products", Description: "Deterministic fake products.", PrimaryKey: []string{"id"}, Fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "sku", Type: "string"}, {Name: "name", Type: "string"}, {Name: "price", Type: "number"}}},
	}}, nil
}

func (Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	stream := req.Stream
	if stream == "" {
		stream = "users"
	}
	count, err := positiveInt(req.Config.Config["count"], 1000)
	if err != nil {
		return err
	}
	seed, err := strconv.Atoi(strings.TrimSpace(req.Config.Config["seed"]))
	if err != nil && strings.TrimSpace(req.Config.Config["seed"]) != "" {
		return fmt.Errorf("faker config seed must be an integer: %w", err)
	}
	if seed < 0 {
		seed = 0
	}
	switch stream {
	case "users":
		for i := 1; i <= count; i++ {
			if err := ctx.Err(); err != nil {
				return err
			}
			n := seed + i
			if err := emit(connectors.Record{"id": fmt.Sprintf("user_%03d", n), "name": fmt.Sprintf("User %03d", n), "email": fmt.Sprintf("user%03d@example.com", n), "updated_at": timestamp(i)}); err != nil {
				return err
			}
		}
	case "purchases":
		for i := 1; i <= count; i++ {
			if err := ctx.Err(); err != nil {
				return err
			}
			n := seed + i
			if err := emit(connectors.Record{"id": fmt.Sprintf("purchase_%03d", n), "user_id": fmt.Sprintf("user_%03d", n), "product_id": fmt.Sprintf("product_%03d", (i%10)+1), "amount": float64((i%10)+1) * 9.99, "updated_at": timestamp(i)}); err != nil {
				return err
			}
		}
	case "products":
		for i := 1; i <= 10; i++ {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(connectors.Record{"id": fmt.Sprintf("product_%03d", i), "sku": fmt.Sprintf("SKU-%03d", i), "name": fmt.Sprintf("Product %03d", i), "price": float64(i) * 4.25}); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("faker stream %q not found", stream)
	}
	return nil
}

func (Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func positiveInt(raw string, def int) (int, error) {
	if strings.TrimSpace(raw) == "" {
		return def, nil
	}
	n, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || n < 1 {
		return 0, fmt.Errorf("count must be a positive integer")
	}
	return n, nil
}

func timestamp(i int) string { return fmt.Sprintf("2026-01-%02dT00:00:00Z", ((i-1)%28)+1) }
