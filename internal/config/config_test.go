package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

type configKeyCase struct {
	name       string
	fileYAML   string
	primaryEnv string
	aliasEnv   string
	envValue   string
	fileWant   any
	envWant    any
	get        func(Config) any
}

var allBoundEnvVars = []string{
	"POLYMETRICS_ROOT", "PM_ROOT",
	"POLYMETRICS_JSON", "PM_JSON",
	"POLYMETRICS_PROJECT", "PM_PROJECT",
	"POLYMETRICS_WAREHOUSE_CONNECTOR", "PM_WAREHOUSE_CONNECTOR",
	"POLYMETRICS_WAREHOUSE_PATH", "PM_WAREHOUSE_PATH",
	"POLYMETRICS_POSTGRES_URL", "PM_POSTGRES_URL",
	"POLYMETRICS_DRAGONFLY_ADDR", "PM_DRAGONFLY_ADDR",
	"POLYMETRICS_TEMPORAL_ADDR", "PM_TEMPORAL_ADDR",
	"POLYMETRICS_RLM_IMAGE", "PM_RLM_IMAGE",
	"POLYMETRICS_PODMAN_BIN", "PM_PODMAN_BIN",
	"POLYMETRICS_RLM_FAKE_RUNNER", "PM_RLM_FAKE_RUNNER",
	"POLYMETRICS_RLM_EMBEDDED_WORKER", "PM_RLM_EMBEDDED_WORKER",
	"POLYMETRICS_LLM_PROVIDER", "PM_LLM_PROVIDER",
	"POLYMETRICS_LLM_BASE_URL", "PM_LLM_BASE_URL",
	"POLYMETRICS_LLM_MODEL", "PM_LLM_MODEL",
	"POLYMETRICS_CRONTAB_FILE", "PM_CRONTAB_FILE",
	"POLYMETRICS_TELEMETRY", "PM_TELEMETRY",
	"POLYMETRICS_TELEMETRY_ENDPOINT", "PM_TELEMETRY_ENDPOINT",
	"POLYMETRICS_OTEL_EXPORTER_OTLP_ENDPOINT", "OTEL_EXPORTER_OTLP_ENDPOINT",
	"POLYMETRICS_TELEMETRY_DIR", "PM_TELEMETRY_DIR",
	"POLYMETRICS_TELEMETRY_CAPTURE", "PM_TELEMETRY_CAPTURE",
}

func TestLoadDefaultsAndMissingFile(t *testing.T) {
	clearBoundEnv(t)
	root := t.TempDir()

	cfg, err := Load(Options{Root: root})
	if err != nil {
		t.Fatalf("Load missing file: %v", err)
	}

	if cfg.Root != root {
		t.Fatalf("Root = %q, want %q", cfg.Root, root)
	}
	if cfg.JSON {
		t.Fatal("JSON = true, want false")
	}
	if cfg.Version != 1 {
		t.Fatalf("Version = %d, want 1", cfg.Version)
	}
	if cfg.Project != "polymetrics-local" {
		t.Fatalf("Project = %q, want polymetrics-local", cfg.Project)
	}
	if cfg.Warehouse.Connector != "warehouse" || cfg.Warehouse.Path != ".polymetrics/warehouse" {
		t.Fatalf("Warehouse = %#v, want connector=warehouse path=.polymetrics/warehouse", cfg.Warehouse)
	}
	if cfg.Runtime.PostgresURL == "" || cfg.Runtime.DragonflyAddr != "localhost:6379" || cfg.Runtime.TemporalAddr != "localhost:7233" {
		t.Fatalf("Runtime defaults = %#v", cfg.Runtime)
	}
	if cfg.RLM.Image != "ghcr.io/polymetrics/rlm-agent:latest" || cfg.RLM.PodmanBin != "podman" {
		t.Fatalf("RLM defaults = %#v", cfg.RLM)
	}
	if cfg.RLM.FakeRunner || cfg.RLM.EmbeddedWorker {
		t.Fatalf("RLM bool defaults = fake:%v embedded:%v, want false/false", cfg.RLM.FakeRunner, cfg.RLM.EmbeddedWorker)
	}
	if cfg.RLM.LLM.Provider != "openrouter" || cfg.RLM.LLM.BaseURL != "https://openrouter.ai/api/v1" || cfg.RLM.LLM.Model != "" {
		t.Fatalf("LLM defaults = %#v", cfg.RLM.LLM)
	}
	if cfg.Schedule.CrontabFile != "" {
		t.Fatalf("Schedule.CrontabFile = %q, want empty", cfg.Schedule.CrontabFile)
	}
	if cfg.Telemetry.Exporter != "none" || cfg.Telemetry.Directory != ".polymetrics/telemetry" || cfg.Telemetry.Capture != "default" || cfg.Telemetry.Endpoint != "" {
		t.Fatalf("Telemetry defaults = %#v", cfg.Telemetry)
	}
	if cfg.ConfigFile != filepath.Join(root, ".polymetrics", "config.yaml") {
		t.Fatalf("ConfigFile = %q, want invocation-root config path", cfg.ConfigFile)
	}
}

func TestLoadFileEnvAndAliasPrecedence(t *testing.T) {
	cases := []configKeyCase{
		{name: "root", fileYAML: "root: file-root\n", primaryEnv: "POLYMETRICS_ROOT", aliasEnv: "PM_ROOT", envValue: "env-root", fileWant: "file-root", envWant: "env-root", get: func(c Config) any { return c.Root }},
		{name: "json", fileYAML: "json: false\n", primaryEnv: "POLYMETRICS_JSON", aliasEnv: "PM_JSON", envValue: "true", fileWant: false, envWant: true, get: func(c Config) any { return c.JSON }},
		{name: "version", fileYAML: "version: 2\n", primaryEnv: "POLYMETRICS_VERSION", aliasEnv: "PM_VERSION", envValue: "3", fileWant: 2, envWant: 3, get: func(c Config) any { return c.Version }},
		{name: "project", fileYAML: "project: file-project\n", primaryEnv: "POLYMETRICS_PROJECT", aliasEnv: "PM_PROJECT", envValue: "env-project", fileWant: "file-project", envWant: "env-project", get: func(c Config) any { return c.Project }},
		{name: "warehouse.connector", fileYAML: "warehouse:\n  connector: file-warehouse\n", primaryEnv: "POLYMETRICS_WAREHOUSE_CONNECTOR", aliasEnv: "PM_WAREHOUSE_CONNECTOR", envValue: "env-warehouse", fileWant: "file-warehouse", envWant: "env-warehouse", get: func(c Config) any { return c.Warehouse.Connector }},
		{name: "warehouse.path", fileYAML: "warehouse:\n  path: file/warehouse\n", primaryEnv: "POLYMETRICS_WAREHOUSE_PATH", aliasEnv: "PM_WAREHOUSE_PATH", envValue: "env/warehouse", fileWant: "file/warehouse", envWant: "env/warehouse", get: func(c Config) any { return c.Warehouse.Path }},
		{name: "runtime.postgres_url", fileYAML: "runtime:\n  postgres_url: postgres://file-host/db\n", primaryEnv: "POLYMETRICS_POSTGRES_URL", aliasEnv: "PM_POSTGRES_URL", envValue: "postgres://env-host/db", fileWant: "postgres://file-host/db", envWant: "postgres://env-host/db", get: func(c Config) any { return c.Runtime.PostgresURL }},
		{name: "runtime.dragonfly_addr", fileYAML: "runtime:\n  dragonfly_addr: file-dragonfly:6379\n", primaryEnv: "POLYMETRICS_DRAGONFLY_ADDR", aliasEnv: "PM_DRAGONFLY_ADDR", envValue: "env-dragonfly:6379", fileWant: "file-dragonfly:6379", envWant: "env-dragonfly:6379", get: func(c Config) any { return c.Runtime.DragonflyAddr }},
		{name: "runtime.temporal_addr", fileYAML: "runtime:\n  temporal_addr: file-temporal:7233\n", primaryEnv: "POLYMETRICS_TEMPORAL_ADDR", aliasEnv: "PM_TEMPORAL_ADDR", envValue: "env-temporal:7233", fileWant: "file-temporal:7233", envWant: "env-temporal:7233", get: func(c Config) any { return c.Runtime.TemporalAddr }},
		{name: "rlm.image", fileYAML: "rlm:\n  image: file/image:tag\n", primaryEnv: "POLYMETRICS_RLM_IMAGE", aliasEnv: "PM_RLM_IMAGE", envValue: "env/image:tag", fileWant: "file/image:tag", envWant: "env/image:tag", get: func(c Config) any { return c.RLM.Image }},
		{name: "rlm.podman_bin", fileYAML: "rlm:\n  podman_bin: file-podman\n", primaryEnv: "POLYMETRICS_PODMAN_BIN", aliasEnv: "PM_PODMAN_BIN", envValue: "env-podman", fileWant: "file-podman", envWant: "env-podman", get: func(c Config) any { return c.RLM.PodmanBin }},
		{name: "rlm.fake_runner", fileYAML: "rlm:\n  fake_runner: false\n", primaryEnv: "POLYMETRICS_RLM_FAKE_RUNNER", aliasEnv: "PM_RLM_FAKE_RUNNER", envValue: "true", fileWant: false, envWant: true, get: func(c Config) any { return c.RLM.FakeRunner }},
		{name: "rlm.embedded_worker", fileYAML: "rlm:\n  embedded_worker: false\n", primaryEnv: "POLYMETRICS_RLM_EMBEDDED_WORKER", aliasEnv: "PM_RLM_EMBEDDED_WORKER", envValue: "true", fileWant: false, envWant: true, get: func(c Config) any { return c.RLM.EmbeddedWorker }},
		{name: "rlm.llm.provider", fileYAML: "rlm:\n  llm:\n    provider: file-provider\n", primaryEnv: "POLYMETRICS_LLM_PROVIDER", aliasEnv: "PM_LLM_PROVIDER", envValue: "env-provider", fileWant: "file-provider", envWant: "env-provider", get: func(c Config) any { return c.RLM.LLM.Provider }},
		{name: "rlm.llm.base_url", fileYAML: "rlm:\n  llm:\n    base_url: http://file-llm/v1\n", primaryEnv: "POLYMETRICS_LLM_BASE_URL", aliasEnv: "PM_LLM_BASE_URL", envValue: "http://env-llm/v1", fileWant: "http://file-llm/v1", envWant: "http://env-llm/v1", get: func(c Config) any { return c.RLM.LLM.BaseURL }},
		{name: "rlm.llm.model", fileYAML: "rlm:\n  llm:\n    model: file-model\n", primaryEnv: "POLYMETRICS_LLM_MODEL", aliasEnv: "PM_LLM_MODEL", envValue: "env-model", fileWant: "file-model", envWant: "env-model", get: func(c Config) any { return c.RLM.LLM.Model }},
		{name: "schedule.crontab_file", fileYAML: "schedule:\n  crontab_file: file-crontab\n", primaryEnv: "POLYMETRICS_CRONTAB_FILE", aliasEnv: "PM_CRONTAB_FILE", envValue: "env-crontab", fileWant: "file-crontab", envWant: "env-crontab", get: func(c Config) any { return c.Schedule.CrontabFile }},
		{name: "telemetry.exporter", fileYAML: "telemetry:\n  exporter: file\n", primaryEnv: "POLYMETRICS_TELEMETRY", aliasEnv: "PM_TELEMETRY", envValue: "otlp", fileWant: "file", envWant: "otlp", get: func(c Config) any { return c.Telemetry.Exporter }},
		{name: "telemetry.endpoint", fileYAML: "telemetry:\n  endpoint: https://file-collector.example.test/v1/traces\n", primaryEnv: "POLYMETRICS_TELEMETRY_ENDPOINT", aliasEnv: "PM_TELEMETRY_ENDPOINT", envValue: "https://env-collector.example.test/v1/traces", fileWant: "https://file-collector.example.test/v1/traces", envWant: "https://env-collector.example.test/v1/traces", get: func(c Config) any { return c.Telemetry.Endpoint }},
		{name: "telemetry.directory", fileYAML: "telemetry:\n  directory: file-telemetry\n", primaryEnv: "POLYMETRICS_TELEMETRY_DIR", aliasEnv: "PM_TELEMETRY_DIR", envValue: "env-telemetry", fileWant: "file-telemetry", envWant: "env-telemetry", get: func(c Config) any { return c.Telemetry.Directory }},
		{name: "telemetry.capture", fileYAML: "telemetry:\n  capture: minimal\n", primaryEnv: "POLYMETRICS_TELEMETRY_CAPTURE", aliasEnv: "PM_TELEMETRY_CAPTURE", envValue: "default", fileWant: "minimal", envWant: "default", get: func(c Config) any { return c.Telemetry.Capture }},
	}

	for _, tt := range cases {
		t.Run(tt.name+"/file_beats_default", func(t *testing.T) {
			clearBoundEnv(t)
			root := writeConfig(t, tt.fileYAML)
			cfg, err := Load(Options{Root: root})
			if err != nil {
				t.Fatalf("Load: %v", err)
			}
			if got := tt.get(cfg); got != tt.fileWant {
				t.Fatalf("file value = %#v, want %#v", got, tt.fileWant)
			}
		})

		t.Run(tt.name+"/primary_env_beats_file", func(t *testing.T) {
			clearBoundEnv(t)
			root := writeConfig(t, tt.fileYAML)
			t.Setenv(tt.primaryEnv, tt.envValue)
			t.Setenv(tt.aliasEnv, "alias-should-not-win")
			cfg, err := Load(Options{Root: root})
			if err != nil {
				t.Fatalf("Load: %v", err)
			}
			if got := tt.get(cfg); got != tt.envWant {
				t.Fatalf("primary env value = %#v, want %#v", got, tt.envWant)
			}
		})

		t.Run(tt.name+"/pm_alias_beats_file_when_primary_absent", func(t *testing.T) {
			clearBoundEnv(t)
			root := writeConfig(t, tt.fileYAML)
			t.Setenv(tt.aliasEnv, tt.envValue)
			cfg, err := Load(Options{Root: root})
			if err != nil {
				t.Fatalf("Load: %v", err)
			}
			if got := tt.get(cfg); got != tt.envWant {
				t.Fatalf("alias env value = %#v, want %#v", got, tt.envWant)
			}
		})
	}
}

func TestLoadBoundGlobalFlagsBeatEnvAndFile(t *testing.T) {
	clearBoundEnv(t)
	root := writeConfig(t, "root: file-root\njson: false\n")
	t.Setenv("POLYMETRICS_ROOT", "env-root")
	t.Setenv("POLYMETRICS_JSON", "false")

	flags := map[string]FlagValue{
		"root": StaticFlag{FlagName: "root", Value: "flag-root", Type: "string", Changed: true},
		"json": StaticFlag{FlagName: "json", Value: "true", Type: "bool", Changed: true},
	}

	cfg, err := Load(Options{Root: root, Flags: flags})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Root != "flag-root" {
		t.Fatalf("Root = %q, want flag-root", cfg.Root)
	}
	if !cfg.JSON {
		t.Fatal("JSON = false, want true from bound flag")
	}
}

func TestLoadMalformedFile(t *testing.T) {
	clearBoundEnv(t)
	root := writeConfig(t, "runtime:\n  postgres_url: [unterminated\n")

	_, err := Load(Options{Root: root})
	if err == nil {
		t.Fatal("Load malformed file succeeded, want error")
	}
	var loadErr *LoadError
	if !errors.As(err, &loadErr) {
		t.Fatalf("Load error = %T %[1]v, want *LoadError", err)
	}
	if loadErr.Path != filepath.Join(root, ".polymetrics", "config.yaml") {
		t.Fatalf("LoadError.Path = %q, want config path", loadErr.Path)
	}
}

func TestResolveBootstrapRootAndJSONPrecedence(t *testing.T) {
	cases := []struct {
		name     string
		root     string
		flags    map[string]FlagValue
		setupEnv func(t *testing.T)
		wantRoot string
		wantJSON bool
	}{
		{
			name:     "defaults use invocation root",
			root:     "default-root",
			wantRoot: "default-root",
		},
		{
			name: "primary env beats alias and default",
			root: "default-root",
			setupEnv: func(t *testing.T) {
				t.Setenv("POLYMETRICS_ROOT", "primary-root")
				t.Setenv("PM_ROOT", "alias-root")
				t.Setenv("POLYMETRICS_JSON", "false")
				t.Setenv("PM_JSON", "true")
			},
			wantRoot: "primary-root",
			wantJSON: false,
		},
		{
			name: "pm alias beats default when primary absent",
			root: "default-root",
			setupEnv: func(t *testing.T) {
				t.Setenv("PM_ROOT", "alias-root")
				t.Setenv("PM_JSON", "true")
			},
			wantRoot: "alias-root",
			wantJSON: true,
		},
		{
			name: "flags beat env",
			root: "default-root",
			flags: map[string]FlagValue{
				"root": StaticFlag{FlagName: "root", Value: "flag-root", Type: "string", Changed: true},
				"json": StaticFlag{FlagName: "json", Value: "true", Type: "bool", Changed: true},
			},
			setupEnv: func(t *testing.T) {
				t.Setenv("POLYMETRICS_ROOT", "primary-root")
				t.Setenv("POLYMETRICS_JSON", "false")
			},
			wantRoot: "flag-root",
			wantJSON: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			clearBoundEnv(t)
			if tt.setupEnv != nil {
				tt.setupEnv(t)
			}

			bootstrap, err := ResolveBootstrap(Options{Root: tt.root, Flags: tt.flags})
			if err != nil {
				t.Fatalf("ResolveBootstrap: %v", err)
			}
			if bootstrap.Root != tt.wantRoot {
				t.Fatalf("Root = %q, want %q", bootstrap.Root, tt.wantRoot)
			}
			if bootstrap.JSON != tt.wantJSON {
				t.Fatalf("JSON = %v, want %v", bootstrap.JSON, tt.wantJSON)
			}
		})
	}
}

func TestLoadUsesBootstrapRootForConfigDiscovery(t *testing.T) {
	clearBoundEnv(t)
	defaultRoot := t.TempDir()
	envRoot := writeConfig(t, "project: env-project\n")
	t.Setenv("POLYMETRICS_ROOT", envRoot)

	cfg, err := Load(Options{Root: defaultRoot})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.ConfigFile != filepath.Join(envRoot, ".polymetrics", "config.yaml") {
		t.Fatalf("ConfigFile = %q, want env root config", cfg.ConfigFile)
	}
	if cfg.Root != envRoot {
		t.Fatalf("Root = %q, want env root", cfg.Root)
	}
	if cfg.Project != "env-project" {
		t.Fatalf("Project = %q, want env-project", cfg.Project)
	}
}

func TestLoadUsesPMRootAliasForConfigDiscovery(t *testing.T) {
	clearBoundEnv(t)
	defaultRoot := t.TempDir()
	aliasRoot := writeConfig(t, "project: alias-project\n")
	t.Setenv("PM_ROOT", aliasRoot)

	cfg, err := Load(Options{Root: defaultRoot})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.ConfigFile != filepath.Join(aliasRoot, ".polymetrics", "config.yaml") {
		t.Fatalf("ConfigFile = %q, want alias root config", cfg.ConfigFile)
	}
	if cfg.Root != aliasRoot {
		t.Fatalf("Root = %q, want alias root", cfg.Root)
	}
	if cfg.Project != "alias-project" {
		t.Fatalf("Project = %q, want alias-project", cfg.Project)
	}
}

func TestLoadExplicitRootFlagOverridesEnvForConfigDiscovery(t *testing.T) {
	clearBoundEnv(t)
	defaultRoot := t.TempDir()
	envRoot := writeConfig(t, "project: env-project\n")
	flagRoot := writeConfig(t, "project: flag-project\n")
	t.Setenv("POLYMETRICS_ROOT", envRoot)
	flags := map[string]FlagValue{
		"root": StaticFlag{FlagName: "root", Value: flagRoot, Type: "string", Changed: true},
	}

	cfg, err := Load(Options{Root: defaultRoot, Flags: flags})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.ConfigFile != filepath.Join(flagRoot, ".polymetrics", "config.yaml") {
		t.Fatalf("ConfigFile = %q, want flag root config", cfg.ConfigFile)
	}
	if cfg.Root != flagRoot {
		t.Fatalf("Root = %q, want flag root", cfg.Root)
	}
	if cfg.Project != "flag-project" {
		t.Fatalf("Project = %q, want flag-project", cfg.Project)
	}
}

func TestLoadConfigFileRootDoesNotRelocateDiscovery(t *testing.T) {
	clearBoundEnv(t)
	effectiveRoot := t.TempDir()
	effectiveConfigDir := filepath.Join(effectiveRoot, ".polymetrics")
	if err := os.MkdirAll(effectiveConfigDir, 0o700); err != nil {
		t.Fatalf("create effective config dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(effectiveConfigDir, "config.yaml"), []byte("runtime:\n  postgres_url: [unterminated\n"), 0o600); err != nil {
		t.Fatalf("write effective config: %v", err)
	}
	discoveryRoot := writeConfig(t, "root: "+effectiveRoot+"\nproject: discovery-project\n")

	cfg, err := Load(Options{Root: discoveryRoot})
	if err != nil {
		t.Fatalf("Load should not relocate to file root during same load: %v", err)
	}
	if cfg.ConfigFile != filepath.Join(discoveryRoot, ".polymetrics", "config.yaml") {
		t.Fatalf("ConfigFile = %q, want discovery root config", cfg.ConfigFile)
	}
	if cfg.Root != effectiveRoot {
		t.Fatalf("Root = %q, want file root", cfg.Root)
	}
	if cfg.Project != "discovery-project" {
		t.Fatalf("Project = %q, want discovery-project", cfg.Project)
	}
}

func TestLoadRecordsExplicitKeys(t *testing.T) {
	clearBoundEnv(t)
	root := writeConfig(t, "runtime:\n  temporal_addr: configured-temporal:7233\nrlm:\n  fake_runner: false\n")
	t.Setenv("PM_CRONTAB_FILE", "configured-crontab")

	cfg, err := Load(Options{Root: root})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	for _, key := range []string{"runtime.temporal_addr", "rlm.fake_runner", "schedule.crontab_file"} {
		if !cfg.IsExplicit(key) {
			t.Fatalf("IsExplicit(%q) = false, want true", key)
		}
	}
	if cfg.IsExplicit("runtime.dragonfly_addr") {
		t.Fatal("runtime.dragonfly_addr should not be explicit when only defaulted")
	}
}

func TestLoadInvocationIsolationNoStateLeak(t *testing.T) {
	clearBoundEnv(t)
	rootA := writeConfig(t, "project: alpha\njson: true\n")
	rootB := writeConfig(t, "project: beta\n")

	flagsA := map[string]FlagValue{
		"json": StaticFlag{FlagName: "json", Value: "true", Type: "bool", Changed: true},
	}

	cfgA, err := Load(Options{Root: rootA, Flags: flagsA})
	if err != nil {
		t.Fatalf("Load A: %v", err)
	}
	cfgB, err := Load(Options{Root: rootB})
	if err != nil {
		t.Fatalf("Load B: %v", err)
	}

	if cfgA.Project != "alpha" || !cfgA.JSON {
		t.Fatalf("cfgA = %#v", cfgA)
	}
	if cfgB.Project != "beta" || cfgB.JSON {
		t.Fatalf("cfgB leaked state = %#v", cfgB)
	}
}

func TestLoadDoesNotIngestUnboundEnv(t *testing.T) {
	clearBoundEnv(t)
	t.Setenv("POLYMETRICS_UNDOCUMENTED_PROJECT", "should-not-win")
	root := t.TempDir()

	cfg, err := Load(Options{Root: root})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Project != "polymetrics-local" {
		t.Fatalf("Project = %q, want default", cfg.Project)
	}
}

func clearBoundEnv(t *testing.T) {
	t.Helper()
	for _, name := range allBoundEnvVars {
		t.Setenv(name, "")
	}
	// Version is in the table but rarely used by runtime code; keep it isolated too.
	t.Setenv("POLYMETRICS_VERSION", "")
	t.Setenv("PM_VERSION", "")
}

func writeConfig(t *testing.T, yaml string) string {
	t.Helper()
	root := t.TempDir()
	dir := filepath.Join(root, ".polymetrics")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatalf("create config dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(yaml), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return root
}
