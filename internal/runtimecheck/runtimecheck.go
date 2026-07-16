package runtimecheck

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.temporal.io/sdk/client"
	tlog "go.temporal.io/sdk/log"

	pmconfig "polymetrics.ai/internal/config"
)

type Config struct {
	PostgresURL   string        `json:"postgres_url"`
	DragonflyAddr string        `json:"dragonfly_addr"`
	TemporalAddr  string        `json:"temporal_addr"`
	Timeout       time.Duration `json:"timeout"`
}

type CheckResult struct {
	Name     string        `json:"name"`
	Status   string        `json:"status"`
	Endpoint string        `json:"endpoint"`
	Latency  time.Duration `json:"latency"`
	Error    string        `json:"error,omitempty"`
}

type Report struct {
	Mode     string        `json:"mode"`
	Duration time.Duration `json:"duration"`
	Checks   []CheckResult `json:"checks"`
}

func FromConfig(cfg pmconfig.Config) Config {
	return Config{
		PostgresURL:   stringOr(cfg.Runtime.PostgresURL, "postgres://polymetrics:polymetrics@localhost:15433/polymetrics?sslmode=disable"),
		DragonflyAddr: stringOr(cfg.Runtime.DragonflyAddr, "localhost:6379"),
		TemporalAddr:  stringOr(cfg.Runtime.TemporalAddr, "localhost:7233"),
		Timeout:       3 * time.Second,
	}
}

func FromEnv() Config {
	cfg, err := pmconfig.Load(pmconfig.Options{})
	if err != nil {
		return FromConfig(pmconfig.Config{})
	}
	return FromConfig(cfg)
}

func Doctor(ctx context.Context, cfg Config) Report {
	start := time.Now()
	if cfg.Timeout <= 0 {
		cfg.Timeout = 3 * time.Second
	}
	checks := []CheckResult{
		checkPostgres(ctx, cfg),
		checkDragonfly(ctx, cfg),
		checkTemporal(ctx, cfg),
	}
	mode := "runtime"
	for _, check := range checks {
		if check.Status != "ok" {
			mode = "degraded"
			break
		}
	}
	return Report{Mode: mode, Duration: time.Since(start), Checks: checks}
}

func Healthy(report Report) bool {
	for _, check := range report.Checks {
		if check.Status != "ok" {
			return false
		}
	}
	return len(report.Checks) > 0
}

func RedactedConfig(cfg Config) Config {
	cfg.PostgresURL = redactPostgresURL(cfg.PostgresURL)
	return cfg
}

func checkPostgres(ctx context.Context, cfg Config) CheckResult {
	start := time.Now()
	result := CheckResult{Name: "postgres", Endpoint: redactPostgresURL(cfg.PostgresURL)}
	checkCtx, cancel := context.WithTimeout(ctx, cfg.Timeout)
	defer cancel()
	pool, err := pgxpool.New(checkCtx, cfg.PostgresURL)
	if err == nil {
		defer pool.Close()
		err = pool.Ping(checkCtx)
	}
	result.Latency = time.Since(start)
	if err != nil {
		result.Status = "error"
		result.Error = err.Error()
		return result
	}
	result.Status = "ok"
	return result
}

func checkDragonfly(ctx context.Context, cfg Config) CheckResult {
	start := time.Now()
	result := CheckResult{Name: "dragonfly", Endpoint: cfg.DragonflyAddr}
	checkCtx, cancel := context.WithTimeout(ctx, cfg.Timeout)
	defer cancel()
	client := redis.NewClient(&redis.Options{Addr: cfg.DragonflyAddr})
	defer client.Close()
	err := client.Ping(checkCtx).Err()
	result.Latency = time.Since(start)
	if err != nil {
		result.Status = "error"
		result.Error = err.Error()
		return result
	}
	result.Status = "ok"
	return result
}

func checkTemporal(ctx context.Context, cfg Config) CheckResult {
	start := time.Now()
	result := CheckResult{Name: "temporal", Endpoint: cfg.TemporalAddr}
	c, err := client.Dial(client.Options{HostPort: cfg.TemporalAddr, Logger: noopLogger{}})
	if err == nil {
		defer c.Close()
		checkCtx, cancel := context.WithTimeout(ctx, cfg.Timeout)
		defer cancel()
		_, err = c.CheckHealth(checkCtx, &client.CheckHealthRequest{})
	}
	result.Latency = time.Since(start)
	if err != nil {
		result.Status = "error"
		result.Error = err.Error()
		return result
	}
	result.Status = "ok"
	return result
}

type noopLogger struct{}

func (noopLogger) Debug(string, ...interface{}) {}
func (noopLogger) Info(string, ...interface{})  {}
func (noopLogger) Warn(string, ...interface{})  {}
func (noopLogger) Error(string, ...interface{}) {}

var _ tlog.Logger = noopLogger{}

func stringOr(value, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}

func redactPostgresURL(raw string) string {
	if raw == "" {
		return raw
	}
	// Avoid importing net/url just for display; this keeps malformed DSNs redacted too.
	for _, marker := range []string{"://"} {
		if i := index(raw, marker); i >= 0 {
			prefix := raw[:i+len(marker)]
			rest := raw[i+len(marker):]
			if at := index(rest, "@"); at >= 0 {
				return prefix + "***@" + rest[at+1:]
			}
		}
	}
	return fmt.Sprintf("%s", raw)
}

func index(s, substr string) int {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
