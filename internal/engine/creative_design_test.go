package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/tommac/contextless/internal/store"
	"github.com/tommac/contextless/internal/vault"
)

func openDesignTestStore(t *testing.T, root string) *store.Store {
	t.Helper()
	s, err := store.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func designTestCard(key, role, quote string) CreativeCard {
	prefix := "Archive — original\r\n"
	body := prefix + quote + "\r\nOriginal appendix."
	return CreativeCard{Key: key, Kind: "file", Source: key + ".md", Role: role, Relation: role, Excerpt: quote,
		Body: body, BodySHA: vault.SHA(body), Start: len(prefix), End: len(prefix) + len(quote)}
}

func designTestSpec(id string, choices ...CreativeChoice) CreativeDesign {
	return CreativeDesign{ID: id, Title: id, Prediction: "Increase useful cache hits.", Outcome: "Observed in replay.", Choices: choices}
}

func saveTestDesign(t *testing.T, s *store.Store, board string, snap CreativeCanvas, spec CreativeDesign, want int) {
	t.Helper()
	revision, err := SaveCreativeDesign(s, board, snap, spec)
	if err != nil || revision != want {
		t.Fatalf("save %s parent %d: got %d, %v; want %d", spec.ID, spec.Parent, revision, err, want)
	}
}

func TestCreativeDesignOriginalPinAndUnrelatedEdit(t *testing.T) {
	root := t.TempDir()
	s := openDesignTestStore(t, root)
	body := "Introduction: Café.\r\n## Rule\r\nÉviter les doublons.\r\n## Appendix\r\nOld result."
	note := filepath.Join(root, "Cache.md")
	write := func(text string) {
		t.Helper()
		if err := os.WriteFile(note, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	write(body)
	path := writeCreativeBoard(t, root, canvasNode{ID: "file", Type: "file", File: "Cache.md", Subpath: "#Rule"}, "constraint: rule")
	read := func() CreativeCanvas {
		t.Helper()
		snap, err := ReadCreativeCanvas(path, root)
		if err != nil {
			t.Fatal(err)
		}
		return snap
	}
	original := read()
	spec := designTestSpec("read", CreativeChoice{Choice: "Protect repeats", AnyOf: []string{original.Cards[0].Key}})
	saveTestDesign(t, s, "cache.canvas", original, spec, 1)
	// Edits before and after the cited section change its offsets and SHA,
	// while preserving the precise passage used for this design.
	outside := "New introduction.\r\n" + strings.Replace(body, "Old result.", "New result.", 1)
	write(outside)
	now := read()
	if now.Cards[0].BodySHA == original.Cards[0].BodySHA || now.Cards[0].Start == original.Cards[0].Start {
		t.Fatal("fixture did not move the source")
	}
	impacts, err := ReviewCreativeDesigns(s, "cache.canvas", now, nil)
	if err != nil || len(impacts) != 0 {
		t.Fatalf("unrelated edit triggered a review: %+v %v", impacts, err)
	}
	write(strings.Replace(outside, "Éviter les doublons.", "Preserve duplicates.", 1))
	changed := read()
	impacts, err = ReviewCreativeDesigns(s, "cache.canvas", changed, nil)
	if err != nil || len(impacts) != 1 || impacts[0].Status != "review" || len(impacts[0].Old) != 1 || len(impacts[0].New) != 1 {
		t.Fatalf("review: %+v %v", impacts, err)
	}
	old, newPin := impacts[0].Old[0], impacts[0].New[0]
	prior, err := CreativeOriginal(s, old.BodySHA)
	if err != nil || prior != body || prior[old.Start:old.End] != old.Excerpt || old.Excerpt != original.Cards[0].Excerpt || newPin.Excerpt != changed.Cards[0].Excerpt {
		t.Fatalf("source mismatch: %#v %#v %v", old, newPin, err)
	}
	spec.Parent = 1
	saveTestDesign(t, s, "cache.canvas", changed, spec, 2)
	if err := os.Remove(note); err != nil {
		t.Fatal(err)
	}
	for _, pin := range []CreativeCard{old, newPin} {
		recovered, err := CreativeOriginal(s, pin.BodySHA)
		if err != nil || vault.SHA(recovered) != pin.BodySHA || recovered[pin.Start:pin.End] != pin.Excerpt {
			t.Fatalf("cannot reopen revision pin %#v: %v", pin, err)
		}
	}
}

func TestCreativeDesignImpactFollowsDeclaredChoice(t *testing.T) {
	s := openDesignTestStore(t, t.TempDir())
	primary := designTestCard("a", "constraint", "Hold repeats.")
	backup := designTestCard("b", "constraint", "Independent source confirms reuse.")
	inspiration := designTestCard("i", "inspiration", "Borrow the library return shelf.")
	other := designTestCard("u", "constraint", "Finish before midnight.")
	before := CreativeCanvas{Question: "How should notes be stored?", Cards: []CreativeCard{primary, backup, inspiration, other}}
	spec := designTestSpec("cache", CreativeChoice{"Cache repeated notes", []string{"a", "b"}}, CreativeChoice{"Use a return shelf", []string{"i"}})
	saveTestDesign(t, s, "one.canvas", before, spec, 1)
	saveTestDesign(t, s, "one.canvas", before, designTestSpec("schedule", CreativeChoice{"Meet deadline", []string{"u"}}), 1)
	saveTestDesign(t, s, "two.canvas", before, designTestSpec("other-board", CreativeChoice{"Isolated", []string{"a"}}), 1)
	// The first constraint is withdrawn, its independent support survives,
	// and a separate inspiration and unrelated design both change.
	after := CreativeCanvas{Question: before.Question, Cards: []CreativeCard{backup, designTestCard("i", "inspiration", "Borrow time sharing."), designTestCard("u", "constraint", "Run at midnight.")}}
	assert := func(keys []string, want map[string]string) {
		t.Helper()
		rows, err := ReviewCreativeDesigns(s, "one.canvas", after, keys)
		if err != nil {
			t.Fatal(err)
		}
		got := map[string]string{}
		for _, row := range rows {
			got[row.ID+"/"+row.Choice] = row.Status
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("for keys %v got %v; want %v", keys, got, want)
		}
	}
	assert(nil, map[string]string{"cache/Cache repeated notes": "supported", "cache/Use a return shelf": "inspiration", "schedule/Meet deadline": "review"})
	assert([]string{"a"}, map[string]string{"cache/Cache repeated notes": "supported"})
	assert([]string{"u"}, map[string]string{"schedule/Meet deadline": "review"})
	assert([]string{"does-not-exist"}, map[string]string{})
	after.Cards = after.Cards[1:] // neither supporting constraint remains
	assert([]string{"a", "b"}, map[string]string{"cache/Cache repeated notes": "review"})
}

func TestCreativeDesignRetryAndConcurrentParent(t *testing.T) {
	root := t.TempDir()
	dbs := []*store.Store{openDesignTestStore(t, root), openDesignTestStore(t, root)}
	snap := CreativeCanvas{Question: "Which cache?", Cards: []CreativeCard{designTestCard("e", "constraint", "No duplicates.")}}
	first := designTestSpec("bounded", CreativeChoice{"Skip duplicates", []string{"e"}})
	saveTestDesign(t, dbs[0], "a.canvas", snap, first, 1)
	saveTestDesign(t, dbs[1], "a.canvas", snap, first, 1)
	type result struct {
		spec CreativeDesign
		rev  int
		err  error
	}
	start := make(chan struct{})
	results := make(chan result, 2)
	for i, s := range dbs {
		spec := first
		spec.Parent = 1
		spec.Prediction = fmt.Sprintf("Variant %d", i)
		go func() {
			<-start
			rev, err := SaveCreativeDesign(s, "a.canvas", snap, spec)
			results <- result{spec, rev, err}
		}()
	}
	close(start)
	x, y := <-results, <-results
	if x.err != nil {
		x, y = y, x
	}
	if x.err != nil || x.rev != 2 || y.err == nil {
		t.Fatalf("competing submissions %+v %+v", x, y)
	}
	saveTestDesign(t, dbs[0], "a.canvas", snap, x.spec, 2)
	// Redelivery after another writer advanced the revision returns the
	// recorded version; a competing payload with the same parent still fails.
	saveTestDesign(t, dbs[0], "a.canvas", snap, first, 1)
	var count int
	if err := dbs[0].DB.QueryRow("SELECT COUNT(*) FROM creative_revisions").Scan(&count); err != nil || count != 2 {
		t.Fatalf("revisions %d: %v", count, err)
	}
}

func TestCreativeDesignExportPreservesHumanEdits(t *testing.T) {
	root := t.TempDir()
	s := openDesignTestStore(t, root)
	card := designTestCard("source", "constraint", "Keep every café receipt.")
	snap := CreativeCanvas{Question: "What must be kept?", Cards: []CreativeCard{card}}
	first := designTestSpec("cache", CreativeChoice{"Keep receipts", []string{card.Key}})
	saveTestDesign(t, s, "board.canvas", snap, first, 1)
	export := func(revision int) {
		t.Helper()
		if err := ExportCreativeDesign(s, root, "board.canvas", first.ID, revision); err != nil {
			t.Fatal(err)
		}
	}
	export(1)
	files, err := filepath.Glob(filepath.Join(root, "Creative designs", "*", "cache-v1.md"))
	if err != nil || len(files) != 1 {
		t.Fatalf("notes %v %v", files, err)
	}
	base := filepath.Join(root, "Creative designs", "Design comparisons.base")
	human := []byte("---\nreviewer: Ada\n---\nKeep this handwriting.\n")
	for _, path := range []string{files[0], base} {
		if err := os.WriteFile(path, human, 0600); err != nil {
			t.Fatal(err)
		}
	}
	export(1)
	first.Parent = 1
	first.Outcome = "Distinct new observed outcome."
	saveTestDesign(t, s, "board.canvas", snap, first, 2)
	export(2)
	for _, path := range []string{files[0], base} {
		got, err := os.ReadFile(path)
		if err != nil || string(got) != string(human) {
			t.Fatalf("lost annotations on %s: %q %v", path, got, err)
		}
	}
	next, err := os.ReadFile(filepath.Join(filepath.Dir(files[0]), "cache-v2.md"))
	if err != nil || !strings.Contains(string(next), "[[cache-v1]]") || !strings.Contains(string(next), card.BodySHA) || !strings.Contains(string(next), first.Outcome) {
		t.Fatalf("new note: %q %v", next, err)
	}
}

func TestCreativeDesignRejectsMalformedDependency(t *testing.T) {
	s := openDesignTestStore(t, t.TempDir())
	snap := CreativeCanvas{Cards: []CreativeCard{designTestCard("a", "constraint", "Require repetitions."), designTestCard("b", "inspiration", "Borrow a clock.")}, Question: "What works?"}
	create := func(choices ...CreativeChoice) CreativeDesign { return designTestSpec("policy", choices...) }
	for _, choices := range [][]CreativeChoice{
		nil, {{"Blank source set", nil}}, {{"   ", []string{"a"}}}, {{"Unknown", []string{"missing"}}}, {{"Duplicate", []string{"a", "a"}}},
		{{"Duplicate", []string{"a"}}, {"Duplicate", []string{"b"}}}, {{"Mixed roles", []string{"a", "b"}}},
	} {
		if _, err := SaveCreativeDesign(s, "board.canvas", snap, create(choices...)); err == nil {
			t.Fatalf("accepted invalid choices: %#v", choices)
		}
	}
	snap.Cards[0].Start++ // still in bounds but no longer matches the verbatim quote
	if _, err := SaveCreativeDesign(s, "board.canvas", snap, create(CreativeChoice{"Has bad pin", []string{"a"}})); err == nil {
		t.Fatal("accepted an invalid pin")
	}
}

func TestCreativeDesignReviewUsesLatestSourcesAndChecksHistoricalHash(t *testing.T) {
	s := openDesignTestStore(t, t.TempDir())
	old := designTestCard("old", "constraint", "Protect repeats.")
	next := designTestCard("new", "constraint", "Protect expensive reopens.")
	snap := CreativeCanvas{Question: "What should survive a scan?", Cards: []CreativeCard{old, next}}
	spec := designTestSpec("cache", CreativeChoice{"Prioritize a page", []string{old.Key}})
	saveTestDesign(t, s, "board.canvas", snap, spec, 1)
	spec.Parent = 1
	spec.Choices = []CreativeChoice{{"Prioritize a page", []string{next.Key}}}
	saveTestDesign(t, s, "board.canvas", snap, spec, 2)
	empty := CreativeCanvas{Question: snap.Question}
	ignored, err := ReviewCreativeDesigns(s, "board.canvas", empty, []string{old.Key})
	if err != nil || len(ignored) != 0 {
		t.Fatalf("past-only source triggered review: %+v %v", ignored, err)
	}
	review, err := ReviewCreativeDesigns(s, "board.canvas", empty, []string{next.Key})
	if err != nil || len(review) != 1 || review[0].Status != "review" || review[0].Old[0].BodySHA != next.BodySHA {
		t.Fatalf("missing current dependency: %+v %v", review, err)
	}
	if historical, err := CreativeOriginal(s, old.BodySHA); err != nil || historical != old.Body {
		t.Fatalf("lost earlier source: %q %v", historical, err)
	}
	if _, err := s.DB.Exec("UPDATE creative_bodies SET body=? WHERE sha=?", "incomplete", old.BodySHA); err != nil {
		t.Fatal(err)
	}
	if _, err := CreativeOriginal(s, old.BodySHA); err == nil {
		t.Fatal("historical read accepted bytes that differ from their recorded hash")
	}
}
