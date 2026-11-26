// Package ui embeds a zero-dependency single-page Context X-Ray web dashboard
// served directly from the cx binary (`cx ui <workspace>`). It visualizes
// compiled packets, ranking lanes, cache-prefix savings, causal ablations,
// Living Answer receipts with verified diffs, typed recall verbs, and archive
// telemetry.
package ui

import (
	"context"
	"encoding/json"
	"mime"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/tommac/contextless/internal/engine"
	"github.com/tommac/contextless/internal/packet"
	"github.com/tommac/contextless/internal/providers"
	"github.com/tommac/contextless/internal/store"
	"github.com/tommac/contextless/internal/vault"
)

// localHosts are the names a browser on this machine reaches the dashboard by.
var localHosts = []string{"localhost", "127.0.0.1", "::1"}

// hostOf is the host of a Host header or URL host, without its port.
func hostOf(hostport string) string {
	if h, _, err := net.SplitHostPort(hostport); err == nil {
		return strings.ToLower(h)
	}
	return strings.ToLower(strings.Trim(hostport, "[]"))
}

// guard keeps other web pages away from the dashboard, which serves the
// conversation's text and runs queries that write files. A page on another
// site can make a browser send requests to a local port, so:
//
//   - the Host header must name this server (a loopback name, or the host
//     the user chose to listen on). A request that arrives under any other
//     name is a DNS rebinding attempt or a mistake, and gets 403.
//   - a request that changes or computes something (anything but GET, HEAD
//     and OPTIONS) must carry a JSON body type, which a cross-site form
//     cannot send without a preflight this server never approves, and an
//     Origin, when the browser sends one, that is this server.
//   - a browser that says the request comes from another site
//     (Sec-Fetch-Site) is refused for the API whatever the method.
func guard(next http.Handler, allowed map[string]bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		deny := func(why string) {
			http.Error(w, "forbidden: "+why, http.StatusForbidden)
		}
		if !allowed[hostOf(r.Host)] {
			deny("this dashboard answers only under the address it listens on")
			return
		}
		api := strings.HasPrefix(r.URL.Path, "/api/")
		if site := r.Header.Get("Sec-Fetch-Site"); api && site != "" && site != "same-origin" && site != "none" {
			deny("cross-site request")
			return
		}
		if origin := r.Header.Get("Origin"); origin != "" {
			u, err := url.Parse(origin)
			if err != nil || !allowed[hostOf(u.Host)] {
				deny("cross-origin request")
				return
			}
		}
		if r.Method != http.MethodGet && r.Method != http.MethodHead && r.Method != http.MethodOptions {
			ct, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
			if ct != "application/json" {
				deny("requests with a body must be application/json")
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

// NewHandler builds the HTTP handler for the Context X-Ray dashboard backed
// by e, answering only under a loopback name.
func NewHandler(e *engine.Engine) http.Handler { return NewHandlerFor(e, "") }

// NewHandlerFor is NewHandler for a server listening on addr: when addr
// names a host that is not loopback (the user asked for it), requests under
// that host are answered too.
func NewHandlerFor(e *engine.Engine, addr string) http.Handler {
	allowed := map[string]bool{}
	for _, h := range localHosts {
		allowed[h] = true
	}
	if h := hostOf(addr); h != "" && h != "0.0.0.0" && h != "::" {
		allowed[h] = true
	}
	return guard(newMux(e), allowed)
}

func newMux(e *engine.Engine) http.Handler {
	var explainMu sync.Mutex
	mux := http.NewServeMux()

	writeJSON := func(w http.ResponseWriter, code int, v any) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(code)
		_ = json.NewEncoder(w).Encode(v)
	}

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(indexHTML))
	})

	mux.HandleFunc("GET /api/overview", func(w http.ResponseWriter, r *http.Request) {
		sess, _ := e.V.Session()
		turns, _ := e.V.Turns()
		var rcpts []store.Receipt
		var sigs map[string]int
		var stats []store.Stat
		if e.S != nil {
			rcpts, _ = e.S.ListReceipts("")
			sigs, _ = e.S.SignalSummary()
			stats = e.S.Stats()
		}
		rcptCounts := map[string]int{"current": 0, "stale": 0, "superseded": 0}
		for _, rc := range rcpts {
			rcptCounts[rc.Status]++
		}
		snapSeq, snapRoot := 0, ""
		if e.S != nil {
			snapSeq, snapRoot, _ = e.S.SnapshotRoot(0)
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"session":        sess,
			"turn_count":     len(turns),
			"fact_count":     len(e.V.Facts()),
			"decision_count": len(e.V.Decisions()),
			"thread_count":   len(e.V.Threads()),
			"entity_count":   len(e.V.Entities()),
			"receipt_counts": rcptCounts,
			"snapshot_seq":   snapSeq,
			"snapshot_root":  snapRoot,
			"signals":        sigs,
			"stats":          stats,
			"config": map[string]any{
				"relevant_tokens":           e.Cfg.RelevantTokens,
				"whole_history_tokens":      e.Cfg.WholeHistoryTokens,
				"recent_turns":              e.Cfg.RecentTurns,
				"adaptive_pack":             e.Cfg.AdaptivePack,
				"working_set_window":        e.Cfg.WorkingSetWindow,
				"omitted_catalog_threshold": e.Cfg.OmittedCatalogThreshold,
				"stable_workspace":          e.Cfg.StableWorkspace,
				"stats_routing":             e.Cfg.StatsRouting,
			},
		})
	})

	mux.HandleFunc("GET /api/turns", func(w http.ResponseWriter, r *http.Request) {
		turns, err := e.V.Turns()
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		q := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))
		type turnRow struct {
			ID        string    `json:"id"`
			Seq       int       `json:"seq"`
			Role      string    `json:"role"`
			Author    string    `json:"author,omitempty"`
			Created   time.Time `json:"created"`
			Bytes     int       `json:"bytes"`
			Tokens    int       `json:"tokens"`
			AtomCount int       `json:"atom_count"`
			BodySHA   string    `json:"body_sha"`
			Preview   string    `json:"preview"`
		}
		rows := make([]turnRow, 0, len(turns))
		for i := len(turns) - 1; i >= 0; i-- {
			t := turns[i]
			if q != "" && !strings.Contains(strings.ToLower(t.Body), q) && !strings.Contains(strings.ToLower(t.ID), q) {
				continue
			}
			prev := strings.Join(strings.Fields(t.Body), " ")
			if rs := []rune(prev); len(rs) > 160 {
				prev = string(rs[:160]) + "…"
			}
			rows = append(rows, turnRow{
				ID:        t.ID,
				Seq:       t.Seq,
				Role:      t.Role,
				Author:    t.Author,
				Created:   t.Created,
				Bytes:     len(t.Body),
				Tokens:    packet.Tokens(providers.Profile{}, len(t.Body)),
				AtomCount: len(packet.AtomsOf(t)),
				BodySHA:   t.BodySHA,
				Preview:   prev,
			})
		}
		writeJSON(w, http.StatusOK, map[string]any{"turns": rows, "total": len(turns)})
	})

	mux.HandleFunc("GET /api/turn/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		ref, ok := vault.ParseRef(id)
		if !ok {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid turn/atom id"})
			return
		}
		t, err := e.V.Turn(ref.Turn)
		if err != nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
			return
		}
		atoms := packet.AtomsOf(t)
		type atomRow struct {
			ID     string `json:"id"`
			Index  int    `json:"index"`
			Start  int    `json:"start"`
			End    int    `json:"end"`
			Tokens int    `json:"tokens"`
			Text   string `json:"text"`
		}
		arows := make([]atomRow, 0, len(atoms))
		for _, a := range atoms {
			txt := a.Text(t.Body)
			arows = append(arows, atomRow{
				ID:     a.ID(),
				Index:  a.K,
				Start:  a.Start,
				End:    a.End,
				Tokens: packet.Tokens(providers.Profile{}, len(txt)),
				Text:   txt,
			})
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"turn":       t,
			"atoms":      arows,
			"focus_atom": ref.Atom,
		})
	})

	mux.HandleFunc("POST /api/explain", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Query    string   `json:"query"`
			Adaptive *bool    `json:"adaptive,omitempty"`
			DropIDs  []string `json:"drop_ids,omitempty"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		if strings.TrimSpace(req.Query) == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "query is required"})
			return
		}
		explainMu.Lock()
		oldAdaptive := e.Cfg.AdaptivePack
		if req.Adaptive != nil {
			e.Cfg.AdaptivePack = *req.Adaptive
		}
		plan, err := e.Explain(req.Query, req.DropIDs)
		e.Cfg.AdaptivePack = oldAdaptive
		explainMu.Unlock()
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		// Attach verbatim previews for top candidates so the UI renders inspectable rows without extra round-trips.
		snippets := map[string]string{}
		for _, c := range plan.TopCandidates {
			if ref, ok := vault.ParseRef(c.ID); ok {
				if t, terr := e.V.Turn(ref.Turn); terr == nil {
					txt := t.Body
					if ref.Atom > 0 {
						atoms := packet.AtomsOf(t)
						if ref.Atom <= len(atoms) {
							txt = atoms[ref.Atom-1].Text(t.Body)
						}
					}
					prev := strings.Join(strings.Fields(txt), " ")
					if rs := []rune(prev); len(rs) > 220 {
						prev = string(rs[:220]) + "…"
					}
					snippets[c.ID] = prev
				}
			}
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"plan":     plan,
			"text":     packet.FormatExplain(plan),
			"snippets": snippets,
		})
	})

	mux.HandleFunc("POST /api/recall", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Op     string `json:"op"`
			Target string `json:"target"`
			Radius int    `json:"radius,omitempty"`
			Limit  int    `json:"limit,omitempty"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		lastSeq, _ := e.V.LastSeq()
		focus := vault.Turn{Seq: lastSeq + 1}
		ids, extras, notes := e.ExecuteRecallOps([]packet.RecallOp{{
			Op: req.Op, Target: req.Target, Radius: req.Radius, Limit: req.Limit,
		}}, focus, nil)
		type excerpt struct {
			ID   string `json:"id"`
			Text string `json:"text"`
		}
		var excerpts []excerpt
		for _, id := range ids {
			ref, ok := vault.ParseRef(id)
			if !ok {
				continue
			}
			t, err := e.V.Turn(ref.Turn)
			if err != nil {
				continue
			}
			if ref.Whole() {
				excerpts = append(excerpts, excerpt{ID: id, Text: t.Body})
			} else {
				atoms := packet.AtomsOf(t)
				lo := max(1, ref.Atom)
				hi := min(len(atoms), ref.AtomEnd)
				if hi >= lo && lo <= len(atoms) {
					excerpts = append(excerpts, excerpt{ID: id, Text: t.Body[atoms[lo-1].Start:atoms[hi-1].End]})
				}
			}
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"pinned_ids": ids,
			"extras":     extras,
			"notes":      notes,
			"excerpts":   excerpts,
		})
	})

	mux.HandleFunc("GET /api/receipts", func(w http.ResponseWriter, r *http.Request) {
		if e.S == nil {
			writeJSON(w, http.StatusOK, map[string]any{"receipts": []store.Receipt{}})
			return
		}
		status := strings.TrimSpace(r.URL.Query().Get("status"))
		list, err := e.S.ListReceipts(status)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"receipts": list})
	})

	mux.HandleFunc("GET /api/receipt/{id}/diff", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		diff, err := e.DiffReceiptSince(id)
		if err != nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, diff)
	})

	mux.HandleFunc("GET /api/runs", func(w http.ResponseWriter, r *http.Request) {
		if e.S == nil {
			writeJSON(w, http.StatusOK, map[string]any{"runs": []store.Run{}})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"runs": e.S.RecentRuns(40)})
	})

	mux.HandleFunc("POST /api/ledger-query", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ResumeID string                 `json:"resume_id,omitempty"`
			Spec     engine.LedgerQuerySpec `json:"spec"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		var (
			res engine.LedgerQueryResult
			err error
		)
		if strings.TrimSpace(req.ResumeID) != "" {
			res, err = e.ResumeLedgerQuery(strings.TrimSpace(req.ResumeID))
		} else {
			res, err = e.RunLedgerQuery(req.Spec)
		}
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"result": res, "text": engine.FormatLedgerQuery(res)})
	})

	mux.HandleFunc("GET /api/views", func(w http.ResponseWriter, r *http.Request) {
		if e.S == nil {
			writeJSON(w, http.StatusOK, map[string]any{"active_keys": []store.KeyStateRow{}})
			return
		}
		key := strings.TrimSpace(r.URL.Query().Get("key"))
		filter := strings.TrimSpace(r.URL.Query().Get("filter"))
		asOf, _ := strconv.Atoi(r.URL.Query().Get("as_of"))
		if key != "" {
			rows, err := e.S.QueryKeyHistory(key, asOf)
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"key": key, "as_of": asOf, "history": rows})
			return
		}
		rows, err := e.S.ListActiveKeys(filter, asOf, 200)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"as_of": asOf, "active_keys": rows})
	})

	mux.HandleFunc("GET /api/stats", func(w http.ResponseWriter, r *http.Request) {
		if e.S == nil {
			writeJSON(w, http.StatusOK, map[string]any{"stats": []store.Stat{}})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"stats": e.S.Stats()})
	})

	return mux
}

// Serve runs the local Context X-Ray web dashboard until ctx is canceled.
func Serve(ctx context.Context, e *engine.Engine, addr string) error {
	if strings.TrimSpace(addr) == "" {
		addr = "127.0.0.1:8787"
	}
	srv := &http.Server{
		Addr:              addr,
		Handler:           NewHandlerFor(e, addr),
		ReadHeaderTimeout: 10 * time.Second,
	}
	errCh := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
		close(errCh)
	}()
	select {
	case <-ctx.Done():
		shutCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutCtx)
		return nil
	case err := <-errCh:
		return err
	}
}

const indexHTML = `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>ContextlessAI · Context X-Ray</title>
<style>
  :root {
    --bg: #0b0f17; --panel: #131926; --panel-2: #192234; --border: #25324a;
    --text: #e6edf7; --muted: #93a4bd; --accent: #38bdf8; --ok: #34d399;
    --warn: #fbbf24; --bad: #f87171; --purple: #a78bfa;
    --mono: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
    --sans: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
  }
  * { box-sizing: border-box; }
  body { margin: 0; font-family: var(--sans); background: var(--bg); color: var(--text); line-height: 1.5; }
  header { display: flex; flex-wrap: wrap; align-items: center; justify-content: space-between; gap: 12px; padding: 14px 24px; background: var(--panel); border-bottom: 1px solid var(--border); position: sticky; top: 0; z-index: 10; }
  header h1 { margin: 0; font-size: 16px; letter-spacing: .02em; display: flex; align-items: center; gap: 10px; }
  .pill { font-size: 11px; padding: 2px 8px; border-radius: 999px; background: var(--panel-2); border: 1px solid var(--border); color: var(--muted); font-family: var(--mono); }
  .pill.ok { color: var(--ok); border-color: rgba(52,211,153,.35); background: rgba(52,211,153,.08); }
  .pill.warn { color: var(--warn); border-color: rgba(251,191,36,.35); background: rgba(251,191,36,.08); }
  .pill.bad { color: var(--bad); border-color: rgba(248,113,113,.35); background: rgba(248,113,113,.08); }
  .pill.accent { color: var(--accent); border-color: rgba(56,189,248,.35); background: rgba(56,189,248,.08); }
  nav { display: flex; gap: 6px; }
  nav button { background: transparent; color: var(--muted); border: 1px solid transparent; padding: 6px 12px; border-radius: 6px; cursor: pointer; font-size: 13px; font-weight: 500; }
  nav button.active { background: var(--panel-2); color: var(--text); border-color: var(--border); }
  main { max-width: 1280px; margin: 0 auto; padding: 20px 24px 60px; }
  .grid { display: grid; gap: 14px; }
  .grid-4 { grid-template-columns: repeat(auto-fit, minmax(190px, 1fr)); }
  .grid-2 { grid-template-columns: repeat(auto-fit, minmax(380px, 1fr)); }
  .card { background: var(--panel); border: 1px solid var(--border); border-radius: 10px; padding: 16px; }
  .card h2 { margin: 0 0 10px; font-size: 14px; color: var(--muted); text-transform: uppercase; letter-spacing: .06em; }
  .metric-val { font-size: 24px; font-weight: 700; font-family: var(--mono); }
  .metric-sub { font-size: 12px; color: var(--muted); margin-top: 2px; }
  .row { display: flex; flex-wrap: wrap; gap: 10px; align-items: center; }
  input[type=text], select { background: var(--bg); color: var(--text); border: 1px solid var(--border); border-radius: 7px; padding: 9px 12px; font-size: 13px; }
  input[type=text] { flex: 1; min-width: 220px; }
  button.primary { background: var(--accent); color: #05131f; border: none; border-radius: 7px; padding: 9px 16px; font-weight: 600; cursor: pointer; font-size: 13px; }
  button.ghost { background: var(--panel-2); color: var(--text); border: 1px solid var(--border); border-radius: 6px; padding: 4px 9px; font-size: 11px; cursor: pointer; font-family: var(--mono); }
  button.ghost:hover { border-color: var(--accent); color: var(--accent); }
  .bar { display: flex; height: 22px; border-radius: 6px; overflow: hidden; background: var(--bg); border: 1px solid var(--border); margin: 10px 0; }
  .bar > div { height: 100%; transition: width .2s; }
  .legend { display: flex; flex-wrap: wrap; gap: 10px; font-size: 12px; color: var(--muted); }
  .dot { display: inline-block; width: 10px; height: 10px; border-radius: 3px; margin-right: 5px; vertical-align: middle; }
  table { width: 100%; border-collapse: collapse; font-size: 13px; }
  th, td { text-align: left; padding: 8px 10px; border-bottom: 1px solid var(--border); vertical-align: top; }
  th { color: var(--muted); font-size: 11px; text-transform: uppercase; letter-spacing: .05em; }
  pre.code { background: var(--bg); border: 1px solid var(--border); border-radius: 8px; padding: 12px; overflow-x: auto; font-family: var(--mono); font-size: 12px; white-space: pre-wrap; margin: 8px 0 0; max-height: 360px; overflow-y: auto; }
  .hidden { display: none; }
  .clickable { cursor: pointer; }
  .clickable:hover { background: rgba(56,189,248,.06); }
  label.check { display: flex; align-items: center; gap: 6px; font-size: 13px; color: var(--muted); }
</style>
</head>
<body>
<header>
  <h1>
    <span>ContextlessAI</span>
    <span class="pill accent">Context X-Ray</span>
    <span id="hdr-title" class="pill">loading…</span>
    <span id="hdr-snap" class="pill"></span>
  </h1>
  <nav>
    <button class="active" data-tab="xray">Packet X-Ray</button>
    <button data-tab="receipts">Living Answers <span id="stale-badge"></span></button>
    <button data-tab="archive">Archive &amp; Recall Lab</button>
  </nav>
</header>

<main>
  <section id="tab-xray" class="grid">
    <div class="card">
      <h2>Context Query Planner (Zero-Model Dry Run)</h2>
      <div class="row">
        <input id="q-input" type="text" placeholder="Ask a question to inspect packet compilation, ranking lanes, cache prefix, and causal drops…">
        <input id="drop-input" type="text" style="max-width:240px" placeholder="Causal -drop IDs (e.g. t000001)">
        <label class="check"><input id="adaptive-chk" type="checkbox"> Adaptive Pack</label>
        <button class="primary" id="btn-explain">Explain Packet</button>
      </div>
    </div>

    <div id="xray-results" class="grid hidden">
      <div class="grid grid-4" id="xray-metrics"></div>
      <div class="card">
        <h2>Section Token Breakdown</h2>
        <div class="bar" id="sec-bar"></div>
        <div class="legend" id="sec-legend"></div>
        <div id="prefix-callout" class="metric-sub" style="margin-top:12px"></div>
      </div>
      <div class="card">
        <h2>Top Ranked Atoms &amp; Causal Controls</h2>
        <table>
          <thead><tr><th>Atom</th><th>Score</th><th>Status</th><th>Lanes / Reason</th><th>Verbatim Snippet</th><th>Inspect</th></tr></thead>
          <tbody id="cand-body"></tbody>
        </table>
      </div>
      <div class="grid grid-2">
        <div class="card">
          <h2>Profile Cost &amp; Latency Estimates</h2>
          <table>
            <thead><tr><th>Profile</th><th>Fits</th><th>Planned / Cap</th><th>Uncached</th><th>Cached</th><th>History</th></tr></thead>
            <tbody id="prof-body"></tbody>
          </table>
        </div>
        <div class="card">
          <h2>Verbatim Inspector</h2>
          <div id="inspector-meta" class="metric-sub">Click any atom's Open Block or Exchange button above to inspect lossless vault text.</div>
          <pre id="inspector-pre" class="code hidden"></pre>
        </div>
      </div>
    </div>
  </section>

  <section id="tab-receipts" class="grid hidden">
    <div class="grid grid-2">
      <div class="card">
        <h2>Query Receipts (Snapshot-Bound Answers)</h2>
        <table>
          <thead><tr><th>ID</th><th>Status</th><th>Snapshot</th><th>Question</th></tr></thead>
          <tbody id="rcpt-body"></tbody>
        </table>
      </div>
      <div class="card">
        <h2>Snapshot Verification &amp; Delta Diff</h2>
        <div id="rcpt-detail" class="metric-sub">Select a receipt on the left to verify its original byte anchors and inspect delta turns.</div>
      </div>
    </div>
  </section>

  <section id="tab-archive" class="grid hidden">
    <div class="card">
      <h2>Typed Recall Lab (Zero-LLM Deterministic Verbs)</h2>
      <div class="row">
        <select id="rec-op">
          <option value="grep">grep (literal or /regex/)</option>
          <option value="timeline">timeline (facts &amp; decisions)</option>
          <option value="open_block">open_block (t000001#a1)</option>
          <option value="exchange">exchange (t000001)</option>
          <option value="key_history">key_history (bitemporal key)</option>
          <option value="bridge">bridge (2-hop PPR from turn)</option>
          <option value="cover_window">cover_window (minimal span join)</option>
          <option value="list_omitted">list_omitted (catalog slice)</option>
          <option value="check_omitted">check_omitted (keyword probe)</option>
        </select>
        <input id="rec-target" type="text" placeholder="Target (pattern, topic, atom ID, or turn ID)">
        <button class="primary" id="btn-recall">Run Verb</button>
      </div>
      <div id="rec-notes" class="metric-sub" style="margin-top:8px"></div>
      <pre id="rec-out" class="code hidden"></pre>
    </div>

    <div class="grid grid-2">
      <div class="card">
        <h2>Verbatim Turns Archive</h2>
        <div class="row" style="margin-bottom:10px">
          <input id="turn-filter" type="text" placeholder="Filter turns by text or ID…">
          <button class="ghost" id="btn-filter">Filter</button>
        </div>
        <table>
          <thead><tr><th>Turn</th><th>Role</th><th>Atoms</th><th>Tokens</th><th>Preview</th></tr></thead>
          <tbody id="turn-body"></tbody>
        </table>
      </div>
      <div class="card">
        <h2>Turn Verbatim Atoms &amp; Telemetry</h2>
        <div id="turn-detail" class="metric-sub">Click a turn to inspect its immutable atoms, byte spans, and SHA-256 hash.</div>
        <h2 style="margin-top:18px">Provider Stats</h2>
        <table>
          <thead><tr><th>Profile</th><th>Kind</th><th>Runs</th><th>OK</th><th>Median</th></tr></thead>
          <tbody id="stats-body"></tbody>
        </table>
      </div>
    </div>
  </section>
</main>

<script>
const COLORS = {
  session:'#38bdf8', decisions:'#a78bfa', recent:'#34d399', relevant:'#fbbf24',
  threads:'#fb7185', entities:'#f472b6', facts:'#2dd4bf', catalog:'#64748b',
  recalled:'#f97316', extra:'#c084fc', focus:'#e2e8f0'
};
const esc = s => String(s ?? '').replace(/[&<>"]/g, c => ({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;'}[c]));

document.querySelectorAll('nav button').forEach(btn => {
  btn.onclick = () => {
    document.querySelectorAll('nav button').forEach(b => b.classList.remove('active'));
    btn.classList.add('active');
    ['xray','receipts','archive'].forEach(t => {
      document.getElementById('tab-'+t).classList.toggle('hidden', t !== btn.dataset.tab);
    });
    if (btn.dataset.tab === 'receipts') loadReceipts();
    if (btn.dataset.tab === 'archive') loadTurns();
  };
});

async function loadOverview() {
  const res = await fetch('/api/overview');
  const d = await res.json();
  document.getElementById('hdr-title').textContent = (d.session.title || 'Workspace') + ' · ' + d.turn_count + ' turns';
  document.getElementById('hdr-snap').textContent = 'snap #' + d.snapshot_seq + (d.snapshot_root ? ' · ' + d.snapshot_root.slice(0,10) : '');
  document.getElementById('adaptive-chk').checked = !!d.config.adaptive_pack;
  if (d.receipt_counts && d.receipt_counts.stale > 0) {
    document.getElementById('stale-badge').innerHTML = ' <span class="pill warn">' + d.receipt_counts.stale + ' stale</span>';
  }
  const sb = document.getElementById('stats-body');
  sb.innerHTML = (d.stats || []).length ? d.stats.map(s =>
    '<tr><td><code>'+esc(s.Profile)+'</code></td><td>'+esc(s.Kind)+'</td><td>'+s.Runs+'</td><td>'+Math.round(100*s.OK/Math.max(1,s.Runs))+'%</td><td>'+(s.MedianWall/1000).toFixed(1)+'s</td></tr>'
  ).join('') : '<tr><td colspan="5" class="metric-sub">No provider runs recorded yet.</td></tr>';
}

async function runExplain() {
  const query = document.getElementById('q-input').value.trim();
  if (!query) return;
  const dropRaw = document.getElementById('drop-input').value.trim();
  const drop_ids = dropRaw ? dropRaw.split(',').map(x => x.trim()).filter(Boolean) : [];
  const adaptive = document.getElementById('adaptive-chk').checked;
  const res = await fetch('/api/explain', {
    method: 'POST', headers: {'Content-Type':'application/json'},
    body: JSON.stringify({query, adaptive, drop_ids})
  });
  const d = await res.json();
  if (d.error) { alert(d.error); return; }
  const p = d.plan;
  document.getElementById('xray-results').classList.remove('hidden');
  const optStr = p.submodular_cert ? ('>= ' + p.submodular_cert.optimality_pct.toFixed(1) + '% OPT') : 'whole / n.a.';
  document.getElementById('xray-metrics').innerHTML = [
    ['Query Class', p.class, 'profile: ' + p.selected_profile],
    ['Planned Budget', p.budget.planned + ' / ' + p.budget.cap, 'brief ' + p.budget.brief + ' tok · cert ' + optStr],
    ['Cache Prefix Share', p.cache_share_pct + '%', p.cache_prefix_tokens + ' cacheable tokens · ' + (p.cache_advice?.recommend_stable ? 'stable prefix pays off' : 'sliding window')],
    ['Omitted Records', String(p.omitted_count), p.omitted_handle ? 'paged catalog attached' : 'inline catalog / none']
  ].map(([k,v,sub]) => '<div class="card"><h2>'+esc(k)+'</h2><div class="metric-val">'+esc(v)+'</div><div class="metric-sub">'+esc(sub)+'</div></div>').join('');

  const totalSec = (p.sections || []).reduce((a,s) => a + s.tokens, 0) || 1;
  document.getElementById('sec-bar').innerHTML = (p.sections || []).map(s => {
    const pct = Math.max(1, (100 * s.tokens / totalSec));
    const col = COLORS[s.key] || '#38bdf8';
    return '<div title="'+esc(s.key)+': '+s.tokens+' tok ('+s.count+')" style="width:'+pct+'%;background:'+col+'"></div>';
  }).join('');
  document.getElementById('sec-legend').innerHTML = (p.sections || []).map(s =>
    '<span><i class="dot" style="background:'+(COLORS[s.key]||'#38bdf8')+'"></i>'+esc(s.key)+' '+s.tokens+'t ('+s.count+')</span>'
  ).join('');
  document.getElementById('prefix-callout').textContent = p.prefix_opt
    ? 'Prefix-Closed Optimizer: ' + p.prefix_opt.prefix_name + ' (' + p.prefix_opt.prefix_tokens + ' prefix tok + ' + p.prefix_opt.suffix_tokens + ' suffix tok · billed ' + p.prefix_opt.billed_tokens.toFixed(0) + ' · utility ' + p.prefix_opt.utility.toFixed(2) + ')'
    : (p.cache_advice?.reason || '');

  document.getElementById('cand-body').innerHTML = (p.top_candidates || []).map(c =>
    '<tr><td><code>'+esc(c.id)+'</code></td><td>'+c.score.toFixed(4)+'</td>' +
    '<td><span class="pill '+(c.included?'ok':'warn')+'">'+(c.included?'included':'omitted')+'</span></td>' +
    '<td>'+esc(c.reason)+(c.provenance?'<div class="metric-sub">'+esc(c.provenance)+'</div>':'')+'</td><td class="metric-sub">'+esc(d.snippets[c.id]||'')+'</td>' +
    '<td class="row"><button class="ghost" onclick="inspectVerb(\'open_block\',\''+esc(c.id)+'\')">open_block</button>' +
    '<button class="ghost" onclick="inspectVerb(\'exchange\',\''+esc(c.turn)+'\')">exchange</button>' +
    '<button class="ghost" onclick="ablate(\''+esc(c.turn)+'\')">-drop</button></td></tr>'
  ).join('') || '<tr><td colspan="6" class="metric-sub">No ranked atoms (whole history sent verbatim or empty archive).</td></tr>';

  document.getElementById('prof-body').innerHTML = (p.profile_estimates || []).map(pe =>
    '<tr><td><code>'+esc(pe.profile)+'</code></td>' +
    '<td><span class="pill '+(pe.fits?'ok':'bad')+'">'+(pe.fits?'fits':'over')+'</span></td>' +
    '<td>'+pe.planned_tokens+' / '+pe.cap_tokens+'</td>' +
    '<td>$'+(pe.est_usd||0).toFixed(4)+'</td><td>$'+(pe.est_cached_usd||0).toFixed(4)+'</td>' +
    '<td>'+(pe.observed_runs ? pe.observed_runs+' runs ('+Math.round(100*pe.success_rate)+'%, '+(pe.median_wall_ms/1000).toFixed(1)+'s)' : '-')+'</td></tr>'
  ).join('');
}

async function inspectVerb(op, target) {
  const res = await fetch('/api/recall', {
    method:'POST', headers:{'Content-Type':'application/json'},
    body: JSON.stringify({op, target})
  });
  const d = await res.json();
  document.getElementById('inspector-meta').textContent = (d.notes || []).join(' ');
  const pre = document.getElementById('inspector-pre');
  const txt = (d.excerpts || []).map(e => '=== ' + e.id + ' ===\n' + e.text).join('\n\n')
    || (d.extras || []).map(x => '=== ' + x.Header + ' ===\n' + x.Text).join('\n\n');
  pre.textContent = txt || '(no verbatim text returned)';
  pre.classList.remove('hidden');
}

function ablate(id) {
  const el = document.getElementById('drop-input');
  const cur = el.value.split(',').map(x=>x.trim()).filter(Boolean);
  if (!cur.includes(id)) cur.push(id);
  el.value = cur.join(',');
  runExplain();
}

async function loadReceipts() {
  const res = await fetch('/api/receipts');
  const d = await res.json();
  const tb = document.getElementById('rcpt-body');
  tb.innerHTML = (d.receipts || []).length ? d.receipts.map(r => {
    const cls = r.status === 'current' ? 'ok' : (r.status === 'stale' ? 'warn' : '');
    const mode = r.as_of_seq > 0 ? 'as-of #' + r.as_of_seq : 'live #' + r.snapshot_seq;
    return '<tr class="clickable" onclick="loadReceiptDiff(\''+esc(r.id)+'\')">' +
      '<td><code>'+esc(r.id)+'</code></td><td><span class="pill '+cls+'">'+esc(r.status)+'</span></td>' +
      '<td>'+esc(mode)+'</td><td>'+esc(r.question)+'</td></tr>';
  }).join('') : '<tr><td colspan="4" class="metric-sub">No Query Receipts yet. Create one with <code>cx watch</code> or MCP <code>cx_create_receipt</code>.</td></tr>';
}

async function loadReceiptDiff(id) {
  const res = await fetch('/api/receipt/' + encodeURIComponent(id) + '/diff');
  const d = await res.json();
  const r = d.receipt;
  let html = '<div class="row" style="margin-bottom:10px">' +
    '<span class="pill '+(d.original_verified?'ok':'bad')+'">'+(d.original_verified?'original quotes SHA-verified':'verification warning')+'</span>' +
    '<span class="pill '+(d.snapshot_match?'ok':'warn')+'">'+(d.snapshot_match?'snapshot root matches':'snapshot root mismatch')+'</span>' +
    '<span class="pill">seq '+r.snapshot_seq+' → '+d.current_seq+'</span></div>' +
    '<p><strong>Question:</strong> '+esc(r.question)+'</p>' +
    '<p><strong>Recorded Answer:</strong> '+esc(r.answer_text)+'</p>';
  if (r.stale_reason) {
    html += '<p><span class="pill warn">Stale Trigger</span> '+esc(r.stale_reason)+'</p>';
  }
  if ((d.verified_quotes || []).length) {
    html += '<h2>Verified Original Evidence</h2><pre class="code">' + esc(d.verified_quotes.join('\n')) + '</pre>';
  }
  if ((d.delta_turns || []).length) {
    html += '<h2 style="margin-top:14px">New Delta Turns Since Snapshot</h2>' +
      d.delta_turns.map(t => '<pre class="code">### '+esc(t.ID)+' ('+esc(t.Role)+')\n'+esc(t.Body)+'</pre>').join('');
  }
  document.getElementById('rcpt-detail').innerHTML = html;
}

async function loadTurns() {
  const q = document.getElementById('turn-filter').value.trim();
  const res = await fetch('/api/turns' + (q ? '?q=' + encodeURIComponent(q) : ''));
  const d = await res.json();
  document.getElementById('turn-body').innerHTML = (d.turns || []).map(t =>
    '<tr class="clickable" onclick="loadTurn(\''+esc(t.id)+'\')">' +
    '<td><code>'+esc(t.id)+'</code></td><td>'+esc(t.role)+'</td><td>'+t.atom_count+'</td><td>'+t.tokens+'</td>' +
    '<td class="metric-sub">'+esc(t.preview)+'</td></tr>'
  ).join('') || '<tr><td colspan="5" class="metric-sub">No turns found.</td></tr>';
}

async function loadTurn(id) {
  const res = await fetch('/api/turn/' + encodeURIComponent(id));
  const d = await res.json();
  const t = d.turn;
  let html = '<div class="row"><span class="pill accent">'+esc(t.ID)+'</span><span class="pill">'+esc(t.Role)+'</span><span class="pill">sha '+esc((t.BodySHA||'').slice(0,12))+'</span></div>';
  html += (d.atoms || []).map(a =>
    '<pre class="code">['+esc(a.id)+' bytes '+a.start+':'+a.end+' · '+a.tokens+' tok]\n'+esc(a.text)+'</pre>'
  ).join('');
  document.getElementById('turn-detail').innerHTML = html;
}

document.getElementById('btn-explain').onclick = runExplain;
document.getElementById('q-input').addEventListener('keydown', e => { if (e.key === 'Enter') runExplain(); });
document.getElementById('btn-filter').onclick = loadTurns;
document.getElementById('btn-recall').onclick = async () => {
  const op = document.getElementById('rec-op').value;
  const target = document.getElementById('rec-target').value.trim();
  if (!target) return;
  const res = await fetch('/api/recall', {
    method:'POST', headers:{'Content-Type':'application/json'},
    body: JSON.stringify({op, target})
  });
  const d = await res.json();
  document.getElementById('rec-notes').textContent = (d.notes || []).join(' ');
  const out = document.getElementById('rec-out');
  const parts = [];
  for (const x of (d.extras || [])) parts.push('=== ' + x.Header + ' ===\n' + x.Text);
  for (const e of (d.excerpts || [])) parts.push('=== ' + e.id + ' ===\n' + e.text);
  out.textContent = parts.join('\n\n') || '(no matches)';
  out.classList.remove('hidden');
};

loadOverview();
</script>
</body>
</html>
`
