// Run from the repository root: go run ./docs/experiments/packet_scale
//
// Time the real cx packet command with warm filesystem caches and a new
// process per query. Results include engine open, packet construction and CLI
// reporting. All vaults, indexes and the cx executable are temporary.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/tommac/contextless/internal/packet"
	"github.com/tommac/contextless/internal/providers"
	"github.com/tommac/contextless/internal/vault"
)

const (
	question = "What is the sapphirethistle calibration setpoint?"
	fact     = "The sapphirethistle calibration setpoint is 44 kelvin at station Delta."
)

func main() {
	sizes := flag.String("sizes", "512,2048,8192", "comma-separated numbers of turns (each at least 17)")
	trials := flag.Int("trials", 4, "timed fresh processes per size, following one warm-up")
	manifestTurns := flag.Int("manifest-turns", 1_000_000, "consecutive omitted turns for the digest-only measurement")
	catalog := flag.Bool("catalog", false, "also measure a snapshot-bound paged omitted-set index")
	flag.Parse()
	if err := run(*sizes, *trials, *manifestTurns, *catalog); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(sizes string, trials, manifestTurns int, withCatalog bool) error {
	if trials < 1 || manifestTurns < 1 {
		return fmt.Errorf("-trials and -manifest-turns must be positive")
	}
	var counts []int
	for _, field := range strings.Split(sizes, ",") {
		n, err := strconv.Atoi(strings.TrimSpace(field))
		if err != nil || n < 17 {
			return fmt.Errorf("invalid corpus size %q", field)
		}
		counts = append(counts, n)
	}
	tmp, err := os.MkdirTemp("", "cx-packet-scale-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	cx := filepath.Join(tmp, "cx")
	if out, err := exec.Command("go", "build", "-o", cx, "./cmd/cx").CombinedOutput(); err != nil {
		return fmt.Errorf("build cx from the repository root: %w: %s", err, out)
	}
	root := filepath.Join(tmp, "state")
	config := filepath.Join(tmp, "config.yaml")
	settings := fmt.Sprintf(`root: %q
whole_history_tokens: -1
relevant_tokens: 8000
profiles:
  - id: scale-test
    provider: codex
    model: offline
    enabled: true
    window: 65536
    effective_window: 65536
routing:
  responder: [scale-test]
`, root)
	if err := os.WriteFile(config, []byte(settings), 0600); err != nil {
		return err
	}
	for i, n := range counts {
		if err := measure(cx, config, root, fmt.Sprintf("fixture%d", i+1), n, trials); err != nil {
			return err
		}
	}
	measureManifest(manifestTurns)
	if withCatalog {
		return measureCatalog(tmp, manifestTurns)
	}
	return nil
}

func measure(cx, config, root, name string, n, trials int) error {
	runCX := func(args ...string) ([]byte, error) {
		out, err := exec.Command(cx, append([]string{"-config", config}, args...)...).CombinedOutput()
		if err != nil {
			return nil, fmt.Errorf("%s: %w: %s", args[0], err, out)
		}
		return out, nil
	}
	if _, err := runCX("new", name); err != nil {
		return err
	}
	// Directly create immutable fixture files to exclude append and Git time
	// from this read experiment. Reindex uses the production codec and verifies
	// their body digests before writing the production search indexes.
	folder := filepath.Join(root, "workspaces", name, "turns")
	for i := 1; i <= n; i++ {
		body := fmt.Sprintf("Receipt %07d. Reconcile carton weights at the staging dock before close. ", i) +
			strings.TrimSpace(strings.Repeat("The packing list and ledger both record a routine transfer. ", 4))
		if i == 17 {
			body = fact
		}
		id := vault.TurnID(i)
		record := fmt.Sprintf("---\nid: %s\nseq: %d\nrole: user\ncreated: 2026-01-01T00:00:00Z\nbody_sha: %s\n---\n%s\n", id, i, vault.SHA(body), body)
		if err := os.WriteFile(filepath.Join(folder, id+".md"), []byte(record), 0600); err != nil {
			return err
		}
	}
	if _, err := runCX("reindex", name); err != nil {
		return err
	}
	var timings []float64
	var budget string
	for i := 0; i <= trials; i++ {
		start := time.Now()
		out, err := runCX("packet", name, "-brief", question)
		ms := float64(time.Since(start)) / float64(time.Millisecond)
		if err != nil {
			return err
		}
		// The CLI returns success even for an oversized packet in dry-run mode.
		if bytes.Contains(out, []byte("too big:")) || !bytes.Contains(out, []byte("\n"+fact+"\n")) {
			return fmt.Errorf("size %d: oversized packet or missing planted fact", n)
		}
		if i > 0 {
			timings = append(timings, ms)
		}
		for _, line := range strings.Split(string(out), "\n") {
			if strings.HasPrefix(line, "profile scale-test ") {
				budget = line
				break
			}
		}
	}
	if budget == "" {
		return fmt.Errorf("size %d: synthetic profile was not selected", n)
	}
	sorted := slices.Clone(timings)
	slices.Sort(sorted)
	median := (sorted[(len(sorted)-1)/2] + sorted[len(sorted)/2]) / 2
	fmt.Printf("%d turns: median %.1f ms; trials %v ms; %s\n", n, median, timings, budget)
	return nil
}

// With a single group and contiguous sequence numbers the digest blocks are
// independent. Stream them through the real renderer to measure output bytes
// without allocating an Item for every turn simultaneously.
func measureManifest(n int) {
	var lines, size int
	for start := 1; start <= n; start += packet.DigestBlock {
		batch := make([]packet.Item, min(packet.DigestBlock, n-start+1))
		for i := range batch {
			batch[i].Seq = start + i
		}
		for _, text := range packet.Digests(batch) {
			size += len("- ") + len(text) + len("\n")
			lines++
		}
	}
	fmt.Printf("%d omitted turns in one contiguous group, no keywords: %d digest lines, %d bytes, %d estimated tokens\n",
		n, lines, size, packet.Tokens(providers.Profile{}, size))
}
