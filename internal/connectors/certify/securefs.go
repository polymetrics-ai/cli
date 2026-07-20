package certify

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
)

func secureDir(root string, parts ...string) (string, error) {
	if root == "" {
		return "", fmt.Errorf("certify: secure directory root is empty")
	}
	if err := os.MkdirAll(root, 0o700); err != nil {
		return "", err
	}
	if err := rejectSymlink(root); err != nil {
		return "", err
	}
	current := root
	for _, part := range parts {
		if part == "" || part == "." || part == ".." || filepath.Base(part) != part {
			return "", fmt.Errorf("certify: unsafe path component %q", part)
		}
		current = filepath.Join(current, part)
		if err := os.Mkdir(current, 0o700); err != nil && !os.IsExist(err) {
			return "", err
		}
		if err := rejectSymlink(current); err != nil {
			return "", err
		}
		if err := os.Chmod(current, 0o700); err != nil {
			return "", err
		}
	}
	return current, nil
}

func rejectSymlink(path string) error {
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return fmt.Errorf("certify: path %s must be a non-symlink directory", path)
	}
	return nil
}

// DurableLedgerBase returns the validated certification-ledger directory
// beneath a caller's project root, creating only that fixed local layout.
func DurableLedgerBase(root string) (string, error) {
	return secureDir(root, ".polymetrics", certificationsDirName, "ledger")
}

// DurableLedgerRoot returns one validated connector's durable ledger root.
func DurableLedgerRoot(root, connector string) (string, error) {
	if !connectorNamePattern.MatchString(connector) {
		return "", fmt.Errorf("certify: invalid ledger connector %q", connector)
	}
	base, err := DurableLedgerBase(root)
	if err != nil {
		return "", err
	}
	return secureDir(base, connector)
}

func atomicWritePrivate(dir, name string, data []byte) error {
	if name == "" || filepath.Base(name) != name {
		return fmt.Errorf("certify: unsafe output filename %q", name)
	}
	path := filepath.Join(dir, name)
	if info, err := os.Lstat(path); err == nil {
		if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
			return fmt.Errorf("certify: refusing to replace non-regular output %s", path)
		}
	} else if !os.IsNotExist(err) {
		return err
	}

	var nonce [8]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		return err
	}
	tmpName := "." + name + ".tmp-" + hex.EncodeToString(nonce[:])
	tmpPath := filepath.Join(dir, tmpName)
	f, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	removeTmp := true
	defer func() {
		_ = f.Close()
		if removeTmp {
			_ = os.Remove(tmpPath)
		}
	}()
	if _, err := f.Write(data); err != nil {
		return err
	}
	if err := f.Sync(); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return err
	}
	removeTmp = false
	dirFile, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer func() { _ = dirFile.Close() }()
	return dirFile.Sync()
}
