package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tommac/contextless/internal/packet"
	"github.com/tommac/contextless/internal/vault"
)

const (
	obsidianMachineStart = "<!-- cx:machine:start -->"
	obsidianMachineEnd   = "<!-- cx:machine:end -->"
)

// ObsidianExportStats reports what ExportObsidian wrote.
type ObsidianExportStats struct {
	Turns     int `json:"turns"`
	Threads   int `json:"threads"`
	Entities  int `json:"entities"`
	Facts     int `json:"facts"`
	Decisions int `json:"decisions"`
	Receipts  int `json:"receipts"`
	Preserved int `json:"preserved_human_notes"`
}

// ExportObsidian exports the workspace into an Obsidian-friendly Markdown
// directory with wikilinks, Bases-compatible YAML properties, and durable
// content-addressed EvidenceAnchors. Machine-owned sections are wrapped in
// `<!-- cx:machine:start --> ... <!-- cx:machine:end -->` so re-exporting
// never clobbers human annotations written outside the machine block.
func (e *Engine) ExportObsidian(outDir string) (ObsidianExportStats, error) {
	var st ObsidianExportStats
	for _, sub := range []string{"turns", "threads", "entities", "facts", "decisions", "receipts"} {
		if err := os.MkdirAll(filepath.Join(outDir, sub), 0o755); err != nil {
			return st, err
		}
	}
	writeNote := func(rel, frontmatter, machineBody string) error {
		full := filepath.Join(outDir, rel)
		block := obsidianMachineStart + "\n" + strings.TrimSpace(machineBody) + "\n" + obsidianMachineEnd
		if oldBytes, err := os.ReadFile(full); err == nil {
			old := string(oldBytes)
			sIdx := strings.Index(old, obsidianMachineStart)
			eIdx := strings.Index(old, obsidianMachineEnd)
			if sIdx >= 0 && eIdx > sIdx {
				// Preserve human notes before and after the machine-owned section.
				updated := old[:sIdx] + block + old[eIdx+len(obsidianMachineEnd):]
				st.Preserved++
				return os.WriteFile(full, []byte(updated), 0o644)
			}
		}
		doc := "---\n" + strings.TrimSpace(frontmatter) + "\n---\n\n" + block + "\n\n## Notes\n\n"
		return os.WriteFile(full, []byte(doc), 0o644)
	}

	turns, err := e.V.Turns()
	if err != nil {
		return st, err
	}
	cards := e.V.Cards()
	turnSHA := map[string]string{}
	for _, t := range turns {
		turnSHA[t.ID] = t.BodySHA
		var links []string
		if c, ok := cards[t.ID]; ok {
			for _, th := range c.Threads {
				links = append(links, "[["+th+"]]")
			}
			for _, en := range c.Mentions {
				links = append(links, "[["+en+"]]")
			}
		}
		fm := fmt.Sprintf("type: turn\nid: %s\nseq: %d\nrole: %s\ncreated: %s\nbody_sha: %s\nseg: %d\natoms: %d",
			t.ID, t.Seq, t.Role, t.Created.Format("2006-01-02T15:04:05Z"), t.BodySHA, vault.AtomSeg, len(packet.AtomsOf(t)))
		var body strings.Builder
		fmt.Fprintf(&body, "# %s (%s)\n\n", t.ID, t.Role)
		if len(links) > 0 {
			fmt.Fprintf(&body, "Links: %s\n\n", strings.Join(links, " · "))
		}
		body.WriteString(t.Body)
		if err := writeNote(filepath.Join("turns", t.ID+".md"), fm, body.String()); err != nil {
			return st, err
		}
		st.Turns++
	}

	for _, th := range e.V.Threads() {
		fm := fmt.Sprintf("type: thread\nid: %s\nstatus: %s", th.ID, th.Status)
		body := fmt.Sprintf("# %s\n\nGoal: %s\n\n%s", th.Title, th.Goal, th.Body)
		if err := writeNote(filepath.Join("threads", th.ID+".md"), fm, body); err != nil {
			return st, err
		}
		st.Threads++
	}
	for _, en := range e.V.Entities() {
		fm := fmt.Sprintf("type: entity\nid: %s\nkind: %s", en.ID, en.Kind)
		body := fmt.Sprintf("# %s\n\nAliases: %s\n\n%s", en.Name, strings.Join(en.Aliases, ", "), en.Body)
		if err := writeNote(filepath.Join("entities", en.ID+".md"), fm, body); err != nil {
			return st, err
		}
		st.Entities++
	}
	for _, f := range e.V.Facts() {
		fm := fmt.Sprintf("type: fact\nid: %s\nstatus: %s\nsource: %s\nsource_sha: %s\nseg: %d",
			f.ID, f.Status, f.Source, turnSHA[f.Source], vault.AtomSeg)
		var body strings.Builder
		fmt.Fprintf(&body, "# %s\n\n> %s\n\nSource: [[%s]]\n", f.Statement, f.Quote, f.Source)
		if len(f.Supersedes) > 0 {
			var sups []string
			for _, s := range f.Supersedes {
				sups = append(sups, "[["+s+"]]")
			}
			fmt.Fprintf(&body, "Supersedes: %s\n", strings.Join(sups, ", "))
		}
		if err := writeNote(filepath.Join("facts", f.ID+".md"), fm, body.String()); err != nil {
			return st, err
		}
		st.Facts++
	}
	for _, d := range e.V.Decisions() {
		fm := fmt.Sprintf("type: decision\nid: %s\nstatus: %s\nsource: %s\nsource_sha: %s\nseg: %d",
			d.ID, d.Status, d.Source, turnSHA[d.Source], vault.AtomSeg)
		body := fmt.Sprintf("# %s\n\n> %s\n\nSource: [[%s]]\n", d.Statement, d.Quote, d.Source)
		if err := writeNote(filepath.Join("decisions", d.ID+".md"), fm, body); err != nil {
			return st, err
		}
		st.Decisions++
	}
	if e.S != nil {
		if rs, err := e.S.ListReceipts(""); err == nil {
			for _, r := range rs {
				fm := fmt.Sprintf("type: receipt\nid: %s\nstatus: %s\nsnapshot_seq: %d\nsnapshot_root: %s\nas_of_seq: %d",
					r.ID, r.Status, r.SnapshotSeq, r.SnapshotRoot, r.AsOfSeq)
				var body strings.Builder
				fmt.Fprintf(&body, "# %s\n\nStatus: **%s**\n\n%s\n", r.Question, r.Status, r.AnswerText)
				if r.StaleReason != "" {
					fmt.Fprintf(&body, "\n> [!WARNING] Stale: %s\n", r.StaleReason)
				}
				if len(r.Evidence) > 0 {
					body.WriteString("\n## Verified Evidence\n\n")
					for _, ev := range r.Evidence {
						fmt.Fprintf(&body, "- [[%s]] (`%s` bytes `%d:%d` sha `%s`): %s\n",
							ev.Turn, ev.AtomRef, ev.Start, ev.End, short(ev.BodySHA), ev.Quote)
					}
				}
				if err := writeNote(filepath.Join("receipts", r.ID+".md"), fm, body.String()); err != nil {
					return st, err
				}
				st.Receipts++
			}
		}
	}
	return st, nil
}
