package main

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/sys/unix"
)

func TestCP11CompoundParentAndRawCloseOwners(t *testing.T) {
	for _, created := range []bool{false, true} {
		name := "failed-open"
		if created {
			name = "created-open"
		}
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			parentFailure := errors.New("parent actual Close completion")
			rawFailure := errors.New("raw actual Close completion")
			parents, rawCloses := 0, 0
			var parentInstance *os.File
			var createdIdentity vNextPublicationIdentity
			directory, err := vNextPublicationOpenDirectoryWithCloseForTest(root, "retained owner root", func(file *os.File, label string) error {
				if label == "owned open" {
					parents++
					parentInstance = file
					if created {
						var stat unix.Stat_t
						if err := unix.Fstatat(int(file.Fd()), "record", &stat, unix.AT_SYMLINK_NOFOLLOW); err != nil {
							t.Fatal(err)
						}
						createdIdentity = vNextPublicationIdentityFromStat(stat)
						if stat.Size != 0 || createdIdentity.mode != unix.S_IFREG {
							t.Fatal("missing actual exclusive creation")
						}
					}
					return errors.Join(file.Close(), parentFailure)
				}
				return file.Close()
			})
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				if err := directory.Close(); err != nil {
					t.Errorf("close test-owned directory: %v", err)
				}
			}()
			vNextPublicationCloseOpenedFileAfterParentCloseForTest = func(fd int, label string) error {
				rawCloses++
				var stat unix.Stat_t
				if err := unix.Fstat(fd, &stat); err != nil {
					t.Fatal(err)
				}
				if label != "owned open" || vNextPublicationIdentityFromStat(stat) != createdIdentity {
					t.Fatal("raw Close adopted another object")
				}
				err := unix.Close(fd)
				if probe := unix.Fstat(fd, &stat); !errors.Is(probe, unix.EBADF) {
					t.Fatalf("raw descriptor remains open: %v", probe)
				}
				return errors.Join(err, rawFailure)
			}
			t.Cleanup(func() { vNextPublicationCloseOpenedFileAfterParentCloseForTest = nil })
			flags := unix.O_RDONLY
			if created {
				flags = unix.O_CREAT | unix.O_EXCL | unix.O_WRONLY
			}
			file, err := directory.openFile("record", "owned open", flags, 0o600, false)
			if file != nil || parents != 1 || !errors.Is(err, parentFailure) {
				t.Fatalf("parent/file ownership file=%v parents=%d err=%v", file, parents, err)
			}
			if _, err := parentInstance.Stat(); !errors.Is(err, fs.ErrClosed) {
				t.Fatalf("parent descriptor remains open: %v", err)
			}
			if created {
				if rawCloses != 1 || !errors.Is(err, rawFailure) {
					t.Fatalf("raw owner count=%d err=%v", rawCloses, err)
				}
				observed, err := vNextPublicationObserveExpectedTree(filepath.Join(root, "record"))
				if err != nil {
					t.Fatal(err)
				}
				if observed["."].identity != createdIdentity || len(observed["."].payload) != 0 {
					t.Fatal("closed raw effect was erased or replaced")
				}
			} else if rawCloses != 0 || !errors.Is(err, fs.ErrNotExist) {
				t.Fatalf("failed Open owns no raw descriptor: closes=%d err=%v", rawCloses, err)
			}
		})
	}
}

func TestCP11CompoundReadDirCopyCloseOwner(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "known"), []byte("retained"), 0o600); err != nil {
		t.Fatal(err)
	}
	completion := errors.New("readDir duplicate actual Close completion")
	var copyInstance *os.File
	closes := 0
	var directory *vNextPublicationDirectory
	directory, err := vNextPublicationOpenDirectoryWithCloseForTest(root, "readDir owner", func(file *os.File, _ string) error {
		if file == directory.file {
			return file.Close()
		}
		closes++
		copyInstance = file
		if identity, err := vNextPublicationIdentityFromFile(file, "readDir copy"); err != nil || identity.mode != unix.S_IFDIR {
			t.Fatalf("actual readDir copy identity=%#v err=%v", identity, err)
		}
		return errors.Join(file.Close(), completion)
	})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := directory.Close(); err != nil {
			t.Errorf("close test-owned directory: %v", err)
		}
	}()
	expected, err := vNextPublicationObserveExpectedTree(root)
	if err != nil {
		t.Fatal(err)
	}
	entries, err := directory.readDir()
	if entries != nil || closes != 1 || !errors.Is(err, completion) {
		t.Fatalf("readDir consumer entries=%v closes=%d err=%v", entries, closes, err)
	}
	if _, err := copyInstance.Stat(); !errors.Is(err, fs.ErrClosed) {
		t.Fatalf("readDir duplicate remains open: %v", err)
	}
	if _, err := directory.file.Stat(); err != nil {
		t.Fatalf("readDir closed borrowed root: %v", err)
	}
	vNextPublicationAssertExpectedTree(t, root, expected)
}
