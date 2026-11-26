package packet

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
)

// External paged omitted-set catalog (promoted from
// docs/experiments/packet_scale/catalog.go). When an archive grows past
// thousands of omitted turns, naming every block inline in brief.md explodes
// the token budget (264k tokens at 1M turns). Moving the omitted membership
// directory into content-addressed 4,096-turn pages on disk shrinks the
// in-packet representation to a ~215-byte snapshot-pinned JSON handle while
// retaining exact, verified membership lookup and cursor pagination.

const (
	omittedPageTurns       = 4096
	omittedDescriptorBytes = 35 // encoding(1) + payloadLen(2) + sha256(32)
)

const (
	omittedEmptyPage byte = iota
	omittedFullPage
	omittedIncludeList
	omittedExcludeList
	omittedBitmapPage
)

// OmittedHandle is the compact snapshot-bound pointer placed in brief.md when
// omitted turns are offloaded to an external paged catalog file.
type OmittedHandle struct {
	Format   string `json:"format"`
	Snapshot string `json:"snapshot"`
	Index    string `json:"index"`
	Last     int    `json:"last_turn"`
	Omitted  int    `json:"omitted"`
}

// JSON renders the handle as a single compact line for brief.md.
func (h OmittedHandle) JSON() string {
	b, _ := json.Marshal(h)
	return string(b)
}

// OmittedCursor tracks resumable pagination over an OmittedCatalog.
type OmittedCursor struct {
	Index string `json:"index"`
	After int    `json:"after"`
}

type omittedPageDesc struct {
	kind   byte
	size   int
	offset int64
	sum    [32]byte
}

func encodeOmittedPage(mask []byte, width, ones int) (byte, []byte) {
	if ones == 0 {
		return omittedEmptyPage, nil
	}
	if ones == width {
		return omittedFullPage, nil
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
			return omittedIncludeList, payload
		}
		return omittedExcludeList, payload
	}
	return omittedBitmapPage, mask
}

func digestOmittedDirectory(snapshot []byte, last, total int, descriptors []byte) string {
	h := sha256.New()
	h.Write([]byte("cx-omitted/v1\x00"))
	h.Write(snapshot)
	h.Write(binary.LittleEndian.AppendUint64(nil, uint64(last)))
	h.Write(binary.LittleEndian.AppendUint64(nil, uint64(total)))
	h.Write(descriptors)
	return hex.EncodeToString(h.Sum(nil))
}

// BuildOmittedCatalog writes an immutable, content-addressed paged omission
// index under dir and returns its compact handle pinned to snapshotRoot.
func BuildOmittedCatalog(dir, snapshotRoot string, last int, omitted func(int) bool) (OmittedHandle, error) {
	var zero OmittedHandle
	if last < 0 || int64(last) > 1<<32-1 {
		return zero, fmt.Errorf("packet: turn domain %d out of range", last)
	}
	snapBytes, err := hex.DecodeString(snapshotRoot)
	if err != nil || len(snapBytes) != sha256.Size {
		sum := sha256.Sum256([]byte(snapshotRoot))
		snapBytes = sum[:]
		snapshotRoot = hex.EncodeToString(snapBytes)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return zero, err
	}
	var directory, payload []byte
	total := 0
	for base := 0; base < last; base += omittedPageTurns {
		width := min(omittedPageTurns, last-base)
		mask := make([]byte, (width+7)/8)
		ones := 0
		for i := 0; i < width; i++ {
			if omitted(base + i + 1) {
				mask[i/8] |= 1 << uint(i%8)
				ones++
			}
		}
		total += ones
		kind, data := encodeOmittedPage(mask, width, ones)
		hash := sha256.Sum256(data)
		directory = append(directory, kind)
		directory = binary.LittleEndian.AppendUint16(directory, uint16(len(data)))
		directory = append(directory, hash[:]...)
		payload = append(payload, data...)
	}
	key := digestOmittedDirectory(snapBytes, last, total, directory)
	path := filepath.Join(dir, key+".cxi")
	if _, err := os.Stat(path); err == nil {
		return OmittedHandle{"cx-omitted/v1", snapshotRoot, key, last, total}, nil
	}
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
	return OmittedHandle{"cx-omitted/v1", snapshotRoot, key, last, total}, nil
}

// OmittedCatalog is an open, digest-verified external omitted-turn directory.
type OmittedCatalog struct {
	file  *os.File
	h     OmittedHandle
	pages []omittedPageDesc
}

// OpenOmittedCatalog verifies h against currentSnapshotRoot and opens the
// paged catalog file for O(1) point lookups and resumable enumeration.
func OpenOmittedCatalog(dir string, h OmittedHandle, currentSnapshotRoot string) (*OmittedCatalog, error) {
	snapBytes, err := hex.DecodeString(currentSnapshotRoot)
	if err != nil || len(snapBytes) != sha256.Size {
		sum := sha256.Sum256([]byte(currentSnapshotRoot))
		snapBytes = sum[:]
		currentSnapshotRoot = hex.EncodeToString(snapBytes)
	}
	key, err := hex.DecodeString(h.Index)
	if err != nil || len(key) != sha256.Size || hex.EncodeToString(key) != h.Index ||
		h.Format != "cx-omitted/v1" || h.Snapshot != currentSnapshotRoot ||
		h.Last < 0 || int64(h.Last) > 1<<32-1 || h.Omitted < 0 || h.Omitted > h.Last {
		return nil, fmt.Errorf("packet: invalid handle or stale snapshot")
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
	count := (h.Last + omittedPageTurns - 1) / omittedPageTurns
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if info.Size() < int64(count*omittedDescriptorBytes) {
		return nil, io.ErrUnexpectedEOF
	}
	directory := make([]byte, count*omittedDescriptorBytes)
	if len(directory) > 0 {
		if _, err := file.ReadAt(directory, 0); err != nil {
			return nil, err
		}
	}
	if digestOmittedDirectory(snapBytes, h.Last, h.Omitted, directory) != h.Index {
		return nil, fmt.Errorf("packet: omitted catalog directory digest mismatch")
	}
	c := &OmittedCatalog{file: file, h: h}
	offset := int64(len(directory))
	for i := 0; i < count; i++ {
		d := directory[i*omittedDescriptorBytes : (i+1)*omittedDescriptorBytes]
		p := omittedPageDesc{kind: d[0], size: int(binary.LittleEndian.Uint16(d[1:])), offset: offset}
		copy(p.sum[:], d[3:])
		c.pages = append(c.pages, p)
		offset += int64(p.size)
	}
	if offset != info.Size() {
		return nil, fmt.Errorf("packet: omitted catalog length mismatch")
	}
	ok = true
	return c, nil
}

// Close closes the underlying catalog file.
func (c *OmittedCatalog) Close() error { return c.file.Close() }

func (c *OmittedCatalog) page(number int) ([]byte, error) {
	p := c.pages[number]
	if p.size == 0 {
		return nil, nil
	}
	data := make([]byte, p.size)
	if _, err := c.file.ReadAt(data, p.offset); err != nil {
		return nil, err
	}
	if sha256.Sum256(data) != p.sum {
		return nil, fmt.Errorf("packet: corrupt omitted catalog page %d", number)
	}
	return data, nil
}

func omittedMember(kind byte, data []byte, bit int) bool {
	switch kind {
	case omittedFullPage:
		return true
	case omittedEmptyPage:
		return false
	case omittedBitmapPage:
		return data[bit/8]&(1<<uint(bit%8)) != 0
	default:
		at := sort.Search(len(data)/2, func(i int) bool { return int(binary.LittleEndian.Uint16(data[2*i:])) >= bit })
		found := at < len(data)/2 && int(binary.LittleEndian.Uint16(data[2*at:])) == bit
		return found == (kind == omittedIncludeList)
	}
}

// Contains reports whether turn seq was omitted from the snapshot.
func (c *OmittedCatalog) Contains(seq int) (bool, error) {
	if seq < 1 || seq > c.h.Last {
		return false, fmt.Errorf("packet: turn %d outside snapshot 1..%d", seq, c.h.Last)
	}
	num := (seq - 1) / omittedPageTurns
	data, err := c.page(num)
	if err != nil {
		return false, err
	}
	return omittedMember(c.pages[num].kind, data, (seq-1)%omittedPageTurns), nil
}

// Enumerate returns up to limit ascending omitted turn sequence numbers after
// cursor, along with the next cursor (nil when exhausted).
func (c *OmittedCatalog) Enumerate(cursor *OmittedCursor, limit int) ([]int, *OmittedCursor, error) {
	if limit < 1 || limit > omittedPageTurns {
		return nil, nil, fmt.Errorf("packet: invalid page limit %d", limit)
	}
	after := 0
	if cursor != nil {
		if cursor.Index != c.h.Index {
			return nil, nil, fmt.Errorf("packet: cursor belongs to another omitted set")
		}
		after = cursor.After
	}
	var ids []int
	for num := after / omittedPageTurns; num < len(c.pages); num++ {
		kind := c.pages[num].kind
		if kind == omittedEmptyPage {
			continue
		}
		data, err := c.page(num)
		if err != nil {
			return nil, nil, err
		}
		base := num * omittedPageTurns
		width := min(omittedPageTurns, c.h.Last-base)
		start := max(0, after-base)
		for bit := start; bit < width; bit++ {
			if omittedMember(kind, data, bit) {
				seq := base + bit + 1
				ids = append(ids, seq)
				if len(ids) == limit && seq < c.h.Last {
					return ids, &OmittedCursor{Index: c.h.Index, After: seq}, nil
				}
			}
		}
	}
	return ids, nil, nil
}
