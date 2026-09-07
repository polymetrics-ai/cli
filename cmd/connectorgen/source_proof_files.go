package main

import (
	"context"
	"encoding/hex"
	"errors"
	"io/fs"
	"os"
	"path"
	"strings"
)

type sourceProofReadEvent struct {
	Path, Phase    string
	Bytes, Charged int64
	Success        bool
	Before, After  os.FileInfo
}
type sourceProofReadStats struct {
	UniqueFiles, ReadCalls, FinalReads int
	UniqueBytes, PhysicalCharged       int64
}

type sourceProofFile struct {
	hash, code string
	size       int64
	info       os.FileInfo
	result     sourceProofResult
	parsed     bool
}

// sourceProofFileCache owns only bounded file observations and parsed results.
// It has no source keys, lanes, targets, review catalogue or acceptance state.
type sourceProofFileCache struct {
	ctx       context.Context
	root      *os.Root
	limits    sourceProofFileLimits
	afterRead func(sourceProofReadEvent)
	files     map[string]*sourceProofFile
	order     []string
	stats     sourceProofReadStats
}
type sourceProofFileLimits struct {
	UniqueBytes int64
	Files       int
}

func sourceProofFileInfo(root *os.Root, name string) (os.FileInfo, string) {
	if !sourceProofSafePath(name) {
		return nil, "proof_input_invalid"
	}
	prefix := ""
	var info os.FileInfo
	for _, part := range strings.Split(name, "/") {
		prefix = path.Join(prefix, part)
		var err error
		info, err = root.Lstat(prefix)
		if errors.Is(err, fs.ErrNotExist) {
			return nil, "missing"
		}
		if err != nil || info.Mode()&os.ModeSymlink != 0 {
			return nil, "proof_input_invalid"
		}
	}
	if !info.Mode().IsRegular() {
		return nil, "proof_input_invalid"
	}
	return info, ""
}

// A successful call refunds proven unused reservation. A failed read retains
// limit+lookahead as its conservative physical charge, even when bytes are nil.
// Final content revalidation spends the same physical allowance, never free I/O.
func (c *sourceProofFileCache) read(name, phase string, cap int64) ([]byte, *sourceProofFile) {
	f := &sourceProofFile{}
	if c.ctx.Err() != nil {
		f.code = "proof_cancelled"
		return nil, f
	}
	before, code := sourceProofFileInfo(c.root, name)
	if code != "" {
		f.code = code
		return nil, f
	}
	limit := cap
	if phase == "initial" {
		limit = min(limit, c.limits.UniqueBytes-c.stats.UniqueBytes)
	}
	physicalLimit := 2*c.limits.UniqueBytes + 2*int64(c.limits.Files)
	limit = min(limit, physicalLimit-c.stats.PhysicalCharged-1)
	if limit < 0 || limit == 0 && before.Size() > 0 {
		f.code = "proof_read_budget_exceeded"
		return nil, f
	}
	c.stats.PhysicalCharged += limit + 1
	if phase == "initial" {
		c.stats.UniqueBytes += limit
	} else {
		c.stats.FinalReads++
	}
	c.stats.ReadCalls++
	raw, err := readSourceInput(c.root, name, limit)
	after, afterCode := sourceProofFileInfo(c.root, name)
	charged := limit + 1
	if err == nil {
		charged = int64(len(raw))
		c.stats.PhysicalCharged -= limit + 1 - charged
		if phase == "initial" {
			c.stats.UniqueBytes -= limit - charged
		}
	}
	if err != nil {
		f.code = "proof_input_invalid"
	} else if afterCode != "" || !os.SameFile(before, after) {
		f.code = "proof_file_changed"
	} else {
		f.hash = sourceBytesHash(raw)
		f.size = int64(len(raw))
		f.info = after
	}
	if c.afterRead != nil {
		c.afterRead(sourceProofReadEvent{Path: name, Phase: phase, Bytes: int64(len(raw)), Charged: charged, Success: err == nil, Before: before, After: after})
	}
	return raw, f
}

func (c *sourceProofFileCache) get(name string, cap int64, receipt bool) ([]byte, *sourceProofFile) {
	if f, ok := c.files[name]; ok {
		// Every successful unique path is identity-checked and content-rehashed
		// at final return. Hits do not multiply I/O by the number of claims.
		return nil, f
	}
	if len(c.files) >= c.limits.Files {
		return nil, &sourceProofFile{code: "proof_file_limit_exceeded"}
	}
	raw, f := c.read(name, "initial", cap)
	c.files[name] = f
	c.order = append(c.order, name)
	c.stats.UniqueFiles++
	if receipt && f.code == "" {
		f.result = parseSourceProofResult(raw)
		f.parsed = true
	}
	return raw, f
}

func (c *sourceProofFileCache) finalize() {
	for _, name := range c.order {
		f := c.files[name]
		if f.code != "" {
			continue
		}
		if c.ctx.Err() != nil {
			f.code = "proof_cancelled"
			continue
		}
		info, code := sourceProofFileInfo(c.root, name)
		if code != "" || !os.SameFile(f.info, info) {
			f.code = "proof_file_changed"
			continue
		}
		_, current := c.read(name, "final", f.size)
		if current.code != "" || current.hash != f.hash || !os.SameFile(current.info, f.info) {
			if current.code == "proof_cancelled" {
				f.code = current.code
			} else {
				f.code = "proof_file_changed"
			}
		}
	}
}

func sourceProofSafePath(p string) bool {
	return validSourceID(p) && p != "." && path.Clean(p) == p && !path.IsAbs(p) && !strings.HasPrefix(p, "../") && !strings.Contains(p, "\\")
}

func sourceProofDigest(s string) bool {
	b, err := hex.DecodeString(s)
	return err == nil && len(b) == 32 && strings.ToLower(s) == s
}
