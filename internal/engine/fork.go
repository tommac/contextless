package engine

import (
	"fmt"

	"github.com/tommac/contextless/internal/config"
	"github.com/tommac/contextless/internal/vault"
)

// ForkOptions controls time-travel workspace branching (`cx fork`).
type ForkOptions struct {
	// AsOfSeq forks the workspace as it stood at turn AsOfSeq (0 copies all turns).
	AsOfSeq int    `json:"as_of_seq,omitempty"`
	Title   string `json:"title,omitempty"`
}

// ForkResult summarizes a newly created counterfactual / time-travel workspace.
type ForkResult struct {
	Source        string `json:"source"`
	Target        string `json:"target"`
	Dir           string `json:"dir"`
	AsOfSeq       int    `json:"as_of_seq"`
	TurnsCopied   int    `json:"turns_copied"`
	FactsKept     int    `json:"facts_kept"`
	DecisionsKept int    `json:"decisions_kept"`
	SnapshotRoot  string `json:"snapshot_root"`
}

// ForkWorkspace creates a new git-backed workspace dstName cloned from srcName
// at turn opt.AsOfSeq (or the latest turn when AsOfSeq == 0). Unlike a naive
// directory copy, time-travel forking at AsOfSeq > 0 prunes future turns/cards
// AND reconstructs true point-in-time fact/decision status (restoring records
// that were only superseded after AsOfSeq back to "active"), rebuilds a clean
// SQLite index + materialized views, and seals an initial git commit.
func ForkWorkspace(cfg config.Config, srcName, dstName string, opt ForkOptions) (ForkResult, error) {
	var res ForkResult
	src, err := OpenReadOnly(cfg, srcName)
	if err != nil {
		return res, err
	}
	defer src.Close()

	sess, err := src.V.Session()
	if err != nil {
		return res, err
	}
	turns, err := src.V.Turns()
	if err != nil {
		return res, err
	}
	cutoff := opt.AsOfSeq
	if cutoff <= 0 && len(turns) > 0 {
		cutoff = turns[len(turns)-1].Seq
	}
	title := opt.Title
	if title == "" {
		if opt.AsOfSeq > 0 {
			title = fmt.Sprintf("%s (fork @ t%06d)", sess.Title, cutoff)
		} else {
			title = sess.Title + " (fork)"
		}
	}

	dst, err := Create(cfg, dstName, title, sess.Purpose)
	if err != nil {
		return res, err
	}
	defer dst.Close()

	// Preserve standing session body & attached repo definitions.
	dstSess, _ := dst.V.Session()
	dstSess.Body = sess.Body
	dstSess.Repos = sess.Repos
	if err := dst.V.PutSession(dstSess); err != nil {
		return res, err
	}

	cards := src.V.Cards()
	copied := 0
	for _, t := range turns {
		if cutoff > 0 && t.Seq > cutoff {
			break
		}
		nt, err := dst.V.AppendTurnAt(t.Role, t.Author, t.Run, t.ReplyTo, t.Body, t.Created)
		if err != nil {
			return res, err
		}
		if c, ok := cards[t.ID]; ok && c.BodySHA == nt.BodySHA {
			_ = dst.V.PutCard(c)
		}
		copied++
	}

	// Copy threads and entities referenced up to cutoff (or all threads/entities).
	for _, th := range src.V.Threads() {
		_ = dst.V.PutThread(th)
	}
	for _, en := range src.V.Entities() {
		_ = dst.V.PutEntity(en)
	}

	// Reconstruct point-in-time facts at cutoff.
	allFacts := src.V.Facts()
	supFactAt := map[string]bool{}
	for _, f := range allFacts {
		seq, _ := vault.TurnSeq(f.Source)
		if cutoff > 0 && seq > cutoff {
			continue
		}
		for _, old := range f.Supersedes {
			supFactAt[old] = true
		}
	}
	factsKept := 0
	for _, f := range allFacts {
		seq, _ := vault.TurnSeq(f.Source)
		if cutoff > 0 && seq > cutoff {
			continue
		}
		if opt.AsOfSeq > 0 {
			if supFactAt[f.ID] {
				f.Status = "superseded"
			} else if f.Status == "superseded" {
				f.Status = "active" // superseding turn has not happened yet on this timeline!
			}
		}
		if err := dst.V.PutFact(f); err == nil {
			factsKept++
		}
	}

	// Reconstruct point-in-time decisions at cutoff.
	allDecs := src.V.Decisions()
	supDecAt := map[string]bool{}
	for _, d := range allDecs {
		seq, _ := vault.TurnSeq(d.Source)
		if cutoff > 0 && seq > cutoff {
			continue
		}
		for _, old := range d.Supersedes {
			supDecAt[old] = true
		}
	}
	decsKept := 0
	for _, d := range allDecs {
		seq, _ := vault.TurnSeq(d.Source)
		if cutoff > 0 && seq > cutoff {
			continue
		}
		if opt.AsOfSeq > 0 {
			if supDecAt[d.ID] {
				d.Status = "superseded"
			} else if d.Status == "superseded" {
				d.Status = "active"
			}
		}
		if err := dst.V.PutDecision(d); err == nil {
			decsKept++
		}
	}

	if err := dst.Reindex(); err != nil {
		return res, err
	}
	dst.V.Commit(fmt.Sprintf("Fork workspace from %s at seq %d", srcName, cutoff))
	_, root, _ := dst.S.SnapshotRoot(0)
	return ForkResult{
		Source:        srcName,
		Target:        dstName,
		Dir:           dst.V.Dir,
		AsOfSeq:       cutoff,
		TurnsCopied:   copied,
		FactsKept:     factsKept,
		DecisionsKept: decsKept,
		SnapshotRoot:  root,
	}, nil
}
