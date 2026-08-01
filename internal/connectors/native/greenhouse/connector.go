package greenhouse

import (
	"context"
	"fmt"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
	greenhousehooks "polymetrics.ai/internal/connectors/hooks/greenhouse"
)

type Connector struct {
	*engine.Connector
}

func New() connectors.Connector {
	bundle, err := engine.Load(defs.FS, "greenhouse")
	if err != nil {
		panic("native/greenhouse: failed to load defs/greenhouse bundle: " + err.Error())
	}
	return Connector{Connector: engine.New(bundle, greenhousehooks.New())}
}

func (c Connector) ValidateWrite(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) error {
	if err := c.Connector.ValidateWrite(ctx, req, records); err != nil {
		return err
	}
	for i, rec := range records {
		if err := greenhousehooks.ValidateWriteRecord(req.Action, rec); err != nil {
			return fmt.Errorf("record %d: %w", i, err)
		}
	}
	return nil
}

func (c Connector) DryRunWrite(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WritePreview, error) {
	if err := c.ValidateWrite(ctx, req, records); err != nil {
		return connectors.WritePreview{}, err
	}
	return c.Connector.DryRunWrite(ctx, req, records)
}
