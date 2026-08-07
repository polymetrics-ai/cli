package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/durability"
	"polymetrics.ai/internal/warehouse"
)

const catalogFileVersion = 1

// catalogStorage persists one schema-only document per connector/account. The
// state document contains an opaque index reference, never the discovery
// payload itself, so a HubSpot-sized catalog does not amplify unrelated state
// writes. Its narrow interface is also reusable when a dynamic destination
// needs a target catalog in a later phase.
type catalogStorage struct {
	root          string
	mkdir         func(string, fs.FileMode) error
	createTemp    func(string, string) (*os.File, error)
	rename        func(string, string) error
	readFile      func(string) ([]byte, error)
	syncDirectory func(string) error
}

type storedCatalog struct {
	Version   int                `json:"version"`
	Catalog   connectors.Catalog `json:"catalog"`
	UpdatedAt time.Time          `json:"updated_at"`
}

func newCatalogStorage(projectDir string) catalogStorage {
	return catalogStorage{
		root:          projectDir,
		mkdir:         os.Mkdir,
		createTemp:    os.CreateTemp,
		rename:        os.Rename,
		readFile:      os.ReadFile,
		syncDirectory: durability.SyncDirectory,
	}
}

func (a *App) catalogReference(connector string, runtime connectors.RuntimeConfig) (catalogReference, error) {
	key := runtime.CoordinationIdentity.AuthCohortKey()
	if key == "" {
		return catalogReference{}, errors.New("catalog coordination identity is unavailable")
	}
	file, err := catalogRelativePath(connector, key)
	if err != nil {
		return catalogReference{}, err
	}
	return catalogReference{Connector: connector, AccountKey: key, File: file}, nil
}

func catalogRelativePath(connector string, accountKey connectors.AuthCohortKey) (string, error) {
	if !safeCatalogPathPart(connector) {
		return "", errors.New("catalog connector identity is invalid")
	}
	if !safeCatalogPathPart(string(accountKey)) {
		return "", errors.New("catalog account identity is invalid")
	}
	return filepath.Join("catalogs", connector, string(accountKey)+".json"), nil
}

func validCatalogReference(reference catalogReference) bool {
	path, err := catalogRelativePath(reference.Connector, reference.AccountKey)
	return err == nil && reference.File == path
}

// safeCatalogPathPart is the #3892 guard, now shared with the warehouse layout
// so one rule governs every generated path component. Restating it here would
// let the two drift apart.
func safeCatalogPathPart(value string) bool {
	return warehouse.SafePathPart(value)
}

func (s catalogStorage) write(reference catalogReference, catalog connectors.Catalog, updatedAt time.Time) (err error) {
	if !validCatalogReference(reference) {
		return errors.New("catalog reference is invalid")
	}
	if strings.TrimSpace(catalog.Connector) != reference.Connector {
		return errors.New("catalog connector does not match its reference")
	}
	path := filepath.Join(s.root, reference.File)
	dir := filepath.Dir(path)
	created, err := s.ensureDirectory(dir)
	if err != nil {
		return fmt.Errorf("create catalog directory: %w", err)
	}
	payload, err := json.MarshalIndent(storedCatalog{Version: catalogFileVersion, Catalog: catalog, UpdatedAt: updatedAt.UTC()}, "", "  ")
	if err != nil {
		return fmt.Errorf("encode catalog file: %w", err)
	}
	payload = append(payload, '\n')
	temp, err := s.createTemp(dir, ".catalog.tmp-*")
	if err != nil {
		return fmt.Errorf("create temporary catalog file: %w", err)
	}
	tempPath := temp.Name()
	defer func() {
		if temp != nil {
			_ = temp.Close()
		}
		if err != nil {
			_ = os.Remove(tempPath)
		}
	}()
	if _, err := temp.Write(payload); err != nil {
		return fmt.Errorf("write temporary catalog file: %w", err)
	}
	if err := temp.Sync(); err != nil {
		return fmt.Errorf("sync temporary catalog file: %w", err)
	}
	if err := temp.Close(); err != nil {
		return fmt.Errorf("close temporary catalog file: %w", err)
	}
	temp = nil
	if err := s.rename(tempPath, path); err != nil {
		return fmt.Errorf("replace catalog file: %w", err)
	}
	if err := s.syncDirectoryChain(dir, created); err != nil {
		return fmt.Errorf("sync durable catalog file: %w", err)
	}
	return nil
}

func (s catalogStorage) read(reference catalogReference) (storedCatalog, error) {
	if !validCatalogReference(reference) {
		return storedCatalog{}, errors.New("catalog reference is invalid")
	}
	raw, err := s.readFile(filepath.Join(s.root, reference.File))
	if err != nil {
		return storedCatalog{}, err
	}
	var stored storedCatalog
	if err := json.Unmarshal(raw, &stored); err != nil {
		return storedCatalog{}, fmt.Errorf("decode catalog file: %w", err)
	}
	if stored.Version != catalogFileVersion || stored.Catalog.Connector != reference.Connector {
		return storedCatalog{}, errors.New("catalog file does not match its reference")
	}
	return stored, nil
}

// ensureDirectory records exactly which mkdir operations created a directory.
// There is deliberately no preflight Stat: syncDirectoryChain uses the actual
// successful mkdir results, including every parent whose new entry a durable
// file depends on.
func (s catalogStorage) ensureDirectory(dir string) ([]string, error) {
	chain := directoryChain(dir)
	created := make([]string, 0, len(chain))
	for _, candidate := range chain {
		err := s.mkdir(candidate, 0o700)
		if err == nil {
			created = append(created, candidate)
			continue
		}
		if errors.Is(err, fs.ErrExist) {
			continue
		}
		return nil, err
	}
	return created, nil
}

func directoryChain(dir string) []string {
	clean := filepath.Clean(dir)
	chain := make([]string, 0)
	for current := clean; ; current = filepath.Dir(current) {
		chain = append(chain, current)
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
	}
	for left, right := 0, len(chain)-1; left < right; left, right = left+1, right-1 {
		chain[left], chain[right] = chain[right], chain[left]
	}
	return chain
}

func (s catalogStorage) syncDirectoryChain(fileDir string, created []string) error {
	dirs := append([]string{fileDir}, created...)
	seen := make(map[string]struct{}, len(dirs))
	for index := len(dirs) - 1; index >= 0; index-- {
		dir := dirs[index]
		if _, already := seen[dir]; already {
			continue
		}
		seen[dir] = struct{}{}
		if err := s.syncDirectory(dir); err != nil {
			return err
		}
	}
	return nil
}
