//go:build darwin

package main

import (
	"errors"

	"golang.org/x/sys/unix"
)

func vNextPublicationRenameNoReplace(oldDirectoryFD int, oldName string, newDirectoryFD int, newName string) error {
	return vNextPublicationNoReplaceRenameError(unix.RenameatxNp(oldDirectoryFD, oldName, newDirectoryFD, newName, unix.RENAME_EXCL))
}

func vNextPublicationNoReplaceUnsupported(err error) bool {
	return errors.Is(err, unix.ENOSYS) ||
		errors.Is(err, unix.EINVAL) ||
		errors.Is(err, unix.ENOTSUP) ||
		errors.Is(err, unix.EOPNOTSUPP) ||
		errors.Is(err, unix.EXDEV)
}
