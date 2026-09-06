//go:build linux

package main

import (
	"errors"

	"golang.org/x/sys/unix"
)

func vNextPublicationRenameNoReplace(oldDirectoryFD int, oldName string, newDirectoryFD int, newName string) error {
	return vNextPublicationNoReplaceRenameError(unix.Renameat2(oldDirectoryFD, oldName, newDirectoryFD, newName, unix.RENAME_NOREPLACE))
}

func vNextPublicationNoReplaceUnsupported(err error) bool {
	return errors.Is(err, unix.ENOSYS) ||
		errors.Is(err, unix.EINVAL) ||
		errors.Is(err, unix.ENOTSUP) ||
		errors.Is(err, unix.EOPNOTSUPP) ||
		errors.Is(err, unix.EXDEV)
}
