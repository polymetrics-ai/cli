// Package config loads invocation-scoped Polymetrics CLI configuration.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

const (
	defaultProject          = "polymetrics-local"
	defaultWarehouse        = "warehouse"
	defaultWarehousePath    = ".polymetrics/warehouse"
	defaultPostgresURL      = "postgres://polymetrics:polymetrics@localhost:15433/polymetrics?sslmode=disable"
	defaultDragonflyAddr    = "localhost:6379"
	defaultTemporalAddr     = "localhost:7233"
	defaultRLMImage         = "ghcr.io/polymetrics/rlm-agent:latest"
	defaultPodmanBin        = "podman"
	defaultLLMProvider      = "openrouter"
	defaultOpenRouterBase   = "https://openrouter.ai/api/v1"
	configRelativeDirectory = ".polymetrics"
	configFileName          = "config.yaml"
)

// Options controls one invocation-scoped config load.
type Options struct {
	// Root is the invocation project root used to find .polymetrics/config.yaml.
	// The config file's own root key does not relocate discovery for the same load.
	Root string
	// Flags optionally supplies already-parsed global flag values. Only root and
	// json are bound by this package; command-specific flags remain legacy-owned.
	Flags map[string]FlagValue
}

// FlagValue is the minimal Viper-compatible view of a bound flag.
type FlagValue interface {
	HasChanged() bool
	Name() string
	ValueString() string
	ValueType() string
}

// StaticFlag is an immutable invocation-scoped flag value for config binding.
type StaticFlag struct {
	FlagName string
	Value    string
	Type     string
	Changed  bool
}

func (f StaticFlag) HasChanged() bool    { return f.Changed }
func (f StaticFlag) Name() string        { return f.FlagName }
func (f StaticFlag) ValueString() string { return f.Value }
func (f StaticFlag) ValueType() string   { return f.Type }

// Config is the typed app configuration resolved for one CLI invocation.
type Config struct {
	Root       string          `json:"root" mapstructure:"root"`
	JSON       bool            `json:"json" mapstructure:"json"`
	Version    int             `json:"version" mapstructure:"version"`
	Project    string          `json:"project" mapstructure:"project"`
	Warehouse  WarehouseConfig `json:"warehouse" mapstructure:"warehouse"`
	Runtime    RuntimeConfig   `json:"runtime" mapstructure:"runtime"`
	RLM        RLMConfig       `json:"rlm" mapstructure:"rlm"`
	Schedule   ScheduleConfig  `json:"schedule" mapstructure:"schedule"`
	ConfigFile string          `json:"config_file" mapstructure:"-"`
}

// WarehouseConfig mirrors the non-secret workspace warehouse defaults written by pm init.
type WarehouseConfig struct {
	Connector string `json:"connector" mapstructure:"connector"`
	Path      string `json:"path" mapstructure:"path"`
}

// RuntimeConfig configures optional PostgreSQL, DragonflyDB, and Temporal services.
type RuntimeConfig struct {
	PostgresURL   string `json:"postgres_url" mapstructure:"postgres_url"`
	DragonflyAddr string `json:"dragonfly_addr" mapstructure:"dragonfly_addr"`
	TemporalAddr  string `json:"temporal_addr" mapstructure:"temporal_addr"`
}

// RLMConfig configures optional runtime-backed RLM agent execution.
type RLMConfig struct {
	Image          string    `json:"image" mapstructure:"image"`
	PodmanBin      string    `json:"podman_bin" mapstructure:"podman_bin"`
	FakeRunner     bool      `json:"fake_runner" mapstructure:"fake_runner"`
	EmbeddedWorker bool      `json:"embedded_worker" mapstructure:"embedded_worker"`
	LLM            LLMConfig `json:"llm" mapstructure:"llm"`
}

// LLMConfig contains non-secret LLM client configuration.
type LLMConfig struct {
	Provider string `json:"provider" mapstructure:"provider"`
	BaseURL  string `json:"base_url" mapstructure:"base_url"`
	Model    string `json:"model" mapstructure:"model"`
}

// ScheduleConfig configures local schedule installation seams.
type ScheduleConfig struct {
	CrontabFile string `json:"crontab_file" mapstructure:"crontab_file"`
}

// LoadError reports a config file read/decode/unmarshal failure.
type LoadError struct {
	Path string
	Err  error
}

func (e *LoadError) Error() string {
	if e == nil {
		return "config: load error"
	}
	if e.Path == "" {
		return fmt.Sprintf("config: %v", e.Err)
	}
	return fmt.Sprintf("config: read %s: %v", e.Path, e.Err)
}

func (e *LoadError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

// Load resolves config for one invocation. Precedence is bound flags, explicit
// environment variables, .polymetrics/config.yaml, then defaults.
func Load(opts Options) (Config, error) {
	root := opts.Root
	if root == "" {
		root = "."
	}
	configPath := filepath.Join(root, configRelativeDirectory, configFileName)

	v := viper.New()
	v.SetConfigFile(configPath)
	setDefaults(v, root)
	bindEnv(v)
	if err := bindFlags(v, opts.Flags); err != nil {
		return Config{}, &LoadError{Path: configPath, Err: err}
	}

	if err := v.ReadInConfig(); err != nil && !isConfigMissing(err) {
		return Config{}, &LoadError{Path: configPath, Err: err}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return Config{}, &LoadError{Path: configPath, Err: err}
	}
	cfg.ConfigFile = configPath
	return cfg, nil
}

func setDefaults(v *viper.Viper, root string) {
	v.SetDefault("root", root)
	v.SetDefault("json", false)
	v.SetDefault("version", 1)
	v.SetDefault("project", defaultProject)
	v.SetDefault("warehouse.connector", defaultWarehouse)
	v.SetDefault("warehouse.path", defaultWarehousePath)
	v.SetDefault("runtime.postgres_url", defaultPostgresURL)
	v.SetDefault("runtime.dragonfly_addr", defaultDragonflyAddr)
	v.SetDefault("runtime.temporal_addr", defaultTemporalAddr)
	v.SetDefault("rlm.image", defaultRLMImage)
	v.SetDefault("rlm.podman_bin", defaultPodmanBin)
	v.SetDefault("rlm.fake_runner", false)
	v.SetDefault("rlm.embedded_worker", false)
	v.SetDefault("rlm.llm.provider", defaultLLMProvider)
	v.SetDefault("rlm.llm.base_url", defaultOpenRouterBase)
	v.SetDefault("rlm.llm.model", "")
	v.SetDefault("schedule.crontab_file", "")
}

func bindEnv(v *viper.Viper) {
	bindings := map[string][]string{
		"root":                   {"POLYMETRICS_ROOT", "PM_ROOT"},
		"json":                   {"POLYMETRICS_JSON", "PM_JSON"},
		"version":                {"POLYMETRICS_VERSION", "PM_VERSION"},
		"project":                {"POLYMETRICS_PROJECT", "PM_PROJECT"},
		"warehouse.connector":    {"POLYMETRICS_WAREHOUSE_CONNECTOR", "PM_WAREHOUSE_CONNECTOR"},
		"warehouse.path":         {"POLYMETRICS_WAREHOUSE_PATH", "PM_WAREHOUSE_PATH"},
		"runtime.postgres_url":   {"POLYMETRICS_POSTGRES_URL", "PM_POSTGRES_URL"},
		"runtime.dragonfly_addr": {"POLYMETRICS_DRAGONFLY_ADDR", "PM_DRAGONFLY_ADDR"},
		"runtime.temporal_addr":  {"POLYMETRICS_TEMPORAL_ADDR", "PM_TEMPORAL_ADDR"},
		"rlm.image":              {"POLYMETRICS_RLM_IMAGE", "PM_RLM_IMAGE"},
		"rlm.podman_bin":         {"POLYMETRICS_PODMAN_BIN", "PM_PODMAN_BIN"},
		"rlm.fake_runner":        {"POLYMETRICS_RLM_FAKE_RUNNER", "PM_RLM_FAKE_RUNNER"},
		"rlm.embedded_worker":    {"POLYMETRICS_RLM_EMBEDDED_WORKER", "PM_RLM_EMBEDDED_WORKER"},
		"rlm.llm.provider":       {"POLYMETRICS_LLM_PROVIDER", "PM_LLM_PROVIDER"},
		"rlm.llm.base_url":       {"POLYMETRICS_LLM_BASE_URL", "PM_LLM_BASE_URL"},
		"rlm.llm.model":          {"POLYMETRICS_LLM_MODEL", "PM_LLM_MODEL"},
		"schedule.crontab_file":  {"POLYMETRICS_CRONTAB_FILE", "PM_CRONTAB_FILE"},
	}
	for key, envNames := range bindings {
		args := append([]string{key}, envNames...)
		_ = v.BindEnv(args...)
	}
}

func bindFlags(v *viper.Viper, flags map[string]FlagValue) error {
	if len(flags) == 0 {
		return nil
	}
	for _, name := range []string{"root", "json"} {
		flag := flags[name]
		if flag == nil {
			continue
		}
		if err := v.BindFlagValue(name, flag); err != nil {
			return fmt.Errorf("bind --%s: %w", name, err)
		}
	}
	return nil
}

func isConfigMissing(err error) bool {
	var notFound viper.ConfigFileNotFoundError
	return errors.As(err, &notFound) || errors.Is(err, os.ErrNotExist)
}
