package faker

import (
	"context"

	"polymetrics.ai/internal/connectors"
)

// Catalog returns the fixed 3-stream catalog, ported field-for-field from
// legacy's Catalog (faker.go:36-45). This is not schema discovery against
// any external source — it is a hand-written, static catalog — but is
// exposed via the same connectors.Catalog shape every connector uses.
func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: []connectors.Stream{
		{
			Name:         "users",
			Description:  "Deterministic fake users.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields: []connectors.Field{
				{Name: "id", Type: "string"},
				{Name: "name", Type: "string"},
				{Name: "email", Type: "string"},
				{Name: "updated_at", Type: "timestamp"},
			},
		},
		{
			Name:         "purchases",
			Description:  "Deterministic fake purchases tied to generated users and products.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields: []connectors.Field{
				{Name: "id", Type: "string"},
				{Name: "user_id", Type: "string"},
				{Name: "product_id", Type: "string"},
				{Name: "amount", Type: "number"},
				{Name: "updated_at", Type: "timestamp"},
			},
		},
		{
			Name:        "products",
			Description: "Deterministic fake products.",
			PrimaryKey:  []string{"id"},
			Fields: []connectors.Field{
				{Name: "id", Type: "string"},
				{Name: "sku", Type: "string"},
				{Name: "name", Type: "string"},
				{Name: "price", Type: "number"},
			},
		},
	}}, nil
}
