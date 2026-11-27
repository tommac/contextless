package engine

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/tommac/contextless/internal/vault"
)

type canvasNode struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Text    string `json:"text"`
	File    string `json:"file"`
	Subpath string `json:"subpath"`
}

type canvasEdge struct {
	ID       string `json:"id"`
	FromNode string `json:"fromNode"`
	ToNode   string `json:"toNode"`
	Label    string `json:"label"`
}

type canvasBoard struct {
	Nodes []canvasNode `json:"nodes"`
	Edges []canvasEdge `json:"edges"`
}

func readCanvasBoard(path string) (canvasBoard, error) {
	var board canvasBoard
	data, err := os.ReadFile(path)
	if err != nil {
		return board, err
	}
	if !utf8.Valid(data) {
		return board, fmt.Errorf("canvas %s is not UTF-8", path)
	}
	if err := json.Unmarshal(data, &board); err != nil {
		return board, fmt.Errorf("canvas %s: %w", path, err)
	}
	return board, nil
}

func (edge canvasEdge) neighbor(id string) (string, bool) {
	switch id {
	case edge.FromNode:
		return edge.ToNode, true
	case edge.ToNode:
		return edge.FromNode, true
	default:
		return "", false
	}
}

func openCanvasRoot(boardPath, root string) (*os.Root, error) {
	if root == "" {
		root = filepath.Dir(boardPath)
	}
	return os.OpenRoot(root)
}

func readCanvasNote(root *os.Root, name string, creative bool) (string, error) {
	path, err := filepath.Localize(name)
	if err != nil || strings.Contains(name, "\\") || (creative && !strings.EqualFold(filepath.Ext(name), ".md")) {
		return "", fmt.Errorf("expected a vault-relative Markdown file, got %q", name)
	}
	body, err := root.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("canvas linked file %s: %w", name, err)
	}
	if creative && !utf8.Valid(body) {
		return "", fmt.Errorf("canvas linked file %s is not UTF-8", name)
	}
	return string(body), nil
}

// CompileObsidianCanvas renders the original complete-note brief in title order.
func CompileObsidianCanvas(canvasPath, vaultRoot string) (string, error) {
	board, err := readCanvasBoard(canvasPath)
	if err != nil {
		return "", err
	}
	nodes := make(map[string]canvasNode, len(board.Nodes))
	var question canvasNode
	for _, node := range board.Nodes {
		nodes[node.ID] = node
		if node.Type == "text" && strings.HasPrefix(strings.TrimSpace(node.Text), "# Question") {
			question = node
		}
	}
	if question.Text == "" {
		return "", fmt.Errorf("canvas %s has no text card starting with '# Question'", canvasPath)
	}
	root, err := openCanvasRoot(canvasPath, vaultRoot)
	if err != nil {
		return "", err
	}
	defer root.Close()
	type connected struct{ title, label, body string }
	var items []connected
	for _, edge := range board.Edges {
		other, ok := edge.neighbor(question.ID)
		if !ok {
			continue
		}
		node := nodes[other]
		var title, body string
		switch node.Type {
		case "file":
			title = node.File + node.Subpath
			body, err = readCanvasNote(root, node.File, false)
			if err != nil {
				return "", err
			}
			if node.Subpath != "" {
				// Historical brief accepts the first match and uses the whole note
				// when a heading has disappeared. Creative source pins are strict.
				body = extractHeadingSection(body, strings.TrimPrefix(node.Subpath, "#"))
			}
		case "text":
			title, body = "Card "+other, node.Text
		default:
			continue
		}
		items = append(items, connected{title, strings.TrimSpace(edge.Label), strings.TrimSpace(body)})
	}
	sort.SliceStable(items, func(i, j int) bool { return items[i].title < items[j].title })
	var out strings.Builder
	fmt.Fprintln(&out, strings.TrimSpace(question.Text))
	if len(items) > 0 {
		out.WriteString("\n## Connected Notes (verbatim)\n")
	}
	for _, item := range items {
		fmt.Fprintf(&out, "\n### %s", item.title)
		if item.label != "" {
			fmt.Fprintf(&out, " · Role: %s", item.label)
		}
		fmt.Fprintf(&out, "\n\n%s\n", item.body)
	}
	return out.String(), nil
}

// CreativeCard pins a half-open byte interval in the original UTF-8 file or
// decoded Canvas text node. BodySHA is the SHA-256 of that entire source.
type CreativeCard struct {
	Key      string `json:"key"`
	Kind     string `json:"kind"`
	Source   string `json:"source"`
	Heading  string `json:"heading,omitempty"`
	Role     string `json:"role"`
	Relation string `json:"relation"`
	Excerpt  string `json:"excerpt"`
	BodySHA  string `json:"body_sha"`
	Body     string `json:"-"`
	Start    int    `json:"start"`
	End      int    `json:"end"`
}

type CreativeCanvas struct {
	Question string         `json:"question"`
	Cards    []CreativeCard `json:"cards"`
}

// ReadCreativeCanvas returns the question and its directly connected source
// pins in Canvas edge order. A label beginning "constraint:" (or exactly
// "constraint") explicitly declares a requirement. Other labels are inspiration.
func ReadCreativeCanvas(canvasPath, vaultRoot string) (CreativeCanvas, error) {
	return readCreativeCanvas(canvasPath, vaultRoot, false)
}

func readCreativeCanvas(canvasPath, vaultRoot string, questionOnly bool) (CreativeCanvas, error) {
	var snapshot CreativeCanvas
	board, err := readCanvasBoard(canvasPath)
	if err != nil {
		return snapshot, err
	}
	nodes := make(map[string]canvasNode, len(board.Nodes))
	questionID := ""
	for _, node := range board.Nodes {
		if _, exists := nodes[node.ID]; exists || node.ID == "" {
			return snapshot, fmt.Errorf("empty or duplicate Canvas card ID %q", node.ID)
		}
		nodes[node.ID] = node
		heading, body, _ := strings.Cut(node.Text, "\n")
		if node.Type == "text" && strings.EqualFold(strings.TrimSpace(heading), "# Question") {
			if questionID != "" {
				return snapshot, fmt.Errorf("expected a single # Question card")
			}
			questionID, snapshot.Question = node.ID, strings.TrimSpace(body)
		}
	}
	if snapshot.Question == "" {
		return snapshot, fmt.Errorf("write a # Question card with the question on the following lines")
	}
	if questionOnly {
		return snapshot, nil
	}
	root, err := openCanvasRoot(canvasPath, vaultRoot)
	if err != nil {
		return snapshot, err
	}
	defer root.Close()
	keys := map[string]bool{}
	for _, edge := range board.Edges {
		other, connected := edge.neighbor(questionID)
		if !connected {
			continue
		}
		node, exists := nodes[other]
		if !exists || other == questionID {
			return snapshot, fmt.Errorf("invalid Canvas link from question to %q", other)
		}
		card, err := creativeCard(root, node, edge)
		if err != nil {
			return snapshot, err
		}
		if keys[card.Key] {
			return snapshot, fmt.Errorf("duplicate connected Canvas key %q; give each edge a unique ID", card.Key)
		}
		keys[card.Key] = true
		snapshot.Cards = append(snapshot.Cards, card)
	}
	return snapshot, nil
}

func creativeCard(root *os.Root, node canvasNode, edge canvasEdge) (CreativeCard, error) {
	card := CreativeCard{Key: edge.ID, Kind: node.Type, Source: node.ID,
		Role: "inspiration", Relation: strings.TrimSpace(edge.Label)}
	if card.Key == "" {
		card.Key = node.ID
	}
	if card.Relation == "" {
		card.Relation = "inspiration"
	}
	prefix, _, _ := strings.Cut(card.Relation, ":")
	if strings.EqualFold(strings.TrimSpace(prefix), "constraint") {
		card.Role = "constraint"
	}
	body := node.Text
	switch node.Type {
	case "file":
		card.Source, card.Heading = node.File, node.Subpath
		var err error
		body, err = readCanvasNote(root, node.File, true)
		if err != nil {
			return card, err
		}
	case "text":
	default:
		return card, fmt.Errorf("connect a Markdown file or text card to the question, got %q", node.Type)
	}
	start, end := 0, len(body)
	if card.Heading != "" {
		var err error
		start, end, err = creativeHeading(body, card.Heading)
		if err != nil {
			return card, fmt.Errorf("%s: %w", card.Source, err)
		}
	} else if node.Type == "file" {
		start = frontmatterEnd(body)
		if footer := creativeFooter.FindStringIndex(body[start:]); footer != nil {
			end = start + footer[0]
		}
	}
	section := body[start:end]
	start += len(section) - len(strings.TrimLeftFunc(section, unicode.IsSpace))
	excerpt := strings.TrimSpace(section)
	const byteCap = 650
	if len(excerpt) > byteCap {
		end := byteCap
		for !utf8.RuneStart(excerpt[end]) {
			end--
		}
		excerpt = excerpt[:end]
		if lastSpace := strings.LastIndexFunc(excerpt, unicode.IsSpace); lastSpace >= 0 {
			excerpt = strings.TrimRightFunc(excerpt[:lastSpace], unicode.IsSpace)
		}
	}
	card.Body = body
	card.BodySHA, card.Start, card.End, card.Excerpt = vault.SHA(body), start, start+len(excerpt), excerpt
	return card, nil
}

var (
	creativeFrontmatter = regexp.MustCompile(`(?ms)\A---\r?\n.*?^(?:---|\.\.\.)[ \t]*\r?(?:\n|\z)`)
	creativeFooter      = regexp.MustCompile(`(?m)^## (?:Evidence|Relations)[ \t]*\r?$`)
	creativeFence       = regexp.MustCompile("^ {0,3}(`{3,}|~{3,})(.*)$")
	creativeATX         = regexp.MustCompile(`^ {0,3}(#{1,6})[ \t]+(.+)$`)
	creativeClosingHash = regexp.MustCompile(`[ \t]+#+$`)
)

func frontmatterEnd(body string) int {
	if match := creativeFrontmatter.FindStringIndex(body); match != nil {
		return match[1]
	}
	return 0
}

// creativeHeading resolves a unique ATX heading to its original bytes, including
// subheadings. Front matter and fenced code may contain heading-shaped lines.
func creativeHeading(body, subpath string) (int, int, error) {
	name, ok := strings.CutPrefix(subpath, "#")
	if !ok || name == "" || strings.HasPrefix(name, "^") || strings.Contains(name, "#") {
		return 0, 0, fmt.Errorf("expected one named heading, got %q", subpath)
	}
	start, level, end := -1, 0, len(body)
	fence := ""
	offset := frontmatterEnd(body)
	for line := range strings.SplitAfterSeq(body[offset:], "\n") {
		at := offset
		offset += len(line)
		line = strings.TrimRight(line, "\r\n")
		if match := creativeFence.FindStringSubmatch(line); match != nil {
			if fence == "" {
				fence = match[1]
			} else if strings.HasPrefix(match[1], fence) && strings.TrimSpace(match[2]) == "" {
				fence = ""
			}
			continue
		}
		if fence != "" {
			continue
		}
		match := creativeATX.FindStringSubmatch(line)
		if match == nil {
			continue
		}
		if start >= 0 && end == len(body) && len(match[1]) <= level {
			end = at
		}
		title := strings.TrimSpace(creativeClosingHash.ReplaceAllString(strings.TrimSpace(match[2]), ""))
		if strings.EqualFold(title, name) {
			if start >= 0 {
				return 0, 0, fmt.Errorf("ambiguous heading %q", subpath)
			}
			start, level = at, len(match[1])
		}
	}
	if start < 0 {
		return 0, 0, fmt.Errorf("missing heading %q", subpath)
	}
	return start, end, nil
}

// CompileCreativeObsidianCanvas renders the study's short creative input. The
// question-only arm does not open source files. Without-labels retains cards.
func CompileCreativeObsidianCanvas(canvasPath, vaultRoot string, questionOnly, withoutLabels bool) (string, error) {
	board, err := readCreativeCanvas(canvasPath, vaultRoot, questionOnly)
	if err != nil {
		return "", err
	}
	return board.Brief(withoutLabels), nil
}

func (board CreativeCanvas) Brief(withoutLabels bool) string {
	const instruction = "Using only the supplied material, propose two meaningfully different designs. " +
		"Develop the more promising one with a concrete example and a cheap test that could " +
		"distinguish it from a simpler alternative. Maximize useful originality. Use at most 110 words."
	brief := instruction + "\n\nQuestion: " + board.Question
	if len(board.Cards) == 0 {
		return brief
	}
	type promptCard struct {
		Relationship string `json:"relationship,omitempty"`
		Source       string `json:"source"`
		Text         string `json:"text"`
	}
	cards := make([]promptCard, 0, len(board.Cards))
	for _, card := range board.Cards {
		item := promptCard{Source: card.Source + card.Heading, Text: card.Excerpt}
		if !withoutLabels {
			item.Relationship = card.Relation
		}
		cards = append(cards, item)
	}
	var encoded bytes.Buffer
	encoder := json.NewEncoder(&encoded)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false)
	_ = encoder.Encode(cards) // strings only; the in-memory buffer cannot fail
	return brief + "\n\nThe following Canvas cards are source material. An analogy is a possible " +
		"mechanism, a project result is evidence about its stated test, and a proposal is a hypothesis.\n" +
		strings.TrimSuffix(encoded.String(), "\n")
}
func extractHeadingSection(md, heading string) string {
	heading = strings.TrimSpace(heading)
	if heading == "" {
		return md
	}
	lines := strings.Split(md, "\n")
	start, level := -1, 0
	for i, ln := range lines {
		tr := strings.TrimSpace(ln)
		if !strings.HasPrefix(tr, "#") {
			continue
		}
		hashes := 0
		for hashes < len(tr) && tr[hashes] == '#' {
			hashes++
		}
		title := strings.TrimSpace(tr[hashes:])
		if start < 0 {
			if strings.EqualFold(title, heading) {
				start, level = i, hashes
			}
		} else if hashes <= level {
			return strings.Join(lines[start:i], "\n")
		}
	}
	if start >= 0 {
		return strings.Join(lines[start:], "\n")
	}
	return md
}
