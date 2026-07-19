package certify

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"sort"
	"strings"
	"sync"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/defs"
)

type resumeIdentity struct {
	ManifestHash string
	Fingerprint  string
}

var (
	builtinRegistryOnce sync.Once
	builtinRegistry     *connectors.Registry
)

func reportIdentity(opts Options) (resumeIdentity, error) {
	manifestHash, err := connectorManifestHash(opts.Connector)
	if err != nil {
		return resumeIdentity{}, err
	}
	fingerprint, err := effectiveOptionsFingerprint(opts)
	if err != nil {
		return resumeIdentity{}, err
	}
	return resumeIdentity{ManifestHash: manifestHash, Fingerprint: fingerprint}, nil
}

func connectorManifestHash(name string) (string, error) {
	var paths []string
	walkErr := fs.WalkDir(defs.FS, name, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if path != name && (strings.Contains(path, "/fixtures") || strings.Contains(path, "/docs")) {
				return fs.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(path, ".json") || strings.HasSuffix(path, ".md") {
			paths = append(paths, path)
		}
		return nil
	})
	if walkErr != nil && !errors.Is(walkErr, fs.ErrNotExist) {
		return "", fmt.Errorf("certify: walk connector identity %q: %w", name, walkErr)
	}
	if len(paths) > 0 {
		sort.Strings(paths)
		hash := sha256.New()
		for _, path := range paths {
			raw, err := fs.ReadFile(defs.FS, path)
			if err != nil {
				return "", fmt.Errorf("certify: read connector identity %s: %w", path, err)
			}
			_, _ = hash.Write([]byte(path))
			_, _ = hash.Write([]byte{0})
			_, _ = hash.Write(raw)
			_, _ = hash.Write([]byte{0})
		}
		return "sha256:" + hex.EncodeToString(hash.Sum(nil)), nil
	}

	// Built-in local connectors such as sample have no defs directory. Their
	// manifest is available without loading the full bundle registry.
	builtinRegistryOnce.Do(func() {
		builtinRegistry = connectors.NewEmptyRegistry()
		builtinRegistry.RegisterBuiltins()
	})
	connector, ok := builtinRegistry.Get(name)
	if !ok {
		return "", fmt.Errorf("certify: connector %q is not registered", name)
	}
	raw, err := json.Marshal(connectors.ManifestOf(connector))
	if err != nil {
		return "", fmt.Errorf("certify: marshal connector identity: %w", err)
	}
	sum := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}

func effectiveOptionsFingerprint(opts Options) (string, error) {
	// SecretEnv contains environment variable names only. Config is validated
	// as non-secret before production runners are constructed.
	value := struct {
		Connector string            `json:"connector"`
		Stream    string            `json:"stream,omitempty"`
		Config    map[string]string `json:"config,omitempty"`
		SecretEnv map[string]string `json:"secret_env_refs,omitempty"`
		Write     bool              `json:"write"`
		Full      bool              `json:"full"`
	}{
		Connector: opts.Connector,
		Stream:    opts.Stream,
		Config:    effectiveCredentialConfig(opts.Connector, opts.Config),
		SecretEnv: opts.SecretEnv,
		Write:     opts.Write,
		Full:      opts.Full,
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return "", fmt.Errorf("certify: marshal effective options: %w", err)
	}
	sum := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}
