package mysql

import (
	"context"

	"github.com/go-mysql-org/go-mysql/client"

	"polymetrics.ai/internal/connectors"
)

// DialForTest opens one connection under the configuration's own transport
// security mode so an integration test can ask the live server what it
// actually negotiated. Declaring a mode is not evidence that it took effect;
// this exists so the test can check the wire rather than the spec file.
//
// It is available to tests only, and returns no configuration values in its
// error.
func DialForTest(ctx context.Context, cfg connectors.RuntimeConfig) (*client.Conn, error) {
	conn, err := resolveConfig(cfg)
	if err != nil {
		return nil, err
	}
	return conn.open(ctx)
}
