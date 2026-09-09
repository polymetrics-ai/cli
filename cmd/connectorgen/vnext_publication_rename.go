package main

import (
	"errors"
	"fmt"
	"io/fs"
)

var errVNextPublicationUnsupportedNoReplace = errors.New("publication filesystem does not support no-replace rename")

// vNextPublicationUnsupportedNoReplaceError preserves the syscall cause while
// giving callers one stable failure class. Publication must retain the prior
// authority and never substitute a clobbering rename when this is returned.
type vNextPublicationUnsupportedNoReplaceError struct {
	cause error
}

func (err *vNextPublicationUnsupportedNoReplaceError) Error() string {
	return fmt.Sprintf("%s: %v", errVNextPublicationUnsupportedNoReplace, err.cause)
}

func (err *vNextPublicationUnsupportedNoReplaceError) Unwrap() []error {
	return []error{errVNextPublicationUnsupportedNoReplace, err.cause}
}

func vNextPublicationNoReplaceRenameError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, fs.ErrExist) {
		return fmt.Errorf("publication no-replace destination exists: %w", err)
	}
	if vNextPublicationNoReplaceUnsupported(err) {
		return &vNextPublicationUnsupportedNoReplaceError{cause: err}
	}
	return fmt.Errorf("publication no-replace rename: %w", err)
}
