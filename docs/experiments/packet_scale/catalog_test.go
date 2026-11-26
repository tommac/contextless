package main

import (
	"bytes"
	"crypto/sha256"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

var snapshot = sha256.Sum256([]byte("catalog-test/fixture"))

func catalogFixture(t *testing.T, last int, omitted func(int) bool) (string, *catalog) {
	t.Helper()
	dir := t.TempDir()
	h, err := buildCatalog(dir, snapshot, last, omitted)
	if err != nil {
		t.Fatal(err)
	}
	c, err := openCatalog(dir, h, snapshot)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { c.file.Close() })
	return dir, c
}

func TestCatalogMatchesMembershipOracle(t *testing.T) {
	for _, tc := range []struct {
		name  string
		last  int
		omit  func(int) bool
		kinds []byte
	}{
		{"zero turns", 0, func(int) bool { return false }, nil},
		{"empty set", pageTurns, func(int) bool { return false }, []byte{emptyPage}},
		{"all omitted", pageTurns + 3, func(int) bool { return true }, []byte{fullPage, fullPage}},
		{"single member", pageTurns, func(id int) bool { return id == pageTurns }, []byte{includeList}},
		{"sparse", pageTurns * 2, func(id int) bool { return id%41 == 0 }, []byte{includeList, includeList}},
		{"sparse complement", pageTurns, func(id int) bool { return id%41 != 0 }, []byte{excludeList}},
		{"fragmented", pageTurns + 7, func(id int) bool { return id%2 == 0 }, []byte{bitmapPage, bitmapPage}},
		{"empty middle page", pageTurns*2 + 1, func(id int) bool { return id == 1 || id > pageTurns*2 }, []byte{includeList, emptyPage, fullPage}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, c := catalogFixture(t, tc.last, tc.omit)
			var kinds []byte
			for _, p := range c.pages {
				kinds = append(kinds, p.kind)
			}
			if !slices.Equal(kinds, tc.kinds) {
				t.Fatalf("page kinds %v, want %v", kinds, tc.kinds)
			}
			want := make([]byte, tc.last)
			count := 0
			for id := 1; id <= tc.last; id++ {
				before := c.reads
				member, err := c.contains(id)
				if err != nil || member != tc.omit(id) || c.reads-before > 1 {
					t.Fatalf("id %d: member %t, reads %d, err %v", id, member, c.reads-before, err)
				}
				if tc.omit(id) {
					want[id-1] = 1
					count++
				}
			}
			if c.h.Omitted != count {
				t.Fatalf("count %d, want %d", c.h.Omitted, count)
			}
			for _, id := range []int{0, tc.last + 1} {
				if _, err := c.contains(id); err == nil {
					t.Fatalf("accepted id %d outside domain", id)
				}
			}
			for _, limit := range []int{1, 37, pageTurns} {
				got := make([]byte, tc.last)
				var cursor *catalogCursor
				last := 0
				for {
					ids, next, err := c.enumerate(cursor, limit)
					if err != nil || len(ids) > limit {
						t.Fatalf("limit %d: got %d IDs, err %v", limit, len(ids), err)
					}
					for _, id := range ids {
						if id <= last || id > tc.last {
							t.Fatalf("invalid id %d following %d", id, last)
						}
						got[id-1] = 1
						last = id
					}
					if next == nil {
						break
					}
					if len(ids) != limit || next.After != last || next.Index != c.h.Index {
						t.Fatalf("invalid continuation %+v", next)
					}
					cursor = next
				}
				if !bytes.Equal(got, want) {
					t.Fatalf("reconstructed membership differs at limit %d", limit)
				}
			}
		})
	}
}

func TestCatalogRejectsStaleHandlesAndInvalidCursors(t *testing.T) {
	dir, c := catalogFixture(t, 17, func(id int) bool { return id%2 == 0 })
	reject := func(h catalogHandle, current [32]byte) {
		t.Helper()
		if opened, err := openCatalog(dir, h, current); err == nil {
			opened.file.Close()
			t.Fatalf("accepted invalid handle %+v", h)
		}
	}
	reject(c.h, sha256.Sum256([]byte("new revision")))
	h := c.h
	h.Format += "x"
	reject(h, snapshot)
	h = c.h
	h.Omitted++
	reject(h, snapshot)
	h = c.h
	h.Last++
	reject(h, snapshot)
	h = c.h
	h.Index = "../" + h.Index
	reject(h, snapshot)
	other, err := buildCatalog(dir, snapshot, 17, func(id int) bool { return id%3 == 0 })
	if err != nil {
		t.Fatal(err)
	}
	h = c.h
	h.Index = other.Index
	reject(h, snapshot)
	for _, cursor := range []catalogCursor{{other.Index, 6}, {c.h.Index, -1}, {c.h.Index, 18}, {c.h.Index, 3}} {
		if _, _, err := c.enumerate(&cursor, 2); err == nil {
			t.Fatalf("accepted invalid cursor %+v", cursor)
		}
	}
	for _, limit := range []int{0, -1, pageTurns + 1} {
		if _, _, err := c.enumerate(nil, limit); err == nil {
			t.Fatalf("accepted invalid limit %d", limit)
		}
	}
}

func TestCatalogDetectsChangedBytes(t *testing.T) {
	for _, tc := range []struct {
		name   string
		omit   func(int) bool
		offset int
	}{
		{"directory", func(id int) bool { return id%2 == 0 }, 4},
		{"bitmap", func(id int) bool { return id%2 == 0 }, descriptorBytes},
		{"list", func(id int) bool { return id%89 == 0 }, descriptorBytes},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir, c := catalogFixture(t, pageTurns, tc.omit)
			path := filepath.Join(dir, c.h.Index+".cxi")
			b, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			b[tc.offset] ^= 1
			if err := os.WriteFile(path, b, 0600); err != nil {
				t.Fatal(err)
			}
			if tc.offset < descriptorBytes {
				if opened, err := openCatalog(dir, c.h, snapshot); err == nil {
					opened.file.Close()
					t.Fatal("opened modified directory")
				}
				return
			}
			if _, err := c.contains(1); err == nil {
				t.Fatal("point lookup accepted corrupt page")
			}
			if ids, _, err := c.enumerate(nil, 2); err == nil || len(ids) != 0 {
				t.Fatal("enumeration accepted corrupt page")
			}
		})
	}
}
