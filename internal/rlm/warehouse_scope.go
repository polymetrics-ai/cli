package rlm

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"polymetrics.ai/internal/safety"
)

// WarehouseScope confines RLM table effects to a warehouse beneath a held
// filesystem root. It intentionally exposes no generic file operations.
type WarehouseScope struct {
	fs           *safety.LocalWriteFS
	warehouseDir string
}

// OpenProjectWarehouse holds the selected project root and binds the standard
// local warehouse beneath it. Replacing the warehouse directory after this
// call cannot re-root later table effects outside the project.
func OpenProjectWarehouse(projectRoot string) (*WarehouseScope, error) {
	return openWarehouseScope(projectRoot, filepath.Join(projectRoot, ".polymetrics", "warehouse"))
}

func openWarehouseScope(projectRoot, warehouseDir string) (*WarehouseScope, error) {
	fs, err := safety.OpenLocalWriteFS(projectRoot, false)
	if err != nil {
		return nil, err
	}
	closeOnError := func(err error) (*WarehouseScope, error) {
		if closeErr := fs.Close(); closeErr != nil {
			return nil, errors.Join(err, fmt.Errorf("close RLM warehouse scope: %w", closeErr))
		}
		return nil, err
	}

	resolvedWarehouse, err := safety.ResolveLocalWritePath(projectRoot, warehouseDir, "RLM warehouse path", false)
	if err != nil {
		return closeOnError(err)
	}
	rootAbs, err := filepath.Abs(projectRoot)
	if err != nil {
		return closeOnError(fmt.Errorf("resolve RLM project root: %w", err))
	}
	warehouseRel, err := filepath.Rel(rootAbs, resolvedWarehouse)
	if err != nil {
		return closeOnError(fmt.Errorf("resolve RLM warehouse path: %w", err))
	}
	if !filepath.IsLocal(warehouseRel) {
		return closeOnError(fmt.Errorf("RLM warehouse path is outside the selected project root"))
	}
	return &WarehouseScope{fs: fs, warehouseDir: warehouseRel}, nil
}

// Close releases the held project root.
func (w *WarehouseScope) Close() error {
	if w == nil || w.fs == nil {
		return nil
	}
	return w.fs.Close()
}

func (w *WarehouseScope) openTable(table string) (*os.File, error) {
	if err := validateInTable(table); err != nil {
		return nil, err
	}
	return w.fs.Open(filepath.Join(w.warehouseDir, table+".ndjson"))
}

func (w *WarehouseScope) openOutputTemp(table string) (*os.File, string, string, error) {
	if err := validateOutTable(table); err != nil {
		return nil, "", "", err
	}
	outPath := filepath.Join(w.warehouseDir, table+".ndjson")
	tmpPath := outPath + ".tmp"
	f, err := w.fs.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	return f, tmpPath, outPath, err
}

func (w *WarehouseScope) remove(path string) error {
	return w.fs.Remove(path)
}

func (w *WarehouseScope) commitOutput(tmpPath, outPath string) error {
	return w.fs.Rename(tmpPath, outPath)
}

func (req RunRequest) warehouseScope() (*WarehouseScope, bool, error) {
	if req.Warehouse != nil {
		return req.Warehouse, false, nil
	}
	scope, err := openWarehouseScope(req.WarehouseDir, req.WarehouseDir)
	return scope, true, err
}
