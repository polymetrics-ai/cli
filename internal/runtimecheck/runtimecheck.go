package runtimecheck

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.temporal.io/sdk/client"
	tlog "go.temporal.io/sdk/log"

	pmconfig "polymetrics.ai/internal/config"
	pmlogging "polymetrics.ai/internal/logging"
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

func init() {
	redis.SetLogger(redisLogAdapter{})
}

type redisLogAdapter struct{}

func (redisLogAdapter) Printf(ctx context.Context, format string, v ...interface{}) {
	pmlogging.FromContext(ctx).DebugContext(ctx, "redis diagnostic", "message", fmt.Sprintf(format, v...))
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
		result.Error = pmlogging.RedactText(ctx, err.Error())
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
		result.Error = pmlogging.RedactText(ctx, err.Error())
		return result
	}
	result.Status = "ok"
	return result
}

func checkTemporal(ctx context.Context, cfg Config) CheckResult {
	start := time.Now()
	result := CheckResult{Name: "temporal", Endpoint: cfg.TemporalAddr}
	checkCtx, cancel := context.WithTimeout(ctx, cfg.Timeout)
	defer cancel()
	c, err := client.DialContext(checkCtx, client.Options{HostPort: cfg.TemporalAddr, Logger: tlog.NewStructuredLogger(pmlogging.FromContext(checkCtx)), ConnectionOptions: client.ConnectionOptions{GetSystemInfoTimeout: cfg.Timeout}})
	if err == nil {
		defer c.Close()
		_, err = c.CheckHealth(checkCtx, &client.CheckHealthRequest{})
	}
	result.Latency = time.Since(start)
	if err != nil {
		result.Status = "error"
		result.Error = pmlogging.RedactText(ctx, err.Error())
		return result
	}
	result.Status = "ok"
	return result
}

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
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "[redacted-url]"
	}
	hadUser := parsed.User != nil
	parsed.User = nil
	parsed.RawQuery = ""
	parsed.ForceQuery = false
	parsed.Fragment = ""
	parsed.RawPath = ""
	redacted := parsed.String()
	if hadUser {
		redacted = strings.Replace(redacted, "://", "://***@", 1)
	}
	return redacted
}
