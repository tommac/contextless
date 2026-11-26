package ui

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/tommac/contextless/internal/config"
	"github.com/tommac/contextless/internal/engine"
	"github.com/tommac/contextless/internal/store"
	"github.com/tommac/contextless/internal/vault"
)

// localReq is a request as the dashboard's own page sends it.
func localReq(method, path string, body io.Reader) *http.Request {
	r := httptest.NewRequest(method, "http://127.0.0.1:8787"+path, body)
	if body != nil {
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("Origin", "http://127.0.0.1:8787")
	}
	return r
}

// Another site's page cannot read the conversation or run a query through
// the browser: the dashboard answers only under its own name, to its own
// origin, and takes a body only as JSON.
func TestUIRefusesOtherSites(t *testing.T) {
	cfg := config.Default()
	cfg.Root = t.TempDir()
	cfg.Sandbox = false
	e, err := engine.Create(cfg, "guard", "Guard", "")
	if err != nil {
		t.Fatal(err)
	}
	defer e.Close()
	_, _ = e.V.AppendTurn(vault.RoleUser, "user", "", "", "The vault code is 31-7-22.")
	_ = e.Reindex()
	h := NewHandler(e)
	do := func(r *http.Request) int {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code == http.StatusForbidden && strings.Contains(w.Body.String(), "31-7-22") {
			t.Errorf("a refused request still returned conversation text")
		}
		return w.Code
	}
	rebind := httptest.NewRequest("GET", "http://attacker.example:8787/api/turns", nil)
	if c := do(rebind); c != http.StatusForbidden {
		t.Errorf("a request under another host name got %d", c)
	}
	form := httptest.NewRequest("POST", "http://127.0.0.1:8787/api/ledger-query", strings.NewReader(`{"spec":{"kind":"decisions"}}`))
	form.Header.Set("Content-Type", "text/plain")
	if c := do(form); c != http.StatusForbidden {
		t.Errorf("a text/plain POST (what a cross-site form sends) got %d", c)
	}
	cross := localReq("POST", "/api/ledger-query", strings.NewReader(`{"spec":{"kind":"decisions"}}`))
	cross.Header.Set("Origin", "https://attacker.example")
	if c := do(cross); c != http.StatusForbidden {
		t.Errorf("a POST from another origin got %d", c)
	}
	fetched := localReq("GET", "/api/turns", nil)
	fetched.Header.Set("Sec-Fetch-Site", "cross-site")
	if c := do(fetched); c != http.StatusForbidden {
		t.Errorf("a cross-site GET of the API got %d", c)
	}
	traversal := localReq("POST", "/api/ledger-query", strings.NewReader(`{"resume_id":"../../../session"}`))
	if c := do(traversal); c != http.StatusBadRequest {
		t.Errorf("a resume id that is a path got %d", c)
	}
	if c := do(localReq("GET", "/api/turns", nil)); c != http.StatusOK {
		t.Errorf("the dashboard's own request got %d", c)
	}
	own := httptest.NewRequest("GET", "http://localhost:8787/", nil)
	if c := do(own); c != http.StatusOK {
		t.Errorf("the page itself under localhost got %d", c)
	}
}

func TestUIEndpoints(t *testing.T) {
	cfg := config.Default()
	cfg.Root = t.TempDir()
	cfg.Sandbox = false
	cfg.WholeHistoryTokens = -1

	e, err := engine.Create(cfg, "uitest", "UI Test Vault", "Exercise all dashboard endpoints")
	if err != nil {
		t.Fatal(err)
	}
	defer e.Close()

	t1, _ := e.V.AppendTurn(vault.RoleUser, "user", "", "", "Set canary pool_size=20 in us-east-1.")
	_, _ = e.V.AppendTurn(vault.RoleAssistant, "test", "", "", "Confirmed pool_size=20.")
	_ = e.Reindex()

	rcpt, err := e.CreateReceipt("What is pool_size?", t1.ID, "pool_size=20", 0, []store.ReceiptEvidence{{
		Turn: t1.ID, BodySHA: t1.BodySHA, Start: 0, End: len(t1.Body), Seg: vault.AtomSeg, Quote: t1.Body,
	}}, nil, "", "")
	if err != nil {
		t.Fatal(err)
	}

	h := NewHandler(e)

	// GET /
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, localReq("GET", "/", nil))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "packet inspector") {
		t.Fatalf("GET / failed: code=%d", rec.Code)
	}

	// GET /api/overview, /api/turns, /api/turn/t000001, /api/receipts, /api/receipt/{id}/diff, /api/runs, /api/stats
	for _, path := range []string{
		"/api/overview",
		"/api/turns",
		"/api/turns?q=pool_size",
		"/api/turn/t000001",
		"/api/receipts",
		"/api/receipt/" + rcpt.ID + "/diff",
		"/api/runs",
		"/api/stats",
	} {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, localReq("GET", path, nil))
		if w.Code != http.StatusOK {
			t.Fatalf("GET %s returned %d: %s", path, w.Code, w.Body.String())
		}
		var js map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &js); err != nil {
			t.Fatalf("GET %s returned invalid JSON: %v", path, err)
		}
	}

	// POST /api/explain with adaptive + drop
	body := bytes.NewBufferString(`{"query":"What is pool_size?","adaptive":true,"drop_ids":["t000001"]}`)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, localReq("POST", "/api/explain", body))
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"plan"`) {
		t.Fatalf("POST /api/explain failed: %d %s", w.Code, w.Body.String())
	}

	// POST /api/recall grep
	rbody := bytes.NewBufferString(`{"op":"grep","target":"pool_size="}`)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, localReq("POST", "/api/recall", rbody))
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "t000001") {
		t.Fatalf("POST /api/recall failed: %d %s", w.Code, w.Body.String())
	}

	// POST /api/ledger-query
	lqBody := bytes.NewBufferString(`{"spec":{"kind":"turns","op":"count","filter":"pool_size","strict":true}}`)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, localReq("POST", "/api/ledger-query", lqBody))
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"verdict":"exact"`) {
		t.Fatalf("POST /api/ledger-query failed: %d %s", w.Code, w.Body.String())
	}

	// GET /api/views (active + key history)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, localReq("GET", "/api/views?key=pool_size", nil))
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"20"`) {
		t.Fatalf("GET /api/views failed: %d %s", w.Code, w.Body.String())
	}
}
