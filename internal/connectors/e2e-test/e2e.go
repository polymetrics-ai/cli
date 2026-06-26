package e2etest

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"polymetrics.ai/internal/connectors"
)

const defaultMaxRecords = 2

func init() { connectors.RegisterFactory("e2e-test", New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{}

func (Connector) Name() string { return "e2e-test" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: "e2e-test", DisplayName: "E2E Testing", IntegrationType: "api", Description: "Deterministic in-process read-only source for connector and pipeline tests.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
}

func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	return ctx.Err()
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: []connectors.Stream{{Name: "data", Description: "Deterministic test data.", Fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "column1", Type: "string"}, {Name: "sequence", Type: "integer"}}, PrimaryKey: []string{"id"}}}}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	stream := req.Stream
	if stream == "" {
		stream = "data"
	}
	if stream != "data" {
		return fmt.Errorf("e2e-test stream %q not found", stream)
	}
	max, err := maxRecords(req.Config)
	if err != nil {
		return err
	}
	seed, err := intConfig(req.Config, "seed", 0)
	if err != nil {
		return err
	}
	for i := 0; i < max; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(connectors.Record{"id": fmt.Sprintf("%d-%d", seed, i), "column1": fmt.Sprintf("value-%d-%d", seed, i), "sequence": i}); err != nil {
			return err
		}
		throwAfter, _ := intConfig(req.Config, "throw_after_n_records", 0)
		if throwAfter > 0 && i+1 >= throwAfter {
			return errors.New("e2e-test configured exception after records")
		}
	}
	return nil
}

func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func maxRecords(cfg connectors.RuntimeConfig) (int, error) {
	for _, key := range []string{"max_records", "max_messages"} {
		if strings.TrimSpace(cfg.Config[key]) != "" {
			return intConfig(cfg, key, defaultMaxRecords)
		}
	}
	return defaultMaxRecords, nil
}

func intConfig(cfg connectors.RuntimeConfig, key string, fallback int) (int, error) {
	raw := strings.TrimSpace(cfg.Config[key])
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 {
		return 0, fmt.Errorf("e2e-test config %s must be a non-negative integer", key)
	}
	return value, nil
}
