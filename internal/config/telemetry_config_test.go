package config

import (
	"testing"
)

func TestLoadTelemetryDefaultsAndEnv(t *testing.T) {
	clearBoundEnv(t)
	root := t.TempDir()

	cfg, err := Load(Options{Root: root})
	if err != nil {
		t.Fatalf("Load defaults: %v", err)
	}
	if cfg.Telemetry.Exporter != "none" {
		t.Fatalf("Telemetry.Exporter default = %q, want none", cfg.Telemetry.Exporter)
	}
	if cfg.Telemetry.Endpoint != "" {
		t.Fatalf("Telemetry.Endpoint default = %q, want empty", cfg.Telemetry.Endpoint)
	}
	if cfg.Telemetry.Directory != ".polymetrics/telemetry" {
		t.Fatalf("Telemetry.Directory default = %q, want .polymetrics/telemetry", cfg.Telemetry.Directory)
	}
	if cfg.Telemetry.Capture != "default" {
		t.Fatalf("Telemetry.Capture default = %q, want default", cfg.Telemetry.Capture)
	}

	t.Setenv("PM_TELEMETRY", "file")
	t.Setenv("PM_TELEMETRY_DIR", "custom-telemetry")
	t.Setenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT", "https://collector.example.test/v1/traces")
	t.Setenv("PM_TELEMETRY_CAPTURE", "minimal")
	cfg, err = Load(Options{Root: root})
	if err != nil {
		t.Fatalf("Load env: %v", err)
	}
	if cfg.Telemetry.Exporter != "file" {
		t.Fatalf("Telemetry.Exporter env = %q, want file", cfg.Telemetry.Exporter)
	}
	if cfg.Telemetry.Directory != "custom-telemetry" {
		t.Fatalf("Telemetry.Directory env = %q, want custom-telemetry", cfg.Telemetry.Directory)
	}
	if cfg.Telemetry.Endpoint != "https://collector.example.test/v1/traces" {
		t.Fatalf("Telemetry.Endpoint env = %q, want OTEL traces endpoint", cfg.Telemetry.Endpoint)
	}
	if cfg.Telemetry.Capture != "minimal" {
		t.Fatalf("Telemetry.Capture env = %q, want minimal", cfg.Telemetry.Capture)
	}
}

func TestLoadTelemetryFileBeatsEnv(t *testing.T) {
	clearBoundEnv(t)
	root := writeConfig(t, `telemetry:
  exporter: file
  endpoint: https://file-collector.example.test/v1/traces
  directory: file-telemetry
  capture: minimal
`)
	t.Setenv("POLYMETRICS_TELEMETRY", "otlp")
	t.Setenv("POLYMETRICS_TELEMETRY_ENDPOINT", "https://env-collector.example.test/v1/traces")

	cfg, err := Load(Options{Root: root})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Telemetry.Exporter != "otlp" {
		t.Fatalf("Telemetry.Exporter = %q, want env otlp", cfg.Telemetry.Exporter)
	}
	if cfg.Telemetry.Endpoint != "https://env-collector.example.test/v1/traces" {
		t.Fatalf("Telemetry.Endpoint = %q, want env endpoint", cfg.Telemetry.Endpoint)
	}
	if cfg.Telemetry.Directory != "file-telemetry" {
		t.Fatalf("Telemetry.Directory = %q, want file-telemetry", cfg.Telemetry.Directory)
	}
	if cfg.Telemetry.Capture != "minimal" {
		t.Fatalf("Telemetry.Capture = %q, want file minimal", cfg.Telemetry.Capture)
	}
}
