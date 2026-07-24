// Package config loads invocation-scoped Polymetrics CLI configuration.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

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

// Bootstrap is the minimal invocation view needed before reading the config file.
type Bootstrap struct {
	Root string
	JSON bool
}

// Config is the typed app configuration resolved for one CLI invocation.
type Config struct {
	Root         string          `json:"root" mapstructure:"root"`
	JSON         bool            `json:"json" mapstructure:"json"`
	Version      int             `json:"version" mapstructure:"version"`
	Project      string          `json:"project" mapstructure:"project"`
	Warehouse    WarehouseConfig `json:"warehouse" mapstructure:"warehouse"`
	Runtime      RuntimeConfig   `json:"runtime" mapstructure:"runtime"`
	RLM          RLMConfig       `json:"rlm" mapstructure:"rlm"`
	Schedule     ScheduleConfig  `json:"schedule" mapstructure:"schedule"`
	ConfigFile   string          `json:"config_file" mapstructure:"-"`
	ExplicitKeys map[string]bool `json:"-" mapstructure:"-"`
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

// IsExplicit reports whether a config key was provided by a bound flag,
// environment variable, or the config file rather than only by defaults.
func (c Config) IsExplicit(key string) bool {
	return c.ExplicitKeys != nil && c.ExplicitKeys[key]
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

// ResolveBootstrap resolves the minimal config state needed before config-file
// discovery. Precedence is bound flags, explicit primary environment variables,
// PM_* aliases, then invocation defaults.
func ResolveBootstrap(opts Options) (Bootstrap, error) {
	root := opts.Root
	if root == "" {
		root = "."
	}
	bootstrap := Bootstrap{Root: root}

	if value, ok := changedFlagValue(opts.Flags, "root"); ok {
		bootstrap.Root = value
	} else if value, ok := lookupBoundEnv("root"); ok {
		bootstrap.Root = value
	}

	if value, ok := changedFlagValue(opts.Flags, "json"); ok {
		jsonValue, err := parseBoolBinding("--json", value)
		if err != nil {
			return bootstrap, err
		}
		bootstrap.JSON = jsonValue
	} else if value, ok := lookupBoundEnv("json"); ok {
		jsonValue, err := parseBoolBinding("json environment", value)
		if err != nil {
			return bootstrap, err
		}
		bootstrap.JSON = jsonValue
	}

	return bootstrap, nil
}

// Load resolves config for one invocation. Precedence is bound flags, explicit
// environment variables, .polymetrics/config.yaml, then defaults.
func Load(opts Options) (Config, error) {
	bootstrap, err := ResolveBootstrap(opts)
	configPath := filepath.Join(bootstrap.Root, configRelativeDirectory, configFileName)
	if err != nil {
		return Config{}, &LoadError{Path: configPath, Err: err}
	}

	v := viper.New()
	v.SetConfigFile(configPath)
	setDefaults(v, bootstrap)
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
	cfg.ExplicitKeys = explicitKeys(v, opts.Flags)
	return cfg, nil
}

func setDefaults(v *viper.Viper, bootstrap Bootstrap) {
	v.SetDefault("root", bootstrap.Root)
	v.SetDefault("json", bootstrap.JSON)
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

type envBinding struct {
	key   string
	names []string
}

func allEnvBindings() []envBinding {
	return []envBinding{
		{key: "root", names: []string{"POLYMETRICS_ROOT", "PM_ROOT"}},
		{key: "json", names: []string{"POLYMETRICS_JSON", "PM_JSON"}},
		{key: "version", names: []string{"POLYMETRICS_VERSION", "PM_VERSION"}},
		{key: "project", names: []string{"POLYMETRICS_PROJECT", "PM_PROJECT"}},
		{key: "warehouse.connector", names: []string{"POLYMETRICS_WAREHOUSE_CONNECTOR", "PM_WAREHOUSE_CONNECTOR"}},
		{key: "warehouse.path", names: []string{"POLYMETRICS_WAREHOUSE_PATH", "PM_WAREHOUSE_PATH"}},
		{key: "runtime.postgres_url", names: []string{"POLYMETRICS_POSTGRES_URL", "PM_POSTGRES_URL"}},
		{key: "runtime.dragonfly_addr", names: []string{"POLYMETRICS_DRAGONFLY_ADDR", "PM_DRAGONFLY_ADDR"}},
		{key: "runtime.temporal_addr", names: []string{"POLYMETRICS_TEMPORAL_ADDR", "PM_TEMPORAL_ADDR"}},
		{key: "rlm.image", names: []string{"POLYMETRICS_RLM_IMAGE", "PM_RLM_IMAGE"}},
		{key: "rlm.podman_bin", names: []string{"POLYMETRICS_PODMAN_BIN", "PM_PODMAN_BIN"}},
		{key: "rlm.fake_runner", names: []string{"POLYMETRICS_RLM_FAKE_RUNNER", "PM_RLM_FAKE_RUNNER"}},
		{key: "rlm.embedded_worker", names: []string{"POLYMETRICS_RLM_EMBEDDED_WORKER", "PM_RLM_EMBEDDED_WORKER"}},
		{key: "rlm.llm.provider", names: []string{"POLYMETRICS_LLM_PROVIDER", "PM_LLM_PROVIDER"}},
		{key: "rlm.llm.base_url", names: []string{"POLYMETRICS_LLM_BASE_URL", "PM_LLM_BASE_URL"}},
		{key: "rlm.llm.model", names: []string{"POLYMETRICS_LLM_MODEL", "PM_LLM_MODEL"}},
		{key: "schedule.crontab_file", names: []string{"POLYMETRICS_CRONTAB_FILE", "PM_CRONTAB_FILE"}},
	}
}

func bindEnv(v *viper.Viper) {
	for _, binding := range allEnvBindings() {
		args := append([]string{binding.key}, binding.names...)
		_ = v.BindEnv(args...)
	}
}

func changedFlagValue(flags map[string]FlagValue, name string) (string, bool) {
	if len(flags) == 0 {
		return "", false
	}
	flag := flags[name]
	if flag == nil || !flag.HasChanged() {
		return "", false
	}
	return flag.ValueString(), true
}

func lookupBoundEnv(key string) (string, bool) {
	for _, name := range envNamesForKey(key) {
		value, ok := os.LookupEnv(name)
		if ok && value != "" {
			return value, true
		}
	}
	return "", false
}

func explicitKeys(v *viper.Viper, flags map[string]FlagValue) map[string]bool {
	keys := make(map[string]bool)
	for _, binding := range allEnvBindings() {
		if _, ok := changedFlagValue(flags, binding.key); ok {
			keys[binding.key] = true
			continue
		}
		if _, ok := lookupBoundEnv(binding.key); ok {
			keys[binding.key] = true
			continue
		}
		if v.InConfig(binding.key) {
			keys[binding.key] = true
		}
	}
	return keys
}

func envNamesForKey(key string) []string {
	for _, binding := range allEnvBindings() {
		if binding.key == key {
			return binding.names
		}
	}
	return nil
}

func parseBoolBinding(source string, value string) (bool, error) {
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("parse %s %q: %w", source, value, err)
	}
	return parsed, nil
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
