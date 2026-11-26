package vault

import (
	"path/filepath"
	"sort"
)

// TypeNote is the record type of an idea note.
const TypeNote = "note"

// Note is one idea stated in the conversation, in plain words, written by
// an indexing model and kept only through what code could check: each
// piece of evidence is a quote found word for word in the turn it cites,
// pinned to the paragraphs that hold it. A note is a pointer into the
// verbatim text and never stands in for it: the packet shows note lines
// as a map, and what a model reads is always the paragraph itself.
type Note struct {
	ID       string         `yaml:"id"`
	Title    string         `yaml:"title"`
	Kind     string         `yaml:"kind,omitempty"`
	Topics   []string       `yaml:"topics,omitempty"`
	Evidence []NoteEvidence `yaml:"evidence"`
	// Seq is the first turn the note points at, and Last the latest: a
	// packet for a message shows only notes whose evidence all lies
	// before that message.
	Seq  int    `yaml:"seq"`
	Last int    `yaml:"last"`
	Run  string `yaml:"run,omitempty"`
}

// NoteEvidence is a checked quote: where it is in its turn, and the
// paragraphs (atoms A to B) that hold it. BodySHA is the body the span was
// checked against; evidence whose turn no longer has that body is not used.
type NoteEvidence struct {
	Turn    string `yaml:"turn"`
	BodySHA string `yaml:"body_sha"`
	Quote   string `yaml:"quote"`
	Start   int    `yaml:"start"`
	End     int    `yaml:"end"`
	A       int    `yaml:"a"`
	B       int    `yaml:"b"`
}

// Atoms lists the paragraph ids the note's evidence is pinned to, in
// order of time, a run of paragraphs of one turn as one range id.
func (n Note) Atoms() []string {
	ev := append([]NoteEvidence{}, n.Evidence...)
	sort.SliceStable(ev, func(i, j int) bool {
		if ev[i].Turn != ev[j].Turn {
			return ev[i].Turn < ev[j].Turn
		}
		return ev[i].A < ev[j].A
	})
	var out []string
	seen := map[string]bool{}
	for _, e := range ev {
		if e.A <= 0 || e.B < e.A {
			continue
		}
		id := AtomRangeID(e.Turn, e.A, e.B)
		if !seen[id] {
			seen[id] = true
			out = append(out, id)
		}
	}
	return out
}

// Notes reads every idea note, oldest evidence first.
func (v *Vault) Notes() []Note {
	var out []Note
	for _, f := range v.list("notes") {
		var n Note
		if _, err := v.read(f, &n); err == nil && n.ID != "" {
			out = append(out, n)
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Seq != out[j].Seq {
			return out[i].Seq < out[j].Seq
		}
		return out[i].ID < out[j].ID
	})
	return out
}

// HasNotes reports whether the vault holds any idea note, without reading
// one.
func (v *Vault) HasNotes() bool { return len(v.list("notes")) > 0 }

// PutNote writes a note.
func (v *Vault) PutNote(n Note) error {
	return v.write(filepath.Join("notes", n.ID+".md"), n, "")
}
