package safety

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// LocalWriteFS performs filesystem effects either beneath a held project root
// or, when explicitly allowed, against ordinary OS paths.
type LocalWriteFS struct {
	rootPath      string
	root          *os.Root
	allowExternal bool
}

// OpenLocalWriteFS creates a filesystem effect scope. With allowExternal set,
// paths retain ordinary os package behavior. Otherwise every operation is
// resolved by os.Root beneath projectRoot at effect time.
func OpenLocalWriteFS(projectRoot string, allowExternal bool) (*LocalWriteFS, error) {
	fs := &LocalWriteFS{allowExternal: allowExternal}
	if allowExternal {
		return fs, nil
	}
	rootPath, err := filepath.Abs(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("resolve local-write root: %w", err)
	}
	root, err := os.OpenRoot(rootPath)
	if err != nil {
		return nil, fmt.Errorf("open local-write root: %w", err)
	}
	fs.rootPath = rootPath
	fs.root = root
	return fs, nil
}

// Close releases the held root, if any.
func (fs *LocalWriteFS) Close() error {
	if fs.root == nil {
		return nil
	}
	return fs.root.Close()
}

// MkdirAll creates path and missing parents within the effect scope.
func (fs *LocalWriteFS) MkdirAll(path string, perm os.FileMode) error {
	name, err := fs.effectPath(path)
	if err != nil {
		return err
	}
	if fs.allowExternal {
		return os.MkdirAll(name, perm)
	}
	return fs.root.MkdirAll(name, perm)
}

// Open opens path for reading within the effect scope.
func (fs *LocalWriteFS) Open(path string) (*os.File, error) {
	name, err := fs.effectPath(path)
	if err != nil {
		return nil, err
	}
	if fs.allowExternal {
		return os.Open(name)
	}
	return fs.root.Open(name)
}

// OpenFile opens path with the requested flags within the effect scope.
func (fs *LocalWriteFS) OpenFile(path string, flag int, perm os.FileMode) (*os.File, error) {
	name, err := fs.effectPath(path)
	if err != nil {
		return nil, err
	}
	if fs.allowExternal {
		return os.OpenFile(name, flag, perm)
	}
	return fs.root.OpenFile(name, flag, perm)
}

// Remove removes path within the effect scope.
func (fs *LocalWriteFS) Remove(path string) error {
	name, err := fs.effectPath(path)
	if err != nil {
		return err
	}
	if fs.allowExternal {
		return os.Remove(name)
	}
	return fs.root.Remove(name)
}

// Rename renames oldPath to newPath within the same effect scope.
func (fs *LocalWriteFS) Rename(oldPath, newPath string) error {
	oldName, err := fs.effectPath(oldPath)
	if err != nil {
		return err
	}
	newName, err := fs.effectPath(newPath)
	if err != nil {
		return err
	}
	if fs.allowExternal {
		return os.Rename(oldName, newName)
	}
	return fs.root.Rename(oldName, newName)
}

func (fs *LocalWriteFS) effectPath(path string) (string, error) {
	if fs.allowExternal {
		return path, nil
	}
	name := filepath.Clean(path)
	if filepath.IsAbs(name) {
		var err error
		name, err = filepath.Rel(fs.rootPath, name)
		if err != nil {
			return "", fmt.Errorf("resolve local-write path: %w", err)
		}
	}
	if !filepath.IsLocal(name) {
		return "", errors.New("local-write path is outside the selected root")
	}
	return name, nil
}
