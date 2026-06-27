package schedule

import (
	"context"
	"errors"
)

// TemporalBackend installs schedules as Temporal cron workflow schedules.
// Requires POLYMETRICS_TEMPORAL_ADDR to be set and reachable.
type TemporalBackend struct {
	Addr string
}

func (b TemporalBackend) Kind() BackendKind { return KindTemporal }

func (b TemporalBackend) Install(ctx context.Context, m Manifest, pmBin string) error {
	return errors.New("TemporalBackend.Install: not implemented")
}

func (b TemporalBackend) Remove(ctx context.Context, name string) error {
	return errors.New("TemporalBackend.Remove: not implemented")
}
