package vault

import (
	"bytes"
	"compress/flate"
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

// Seekable, independently verified 4 KiB block-compressed body pack
// (graduated from reports/2026-11-20-large-memory/fixed-window-results.md).
//
// Layout:
//   [0..8)   magic "CXPK0001"
//   [8..12)  block_bytes (uint32 LE, default 4096)
//   [12..20) dir_offset  (uint64 LE)
//   [20..24) dir_len     (uint32 LE)
//   [24..56) dir_sha256  (32 bytes)
//   [56..dir_offset)     concatenated raw DEFLATE blocks
//   [dir_offset..)       UTF-8 JSON packDirectory
//
// Every logical EvidenceAnchor cites (turn_id, body_sha, byte_start, byte_end),
// so repacking or changing block_bytes never breaks any citation or receipt.
// Multi-citation reads coalesce blocks so multiple citations sharing a 4 KiB
// block decompress and SHA-verify that block at most once.

const (
	packMagic        = "CXPK0001"
	packHeaderBytes  = 56
	DefaultPackBlock = 4096
)

type packBlockMeta struct {
	Offset  int64  `json:"offset"`
	CompLen int    `json:"comp_len"`
	RawLen  int    `json:"raw_len"`
	RawSHA  string `json:"raw_sha"`
}

type packTurnMeta struct {
	ID        string `json:"id"`
	Seq       int    `json:"seq"`
	BodySHA   string `json:"body_sha"`
	RawOffset int64  `json:"raw_offset"`
	RawLen    int    `json:"raw_len"`
}

type packDirectory struct {
	BlockBytes int             `json:"block_bytes"`
	TotalRaw   int64           `json:"total_raw"`
	Blocks     []packBlockMeta `json:"blocks"`
	Turns      []packTurnMeta  `json:"turns"`
}

// PackStats reports compression and block metrics for a built BodyPack.
type PackStats struct {
	Path        string  `json:"path"`
	Turns       int     `json:"turns"`
	RawBytes    int64   `json:"raw_bytes"`
	PackedBytes int64   `json:"packed_bytes"`
	BlockCount  int     `json:"block_count"`
	BlockBytes  int     `json:"block_bytes"`
	Ratio       float64 `json:"ratio"` // packed / raw
}

// VerifiedSpan is one byte-exact, SHA-verified text excerpt read from a BodyPack.
type VerifiedSpan struct {
	Anchor EvidenceAnchor `json:"anchor"`
	Text   string         `json:"text"`
}

// PackReadStats measures physical I/O and decompression amplification for a
// coalesced ReadSpans batch.
type PackReadStats struct {
	SpansRequested    int `json:"spans_requested"`
	BlocksDecoded     int `json:"blocks_decoded"`
	PhysicalBytesRead int `json:"physical_bytes_read"`
	DecompressedBytes int `json:"decompressed_bytes"`
	BytesReturned     int `json:"bytes_returned"`
}

// BodyPack is an open, digest-verified seekable compressed archive of turn bodies.
type BodyPack struct {
	file     *os.File
	dir      packDirectory
	turnByID map[string]packTurnMeta
}

// BuildBodyPack packs all immutable turn bodies in v into independently
// decodable, SHA-256 verified DEFLATE blocks of blockBytes (default 4096) at
// packPath. Markdown files in turns/ remain untouched.
func BuildBodyPack(v *Vault, packPath string, blockBytes int) (PackStats, error) {
	var st PackStats
	if blockBytes <= 0 {
		blockBytes = DefaultPackBlock
	}
	turns, err := v.Turns()
	if err != nil {
		return st, err
	}
	if err := os.MkdirAll(filepath.Dir(packPath), 0o755); err != nil {
		return st, err
	}
	tmp := packPath + ".tmp"
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0o644)
	if err != nil {
		return st, err
	}
	defer func() {
		if f != nil {
			f.Close()
			_ = os.Remove(tmp)
		}
	}()
	// Reserve 56-byte header.
	var hdr [packHeaderBytes]byte
	if _, err := f.Write(hdr[:]); err != nil {
		return st, err
	}

	dir := packDirectory{BlockBytes: blockBytes}
	var rawBuf bytes.Buffer
	var compBuf bytes.Buffer
	var fileOff int64 = packHeaderBytes
	var rawOff int64

	flushBlock := func(raw []byte) error {
		compBuf.Reset()
		fw, err := flate.NewWriter(&compBuf, flate.DefaultCompression)
		if err != nil {
			return err
		}
		if _, err := fw.Write(raw); err != nil {
			return err
		}
		if err := fw.Close(); err != nil {
			return err
		}
		comp := compBuf.Bytes()
		if _, err := f.Write(comp); err != nil {
			return err
		}
		sum := sha256.Sum256(raw)
		dir.Blocks = append(dir.Blocks, packBlockMeta{
			Offset:  fileOff,
			CompLen: len(comp),
			RawLen:  len(raw),
			RawSHA:  hex.EncodeToString(sum[:]),
		})
		fileOff += int64(len(comp))
		return nil
	}

	for _, t := range turns {
		bodyBytes := []byte(t.Body)
		sha := t.BodySHA
		if sha == "" {
			sha = SHA(t.Body)
		}
		dir.Turns = append(dir.Turns, packTurnMeta{
			ID:        t.ID,
			Seq:       t.Seq,
			BodySHA:   sha,
			RawOffset: rawOff,
			RawLen:    len(bodyBytes),
		})
		rawOff += int64(len(bodyBytes))
		rawBuf.Write(bodyBytes)
		for rawBuf.Len() >= blockBytes {
			chunk := rawBuf.Next(blockBytes)
			if err := flushBlock(chunk); err != nil {
				return st, err
			}
		}
	}
	if rawBuf.Len() > 0 {
		if err := flushBlock(rawBuf.Bytes()); err != nil {
			return st, err
		}
	}
	dir.TotalRaw = rawOff

	dirBytes, err := json.Marshal(dir)
	if err != nil {
		return st, err
	}
	dirOffset := fileOff
	if _, err := f.Write(dirBytes); err != nil {
		return st, err
	}
	dirSum := sha256.Sum256(dirBytes)
	copy(hdr[0:8], packMagic)
	binary.LittleEndian.PutUint32(hdr[8:12], uint32(blockBytes))
	binary.LittleEndian.PutUint64(hdr[12:20], uint64(dirOffset))
	binary.LittleEndian.PutUint32(hdr[20:24], uint32(len(dirBytes)))
	copy(hdr[24:56], dirSum[:])
	if _, err := f.WriteAt(hdr[:], 0); err != nil {
		return st, err
	}
	if err := f.Sync(); err != nil {
		return st, err
	}
	packedBytes := dirOffset + int64(len(dirBytes))
	if err := f.Close(); err != nil {
		return st, err
	}
	f = nil
	if err := os.Rename(tmp, packPath); err != nil {
		_ = os.Remove(tmp)
		return st, err
	}
	ratio := 1.0
	if rawOff > 0 {
		ratio = float64(packedBytes) / float64(rawOff)
	}
	return PackStats{
		Path:        packPath,
		Turns:       len(turns),
		RawBytes:    rawOff,
		PackedBytes: packedBytes,
		BlockCount:  len(dir.Blocks),
		BlockBytes:  blockBytes,
		Ratio:       ratio,
	}, nil
}

// OpenBodyPack opens and verifies the directory digest of a BodyPack file.
func OpenBodyPack(packPath string) (*BodyPack, error) {
	f, err := os.Open(packPath)
	if err != nil {
		return nil, err
	}
	ok := false
	defer func() {
		if !ok {
			f.Close()
		}
	}()
	var hdr [packHeaderBytes]byte
	if _, err := f.ReadAt(hdr[:], 0); err != nil {
		return nil, fmt.Errorf("vault pack header: %w", err)
	}
	if string(hdr[0:8]) != packMagic {
		return nil, fmt.Errorf("vault pack: invalid magic")
	}
	dirOffset := int64(binary.LittleEndian.Uint64(hdr[12:20]))
	dirLen := int(binary.LittleEndian.Uint32(hdr[20:24]))
	dirBytes := make([]byte, dirLen)
	if _, err := f.ReadAt(dirBytes, dirOffset); err != nil {
		return nil, fmt.Errorf("vault pack directory read: %w", err)
	}
	gotSum := sha256.Sum256(dirBytes)
	if !bytes.Equal(gotSum[:], hdr[24:56]) {
		return nil, fmt.Errorf("vault pack: directory SHA-256 mismatch (corrupted)")
	}
	var dir packDirectory
	if err := json.Unmarshal(dirBytes, &dir); err != nil {
		return nil, fmt.Errorf("vault pack directory decode: %w", err)
	}
	if dir.BlockBytes <= 0 {
		return nil, fmt.Errorf("vault pack: invalid block_bytes %d", dir.BlockBytes)
	}
	byID := make(map[string]packTurnMeta, len(dir.Turns))
	for _, tm := range dir.Turns {
		byID[tm.ID] = tm
	}
	ok = true
	return &BodyPack{file: f, dir: dir, turnByID: byID}, nil
}

// Close closes the underlying pack file.
func (p *BodyPack) Close() error {
	if p == nil || p.file == nil {
		return nil
	}
	return p.file.Close()
}

// TurnCount returns the number of packed turns.
func (p *BodyPack) TurnCount() int { return len(p.dir.Turns) }

func (p *BodyPack) decodeBlock(idx int) ([]byte, int, error) {
	if idx < 0 || idx >= len(p.dir.Blocks) {
		return nil, 0, fmt.Errorf("vault pack: block %d out of range", idx)
	}
	blk := p.dir.Blocks[idx]
	comp := make([]byte, blk.CompLen)
	if _, err := p.file.ReadAt(comp, blk.Offset); err != nil {
		return nil, 0, err
	}
	r := flate.NewReader(bytes.NewReader(comp))
	raw, err := io.ReadAll(r)
	_ = r.Close()
	if err != nil {
		return nil, blk.CompLen, fmt.Errorf("vault pack block %d inflate: %w", idx, err)
	}
	if len(raw) != blk.RawLen {
		return nil, blk.CompLen, fmt.Errorf("vault pack block %d length %d != %d", idx, len(raw), blk.RawLen)
	}
	sum := sha256.Sum256(raw)
	if hex.EncodeToString(sum[:]) != blk.RawSHA {
		return nil, blk.CompLen, fmt.Errorf("vault pack block %d SHA-256 mismatch", idx)
	}
	return raw, blk.CompLen, nil
}

func (p *BodyPack) sliceRaw(start, length int64, cache map[int][]byte, st *PackReadStats) ([]byte, error) {
	if length == 0 {
		return nil, nil
	}
	if start < 0 || start+length > p.dir.TotalRaw {
		return nil, fmt.Errorf("vault pack: raw range [%d:%d] out of bounds (%d)", start, start+length, p.dir.TotalRaw)
	}
	bs := int64(p.dir.BlockBytes)
	firstBlk := int(start / bs)
	lastBlk := int((start + length - 1) / bs)
	out := make([]byte, 0, length)
	for bi := firstBlk; bi <= lastBlk; bi++ {
		raw, ok := cache[bi]
		if !ok {
			var compBytes int
			var err error
			raw, compBytes, err = p.decodeBlock(bi)
			if err != nil {
				return nil, err
			}
			cache[bi] = raw
			if st != nil {
				st.BlocksDecoded++
				st.PhysicalBytesRead += compBytes
				st.DecompressedBytes += len(raw)
			}
		}
		blkStart := int64(bi) * bs
		lo := max(int64(0), start-blkStart)
		hi := min(int64(len(raw)), start+length-blkStart)
		out = append(out, raw[lo:hi]...)
	}
	return out, nil
}

// ReadTurn decompresses ONLY the block(s) spanning turnID, verifies every
// block digest and the turn's BodySHA, and returns the exact verbatim body.
func (p *BodyPack) ReadTurn(turnID string) (string, string, error) {
	tm, ok := p.turnByID[turnID]
	if !ok {
		return "", "", fmt.Errorf("vault pack: turn %q not found", turnID)
	}
	cache := map[int][]byte{}
	raw, err := p.sliceRaw(tm.RawOffset, int64(tm.RawLen), cache, nil)
	if err != nil {
		return "", "", err
	}
	body := string(raw)
	if SHA(body) != tm.BodySHA {
		return "", "", fmt.Errorf("vault pack: turn %s body SHA mismatch", turnID)
	}
	return body, tm.BodySHA, nil
}

// ReadSpans resolves a batch of EvidenceAnchors with block coalescing: any
// citations sharing a 4 KiB compressed block decode and SHA-verify that block
// at most once, returning exact verified text plus physical/decompressed stats.
func (p *BodyPack) ReadSpans(anchors []EvidenceAnchor) ([]VerifiedSpan, PackReadStats, error) {
	st := PackReadStats{SpansRequested: len(anchors)}
	if len(anchors) == 0 {
		return nil, st, nil
	}
	// Pre-sort block indices so sequential disk/OS page reads stay ordered.
	bs := int64(p.dir.BlockBytes)
	blkSet := map[int]bool{}
	for _, a := range anchors {
		tm, ok := p.turnByID[a.Turn]
		if !ok {
			return nil, st, fmt.Errorf("vault pack: anchor turn %q not found", a.Turn)
		}
		if a.BodySHA != "" && a.BodySHA != tm.BodySHA {
			return nil, st, fmt.Errorf("vault pack: anchor %s body_sha mismatch", a.Turn)
		}
		if a.Start < 0 || a.End < a.Start || a.End > tm.RawLen {
			return nil, st, fmt.Errorf("vault pack: anchor %s [%d:%d] out of range (%d)", a.Turn, a.Start, a.End, tm.RawLen)
		}
		if a.End > a.Start {
			first := int((tm.RawOffset + int64(a.Start)) / bs)
			last := int((tm.RawOffset + int64(a.End) - 1) / bs)
			for bi := first; bi <= last; bi++ {
				blkSet[bi] = true
			}
		}
	}
	blks := make([]int, 0, len(blkSet))
	for bi := range blkSet {
		blks = append(blks, bi)
	}
	sort.Ints(blks)
	cache := make(map[int][]byte, len(blks))
	for _, bi := range blks {
		raw, compBytes, err := p.decodeBlock(bi)
		if err != nil {
			return nil, st, err
		}
		cache[bi] = raw
		st.BlocksDecoded++
		st.PhysicalBytesRead += compBytes
		st.DecompressedBytes += len(raw)
	}

	out := make([]VerifiedSpan, 0, len(anchors))
	for _, a := range anchors {
		tm := p.turnByID[a.Turn]
		raw, err := p.sliceRaw(tm.RawOffset+int64(a.Start), int64(a.End-a.Start), cache, nil)
		if err != nil {
			return nil, st, err
		}
		st.BytesReturned += len(raw)
		out = append(out, VerifiedSpan{Anchor: a, Text: string(raw)})
	}
	return out, st, nil
}

// VerifyAll verifies every compressed block SHA-256 and every turn BodySHA.
func (p *BodyPack) VerifyAll() error {
	cache := make(map[int][]byte, len(p.dir.Blocks))
	for i := range p.dir.Blocks {
		raw, _, err := p.decodeBlock(i)
		if err != nil {
			return err
		}
		cache[i] = raw
	}
	for _, tm := range p.dir.Turns {
		raw, err := p.sliceRaw(tm.RawOffset, int64(tm.RawLen), cache, nil)
		if err != nil {
			return err
		}
		if SHA(string(raw)) != tm.BodySHA {
			return fmt.Errorf("vault pack: turn %s SHA verification failed", tm.ID)
		}
	}
	return nil
}
