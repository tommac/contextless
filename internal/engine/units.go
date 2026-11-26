package engine

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/tommac/contextless/internal/packet"
	"github.com/tommac/contextless/internal/store"
	"github.com/tommac/contextless/internal/vault"
)

// A unit's record is what a plan unit, or one part of an oversized
// message, hands back: its output, the claims the output relies on with
// their evidence, the choices it made that another unit could have made
// differently, what it could not settle, and evidence that disagrees with
// it. The engine checks every claim before anything reads the record: an
// evidence id must name stored text at or before the message being
// answered, and a quote must occur in the text it cites. A quote that is
// found is pinned to the paragraphs (atoms) that hold it, so evidence is
// passed by address and checked by code, never restated by a model
// (LLM×MapReduce's structured unit output; docs/ARCHITECTURE.md, plans).

// unitRecord is one unit's result as the engine checked and stored it,
// in artifacts/<plan>/<unit>.json.
type unitRecord struct {
	Plan          string         `json:"plan"`
	Unit          string         `json:"unit"`
	Title         string         `json:"title"`
	Role          string         `json:"role"`
	DependsOn     []string       `json:"depends_on,omitempty"`
	By            string         `json:"by,omitempty"`
	Runs          []string       `json:"runs,omitempty"`
	Status        string         `json:"status"`
	Output        string         `json:"output,omitempty"`
	Claims        []checkedClaim `json:"claims,omitempty"`
	Assumptions   []string       `json:"assumptions,omitempty"`
	OpenQuestions []string       `json:"open_questions,omitempty"`
	Conflicts     []string       `json:"conflicts,omitempty"`
	Notes         string         `json:"notes,omitempty"`
	Files         []fileRef      `json:"files,omitempty"`
	Changes       []repoChange   `json:"changes,omitempty"`
	Error         string         `json:"error,omitempty"`
}

// statusFailed is the status of a unit whose runs produced no record.
const statusFailed = "failed"

// checkedClaim is a unit's claim with the engine's check of it.
type checkedClaim struct {
	ID       string   `json:"id"` // u1.c3
	Claim    string   `json:"claim"`
	Evidence []string `json:"evidence,omitempty"`
	Quote    string   `json:"quote,omitempty"`
	Declared string   `json:"declared"` // support as the unit gave it
	Support  string   `json:"support"`  // support after the check
	Check    string   `json:"check"`
	// Located is the stored text that holds the quote: a turn, the
	// paragraphs of it that hold the quote (tNNNNNN#aK-M), or a file the
	// unit wrote.
	Located string `json:"located,omitempty"`
	// Anchor is the durable, content-addressed byte span of the verified
	// quote inside its immutable turn (survives repacking or splitter updates).
	Anchor *vault.EvidenceAnchor `json:"anchor,omitempty"`
	// Notes say why evidence was set aside or where a quote really is.
	Notes []string `json:"notes,omitempty"`
}

// What the engine found when it checked a claim.
const (
	checkVerified   = "verified"    // the quote is in the stored text the claim cites
	checkNotFound   = "not_found"   // the quote is in none of the texts it cites
	checkExternal   = "external"    // it cites web sources only, which the engine does not fetch
	checkNoQuote    = "no_quote"    // evidence, but no quote to check
	checkNoEvidence = "no_evidence" // nothing the engine could resolve
)

// fileRef is a file a unit wrote, as saved in the vault.
type fileRef struct {
	Path string `json:"path"`
	SHA  string `json:"sha"`
}

// unitFile is a saved file a claim may cite.
type unitFile struct{ path, text string }

// unitRecordOf checks a unit's outcome and makes its record. err is the
// unit's failure, if it produced no record.
func (e *Engine) unitRecordOf(plan, unit, title, role string, focus vault.Turn, o outcome, err error) unitRecord {
	r := unitRecord{Plan: plan, Unit: unit, Title: title, Role: role, By: o.Author, Status: o.Status, Changes: o.Changes}
	for _, ri := range o.Runs {
		r.Runs = append(r.Runs, ri.ID)
	}
	// a unit cites a file it wrote by the path it wrote it at (out/x),
	// or by where the vault keeps it (artifacts/<run>/x)
	files := map[string]unitFile{}
	for _, p := range o.Files {
		data, rerr := os.ReadFile(filepath.Join(e.V.Dir, p))
		if rerr != nil {
			continue
		}
		f := unitFile{path: filepath.ToSlash(p), text: string(data)}
		r.Files = append(r.Files, fileRef{Path: f.path, SHA: vault.SHA(f.text)})
		files[f.path] = f
		if parts := strings.SplitN(f.path, "/", 3); len(parts) == 3 {
			files[parts[2]] = f
			files["out/"+parts[2]] = f
		}
	}
	if err != nil {
		r.Status, r.Error = statusFailed, err.Error()
		return r
	}
	r.Output = o.Answer
	r.Assumptions, r.OpenQuestions, r.Conflicts = nonBlank(o.H.Assumptions), nonBlank(o.H.OpenQuestions), nonBlank(o.H.Conflicts)
	r.Notes = strings.TrimSpace(o.H.Notes)
	r.Claims = e.checkClaims(unit, o.H.Claims, focus, files)
	// Record verified & unverified citations as free ranking telemetry.
	if e.S != nil && !e.S.ReadOnly && len(r.Claims) > 0 {
		var sigs []store.AtomSignal
		for _, c := range r.Claims {
			switch c.Check {
			case checkVerified:
				if c.Located != "" {
					sigs = append(sigs, store.AtomSignal{FocusTurn: focus.ID, AtomID: c.Located, Signal: "cited_verified", Lane: "unit"})
				}
			case checkNotFound:
				for _, ev := range c.Evidence {
					sigs = append(sigs, store.AtomSignal{FocusTurn: focus.ID, AtomID: ev, Signal: "cited_unverified", Lane: "unit"})
				}
			}
		}
		_ = e.S.LogSignals(sigs)
	}
	return r
}

// refToken finds the recall ids inside a longer evidence string, such as
// "turn t000012" or "t000012, t000013".
var refToken = regexp.MustCompile(`t\d{6,}(?:#p\d+|#a\d+(?:-\d+)?)?`)

// checkClaims checks each claim's evidence and quote against the vault
// and the files the unit wrote.
func (e *Engine) checkClaims(unit string, claims []packet.UnitClaim, focus vault.Turn, files map[string]unitFile) []checkedClaim {
	turns := map[string]vault.Turn{}
	turn := func(id string) (vault.Turn, error) {
		if t, ok := turns[id]; ok {
			return t, nil
		}
		t, err := e.V.Turn(id)
		if err == nil {
			turns[id] = t
		}
		return t, err
	}
	var out []checkedClaim
	for i, c := range claims {
		cc := checkedClaim{ID: fmt.Sprintf("%s.c%d", unit, i+1), Claim: strings.TrimSpace(c.Claim), Quote: strings.TrimSpace(c.Quote)}
		cc.Declared = strings.ToLower(strings.TrimSpace(c.Support))
		switch cc.Declared {
		case packet.SupportQuoted, packet.SupportInferred, packet.SupportUnsupported:
		default:
			cc.Notes = append(cc.Notes, fmt.Sprintf("support %q is not quoted, inferred or unsupported", c.Support))
			cc.Declared = packet.SupportUnsupported
		}
		var spans []citedSpan
		var cited []unitFile
		var web []string
		for _, raw := range c.Evidence {
			ev := strings.TrimSpace(raw)
			if ev == "" {
				continue
			}
			cc.Evidence = append(cc.Evidence, ev)
			switch {
			case strings.HasPrefix(ev, "http://") || strings.HasPrefix(ev, "https://"):
				web = append(web, ev)
				continue
			case files[ev].path != "":
				cited = append(cited, files[ev])
				continue
			}
			ids := []string{ev}
			if _, ok := vault.ParseRef(ev); !ok {
				if found := refToken.FindAllString(ev, -1); len(found) > 0 {
					ids = found
				}
			}
			for _, id := range ids {
				sp, why := e.citedSpan(id, focus, turn)
				if why != "" {
					cc.Notes = append(cc.Notes, why)
					continue
				}
				spans = append(spans, sp)
			}
		}
		resolved := len(spans) > 0 || len(cited) > 0
		switch {
		case !resolved && len(web) == 0:
			cc.Check = checkNoEvidence
		case cc.Quote == "":
			cc.Check = checkNoQuote
		default:
			cc.Check = checkNotFound
			if at, anchor, note, ok := locateQuote(cc.Quote, spans, cited); ok {
				cc.Check, cc.Located, cc.Anchor = checkVerified, at, anchor
				if note != "" {
					cc.Notes = append(cc.Notes, note)
				}
			} else if len(web) > 0 {
				cc.Check = checkExternal
			}
		}
		cc.Support = standing(cc.Declared, cc.Check)
		out = append(out, cc)
	}
	return out
}

// standing is a claim's support after the check. A quoted claim whose
// quote is not in the text it cites, or that gives no quote, counts as
// inferred from that evidence; a claim with no evidence the engine could
// resolve is unsupported. A quote from a web source stands as declared,
// marked as not checked. The check never raises a claim above what the
// unit declared.
func standing(declared, check string) string {
	if declared == packet.SupportUnsupported {
		return declared
	}
	switch check {
	case checkVerified, checkExternal:
		return declared
	case checkNotFound, checkNoQuote:
		return packet.SupportInferred
	}
	return packet.SupportUnsupported
}

// citedSpan is the stored text an evidence id names: bytes from to to of
// a turn's body.
type citedSpan struct {
	id       string
	turn     vault.Turn
	from, to int
}

// citedSpan resolves an evidence id, or says why it names nothing a claim
// can rest on.
func (e *Engine) citedSpan(id string, focus vault.Turn, turn func(string) (vault.Turn, error)) (citedSpan, string) {
	ref, ok := vault.ParseRef(id)
	if !ok {
		return citedSpan{}, fmt.Sprintf("%q is not a turn, paragraph or part id, a file this unit wrote, or a web address", id)
	}
	if seq, _ := vault.TurnSeq(ref.Turn); focus.Seq > 0 && seq > focus.Seq {
		return citedSpan{}, fmt.Sprintf("%s comes after the message being answered", ref.Turn)
	}
	t, err := turn(ref.Turn)
	if err != nil {
		return citedSpan{}, fmt.Sprintf("%s does not exist", ref.Turn)
	}
	sp := citedSpan{id: id, turn: t, to: len(t.Body)}
	switch {
	case ref.Part > 0:
		parts := vault.Parts(t.Body)
		if ref.Part > len(parts) {
			return citedSpan{}, fmt.Sprintf("%s has %d part(s), so %s does not exist", t.ID, len(parts), id)
		}
		for _, p := range parts[:ref.Part-1] {
			sp.from += len(p)
		}
		sp.to = sp.from + len(parts[ref.Part-1])
	case ref.Atom > 0:
		atoms := vault.Atoms(t.ID, t.Body)
		if ref.AtomEnd > len(atoms) {
			return citedSpan{}, fmt.Sprintf("%s has %d paragraph(s), so %s does not exist", t.ID, len(atoms), id)
		}
		sp.from, sp.to = atoms[ref.Atom-1].Start, atoms[ref.AtomEnd-1].End
	}
	return sp, ""
}

// locateQuote finds a claim's quote in the spans it cites, else elsewhere
// in a turn it cites (the note says where), else in a file it cites, and
// names the stored text plus its durable byte anchor.
func locateQuote(quote string, spans []citedSpan, cited []unitFile) (at string, anchor *vault.EvidenceAnchor, note string, ok bool) {
	for _, sp := range spans {
		if a, b, ok := findQuote(sp.turn.Body, sp.from, sp.to, quote); ok {
			ref := atomsAt(sp.turn, a, b)
			anc := &vault.EvidenceAnchor{Turn: sp.turn.ID, BodySHA: sp.turn.BodySHA, Start: a, End: b, Seg: vault.AtomSeg, AtomRef: ref}
			return ref, anc, "", true
		}
	}
	for _, sp := range spans {
		if sp.from == 0 && sp.to == len(sp.turn.Body) {
			continue
		}
		if a, b, ok := findQuote(sp.turn.Body, 0, len(sp.turn.Body), quote); ok {
			ref := atomsAt(sp.turn, a, b)
			anc := &vault.EvidenceAnchor{Turn: sp.turn.ID, BodySHA: sp.turn.BodySHA, Start: a, End: b, Seg: vault.AtomSeg, AtomRef: ref}
			return ref, anc, fmt.Sprintf("the quote is in %s, not in the cited %s", ref, sp.id), true
		}
	}
	for _, f := range cited {
		if containsVerbatim(f.text, quote) {
			return f.path, nil, "", true
		}
	}
	return "", nil, "", false
}

// findQuote returns where quote occurs in body[from:to], comparing as
// containsVerbatim does (runs of whitespace collapsed), as byte offsets
// into body.
func findQuote(body string, from, to int, quote string) (start, end int, ok bool) {
	q := strings.Join(strings.Fields(quote), " ")
	if len(q) < 3 {
		return 0, 0, false
	}
	norm, at := normIndex(body[from:to], from)
	i := strings.Index(norm, q)
	if i < 0 {
		return 0, 0, false
	}
	return at[i], at[i+len(q)-1] + 1, true
}

// normIndex is s with every run of whitespace collapsed to one space and
// leading and trailing whitespace dropped (strings.Fields joined by
// spaces), with the byte offset in the body (s starts at base) of every
// byte of the result.
func normIndex(s string, base int) (string, []int) {
	var b strings.Builder
	at := make([]int, 0, len(s))
	space := false
	for i := 0; i < len(s); {
		r, w := utf8.DecodeRuneInString(s[i:])
		if unicode.IsSpace(r) {
			space = b.Len() > 0
			i += w
			continue
		}
		if space {
			b.WriteByte(' ')
			at = append(at, base+i)
			space = false
		}
		b.WriteString(s[i : i+w])
		for k := 0; k < w; k++ {
			at = append(at, base+i+k)
		}
		i += w
	}
	return b.String(), at
}

// atomsAt names the paragraphs of t that hold its bytes start to end, or
// the turn when that is all of it. Uses the bounded cached split.
func atomsAt(t vault.Turn, start, end int) string {
	atoms := packet.AtomsOf(t)
	a, b := 0, 0
	for _, x := range atoms {
		if a == 0 && start < x.End {
			a = x.K
		}
		if a > 0 && end <= x.End {
			b = x.K
			break
		}
	}
	if a == 0 || b == 0 || (a == 1 && b == len(atoms)) {
		return t.ID
	}
	return vault.AtomRangeID(t.ID, a, b)
}

// storeUnitRecord writes the record of unit i into the vault, as JSON for
// code and as the text the assembler reads, and returns it as a supplied
// item for later units, the assembler and the reviewer.
func (e *Engine) storeUnitRecord(dir string, i int, r unitRecord) packet.Item {
	header := unitHeader(r)
	text := renderUnitRecord(r)
	if js, err := json.MarshalIndent(r, "", "  "); err == nil {
		_ = os.WriteFile(filepath.Join(dir, r.Unit+".json"), append(js, '\n'), 0o644)
	}
	_ = os.WriteFile(filepath.Join(dir, r.Unit+".md"), []byte("# "+header+"\n\n"+text+"\n"), 0o644)
	return packet.Item{Key: packet.KeyExtra, ID: "unit-" + r.Unit, Seq: i + 1, Header: header, Text: text}
}

func unitHeader(r unitRecord) string {
	by := r.By
	if r.Error != "" || by == "" {
		by = "failed"
	}
	return fmt.Sprintf("Unit %s · %s · %s · by %s", r.Unit, r.Role, r.Title, by)
}

// renderUnitRecord is a unit's record as the assembler and the reviewer
// read it: the unit's own text verbatim, each claim with what the engine
// found when it checked it, and the files and repository changes the
// engine recorded. A unit that found no information gives no content,
// only what it looked for.
func renderUnitRecord(r unitRecord) string {
	var b strings.Builder
	changes := func() {
		for _, ch := range r.Changes {
			fmt.Fprintf(&b, "\nRepository %s:\n%s\n", ch.Repo, ch.Report)
		}
	}
	switch {
	case r.Error != "":
		fmt.Fprintf(&b, "This unit failed: %s\n", r.Error)
		changes()
		return strings.TrimSpace(b.String())
	case r.Status == packet.StatusNoInformation:
		b.WriteString("This unit found no information for its task.\n")
		if r.Notes != "" {
			fmt.Fprintf(&b, "\nWhat it looked for (its notes):\n%s\n", r.Notes)
		}
		return strings.TrimSpace(b.String())
	}
	if r.Status != packet.StatusCompleted {
		fmt.Fprintf(&b, "Status: %s\n\n", r.Status)
	}
	if out := strings.TrimSpace(r.Output); out != "" {
		fmt.Fprintf(&b, "Output:\n\n%s\n", out)
	}
	if len(r.Claims) > 0 {
		b.WriteString("\nClaims, each checked by the engine against the text it cites:\n")
		for _, c := range r.Claims {
			fmt.Fprintf(&b, "- %s · %s: %s\n", c.ID, claimStanding(c), flat(c.Claim))
			if c.Quote != "" {
				fmt.Fprintf(&b, "  Quote: \"%s\"\n", flat(c.Quote))
			}
			for _, n := range c.Notes {
				fmt.Fprintf(&b, "  Note: %s\n", n)
			}
		}
	}
	list := func(title string, xs []string) {
		if len(xs) == 0 {
			return
		}
		fmt.Fprintf(&b, "\n%s:\n", title)
		for _, x := range xs {
			fmt.Fprintf(&b, "- %s\n", x)
		}
	}
	list("Assumptions", r.Assumptions)
	list("Open questions", r.OpenQuestions)
	list("Conflicts", r.Conflicts)
	if r.Notes != "" {
		fmt.Fprintf(&b, "\nNotes:\n%s\n", r.Notes)
	}
	if len(r.Files) > 0 {
		b.WriteString("\nFiles saved:\n")
		for _, f := range r.Files {
			fmt.Fprintf(&b, "- %s (sha %s)\n", f.Path, short(f.SHA))
		}
	}
	changes()
	return strings.TrimSpace(b.String())
}

// claimStanding says how a claim stands after the check, for the line
// that shows it.
func claimStanding(c checkedClaim) string {
	s := c.Support
	if c.Declared != c.Support {
		s += " (declared " + c.Declared + ")"
	}
	switch c.Check {
	case checkVerified:
		return s + ", quote found in " + c.Located
	case checkNotFound:
		return s + ", quote not found in " + strings.Join(c.Evidence, ", ")
	case checkExternal:
		var web []string
		for _, ev := range c.Evidence {
			if strings.HasPrefix(ev, "http://") || strings.HasPrefix(ev, "https://") {
				web = append(web, ev)
			}
		}
		return s + ", from a web source the engine did not check: " + strings.Join(web, ", ")
	case checkNoQuote:
		return s + ", no quote, evidence " + strings.Join(c.Evidence, ", ")
	}
	return s + ", no evidence"
}

// checkCounts tallies the checks over a plan's records, for the log and
// for measuring how often units cite what they quote.
func checkCounts(recs []unitRecord) map[string]int {
	n := map[string]int{}
	for _, r := range recs {
		n["units"]++
		n[r.Status]++
		for _, c := range r.Claims {
			n["claims"]++
			n[c.Check]++
		}
	}
	return n
}

func (e *Engine) logChecks(recs []unitRecord) {
	n := checkCounts(recs)
	e.logf("unit records: %d of %d completed, %d found no information, %d failed; %d claims: %d quotes found, %d not found, %d from web sources, %d without a quote, %d without evidence",
		n[packet.StatusCompleted], n["units"], n[packet.StatusNoInformation], n[statusFailed],
		n["claims"], n[checkVerified], n[checkNotFound], n[checkExternal], n[checkNoQuote], n[checkNoEvidence])
}

// flat is s with its whitespace collapsed, for one line of a list.
func flat(s string) string { return strings.Join(strings.Fields(s), " ") }

func nonBlank(xs []string) []string {
	var out []string
	for _, x := range xs {
		if x = strings.TrimSpace(x); x != "" {
			out = append(out, x)
		}
	}
	return out
}

// unitRecordGuide tells a unit what its record holds; plan units and the
// parts of an oversized message share it.
const unitRecordGuide = `Your handoff is this unit's record. The engine checks it and passes it verbatim to an assembler that writes the final reply:

- answer: the unit's output, everything the reply needs from this unit, without preamble.
- claims: each statement of fact your output relies on that comes from the conversation, a file or a source. claim is one sentence; evidence lists where it is stated (a turn id such as t000123, paragraphs such as t000123#a4-6, a part such as t000123#p2, a file you wrote such as out/notes.md, or the address of a web page); quote is a short passage copied exactly, character for character, from that evidence, that states the claim; support is quoted (the quote states it), inferred (it follows from the evidence but is not stated there) or unsupported (your own judgment or general knowledge). The engine looks for every quote in the text it cites and marks the quotes it does not find; such a claim counts as inferred. Leave claims empty when the output is your own work (a draft, a design, code) rather than findings.
- assumptions: choices you made that another unit could have made differently (a name, an interface, a format, a unit of measure, a reading of the request).
- open_questions: what you could not settle.
- conflicts: evidence you saw that disagrees with your answer, with your instructions, or with other evidence.
- notes: anything else the assembler should know.`
