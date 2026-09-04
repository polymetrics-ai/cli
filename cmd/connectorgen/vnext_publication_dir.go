package main

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"sort"
	"strings"

	"golang.org/x/sys/unix"
)

// vNextPublicationDirectory owns a directory descriptor. Every child operation
// opens its final component with O_NOFOLLOW, and every multi-component path is
// resolved one verified descriptor at a time.
type vNextPublicationDirectory struct {
	file  *os.File
	label string
}

func vNextPublicationOpenDirectory(path, label string) (*vNextPublicationDirectory, error) {
	fd, err := unix.Open(path, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
	if err != nil {
		return nil, vNextPublicationOpenPathError(path, label, err)
	}
	return vNextPublicationDirectoryFromFD(fd, label)
}

func vNextPublicationDirectoryFromFD(fd int, label string) (*vNextPublicationDirectory, error) {
	file := os.NewFile(uintptr(fd), label)
	if file == nil {
		_ = unix.Close(fd)
		return nil, fmt.Errorf("open %s: invalid file descriptor", label)
	}
	info, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return nil, fmt.Errorf("stat %s: %w", label, err)
	}
	if !info.IsDir() {
		_ = file.Close()
		return nil, fmt.Errorf("%s is not a directory", label)
	}
	return &vNextPublicationDirectory{file: file, label: label}, nil
}

func vNextPublicationOpenError(label string, err error) error {
	if errors.Is(err, unix.ELOOP) {
		return fmt.Errorf("%s is a symlink", label)
	}
	return fmt.Errorf("open %s: %w", label, err)
}

func vNextPublicationOpenPathError(path, label string, err error) error {
	if errors.Is(err, unix.ENOTDIR) {
		var stat unix.Stat_t
		if unix.Lstat(path, &stat) == nil && vNextPublicationStatIsSymlink(stat) {
			return fmt.Errorf("%s is a symlink", label)
		}
	}
	return vNextPublicationOpenError(label, err)
}

func vNextPublicationOpenAtError(directory *vNextPublicationDirectory, name, label string, err error) error {
	if errors.Is(err, unix.ENOTDIR) {
		var stat unix.Stat_t
		if unix.Fstatat(int(directory.file.Fd()), name, &stat, unix.AT_SYMLINK_NOFOLLOW) == nil && vNextPublicationStatIsSymlink(stat) {
			return fmt.Errorf("%s is a symlink", label)
		}
	}
	return vNextPublicationOpenError(label, err)
}

func vNextPublicationDirectNameValid(name string) bool {
	return name != "" && name != "." && name != ".." && !strings.ContainsAny(name, "/\\\x00")
}

func (d *vNextPublicationDirectory) Close() error {
	if d == nil || d.file == nil {
		return nil
	}
	return d.file.Close()
}

func (d *vNextPublicationDirectory) Sync() error {
	if d == nil || d.file == nil {
		return fs.ErrClosed
	}
	return d.file.Sync()
}

func (d *vNextPublicationDirectory) openDirectory(name, label string) (*vNextPublicationDirectory, error) {
	if name != "." && !vNextPublicationDirectNameValid(name) {
		return nil, fmt.Errorf("invalid %s name %q", label, name)
	}
	fd, err := unix.Openat(int(d.file.Fd()), name, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
	if err != nil {
		return nil, vNextPublicationOpenAtError(d, name, label, err)
	}
	return vNextPublicationDirectoryFromFD(fd, label)
}

func (d *vNextPublicationDirectory) ensureDirectory(name, label string) (*vNextPublicationDirectory, error) {
	child, err := d.openDirectory(name, label)
	if err == nil {
		return child, nil
	}
	if !errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}
	if err := unix.Mkdirat(int(d.file.Fd()), name, 0o755); err != nil && !errors.Is(err, unix.EEXIST) {
		return nil, fmt.Errorf("create %s: %w", label, err)
	}
	return d.openDirectory(name, label)
}

func (d *vNextPublicationDirectory) openParent(name, label string, create bool) (*vNextPublicationDirectory, string, error) {
	if !fs.ValidPath(name) {
		return nil, "", fmt.Errorf("invalid %s path %q", label, name)
	}
	parts := strings.Split(name, "/")
	parent, err := d.openDirectory(".", label)
	if err != nil {
		return nil, "", err
	}
	for _, component := range parts[:len(parts)-1] {
		var next *vNextPublicationDirectory
		if create {
			next, err = parent.ensureDirectory(component, label)
		} else {
			next, err = parent.openDirectory(component, label)
		}
		_ = parent.Close()
		if err != nil {
			return nil, "", err
		}
		parent = next
	}
	return parent, parts[len(parts)-1], nil
}

func (d *vNextPublicationDirectory) openFile(name, label string, flags int, perm fs.FileMode, createParents bool) (*os.File, error) {
	parent, base, err := d.openParent(name, label, createParents)
	if err != nil {
		return nil, err
	}
	defer parent.Close()
	fd, err := unix.Openat(int(parent.file.Fd()), base, flags|unix.O_CLOEXEC|unix.O_NOFOLLOW, uint32(perm.Perm()))
	if err != nil {
		return nil, vNextPublicationOpenAtError(parent, base, label, err)
	}
	file := os.NewFile(uintptr(fd), label)
	if file == nil {
		_ = unix.Close(fd)
		return nil, fmt.Errorf("open %s: invalid file descriptor", label)
	}
	return file, nil
}

func (d *vNextPublicationDirectory) openRegular(name, label string, flags int) (*os.File, error) {
	file, err := d.openFile(name, label, flags, 0, false)
	if err != nil {
		return nil, err
	}
	info, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return nil, fmt.Errorf("stat %s: %w", label, err)
	}
	if !info.Mode().IsRegular() {
		_ = file.Close()
		return nil, fmt.Errorf("%s is not a regular file", label)
	}
	return file, nil
}

func (d *vNextPublicationDirectory) readFile(name, label string) ([]byte, error) {
	file, err := d.openRegular(name, label, unix.O_RDONLY)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	payload, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", label, err)
	}
	return payload, nil
}

func (d *vNextPublicationDirectory) readDir() ([]os.DirEntry, error) {
	copy, err := d.openDirectory(".", d.label)
	if err != nil {
		return nil, err
	}
	defer copy.Close()
	entries, err := copy.file.ReadDir(-1)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", d.label, err)
	}
	sort.Slice(entries, func(left, right int) bool { return entries[left].Name() < entries[right].Name() })
	return entries, nil
}

func (d *vNextPublicationDirectory) lstat(name, label string) (unix.Stat_t, error) {
	if !vNextPublicationDirectNameValid(name) {
		return unix.Stat_t{}, fmt.Errorf("invalid %s name %q", label, name)
	}
	var stat unix.Stat_t
	if err := unix.Fstatat(int(d.file.Fd()), name, &stat, unix.AT_SYMLINK_NOFOLLOW); err != nil {
		return unix.Stat_t{}, fmt.Errorf("stat %s: %w", label, err)
	}
	return stat, nil
}

func vNextPublicationStatIsDir(stat unix.Stat_t) bool {
	return stat.Mode&unix.S_IFMT == unix.S_IFDIR
}

func vNextPublicationStatIsRegular(stat unix.Stat_t) bool {
	return stat.Mode&unix.S_IFMT == unix.S_IFREG
}

func vNextPublicationStatIsSymlink(stat unix.Stat_t) bool {
	return stat.Mode&unix.S_IFMT == unix.S_IFLNK
}

func (d *vNextPublicationDirectory) rename(oldName, newName string) error {
	if !vNextPublicationDirectNameValid(oldName) || !vNextPublicationDirectNameValid(newName) {
		return fmt.Errorf("invalid publication rename %q to %q", oldName, newName)
	}
	if err := unix.Renameat(int(d.file.Fd()), oldName, int(d.file.Fd()), newName); err != nil {
		return fmt.Errorf("rename publication member %q to %q: %w", oldName, newName, err)
	}
	return nil
}

func (d *vNextPublicationDirectory) removeRegular(name, label string) error {
	file, err := d.openRegular(name, label, unix.O_RDONLY)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close %s before removal: %w", label, err)
	}
	if err := unix.Unlinkat(int(d.file.Fd()), name, 0); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("remove %s: %w", label, err)
	}
	return nil
}

func (d *vNextPublicationDirectory) removeTree(name, label string) error {
	stat, err := d.lstat(name, label)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if vNextPublicationStatIsSymlink(stat) {
		return fmt.Errorf("%s is a symlink", label)
	}
	if vNextPublicationStatIsRegular(stat) {
		if err := unix.Unlinkat(int(d.file.Fd()), name, 0); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("remove %s: %w", label, err)
		}
		return nil
	}
	if !vNextPublicationStatIsDir(stat) {
		return fmt.Errorf("%s is not a regular file or directory", label)
	}
	child, err := d.openDirectory(name, label)
	if err != nil {
		return err
	}
	entries, err := child.readDir()
	if err == nil {
		for _, entry := range entries {
			if err = child.removeTree(entry.Name(), label+"/"+entry.Name()); err != nil {
				break
			}
		}
	}
	closeErr := child.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return fmt.Errorf("close %s: %w", label, closeErr)
	}
	if err := unix.Unlinkat(int(d.file.Fd()), name, unix.AT_REMOVEDIR); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("remove directory %s: %w", label, err)
	}
	return nil
}

type vNextPublicationDirectoryFS struct {
	root *vNextPublicationDirectory
}

func (f vNextPublicationDirectoryFS) Open(name string) (fs.File, error) {
	if f.root == nil {
		return nil, fs.ErrClosed
	}
	if name == "." {
		directory, err := f.root.openDirectory(".", "publication filesystem root")
		if err != nil {
			return nil, err
		}
		return directory.file, nil
	}
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrInvalid}
	}
	return f.root.openFile(name, "publication filesystem member", unix.O_RDONLY, 0, false)
}
