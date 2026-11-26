// Experimental paged omitted-set directory. The handle binds a trusted corpus
// snapshot and a particular omitted set. The source texts stay in the vault.
package main

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"iter"
	"os"
	"path/filepath"
	"sort"
)

const pageTurns = 4096
const descriptorBytes = 35 // encoding, payload length, SHA-256(payload)

const (
	emptyPage byte = iota
	fullPage
	includeList
	excludeList
	bitmapPage
)

type catalogHandle struct {
	Format   string `json:"format"`
	Snapshot string `json:"snapshot"`
	Index    string `json:"index"`
	Last     int    `json:"last_turn"`
	Omitted  int    `json:"omitted"`
}

type catalogCursor struct {
	Index string
	After int
}

type pageDescriptor struct {
	kind   byte
	size   int
	offset int64
	sum    [32]byte
}

func encodePage(mask []byte, width, ones int) (byte, []byte) {
	if ones == 0 {
		return emptyPage, nil
	}
	if ones == width {
		return fullPage, nil
	}
	if 2*min(ones, width-ones) < len(mask) {
		include := ones <= width-ones
		var payload []byte
		for i := 0; i < width; i++ {
			if (mask[i/8]&(1<<uint(i%8)) != 0) == include {
				payload = binary.LittleEndian.AppendUint16(payload, uint16(i))
			}
		}
		if include {
			return includeList, payload
		}
		return excludeList, payload
	}
	return bitmapPage, mask
}

func digestDirectory(snapshot [32]byte, last, total int, descriptors []byte) string {
	h := sha256.New()
	h.Write([]byte("cx-omitted/v1\x00"))
	h.Write(snapshot[:])
	h.Write(binary.LittleEndian.AppendUint64(nil, uint64(last)))
	h.Write(binary.LittleEndian.AppendUint64(nil, uint64(total)))
	h.Write(descriptors)
	return hex.EncodeToString(h.Sum(nil))
}

// buildCatalog accepts a membership oracle so the experiment can use exactly
// the same interface for real dropped IDs or synthetic distributions.
func buildCatalog(dir string, snapshot [32]byte, last int, omitted func(int) bool) (catalogHandle, error) {
	var zero catalogHandle
	if last < 0 || int64(last) > 1<<32-1 {
		return zero, fmt.Errorf("turn domain out of range")
	}
	var directory, payload []byte
	total := 0
	for base := 0; base < last; base += pageTurns {
		width := min(pageTurns, last-base)
		mask := make([]byte, (width+7)/8)
		ones := 0
		for i := 0; i < width; i++ {
			if omitted(base + i + 1) {
				mask[i/8] |= 1 << uint(i%8)
				ones++
			}
		}
		total += ones
		kind, data := encodePage(mask, width, ones)
		hash := sha256.Sum256(data)
		directory = append(directory, kind)
		directory = binary.LittleEndian.AppendUint16(directory, uint16(len(data)))
		directory = append(directory, hash[:]...)
		payload = append(payload, data...)
	}
	key := digestDirectory(snapshot, last, total, directory)
	path := filepath.Join(dir, key+".cxi")
	// Readers reopening a handle must always see a complete index, even if
	// another builder has just produced the same directory.
	f, err := os.CreateTemp(dir, ".catalog-*")
	if err != nil {
		return zero, err
	}
	defer os.Remove(f.Name())
	_, writeErr := f.Write(append(directory, payload...))
	closeErr := f.Close()
	if writeErr != nil {
		return zero, writeErr
	}
	if closeErr != nil {
		return zero, closeErr
	}
	if err := os.Rename(f.Name(), path); err != nil {
		return zero, err
	}
	return catalogHandle{"cx-omitted/v1", hex.EncodeToString(snapshot[:]), key, last, total}, nil
}

type catalog struct {
	file  *os.File
	h     catalogHandle
	pages []pageDescriptor
	reads int // explicit ReadAt calls, excluding stat/open and the OS cache
}

func openCatalog(dir string, h catalogHandle, current [32]byte) (*catalog, error) {
	key, err := hex.DecodeString(h.Index)
	if err != nil || len(key) != sha256.Size || hex.EncodeToString(key) != h.Index ||
		h.Format != "cx-omitted/v1" || h.Snapshot != hex.EncodeToString(current[:]) ||
		h.Last < 0 || int64(h.Last) > 1<<32-1 || h.Omitted < 0 || h.Omitted > h.Last {
		return nil, fmt.Errorf("invalid handle or stale corpus snapshot")
	}
	file, err := os.Open(filepath.Join(dir, h.Index+".cxi"))
	if err != nil {
		return nil, err
	}
	ok := false
	defer func() {
		if !ok {
			file.Close()
		}
	}()
	count := (h.Last + pageTurns - 1) / pageTurns
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if info.Size() < int64(count*descriptorBytes) {
		return nil, io.ErrUnexpectedEOF
	}
	directory := make([]byte, count*descriptorBytes)
	c := &catalog{file: file, h: h}
	if len(directory) > 0 {
		c.reads++
		if _, err := file.ReadAt(directory, 0); err != nil {
			return nil, err
		}
	}
	if digestDirectory(current, h.Last, h.Omitted, directory) != h.Index {
		return nil, fmt.Errorf("directory digest mismatch")
	}
	offset := int64(len(directory))
	for i := 0; i < count; i++ {
		d := directory[i*descriptorBytes : (i+1)*descriptorBytes]
		p := pageDescriptor{kind: d[0], size: int(binary.LittleEndian.Uint16(d[1:])), offset: offset}
		copy(p.sum[:], d[3:])
		width := min(pageTurns, h.Last-i*pageTurns)
		switch p.kind {
		case emptyPage, fullPage:
			if p.size != 0 {
				return nil, fmt.Errorf("invalid marker length")
			}
		case includeList, excludeList:
			if p.size%2 != 0 || p.size >= (width+7)/8 {
				return nil, fmt.Errorf("invalid list length")
			}
		case bitmapPage:
			if p.size != (width+7)/8 {
				return nil, fmt.Errorf("invalid bitmap length")
			}
		default:
			return nil, fmt.Errorf("unknown page encoding")
		}
		c.pages = append(c.pages, p)
		offset += int64(p.size)
	}
	if offset != info.Size() {
		return nil, fmt.Errorf("index file length mismatch")
	}
	ok = true
	return c, nil
}

func (c *catalog) page(number int) ([]byte, error) {
	p := c.pages[number]
	if p.size == 0 {
		return nil, nil
	}
	data := make([]byte, p.size)
	c.reads++
	if _, err := c.file.ReadAt(data, p.offset); err != nil {
		return nil, err
	}
	if sha256.Sum256(data) != p.sum {
		return nil, fmt.Errorf("corrupt page %d", number)
	}
	return data, nil
}

func member(kind byte, data []byte, bit int) bool {
	switch kind {
	case fullPage:
		return true
	case emptyPage:
		return false
	case bitmapPage:
		return data[bit/8]&(1<<uint(bit%8)) != 0
	default:
		at := sort.Search(len(data)/2, func(i int) bool { return int(binary.LittleEndian.Uint16(data[2*i:])) >= bit })
		found := at < len(data)/2 && int(binary.LittleEndian.Uint16(data[2*at:])) == bit
		return found == (kind == includeList)
	}
}

// Walk sparse postings directly. For a sparse complement, advance through
// the sorted exceptions once as we emit the consecutive runs between them.
func pageMembers(kind byte, data []byte, start, width int) iter.Seq[int] {
	return func(yield func(int) bool) {
		if kind == includeList || kind == excludeList {
			posting := func(i int) int { return int(binary.LittleEndian.Uint16(data[2*i:])) }
			count := len(data) / 2
			i := sort.Search(count, func(i int) bool { return posting(i) >= start })
			if kind == includeList {
				for ; i < count; i++ {
					if !yield(posting(i)) {
						return
					}
				}
				return
			}
			for pos := start; pos < width; pos++ {
				if i < count && pos == posting(i) {
					i++
					continue
				}
				if !yield(pos) {
					return
				}
			}
			return
		}
		for i := start; i < width; i++ {
			if member(kind, data, i) && !yield(i) {
				return
			}
		}
	}
}

func (c *catalog) contains(id int) (bool, error) {
	if id < 1 || id > c.h.Last {
		return false, fmt.Errorf("turn outside snapshot")
	}
	number := (id - 1) / pageTurns
	b, err := c.page(number)
	if err != nil {
		return false, err
	}
	return member(c.pages[number].kind, b, (id-1)%pageTurns), nil
}

// enumerate returns at most limit ascending IDs. A non-nil cursor carries the
// last returned member and is scoped to this handle. A full final page may
// require one terminal empty request. Cursors do not establish reader access
// control; the caller must already have access to the bound workspace.
func (c *catalog) enumerate(cursor *catalogCursor, limit int) ([]int, *catalogCursor, error) {
	if limit < 1 || limit > pageTurns {
		return nil, nil, fmt.Errorf("invalid page size")
	}
	after := 0
	if cursor != nil {
		if cursor.Index != c.h.Index {
			return nil, nil, fmt.Errorf("cursor from different omitted set")
		}
		ok, err := c.contains(cursor.After)
		if err != nil {
			return nil, nil, err
		}
		if !ok {
			return nil, nil, fmt.Errorf("invalid cursor position")
		}
		after = cursor.After
	}
	var ids []int
	for num := after / pageTurns; num < len(c.pages); num++ {
		kind := c.pages[num].kind
		if kind == emptyPage {
			continue
		}
		data, err := c.page(num)
		if err != nil {
			return nil, nil, err
		}
		base := num * pageTurns
		for i := range pageMembers(kind, data, max(0, after-base), min(pageTurns, c.h.Last-base)) {
			id := base + i + 1
			ids = append(ids, id)
			if len(ids) == limit && id < c.h.Last {
				return ids, &catalogCursor{c.h.Index, id}, nil
			}
		}
	}
	return ids, nil, nil
}

// Smoke experiment only. Use simple deterministic membership
// to keep the disjoint case independent of any answer or retrieval algorithm.
func measureCatalog(dir string, n int) error {
	snapshot := sha256.Sum256([]byte(fmt.Sprintf("synthetic-snapshot/v1/%d", n)))
	for _, scenario := range []struct {
		name    string
		omitted func(int) bool
	}{
		{"dense", func(id int) bool { return true }},
		{"sparse", func(id int) bool { return id%1031 == 0 }},
		{"fragmented", func(id int) bool { return id%2 == 1 }},
	} {
		h, err := buildCatalog(dir, snapshot, n, scenario.omitted)
		if err != nil {
			return err
		}
		c, err := openCatalog(dir, h, snapshot)
		if err != nil {
			return err
		}
		info, err := c.file.Stat()
		if err != nil {
			c.file.Close()
			return err
		}
		// Validate each returned ID and all missing intervals. Streaming keeps
		// the output bounded even when every turn is omitted.
		var cursor *catalogCursor
		seen, expected := 0, 1
		for {
			ids, next, err := c.enumerate(cursor, 1024)
			if err != nil {
				c.file.Close()
				return err
			}
			for _, got := range ids {
				for expected <= n && !scenario.omitted(expected) {
					expected++
				}
				if got != expected {
					c.file.Close()
					return fmt.Errorf("%s: omitted ID %d, expected %d", scenario.name, got, expected)
				}
				seen++
				expected++
			}
			if next == nil {
				break
			}
			cursor = next
		}
		for expected <= n && !scenario.omitted(expected) {
			expected++
		}
		if expected != n+1 || seen != h.Omitted {
			c.file.Close()
			return fmt.Errorf("%s incomplete enumeration", scenario.name)
		}
		before := c.reads
		found, err := c.contains(n/2 + 1)
		c.file.Close()
		if err != nil {
			return err
		}
		if found != scenario.omitted(n/2+1) {
			return fmt.Errorf("%s: lookup mismatch", scenario.name)
		}
		encoded, err := json.Marshal(h)
		if err != nil {
			return err
		}
		fmt.Printf("catalog %s: %d of %d omitted, %d packet JSON bytes, %d external bytes, %d directory ReadAt on open, %d page ReadAt for lookup; enumeration verified\n", scenario.name, seen, n, len(encoded), info.Size(), 1, c.reads-before)
	}
	return nil
}
