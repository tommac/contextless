package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/tommac/contextless/internal/creativity"
	"github.com/tommac/contextless/internal/engine"
	"github.com/tommac/contextless/internal/store"
)

func creativeObsidian(command string, args []string) error {
	fs := flag.NewFlagSet("obsidian "+command, flag.ContinueOnError)
	vault := fs.String("vault", "", "vault root (defaults to the board directory)")
	var creative, questionOnly, withoutLabels, asJSON bool
	var edge string
	if command == "brief" {
		fs.BoolVar(&creative, "creative", false, "produce a bounded creative brief")
		fs.BoolVar(&questionOnly, "question-only", false, "omit source cards in a creative brief")
		fs.BoolVar(&withoutLabels, "without-labels", false, "omit relationship labels in a creative brief")
		fs.BoolVar(&asJSON, "json", false, "print exact source pins with edge keys")
	}
	if command == "review" {
		fs.StringVar(&edge, "edge", "", "look up designs citing this edited edge")
	}
	fixed := 1
	if command == "design" || command == "source" {
		fixed = 2
	}
	if len(args) < fixed {
		return fmt.Errorf("obsidian %s needs %d positional argument(s)", command, fixed)
	}
	if err := fs.Parse(args[fixed:]); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return errors.New("unexpected positional argument after flags")
	}
	board := args[0]
	if command == "brief" {
		if asJSON && (questionOnly || withoutLabels) {
			return errors.New("-json prints source pins; use -question-only and -without-labels with text output")
		}
		if asJSON {
			snap, err := engine.ReadCreativeCanvas(board, *vault)
			if err != nil {
				return err
			}
			return printCreativeJSON(snap)
		}
		var text string
		var err error
		if creative || questionOnly || withoutLabels {
			text, err = engine.CompileCreativeObsidianCanvas(board, *vault, questionOnly, withoutLabels)
		} else {
			text, err = engine.CompileObsidianCanvas(board, *vault)
		}
		if err == nil {
			fmt.Print(text)
			if !strings.HasSuffix(text, "\n") {
				fmt.Println()
			}
		}
		return err
	}
	root, key, err := engine.CreativeBoardPath(board, *vault)
	if err != nil {
		return err
	}
	s, err := store.OpenFile(filepath.Join(root, ".cx", "creative.sqlite"))
	if err != nil {
		return err
	}
	defer s.Close()
	if command == "source" {
		body, err := engine.CreativeOriginal(s, args[1])
		if err == nil {
			_, err = io.WriteString(os.Stdout, body)
		}
		return err
	}
	snap, err := engine.ReadCreativeCanvas(board, root)
	if err != nil {
		return err
	}
	if command == "review" {
		var keys []string
		if edge != "" {
			keys = []string{edge}
		}
		impacts, err := engine.ReviewCreativeDesigns(s, key, snap, keys)
		if err != nil {
			return err
		}
		if impacts == nil {
			impacts = []engine.CreativeImpact{}
		}
		return printCreativeJSON(impacts)
	}
	file, err := os.Open(args[1])
	if err != nil {
		return err
	}
	defer file.Close()
	dec := json.NewDecoder(file)
	dec.DisallowUnknownFields()
	var spec engine.CreativeDesign
	if err = dec.Decode(&spec); err != nil {
		return err
	}
	if err := dec.Decode(new(any)); err != io.EOF {
		return errors.New("expected a single design object")
	}
	rev, err := engine.SaveCreativeDesign(s, key, snap, spec)
	if err != nil {
		return err
	}
	if err := engine.ExportCreativeDesign(s, root, key, spec.ID, rev); err != nil {
		return err
	}
	fmt.Printf("Saved %s revision %d. Open Creative designs/Design comparisons.base in the vault.\n", spec.ID, rev)
	return nil
}

func printCreativeJSON(value any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(value)
}

func creativeDemo(args []string) error {
	fs := flag.NewFlagSet("demo creativity", flag.ContinueOnError)
	evaluate := fs.Bool("evaluate", false, "read a study policy on stdin and return its measured outcome")
	scale := fs.Bool("scale", false, "compare source-linked review, keyword watching and reviewing everything at three archive sizes")
	board := fs.String("board", "", "replay the active trace named in the demonstration Canvas")
	workload := fs.String("workload", "", "editable replay JSON")
	output := fs.String("out", "", "keep the replay as an editable Obsidian vault in a new or empty directory")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		return errors.New("unexpected creativity demo argument")
	}
	if (*scale || *evaluate) && (*scale && *evaluate || *workload != "" || *output != "" || *board != "") {
		return errors.New("use -scale or -evaluate as a standalone mode")
	}
	if *evaluate {
		return creativity.EvaluateJSON(os.Stdin, os.Stdout)
	}
	if *scale {
		return engine.RunCreativeScale(os.Stdout)
	}
	if *board != "" {
		if *output != "" {
			return errors.New("-board reads an existing demo vault; -out creates a new one")
		}
		return replayCreativeBoard(os.Stdout, *board, *workload)
	}
	if *workload != "" {
		if *output != "" {
			return errors.New("-out creates the built-in example; edit the emitted workload.json to change it")
		}
		return creativity.RunDemo(os.Stdout, *workload)
	}
	if err := creativity.RunDemo(os.Stdout, *workload); err != nil {
		return err
	}
	root := *output
	if root == "" {
		var err error
		root, err = os.MkdirTemp("", "cx-creativity-*")
		if err != nil {
			return err
		}
		defer os.RemoveAll(root)
	} else {
		if err := os.MkdirAll(root, 0755); err != nil {
			return err
		}
		entries, err := os.ReadDir(root)
		if err != nil {
			return err
		}
		if len(entries) > 0 {
			return errors.New("-out requires a new or empty vault directory")
		}
	}
	rows, err := creativity.DemoRows(*workload)
	if err != nil {
		return err
	}
	if err := saveCreativeDemo(root, rows); err != nil {
		return err
	}
	if *output != "" {
		fmt.Printf("Open %s as a vault and compare revisions in Creative designs/Design comparisons.base.\n", root)
	}
	return nil
}

const demoCanvas = `{
 "nodes":[
  {"id":"q","type":"text","text":"# Question\nWhich passages should the cache admit when scans interrupt a working set?","x":0,"y":0,"width":350,"height":110},
  {"id":"trace","type":"file","file":"Return pattern.md","x":450,"y":-60,"width":350,"height":150},
  {"id":"limit","type":"text","text":"The cache has four resident slots.","x":450,"y":180,"width":350,"height":100}],
 "edges":[
  {"id":"return-pattern","fromNode":"q","fromSide":"right","toNode":"trace","toSide":"left","label":"constraint: observed reuse"},
  {"id":"capacity","fromNode":"q","fromSide":"right","toNode":"limit","toSide":"left","label":"constraint: available slots"}]}`

func saveCreativeDemo(root string, rows []creativity.DemoRow) error {
	data := map[string]string{"Passage cache.canvas": demoCanvas, "Return pattern.md": "Single-use scan passages never return; worthwhile passages recur at least three times.\nTrace: before\n"}
	for name, body := range data {
		if err := os.WriteFile(filepath.Join(root, name), []byte(body), 0644); err != nil {
			return err
		}
	}
	f, err := os.Create(filepath.Join(root, "workload.json"))
	if err != nil {
		return err
	}
	err = creativity.WriteWorkload(f)
	closeErr := f.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	board := "Passage cache.canvas"
	path := filepath.Join(root, board)
	s, err := store.OpenFile(filepath.Join(root, ".cx", "creative.sqlite"))
	if err != nil {
		return err
	}
	defer s.Close()
	snap, err := engine.ReadCreativeCanvas(path, root)
	if err != nil {
		return err
	}
	outcome := func(policy string) string {
		var parts []string
		for _, row := range rows {
			if row.Policy == policy {
				parts = append(parts, fmt.Sprintf("%s = %d", row.Workload, row.Score))
			}
		}
		return "Saved fetch cost × value: " + strings.Join(parts, ", ")
	}
	save := func(snapshot engine.CreativeCanvas, spec engine.CreativeDesign) error {
		rev, err := engine.SaveCreativeDesign(s, board, snapshot, spec)
		if err != nil {
			return err
		}
		return engine.ExportCreativeDesign(s, root, board, spec.ID, rev)
	}
	d := engine.CreativeDesign{ID: "admission", Title: "Admit on a second touch", Prediction: "A one-time scan should leave the recurring working set resident.", Outcome: outcome("two-touch"), Choices: []engine.CreativeChoice{{Choice: "Delay admission until reuse", AnyOf: []string{"return-pattern"}}}}
	if err = save(snap, d); err != nil {
		return err
	}
	control := engine.CreativeDesign{ID: "resident-limit", Title: "Cap residency", Prediction: "At most four passages can be resident.", Outcome: "Four slots allocated in both traces.", Choices: []engine.CreativeChoice{{Choice: "Enforce four resident slots", AnyOf: []string{"capacity"}}}}
	if err = save(snap, control); err != nil {
		return err
	}
	if err = os.WriteFile(filepath.Join(root, "Return pattern.md"), []byte("New high-value passages can recur exactly once before expiry.\nTrace: after\n"), 0644); err != nil {
		return err
	}
	next, err := engine.ReadCreativeCanvas(path, root)
	if err != nil {
		return err
	}
	impacts, err := engine.ReviewCreativeDesigns(s, board, next, []string{"return-pattern"})
	if err != nil {
		return err
	}
	if len(impacts) != 1 || impacts[0].ID != d.ID || impacts[0].Status != "review" {
		return errors.New("unexpected change attribution in the demo")
	}
	change := impacts[0]
	fmt.Printf("\nCanvas revision: %s needs review (%s).\nPrior passage: %s\nRevised passage: %s\n", change.ID, change.Choice, change.Old[0].Excerpt, change.New[0].Excerpt)
	d.Parent = 1
	d.Title = "Admit high-value passages on first touch"
	d.Prediction = "Protecting high-value first reads should pay when they return."
	d.Outcome = outcome("value-first")
	d.Choices[0].Choice = "Protect valuable first reads"
	if err = save(next, d); err != nil {
		return err
	}
	// An editable draft with the current parent lets the user record the next design.
	d.Parent = 2
	example, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(root, "admission.json"), append(example, '\n'), 0644)
}

// The trace label is an explicit experimental assumption, editable directly in Obsidian.
func replayCreativeBoard(out io.Writer, board, workload string) error {
	root, key, err := engine.CreativeBoardPath(board, "")
	if err != nil {
		return err
	}
	snap, err := engine.ReadCreativeCanvas(board, root)
	if err != nil {
		return err
	}
	phase := ""
	for _, card := range snap.Cards {
		if card.Key != "return-pattern" {
			continue
		}
		for _, line := range strings.Split(card.Excerpt, "\n") {
			if trace, ok := strings.CutPrefix(strings.TrimSpace(line), "Trace: "); ok {
				if phase != "" || (trace != "before" && trace != "after") {
					return errors.New("use one Trace: before or Trace: after line in the return-pattern card")
				}
				phase = trace
			}
		}
	}
	if phase == "" {
		return errors.New("the return-pattern card needs a Trace: before or Trace: after line")
	}
	if workload == "" {
		workload = filepath.Join(root, "workload.json")
	}
	rows, err := creativity.DemoRows(workload)
	if err != nil {
		return err
	}
	s, err := store.OpenFile(filepath.Join(root, ".cx", "creative.sqlite"))
	if err != nil {
		return err
	}
	defer s.Close()
	impacts, err := engine.ReviewCreativeDesigns(s, key, snap, []string{"return-pattern"})
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "Selected trace %s\n", phase)
	for _, row := range rows {
		if row.Workload == phase {
			fmt.Fprintf(out, "%s: %d saved fetch value (%d/%d reads hit)\n", row.Policy, row.Score, row.Hits, row.Requests)
		}
	}
	for _, impact := range impacts {
		fmt.Fprintf(out, "%s: %s (%s)\n", impact.ID, impact.Status, impact.Choice)
	}
	return nil
}
